package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

type SyncHandler struct {
	deviceRepo *repository.DeviceRepository
	connRepo   *repository.ConnectionRepository
}

func NewSyncHandler(deviceRepo *repository.DeviceRepository, connRepo *repository.ConnectionRepository) *SyncHandler {
	return &SyncHandler{deviceRepo: deviceRepo, connRepo: connRepo}
}

type syncDeviceRequest struct {
	ID                uuid.UUID    `json:"id"`
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	AndroidID         string       `json:"android_id"`
	Status            string       `json:"status"`
	BasePort          int          `json:"base_port"`
	HTTPPort          int          `json:"http_port"`
	SOCKS5Port        int          `json:"socks5_port"`
	UDPRelayPort      int          `json:"udp_relay_port"`
	OVPNPort          int          `json:"ovpn_port"`
	DeviceModel       string       `json:"device_model"`
	AndroidVersion    string       `json:"android_version"`
	AppVersion        string       `json:"app_version"`
	RelayServerID     *uuid.UUID   `json:"relay_server_id"`
	AuthToken         string       `json:"auth_token"`
	AutoRotateMinutes int          `json:"auto_rotate_minutes"`
}

func (h *SyncHandler) SyncDevice(c *gin.Context) {
	var req syncDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := &domain.Device{
		ID:                req.ID,
		Name:              req.Name,
		Description:       req.Description,
		AndroidID:         req.AndroidID,
		Status:            domain.DeviceStatus(req.Status),
		BasePort:          req.BasePort,
		HTTPPort:          req.HTTPPort,
		SOCKS5Port:        req.SOCKS5Port,
		UDPRelayPort:      req.UDPRelayPort,
		OVPNPort:          req.OVPNPort,
		DeviceModel:       req.DeviceModel,
		AndroidVersion:    req.AndroidVersion,
		AppVersion:        req.AppVersion,
		RelayServerID:     req.RelayServerID,
		AutoRotateMinutes: req.AutoRotateMinutes,
	}

	if err := h.deviceRepo.Upsert(c.Request.Context(), device, req.AuthToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type syncConnectionItem struct {
	ID            uuid.UUID  `json:"id"`
	DeviceID      uuid.UUID  `json:"device_id"`
	CustomerID    *uuid.UUID `json:"customer_id"`
	Username      string     `json:"username"`
	PasswordHash  string     `json:"password_hash"`
	PasswordPlain string     `json:"password_plain"`
	IPWhitelist   []string   `json:"ip_whitelist"`
	BandwidthLimit int64     `json:"bandwidth_limit"`
	BandwidthUsed  int64     `json:"bandwidth_used"`
	Active        bool       `json:"active"`
	ProxyType     string     `json:"proxy_type"`
	BasePort      *int       `json:"base_port"`
	HTTPPort      *int       `json:"http_port"`
	SOCKS5Port    *int       `json:"socks5_port"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type syncConnectionsRequest struct {
	DeviceID    uuid.UUID            `json:"device_id"`
	Connections []syncConnectionItem `json:"connections"`
}

func (h *SyncHandler) SyncConnections(c *gin.Context) {
	var req syncConnectionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conns := make([]domain.ProxyConnection, len(req.Connections))
	for i, ci := range req.Connections {
		conns[i] = domain.ProxyConnection{
			ID:             ci.ID,
			DeviceID:       ci.DeviceID,
			CustomerID:     ci.CustomerID,
			Username:       ci.Username,
			PasswordHash:   ci.PasswordHash,
			// PasswordPlain is deprecated (nullable after migration); not synced
			IPWhitelist:    ci.IPWhitelist,
			BandwidthLimit: ci.BandwidthLimit,
			BandwidthUsed:  ci.BandwidthUsed,
			Active:         ci.Active,
			ProxyType:      ci.ProxyType,
			BasePort:       ci.BasePort,
			HTTPPort:       ci.HTTPPort,
			SOCKS5Port:     ci.SOCKS5Port,
			ExpiresAt:      ci.ExpiresAt,
			CreatedAt:      ci.CreatedAt,
			UpdatedAt:      ci.UpdatedAt,
		}
		if conns[i].IPWhitelist == nil {
			conns[i].IPWhitelist = []string{}
		}
	}

	if err := h.connRepo.ReplaceAllByDeviceID(c.Request.Context(), req.DeviceID, conns); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
