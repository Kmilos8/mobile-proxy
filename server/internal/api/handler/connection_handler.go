package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type ConnectionHandler struct {
	connService *service.ConnectionService
}

func NewConnectionHandler(connService *service.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{connService: connService}
}

func (h *ConnectionHandler) Create(c *gin.Context) {
	var req domain.CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := h.connService.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conn)
}

func (h *ConnectionHandler) List(c *gin.Context) {
	deviceIDStr := c.Query("device_id")
	if deviceIDStr != "" {
		deviceID, err := uuid.Parse(deviceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
			return
		}
		conns, err := h.connService.ListByDevice(c.Request.Context(), deviceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"connections": conns})
		return
	}

	conns, err := h.connService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"connections": conns})
}

func (h *ConnectionHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}

	conn, err := h.connService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}

	c.JSON(http.StatusOK, conn)
}

func (h *ConnectionHandler) SetActive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}

	var body struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.connService.SetActive(c.Request.Context(), id, body.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ConnectionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}

	if err := h.connService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ConnectionHandler) RegeneratePassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}

	newPass, err := h.connService.RegeneratePassword(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": newPass})
}

// BandwidthFlush is an internal endpoint (no JWT) called by the tunnel server every 30s.
// Receives a map of {username -> bytes_used} and updates the DB.
func (h *ConnectionHandler) BandwidthFlush(c *gin.Context) {
	var data map[string]int64 // username -> bytes used
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for username, used := range data {
		if err := h.connService.UpdateBandwidthUsedByUsername(c.Request.Context(), username, used); err != nil {
			log.Printf("[bandwidth-flush] failed for %s: %v", username, err)
		}
	}
	// Forward bandwidth data to peer (dashboard VPS), unless this is already a peer sync
	if c.Query("from_peer") != "1" {
		h.connService.SyncBandwidth(data)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ResetBandwidth resets bandwidth_used to 0 for a connection (DB and tunnel in-memory counter).
func (h *ConnectionHandler) ResetBandwidth(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}
	if err := h.connService.ResetBandwidth(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
