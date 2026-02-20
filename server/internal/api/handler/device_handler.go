package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type lastBytes struct {
	bytesIn  int64
	bytesOut int64
}

type DeviceHandler struct {
	deviceService *service.DeviceService
	bwService     *service.BandwidthService
	wsHub         *WSHub
	lastBytesMu   sync.Mutex
	lastBytesMap  map[uuid.UUID]lastBytes
}

func NewDeviceHandler(deviceService *service.DeviceService, bwService *service.BandwidthService, wsHub *WSHub) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		bwService:     bwService,
		wsHub:         wsHub,
		lastBytesMap:  make(map[uuid.UUID]lastBytes),
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

func (h *DeviceHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	name := device.Name
	description := device.Description
	if body.Name != nil {
		name = *body.Name
	}
	if body.Description != nil {
		description = *body.Description
	}

	if err := h.deviceService.UpdateNameDescription(c.Request.Context(), id, name, description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated device
	updated, err := h.deviceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
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

	// Record bandwidth delta (device sends cumulative bytes)
	if req.BytesIn > 0 || req.BytesOut > 0 {
		h.lastBytesMu.Lock()
		prev, exists := h.lastBytesMap[id]
		h.lastBytesMap[id] = lastBytes{bytesIn: req.BytesIn, bytesOut: req.BytesOut}
		h.lastBytesMu.Unlock()

		if exists {
			deltaIn := req.BytesIn - prev.bytesIn
			deltaOut := req.BytesOut - prev.bytesOut
			// Only record if positive delta (counter may reset on app restart)
			if deltaIn > 0 || deltaOut > 0 {
				if deltaIn < 0 {
					deltaIn = req.BytesIn
				}
				if deltaOut < 0 {
					deltaOut = req.BytesOut
				}
				_ = h.bwService.Record(c.Request.Context(), id, nil, deltaIn, deltaOut)
			}
		}
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

func (h *DeviceHandler) GetBandwidth(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	todayIn, todayOut, err := h.bwService.GetDeviceTodayTotal(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monthIn, monthOut, err := h.bwService.GetDeviceMonthTotal(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"today_in":  todayIn,
		"today_out": todayOut,
		"month_in":  monthIn,
		"month_out": monthOut,
	})
}

func (h *DeviceHandler) GetCommands(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	commands, err := h.deviceService.GetCommandHistory(c.Request.Context(), id, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"commands": commands})
}

func (h *DeviceHandler) GetBandwidthHourly(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().UTC().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	tzOffset := 0
	if tzStr := c.Query("tz_offset"); tzStr != "" {
		fmt.Sscanf(tzStr, "%d", &tzOffset)
	}

	hourly, err := h.bwService.GetDeviceHourly(c.Request.Context(), id, date, tzOffset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hourly": hourly})
}

func (h *DeviceHandler) GetUptime(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().UTC().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	tzOffset := 0
	if tzStr := c.Query("tz_offset"); tzStr != "" {
		fmt.Sscanf(tzStr, "%d", &tzOffset)
	}

	segments, err := h.deviceService.GetUptimeSegments(c.Request.Context(), id, date, tzOffset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"segments": segments})
}
