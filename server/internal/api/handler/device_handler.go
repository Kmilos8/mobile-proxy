package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type DeviceHandler struct {
	deviceService *service.DeviceService
	wsHub         *WSHub
}

func NewDeviceHandler(deviceService *service.DeviceService, wsHub *WSHub) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		wsHub:         wsHub,
	}
}

func (h *DeviceHandler) Register(c *gin.Context) {
	var req domain.DeviceRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.deviceService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *DeviceHandler) List(c *gin.Context) {
	devices, err := h.deviceService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

func (h *DeviceHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	device, err := h.deviceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) Heartbeat(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	var req domain.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.deviceService.Heartbeat(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast device update via WebSocket
	device, _ := h.deviceService.GetByID(c.Request.Context(), id)
	if device != nil {
		h.wsHub.Broadcast(domain.WSMessage{
			Type:    "device_update",
			Payload: device,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *DeviceHandler) SendCommand(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	var req domain.CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd, err := h.deviceService.SendCommand(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cmd)
}

func (h *DeviceHandler) CommandResult(c *gin.Context) {
	cmdID, err := uuid.Parse(c.Param("commandId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid command id"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
		Result string `json:"result"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := domain.CommandStatus(body.Status)
	if err := h.deviceService.UpdateCommandStatus(c.Request.Context(), cmdID, status, body.Result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *DeviceHandler) GetIPHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	history, err := h.deviceService.GetIPHistory(c.Request.Context(), id, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}
