package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type ConnectionHandler struct {
	connService  *service.ConnectionService
	shareService *service.DeviceShareService
}

func NewConnectionHandler(connService *service.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{connService: connService}
}

func (h *ConnectionHandler) SetShareService(ss *service.DeviceShareService) {
	h.shareService = ss
}

func (h *ConnectionHandler) Create(c *gin.Context) {
	var req domain.CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	if roleStr == "customer" {
		userIDVal, _ := c.Get("user_id")
		customerID, _ := userIDVal.(uuid.UUID)
		allowed, err := h.shareService.CanDo(c.Request.Context(), req.DeviceID, customerID, "manage_ports")
		if err != nil || !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		req.CustomerID = &customerID
	}

	conn, err := h.connService.Create(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conn)
}

func (h *ConnectionHandler) List(c *gin.Context) {
	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	deviceIDStr := c.Query("device_id")

	if roleStr == "customer" {
		userIDVal, _ := c.Get("user_id")
		customerID, _ := userIDVal.(uuid.UUID)

		if deviceIDStr != "" {
			deviceID, err := uuid.Parse(deviceIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
				return
			}
			conns, err := h.connService.ListByDeviceForCustomer(c.Request.Context(), deviceID, customerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"connections": conns})
			return
		}

		conns, err := h.connService.ListByCustomer(c.Request.Context(), customerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"connections": conns})
		return
	}

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

	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	if roleStr == "customer" {
		userIDVal, _ := c.Get("user_id")
		customerID, _ := userIDVal.(uuid.UUID)
		conn, err := h.connService.GetByIDForCustomer(c.Request.Context(), id, customerID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusOK, conn)
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

	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	if roleStr == "customer" {
		userIDVal, _ := c.Get("user_id")
		customerID, _ := userIDVal.(uuid.UUID)
		// Fetch connection to get device ID for permission check
		conn, err := h.connService.GetByIDForCustomer(c.Request.Context(), id, customerID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		allowed, err := h.shareService.CanDo(c.Request.Context(), conn.DeviceID, customerID, "manage_ports")
		if err != nil || !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
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

	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	if roleStr == "customer" {
		userIDVal, _ := c.Get("user_id")
		customerID, _ := userIDVal.(uuid.UUID)
		conn, err := h.connService.GetByIDForCustomer(c.Request.Context(), id, customerID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		allowed, err := h.shareService.CanDo(c.Request.Context(), conn.DeviceID, customerID, "manage_ports")
		if err != nil || !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
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
// This is admin-only: customers get 403.
func (h *ConnectionHandler) ResetBandwidth(c *gin.Context) {
	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	if roleStr == "customer" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

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
