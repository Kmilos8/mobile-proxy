package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type PairingHandler struct {
	pairingService *service.PairingService
}

func NewPairingHandler(pairingService *service.PairingService) *PairingHandler {
	return &PairingHandler{pairingService: pairingService}
}

// CreateCode creates a new pairing code (JWT auth required)
func (h *PairingHandler) CreateCode(c *gin.Context) {
	var req domain.CreatePairingCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body — default to 5 minutes
		req.ExpiresInMinutes = 5
	}

	// Get user ID from JWT context
	var createdBy *uuid.UUID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uuid.UUID); ok {
			createdBy = &id
		}
	}

	resp, err := h.pairingService.CreateCode(c.Request.Context(), req.ExpiresInMinutes, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListCodes returns all pairing codes (JWT auth required)
func (h *PairingHandler) ListCodes(c *gin.Context) {
	codes, err := h.pairingService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pairing_codes": codes})
}

// DeleteCode removes a pairing code (JWT auth required)
func (h *PairingHandler) DeleteCode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.pairingService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ClaimCode is the PUBLIC endpoint called by Android app to pair (no auth)
func (h *PairingHandler) ClaimCode(c *gin.Context) {
	var req domain.ClaimPairingCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.pairingService.ClaimCode(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
