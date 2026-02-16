package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	VpnIP      string `json:"vpn_ip"`
}

// Connected is called by the OpenVPN client-connect script
func (h *VPNHandler) Connected(c *gin.Context) {
	var req vpnConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceService.GetByName(c.Request.Context(), req.CommonName)
	if err != nil {
		log.Printf("VPN connected: device not found for CN=%s: %v", req.CommonName, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	if err := h.deviceService.SetVpnIP(c.Request.Context(), device.ID, req.VpnIP); err != nil {
		log.Printf("VPN connected: failed to set VPN IP for %s: %v", req.CommonName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update vpn ip"})
		return
	}

	if err := h.vpnService.OnDeviceConnected(device.BasePort, req.VpnIP); err != nil {
		log.Printf("VPN connected: failed to setup iptables for %s: %v", req.CommonName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to setup iptables"})
		return
	}

	log.Printf("VPN connected: %s (vpn_ip=%s, base_port=%d)", req.CommonName, req.VpnIP, device.BasePort)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Disconnected is called by the OpenVPN client-disconnect script
func (h *VPNHandler) Disconnected(c *gin.Context) {
	var req vpnConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceService.GetByName(c.Request.Context(), req.CommonName)
	if err != nil {
		log.Printf("VPN disconnected: device not found for CN=%s: %v", req.CommonName, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	if err := h.vpnService.OnDeviceDisconnected(device.BasePort, req.VpnIP); err != nil {
		log.Printf("VPN disconnected: failed to teardown iptables for %s: %v", req.CommonName, err)
	}

	log.Printf("VPN disconnected: %s", req.CommonName)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
