package service

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type SyncService struct {
	peerAPIURL string
	client     *http.Client
}

func NewSyncService(peerAPIURL string) *SyncService {
	return &SyncService{
		peerAPIURL: peerAPIURL,
		client:     &http.Client{Timeout: 3 * time.Second},
	}
}

func (s *SyncService) SyncDevice(device *domain.Device, authToken string) {
	payload := map[string]interface{}{
		"id":                  device.ID,
		"name":                device.Name,
		"description":         device.Description,
		"android_id":          device.AndroidID,
		"status":              string(device.Status),
		"base_port":           device.BasePort,
		"http_port":           device.HTTPPort,
		"socks5_port":         device.SOCKS5Port,
		"udp_relay_port":      device.UDPRelayPort,
		"ovpn_port":           device.OVPNPort,
		"device_model":        device.DeviceModel,
		"android_version":     device.AndroidVersion,
		"app_version":         device.AppVersion,
		"relay_server_id":     device.RelayServerID,
		"auth_token":          authToken,
		"auto_rotate_minutes": device.AutoRotateMinutes,
	}

	body, _ := json.Marshal(payload)
	resp, err := s.client.Post(s.peerAPIURL+"/api/internal/sync/device", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[sync] SyncDevice %s to peer failed: %v", device.ID, err)
		return
	}
	resp.Body.Close()
	log.Printf("[sync] SyncDevice %s to peer: %d", device.ID, resp.StatusCode)
}

func (s *SyncService) SyncConnections(deviceID uuid.UUID, connections []domain.ProxyConnection) {
	type connItem struct {
		ID             uuid.UUID  `json:"id"`
		DeviceID       uuid.UUID  `json:"device_id"`
		CustomerID     *uuid.UUID `json:"customer_id"`
		Username       string     `json:"username"`
		PasswordHash   string     `json:"password_hash"`
		PasswordPlain  string     `json:"password_plain"`
		IPWhitelist    []string   `json:"ip_whitelist"`
		BandwidthLimit int64      `json:"bandwidth_limit"`
		BandwidthUsed  int64      `json:"bandwidth_used"`
		Active         bool       `json:"active"`
		ProxyType      string     `json:"proxy_type"`
		BasePort       *int       `json:"base_port"`
		HTTPPort       *int       `json:"http_port"`
		SOCKS5Port     *int       `json:"socks5_port"`
		ExpiresAt      *time.Time `json:"expires_at"`
		CreatedAt      time.Time  `json:"created_at"`
		UpdatedAt      time.Time  `json:"updated_at"`
	}

	items := make([]connItem, len(connections))
	for i, c := range connections {
		items[i] = connItem{
			ID:             c.ID,
			DeviceID:       c.DeviceID,
			CustomerID:     c.CustomerID,
			Username:       c.Username,
			PasswordHash:   c.PasswordHash,
			PasswordPlain:  c.PasswordHash,
			IPWhitelist:    c.IPWhitelist,
			BandwidthLimit: c.BandwidthLimit,
			BandwidthUsed:  c.BandwidthUsed,
			Active:         c.Active,
			ProxyType:      c.ProxyType,
			BasePort:       c.BasePort,
			HTTPPort:       c.HTTPPort,
			SOCKS5Port:     c.SOCKS5Port,
			ExpiresAt:      c.ExpiresAt,
			CreatedAt:      c.CreatedAt,
			UpdatedAt:      c.UpdatedAt,
		}
		if items[i].IPWhitelist == nil {
			items[i].IPWhitelist = []string{}
		}
	}

	payload := map[string]interface{}{
		"device_id":   deviceID,
		"connections": items,
	}

	body, _ := json.Marshal(payload)
	resp, err := s.client.Post(s.peerAPIURL+"/api/internal/sync/connections", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[sync] SyncConnections device=%s to peer failed: %v", deviceID, err)
		return
	}
	resp.Body.Close()
	log.Printf("[sync] SyncConnections device=%s (%d conns) to peer: %d", deviceID, len(connections), resp.StatusCode)
}

// SyncBandwidth forwards bandwidth flush data to the peer API server.
func (s *SyncService) SyncBandwidth(data map[string]int64) {
	body, _ := json.Marshal(data)
	resp, err := s.client.Post(s.peerAPIURL+"/api/internal/bandwidth-flush?from_peer=1", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[sync] SyncBandwidth to peer failed: %v", err)
		return
	}
	resp.Body.Close()
	log.Printf("[sync] SyncBandwidth %d usernames to peer: %d", len(data), resp.StatusCode)
}
