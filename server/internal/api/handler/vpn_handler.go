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
}

func NewVPNHandler(deviceService *service.DeviceService, vpnService *service.VPNService) *VPNHandler {
	return &VPNHandler{
		deviceService: deviceService,
		vpnService:    vpnService,
	}
}

type vpnConnectRequest struct {
	CommonName string `json:"common_name"`
	DeviceID   string `json:"device_id"`
	VpnIP      string `json:"vpn_ip"`
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

	if err := h.vpnService.OnDeviceConnected(device.BasePort, req.VpnIP); err != nil {
		log.Printf("VPN connected: failed to setup iptables: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to setup iptables"})
		return
	}

	identifier := req.DeviceID
	if identifier == "" {
		identifier = req.CommonName
	}
	log.Printf("VPN connected: %s (vpn_ip=%s, base_port=%d)", identifier, req.VpnIP, device.BasePort)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

	if err := h.vpnService.OnDeviceDisconnected(device.BasePort, req.VpnIP); err != nil {
		log.Printf("VPN disconnected: failed to teardown iptables: %v", err)
	}

	identifier := req.DeviceID
	if identifier == "" {
		identifier = req.CommonName
	}
	log.Printf("VPN disconnected: %s", identifier)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
