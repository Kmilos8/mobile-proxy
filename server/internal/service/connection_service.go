package service

import (
	"context"
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
	connRepo      *repository.ConnectionRepository
	deviceRepo    *repository.DeviceRepository
	portService   *PortService
	tunnelPushURL string
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
	if proxyType != "http" && proxyType != "socks5" {
		return nil, fmt.Errorf("invalid proxy_type: must be 'http' or 'socks5'")
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
		PasswordPlain:  req.Password,
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
	if s.portService != nil {
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
	if conn.BasePort != nil && device.VpnIP != "" && s.tunnelPushURL != "" {
		go s.refreshDNAT(device.ID.String(), *conn.BasePort, device.VpnIP, proxyType)
	}

	return conn, nil
}

func (s *ConnectionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProxyConnection, error) {
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	conn.Password = conn.PasswordPlain
	return conn, nil
}

func (s *ConnectionService) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.ProxyConnection, error) {
	conns, err := s.connRepo.ListByDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	for i := range conns {
		conns[i].Password = conns[i].PasswordPlain
	}
	return conns, nil
}

func (s *ConnectionService) List(ctx context.Context) ([]domain.ProxyConnection, error) {
	conns, err := s.connRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range conns {
		conns[i].Password = conns[i].PasswordPlain
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
	if conn.BasePort != nil && s.tunnelPushURL != "" {
		device, err := s.deviceRepo.GetByID(ctx, conn.DeviceID)
		if err == nil && device.VpnIP != "" {
			go s.teardownDNAT(device.ID.String(), *conn.BasePort, device.VpnIP, conn.ProxyType)
		}
	}

	return nil
}

func (s *ConnectionService) refreshDNAT(deviceID string, basePort int, vpnIP string, proxyType string) {
	body, _ := json.Marshal(map[string]interface{}{
		"device_id":  deviceID,
		"base_port":  basePort,
		"vpn_ip":     vpnIP,
		"proxy_type": proxyType,
	})
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(s.tunnelPushURL+"/refresh-dnat", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Refresh DNAT failed (device=%s port=%d): %v", deviceID, basePort, err)
		return
	}
	resp.Body.Close()
	log.Printf("Refresh DNAT sent for device=%s port=%d type=%s", deviceID, basePort, proxyType)
}

func (s *ConnectionService) teardownDNAT(deviceID string, basePort int, vpnIP string, proxyType string) {
	body, _ := json.Marshal(map[string]interface{}{
		"device_id":  deviceID,
		"base_port":  basePort,
		"vpn_ip":     vpnIP,
		"proxy_type": proxyType,
	})
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(s.tunnelPushURL+"/teardown-dnat", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Teardown DNAT failed (device=%s port=%d): %v", deviceID, basePort, err)
		return
	}
	resp.Body.Close()
	log.Printf("Teardown DNAT sent for device=%s port=%d type=%s", deviceID, basePort, proxyType)
}
