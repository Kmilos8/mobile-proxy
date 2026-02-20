package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

type RelayServerHandler struct {
	relayServerService *service.RelayServerService
}

func NewRelayServerHandler(relayServerService *service.RelayServerService) *RelayServerHandler {
	return &RelayServerHandler{relayServerService: relayServerService}
}

func (h *RelayServerHandler) List(c *gin.Context) {
	servers, err := h.relayServerService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"relay_servers": servers})
}

func (h *RelayServerHandler) ListActive(c *gin.Context) {
	servers, err := h.relayServerService.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"relay_servers": servers})
}

func (h *RelayServerHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		IP       string `json:"ip" binding:"required"`
		Location string `json:"location"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rs := &domain.RelayServer{
		ID:       uuid.New(),
		Name:     req.Name,
		IP:       req.IP,
		Location: req.Location,
		Active:   true,
	}

	if err := h.relayServerService.Create(c.Request.Context(), rs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rs)
}
