package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/repository"
	"github.com/mobileproxy/server/internal/service"
)

type OpenVPNHandler struct {
	connRepo      *repository.ConnectionRepository
	deviceService *service.DeviceService
	tunnelPushURL string // e.g. http://127.0.0.1:8081
}

func NewOpenVPNHandler(connRepo *repository.ConnectionRepository, deviceService *service.DeviceService) *OpenVPNHandler {
	tunnelPushURL := "http://127.0.0.1:8081"
	if v := os.Getenv("TUNNEL_PUSH_URL"); v != "" {
		tunnelPushURL = v
	}
	return &OpenVPNHandler{
		connRepo:      connRepo,
		deviceService: deviceService,
		tunnelPushURL: tunnelPushURL,
	}
}

// Auth handles POST /api/internal/openvpn/auth
// Called by OpenVPN auth-user-pass-verify script.
// Validates username/password against proxy connections.
func (h *OpenVPNHandler) Auth(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	conn, err := h.connRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		log.Printf("[openvpn-auth] user %s not found: %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false})
		return
	}

	if conn.PasswordPlain != req.Password {
		log.Printf("[openvpn-auth] password mismatch for user %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false})
		return
	}

	log.Printf("[openvpn-auth] authenticated user %s (connection %s)", req.Username, conn.ID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Connect handles POST /api/internal/openvpn/connect
// Called by OpenVPN client-connect script.
// Looks up the connection -> device, then tells the tunnel server to set up
// policy routing (ip rule) so client traffic routes through the device tunnel.
func (h *OpenVPNHandler) Connect(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		VpnIP    string `json:"vpn_ip"` // 10.9.0.x assigned by OpenVPN
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up connection by username
	conn, err := h.connRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		log.Printf("[openvpn-connect] connection not found for user %s: %v", req.Username, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}

	// Look up the device to get its VPN IP (192.168.255.x)
	device, err := h.deviceService.GetByID(c.Request.Context(), conn.DeviceID)
	if err != nil {
		log.Printf("[openvpn-connect] device not found for connection %s: %v", conn.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	if device.VpnIP == "" {
		log.Printf("[openvpn-connect] device %s has no VPN IP (offline?)", device.ID)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "device not connected to tunnel"})
		return
	}

	// POST to tunnel push API to set up policy routing
	// Use device's relay server IP to reach the correct tunnel server
	pushURL := h.tunnelPushURL
	if device.RelayServerIP != "" {
		pushURL = fmt.Sprintf("http://%s:8081", device.RelayServerIP)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"client_vpn_ip": req.VpnIP,
		"device_vpn_ip": device.VpnIP,
	})
	resp, err := http.Post(pushURL+"/openvpn-client-connect", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[openvpn-connect] failed to notify tunnel: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tunnel notification failed"})
		return
	}
	resp.Body.Close()

	log.Printf("[openvpn-connect] user=%s vpn_ip=%s -> device=%s device_vpn_ip=%s", req.Username, req.VpnIP, device.ID, device.VpnIP)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Disconnect handles POST /api/internal/openvpn/disconnect
// Called by OpenVPN client-disconnect script.
func (h *OpenVPNHandler) Disconnect(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		VpnIP    string `json:"vpn_ip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up connection to find the device's relay server
	pushURL := h.tunnelPushURL
	conn, err := h.connRepo.GetByUsername(c.Request.Context(), req.Username)
	if err == nil {
		device, err := h.deviceService.GetByID(c.Request.Context(), conn.DeviceID)
		if err == nil && device.RelayServerIP != "" {
			pushURL = fmt.Sprintf("http://%s:8081", device.RelayServerIP)
		}
	}

	// POST to tunnel push API to remove policy routing rule
	body, _ := json.Marshal(map[string]interface{}{
		"client_vpn_ip": req.VpnIP,
	})
	resp, err := http.Post(pushURL+"/openvpn-client-disconnect", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[openvpn-disconnect] failed to notify tunnel: %v", err)
	} else {
		resp.Body.Close()
	}

	log.Printf("[openvpn-disconnect] user=%s vpn_ip=%s", req.Username, req.VpnIP)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DownloadOVPN handles GET /api/connections/:id/ovpn
// Generates and returns a .ovpn config file for the given connection.
func (h *OpenVPNHandler) DownloadOVPN(c *gin.Context) {
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}

	conn, err := h.connRepo.GetByID(c.Request.Context(), connID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}

	// Look up device to get relay server IP
	device, err := h.deviceService.GetByID(c.Request.Context(), conn.DeviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	serverIP := device.RelayServerIP
	if serverIP == "" {
		serverIP = os.Getenv("VPN_SERVER_IP")
	}
	if serverIP == "" {
		serverIP = "127.0.0.1"
	}

	// Read CA cert and tls-auth key from PKI directory
	caCert, err := os.ReadFile("/etc/openvpn-client/pki/ca.crt")
	if err != nil {
		log.Printf("[openvpn-ovpn] failed to read CA cert: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PKI not initialized"})
		return
	}

	// Build .ovpn config
	var ovpn strings.Builder
	ovpn.WriteString("client\n")
	ovpn.WriteString("dev tun\n")
	ovpn.WriteString("proto udp\n")
	ovpn.WriteString(fmt.Sprintf("remote %s 1195\n", serverIP))
	ovpn.WriteString("resolv-retry infinite\n")
	ovpn.WriteString("nobind\n")
	ovpn.WriteString("persist-key\n")
	ovpn.WriteString("persist-tun\n")
	ovpn.WriteString("remote-cert-tls server\n")
	ovpn.WriteString("cipher AES-128-GCM\n")
	ovpn.WriteString("auth SHA256\n")
	ovpn.WriteString("auth-user-pass\n")
	ovpn.WriteString("setenv CLIENT_CERT 0\n")
	ovpn.WriteString("sndbuf 0\n")
	ovpn.WriteString("rcvbuf 0\n")
	ovpn.WriteString("verb 3\n")
	ovpn.WriteString("\n")

	// Embed credentials inline so the user isn't prompted
	ovpn.WriteString("<auth-user-pass>\n")
	ovpn.WriteString(conn.Username + "\n")
	ovpn.WriteString(conn.PasswordPlain + "\n")
	ovpn.WriteString("</auth-user-pass>\n")
	ovpn.WriteString("\n")

	ovpn.WriteString("<ca>\n")
	ovpn.WriteString(strings.TrimSpace(string(caCert)))
	ovpn.WriteString("\n</ca>\n")

	filename := fmt.Sprintf("%s.ovpn", conn.Username)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/x-openvpn-profile", []byte(ovpn.String()))
}
