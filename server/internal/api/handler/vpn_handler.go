package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type VPNHandler struct {
	deviceService *service.DeviceService
	vpnService    *service.VPNService
	connService   *service.ConnectionService
}

func NewVPNHandler(deviceService *service.DeviceService, vpnService *service.VPNService, connService *service.ConnectionService) *VPNHandler {
	return &VPNHandler{
		deviceService: deviceService,
		vpnService:    vpnService,
		connService:   connService,
	}
}

type vpnConnectRequest struct {
	CommonName string `json:"common_name"`
	DeviceID   string `json:"device_id"`
	VpnIP      string `json:"vpn_ip"`
}

type connectionPortInfo struct {
	Port      int    `json:"port"`
	ProxyType string `json:"proxy_type"`
}

// Connected is called by the tunnel server or OpenVPN client-connect script
func (h *VPNHandler) Connected(c *gin.Context) {
	var req vpnConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var device *domain.Device
	var err error

	if req.DeviceID != "" {
		// Tunnel server sends device_id (UUID)
		id, parseErr := uuid.Parse(req.DeviceID)
		if parseErr != nil {
			log.Printf("VPN connected: invalid device_id=%s: %v", req.DeviceID, parseErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
			return
		}
		device, err = h.deviceService.GetByID(c.Request.Context(), id)
		if err != nil {
			log.Printf("VPN connected: device not found for id=%s: %v", req.DeviceID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
	} else {
		// Legacy OpenVPN sends common_name
		device, err = h.deviceService.GetByName(c.Request.Context(), req.CommonName)
		if err != nil {
			log.Printf("VPN connected: device not found for CN=%s: %v", req.CommonName, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
	}

	if err := h.deviceService.SetVpnIP(c.Request.Context(), device.ID, req.VpnIP); err != nil {
		log.Printf("VPN connected: failed to set VPN IP: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update vpn ip"})
		return
	}

	// Get connection details for this device
	var connections []connectionPortInfo
	if h.connService != nil {
		conns, err := h.connService.ListByDevice(c.Request.Context(), device.ID)
		if err == nil {
			for _, conn := range conns {
				if conn.BasePort != nil {
					connections = append(connections, connectionPortInfo{
						Port:      *conn.BasePort,
						ProxyType: conn.ProxyType,
					})
				}
			}
		}
	}

	identifier := req.DeviceID
	if identifier == "" {
		identifier = req.CommonName
	}
	log.Printf("VPN connected: %s (vpn_ip=%s, base_port=%d, connections=%v)", identifier, req.VpnIP, device.BasePort, connections)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "base_port": device.BasePort, "connections": connections})
}

// Disconnected is called by the tunnel server or OpenVPN client-disconnect script
func (h *VPNHandler) Disconnected(c *gin.Context) {
	var req vpnConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var device *domain.Device
	var err error

	if req.DeviceID != "" {
		id, parseErr := uuid.Parse(req.DeviceID)
		if parseErr != nil {
			log.Printf("VPN disconnected: invalid device_id=%s: %v", req.DeviceID, parseErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
			return
		}
		device, err = h.deviceService.GetByID(c.Request.Context(), id)
		if err != nil {
			log.Printf("VPN disconnected: device not found for id=%s: %v", req.DeviceID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
	} else {
		device, err = h.deviceService.GetByName(c.Request.Context(), req.CommonName)
		if err != nil {
			log.Printf("VPN disconnected: device not found for CN=%s: %v", req.CommonName, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
	}

	// Get connection details for teardown
	var connections []connectionPortInfo
	if h.connService != nil {
		conns, err := h.connService.ListByDevice(c.Request.Context(), device.ID)
		if err == nil {
			for _, conn := range conns {
				if conn.BasePort != nil {
					connections = append(connections, connectionPortInfo{
						Port:      *conn.BasePort,
						ProxyType: conn.ProxyType,
					})
				}
			}
		}
	}

	identifier := req.DeviceID
	if identifier == "" {
		identifier = req.CommonName
	}
	log.Printf("VPN disconnected: %s (connections=%v)", identifier, connections)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "base_port": device.BasePort, "connections": connections})
}
