package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type ConnectionService struct {
	connRepo        *repository.ConnectionRepository
	deviceRepo      *repository.DeviceRepository
	relayServerRepo *repository.RelayServerRepository
	portService     *PortService
	tunnelPushURL   string // fallback static URL
	syncService     *SyncService
}

func (s *ConnectionService) SetSyncService(ss *SyncService) {
	s.syncService = ss
}

func NewConnectionService(connRepo *repository.ConnectionRepository, deviceRepo *repository.DeviceRepository) *ConnectionService {
	return &ConnectionService{
		connRepo:   connRepo,
		deviceRepo: deviceRepo,
	}
}

func (s *ConnectionService) SetPortService(ps *PortService) {
	s.portService = ps
}

func (s *ConnectionService) SetTunnelPushURL(url string) {
	s.tunnelPushURL = url
}

func (s *ConnectionService) SetRelayServerRepo(repo *repository.RelayServerRepository) {
	s.relayServerRepo = repo
}

// getTunnelPushURL resolves the tunnel push URL for a device based on its relay server.
// Falls back to the static tunnelPushURL if relay server is not available.
func (s *ConnectionService) getTunnelPushURL(ctx context.Context, device *domain.Device) string {
	if device.RelayServerID != nil && s.relayServerRepo != nil {
		if rs, err := s.relayServerRepo.GetByID(ctx, *device.RelayServerID); err == nil {
			return fmt.Sprintf("http://%s:8081", rs.IP)
		}
	}
	return s.tunnelPushURL
}

func (s *ConnectionService) Create(ctx context.Context, req *domain.CreateConnectionRequest) (*domain.ProxyConnection, error) {
	// Verify device exists
	device, err := s.deviceRepo.GetByID(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Default proxy type to "http"
	proxyType := req.ProxyType
	if proxyType == "" {
		proxyType = "http"
	}
	if proxyType != "http" && proxyType != "socks5" && proxyType != "openvpn" {
		return nil, fmt.Errorf("invalid proxy_type: must be 'http', 'socks5', or 'openvpn'")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	conn := &domain.ProxyConnection{
		ID:             uuid.New(),
		DeviceID:       req.DeviceID,
		CustomerID:     req.CustomerID,
		Username:       req.Username,
		PasswordHash:   string(hash),
		Password:       req.Password, // Return plaintext on creation only
		IPWhitelist:    req.IPWhitelist,
		BandwidthLimit: req.BandwidthLimit,
		Active:         true,
		ProxyType:      proxyType,
	}
	if conn.IPWhitelist == nil {
		conn.IPWhitelist = []string{}
	}

	// Allocate a unique port for this connection (single port based on type)
	// OpenVPN uses the shared server port 1195 — no port allocation needed
	if s.portService != nil && proxyType != "openvpn" {
		basePort, err := s.portService.AllocatePort(ctx)
		if err != nil {
			return nil, fmt.Errorf("allocate connection port: %w", err)
		}
		conn.BasePort = &basePort
		if proxyType == "http" {
			conn.HTTPPort = &basePort
		} else {
			conn.SOCKS5Port = &basePort
		}
	}

	if err := s.connRepo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	// If device is online, trigger DNAT refresh for the new connection port
	// Skip DNAT for openvpn — it uses the shared VPN server port, not per-connection ports
	tunnelURL := s.getTunnelPushURL(ctx, device)
	if conn.BasePort != nil && device.VpnIP != "" && tunnelURL != "" && proxyType != "openvpn" {
		go s.refreshDNAT(tunnelURL, device.ID.String(), *conn.BasePort, device.VpnIP, proxyType)
	}

	// Sync all connections for this device to peer server
	if s.syncService != nil {
		conns, err := s.connRepo.ListByDevice(ctx, req.DeviceID)
		if err == nil {
			go s.syncService.SyncConnections(req.DeviceID, conns)
		}
	}

	return conn, nil
}

func (s *ConnectionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProxyConnection, error) {
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *ConnectionService) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.ProxyConnection, error) {
	conns, err := s.connRepo.ListByDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return conns, nil
}

func (s *ConnectionService) List(ctx context.Context) ([]domain.ProxyConnection, error) {
	conns, err := s.connRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	return conns, nil
}

func (s *ConnectionService) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.connRepo.UpdateActive(ctx, id, active)
}

func (s *ConnectionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get connection before deleting to know its ports
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get connection: %w", err)
	}

	if err := s.connRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Tear down DNAT for the connection's port
	if conn.BasePort != nil {
		device, err := s.deviceRepo.GetByID(ctx, conn.DeviceID)
		if err == nil && device.VpnIP != "" {
			tunnelURL := s.getTunnelPushURL(ctx, device)
			if tunnelURL != "" {
				go s.teardownDNAT(tunnelURL, device.ID.String(), *conn.BasePort, device.VpnIP, conn.ProxyType)
			}
		}
	}

	// Sync all connections for this device to peer server
	if s.syncService != nil {
		conns, err := s.connRepo.ListByDevice(ctx, conn.DeviceID)
		if err == nil {
			go s.syncService.SyncConnections(conn.DeviceID, conns)
		}
	}

	return nil
}

func (s *ConnectionService) RegeneratePassword(ctx context.Context, id uuid.UUID) (string, error) {
	// Generate 16-character random password
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	newPass := base64.URLEncoding.EncodeToString(b)[:16]

	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), 12)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	if err := s.connRepo.UpdatePasswordHash(ctx, id, string(hash)); err != nil {
		return "", fmt.Errorf("update password: %w", err)
	}

	// Sync credentials to device via heartbeat (device will get new hash as SOCKS5 credential)
	conn, err := s.connRepo.GetByID(ctx, id)
	if err == nil && s.syncService != nil {
		conns, err := s.connRepo.ListByDevice(ctx, conn.DeviceID)
		if err == nil {
			go s.syncService.SyncConnections(conn.DeviceID, conns)
		}
	}

	return newPass, nil
}

func (s *ConnectionService) refreshDNAT(tunnelURL string, deviceID string, basePort int, vpnIP string, proxyType string) {
	body, _ := json.Marshal(map[string]interface{}{
		"device_id":  deviceID,
		"base_port":  basePort,
		"vpn_ip":     vpnIP,
		"proxy_type": proxyType,
	})
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(tunnelURL+"/refresh-dnat", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Refresh DNAT failed (device=%s port=%d): %v", deviceID, basePort, err)
		return
	}
	resp.Body.Close()
	log.Printf("Refresh DNAT sent for device=%s port=%d type=%s via %s", deviceID, basePort, proxyType, tunnelURL)
}

func (s *ConnectionService) UpdateBandwidthUsedByUsername(ctx context.Context, username string, used int64) error {
	return s.connRepo.UpdateBandwidthUsed(ctx, username, used)
}

func (s *ConnectionService) ResetBandwidth(ctx context.Context, id uuid.UUID) error {
	if err := s.connRepo.ResetBandwidthUsed(ctx, id); err != nil {
		return err
	}
	// Best-effort: reset tunnel in-memory counter to prevent 30s flush overwriting the zero
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return nil // DB reset succeeded; tunnel reset is best-effort
	}
	device, err := s.deviceRepo.GetByID(ctx, conn.DeviceID)
	if err != nil {
		return nil
	}
	tunnelURL := s.getTunnelPushURL(ctx, device)
	if tunnelURL != "" {
		go s.resetTunnelBandwidth(tunnelURL, conn.Username)
	}
	return nil
}

func (s *ConnectionService) resetTunnelBandwidth(tunnelURL, username string) {
	body, _ := json.Marshal(map[string]string{"username": username})
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(tunnelURL+"/openvpn-client-reset-bandwidth", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("[reset-bandwidth] tunnel call failed for %s: %v", username, err)
		return
	}
	resp.Body.Close()
}

func (s *ConnectionService) teardownDNAT(tunnelURL string, deviceID string, basePort int, vpnIP string, proxyType string) {
	body, _ := json.Marshal(map[string]interface{}{
		"device_id":  deviceID,
		"base_port":  basePort,
		"vpn_ip":     vpnIP,
		"proxy_type": proxyType,
	})
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(tunnelURL+"/teardown-dnat", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Teardown DNAT failed (device=%s port=%d): %v", deviceID, basePort, err)
		return
	}
	resp.Body.Close()
	log.Printf("Teardown DNAT sent for device=%s port=%d type=%s via %s", deviceID, basePort, proxyType, tunnelURL)
}
