package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
	"github.com/mobileproxy/server/internal/service"
)

type RotationLinkHandler struct {
	linkRepo      *repository.RotationLinkRepository
	deviceService *service.DeviceService
}

func NewRotationLinkHandler(linkRepo *repository.RotationLinkRepository, deviceService *service.DeviceService) *RotationLinkHandler {
	return &RotationLinkHandler{
		linkRepo:      linkRepo,
		deviceService: deviceService,
	}
}

// generateToken creates a URL-safe random token
func generateToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(token, "=")
}

// Create creates a new rotation link for a device (JWT auth required)
func (h *RotationLinkHandler) Create(c *gin.Context) {
	var body struct {
		DeviceID uuid.UUID `json:"device_id" binding:"required"`
		Name     string    `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify device exists
	_, err := h.deviceService.GetByID(c.Request.Context(), body.DeviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	link := &domain.RotationLink{
		ID:       uuid.New(),
		DeviceID: body.DeviceID,
		Token:    generateToken(),
		Name:     body.Name,
	}

	if err := h.linkRepo.Create(c.Request.Context(), link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, link)
}

// List returns all rotation links for a device (JWT auth required)
func (h *RotationLinkHandler) List(c *gin.Context) {
	deviceID, err := uuid.Parse(c.Query("device_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id query parameter required"})
		return
	}

	links, err := h.linkRepo.ListByDevice(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"links": links})
}

// Delete removes a rotation link (JWT auth required)
func (h *RotationLinkHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	if err := h.linkRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Rotate is the PUBLIC endpoint - no auth needed. Triggers IP rotation via the token.
func (h *RotationLinkHandler) Rotate(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	link, err := h.linkRepo.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid rotation link"})
		return
	}

	// Update last used timestamp
	h.linkRepo.UpdateLastUsed(c.Request.Context(), link.ID)

	// Send rotate_ip command to the device
	cmd, err := h.deviceService.SendCommand(c.Request.Context(), link.DeviceID, &domain.CommandRequest{
		Type:    domain.CommandRotateIP,
		Payload: `{"method":"airplane_mode","source":"rotation_link"}`,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send rotation command"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"message":    "IP rotation triggered",
		"command_id": cmd.ID,
		"device_id":  link.DeviceID,
	})
}
