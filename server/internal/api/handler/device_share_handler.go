package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type DeviceShareHandler struct {
	shareService *service.DeviceShareService
}

func NewDeviceShareHandler(shareService *service.DeviceShareService) *DeviceShareHandler {
	return &DeviceShareHandler{shareService: shareService}
}

// CreateShare creates a new device share.
// The caller must be the device owner (customer_id on device == user_id in JWT).
func (h *DeviceShareHandler) CreateShare(c *gin.Context) {
	var body struct {
		DeviceID           uuid.UUID `json:"device_id" binding:"required"`
		SharedWith         uuid.UUID `json:"shared_with" binding:"required"`
		CanRename          bool      `json:"can_rename"`
		CanManagePorts     bool      `json:"can_manage_ports"`
		CanDownloadConfigs bool      `json:"can_download_configs"`
		CanRotateIP        bool      `json:"can_rotate_ip"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("user_id")
	ownerID, _ := userIDVal.(uuid.UUID)

	share := &domain.DeviceShare{
		DeviceID:           body.DeviceID,
		OwnerID:            ownerID,
		SharedWith:         body.SharedWith,
		CanRename:          body.CanRename,
		CanManagePorts:     body.CanManagePorts,
		CanDownloadConfigs: body.CanDownloadConfigs,
		CanRotateIP:        body.CanRotateIP,
	}

	if err := h.shareService.CreateShare(c.Request.Context(), share); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, share)
}

// ListShares returns all shares for a device.
// Requires device_id query param. Customers must have access to the device.
func (h *DeviceShareHandler) ListShares(c *gin.Context) {
	deviceIDStr := c.Query("device_id")
	if deviceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id query param required"})
		return
	}

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
		return
	}

	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	if roleStr == "customer" {
		userIDVal, _ := c.Get("user_id")
		customerID, _ := userIDVal.(uuid.UUID)
		allowed, err := h.shareService.CanAccess(c.Request.Context(), deviceID, customerID)
		if err != nil || !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	shares, err := h.shareService.ListSharesForDevice(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shares": shares})
}

// UpdateShare updates permission booleans on an existing share.
// The caller must be the device owner.
func (h *DeviceShareHandler) UpdateShare(c *gin.Context) {
	shareID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid share id"})
		return
	}

	var body struct {
		CanRename          bool `json:"can_rename"`
		CanManagePorts     bool `json:"can_manage_ports"`
		CanDownloadConfigs bool `json:"can_download_configs"`
		CanRotateIP        bool `json:"can_rotate_ip"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("user_id")
	callerID, _ := userIDVal.(uuid.UUID)

	share := &domain.DeviceShare{
		ID:                 shareID,
		OwnerID:            callerID,
		CanRename:          body.CanRename,
		CanManagePorts:     body.CanManagePorts,
		CanDownloadConfigs: body.CanDownloadConfigs,
		CanRotateIP:        body.CanRotateIP,
	}

	if err := h.shareService.UpdateShare(c.Request.Context(), share); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DeleteShare removes a share by ID.
// The caller must be the share owner.
func (h *DeviceShareHandler) DeleteShare(c *gin.Context) {
	shareID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid share id"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	callerID, _ := userIDVal.(uuid.UUID)

	if err := h.shareService.DeleteShare(c.Request.Context(), shareID, callerID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
