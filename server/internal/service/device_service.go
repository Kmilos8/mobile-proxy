package service

import (
	"bytes"
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
)

type DeviceService struct {
	deviceRepo      *repository.DeviceRepository
	ipHistRepo      *repository.IPHistoryRepository
	commandRepo     *repository.CommandRepository
	statusLogRepo   *repository.StatusLogRepository
	relayServerRepo *repository.RelayServerRepository
	userRepo        *repository.UserRepository
	portService     *PortService
	vpnService      *VPNService
	tunnelPushURL   string // fallback static URL (e.g. http://178.156.210.156:8081)
}

func NewDeviceService(
	deviceRepo *repository.DeviceRepository,
	ipHistRepo *repository.IPHistoryRepository,
	commandRepo *repository.CommandRepository,
	portService *PortService,
	vpnService *VPNService,
) *DeviceService {
	return &DeviceService{
		deviceRepo:  deviceRepo,
		ipHistRepo:  ipHistRepo,
		commandRepo: commandRepo,
		portService: portService,
		vpnService:  vpnService,
	}
}

// SetTunnelPushURL configures the tunnel server push endpoint for instant command delivery.
func (s *DeviceService) SetTunnelPushURL(url string) {
	s.tunnelPushURL = url
}

// SetStatusLogRepo configures the status log repository for tracking status transitions.
func (s *DeviceService) SetStatusLogRepo(repo *repository.StatusLogRepository) {
	s.statusLogRepo = repo
}

// SetUserRepo configures the user repository for webhook dispatch.
func (s *DeviceService) SetUserRepo(repo *repository.UserRepository) {
	s.userRepo = repo
}

// SetRelayServerRepo configures the relay server repository for dynamic tunnel URL resolution.
func (s *DeviceService) SetRelayServerRepo(repo *repository.RelayServerRepository) {
	s.relayServerRepo = repo
}

// getTunnelPushURL resolves the tunnel push URL for a device based on its relay server.
func (s *DeviceService) getTunnelPushURL(ctx context.Context, deviceID uuid.UUID) string {
	if s.relayServerRepo != nil {
		device, err := s.deviceRepo.GetByID(ctx, deviceID)
		if err == nil && device.RelayServerID != nil {
			if rs, err := s.relayServerRepo.GetByID(ctx, *device.RelayServerID); err == nil {
				return fmt.Sprintf("http://%s:8081", rs.IP)
			}
		}
	}
	return s.tunnelPushURL
}

func (s *DeviceService) Register(ctx context.Context, req *domain.DeviceRegistrationRequest) (*domain.DeviceRegistrationResponse, error) {
	// Check if device already registered
	existing, err := s.deviceRepo.GetByAndroidID(ctx, req.AndroidID)
	if err == nil && existing != nil {
		return &domain.DeviceRegistrationResponse{
			DeviceID: existing.ID,
			BasePort: existing.BasePort,
		}, nil
	}

	// Allocate ports
	basePort, err := s.portService.AllocatePort(ctx)
	if err != nil {
		return nil, fmt.Errorf("allocate port: %w", err)
	}

	device := &domain.Device{
		ID:             uuid.New(),
		Name:           req.Name,
		AndroidID:      req.AndroidID,
		Status:         domain.DeviceStatusOffline,
		BasePort:       basePort,
		HTTPPort:       basePort,
		SOCKS5Port:     basePort + 1,
		UDPRelayPort:   basePort + 2,
		OVPNPort:       basePort + 3,
		DeviceModel:    req.DeviceModel,
		AndroidVersion: req.AndroidVersion,
		AppVersion:     req.AppVersion,
	}
	if device.Name == "" {
		device.Name = fmt.Sprintf("%s-%s", req.DeviceModel, req.AndroidID[:8])
	}

	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}

	// Generate VPN config if VPN service is available
	var vpnConfig string
	if s.vpnService != nil {
		// Assign VPN IP based on port offset (unique per device)
		deviceIndex := (basePort - 30000) / 4
		vpnIP, err := s.vpnService.AssignVPNIP(device.Name, deviceIndex)
		if err != nil {
			return nil, fmt.Errorf("assign vpn ip: %w", err)
		}
		_ = s.deviceRepo.SetVpnIP(ctx, device.ID, vpnIP)
		vpnConfig = s.vpnService.GenerateClientConfig(s.vpnService.ServerIP(), device.Name)
	}

	return &domain.DeviceRegistrationResponse{
		DeviceID:  device.ID,
		VpnConfig: vpnConfig,
		BasePort:  basePort,
	}, nil
}

func (s *DeviceService) Heartbeat(ctx context.Context, deviceID uuid.UUID, req *domain.HeartbeatRequest) (*domain.HeartbeatResponse, error) {
	// Get current device to check for IP change
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	// Log status transition and send recovery webhook if device was offline
	wasOffline := device.Status == domain.DeviceStatusOffline
	if s.statusLogRepo != nil {
		now := time.Now().UTC()
		if device.Status != domain.DeviceStatusOnline {
			// Device was not online, heartbeat sets it online — log the transition
			_ = s.statusLogRepo.Insert(ctx, deviceID, string(domain.DeviceStatusOnline), string(device.Status), now)
		} else {
			// Device already online — ensure there's at least one log entry for today
			if hasLogs, err := s.statusLogRepo.HasLogsForDate(ctx, deviceID, now); err == nil && !hasLogs {
				_ = s.statusLogRepo.Insert(ctx, deviceID, string(domain.DeviceStatusOnline), string(domain.DeviceStatusOnline), now)
			}
		}
	}
	// Send recovery webhook if device transitioned from offline to online
	if wasOffline {
		go s.sendRecoveryWebhook(ctx, *device)
	}

	// Update heartbeat
	if err := s.deviceRepo.UpdateHeartbeat(ctx, deviceID, req); err != nil {
		return nil, fmt.Errorf("update heartbeat: %w", err)
	}

	// Check for IP change
	if req.CellularIP != "" && req.CellularIP != device.CellularIP {
		ipHist := &domain.IPHistory{
			ID:       uuid.New(),
			DeviceID: deviceID,
			IP:       req.CellularIP,
			Method:   "natural",
		}
		_ = s.ipHistRepo.Create(ctx, ipHist)
	}

	// Get pending commands
	commands, err := s.commandRepo.GetPending(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get commands: %w", err)
	}

	// Mark commands as sent
	if len(commands) > 0 {
		ids := make([]uuid.UUID, len(commands))
		for i, cmd := range commands {
			ids[i] = cmd.ID
		}
		_ = s.commandRepo.MarkAsSent(ctx, ids)
	}

	return &domain.HeartbeatResponse{
		Commands: commands,
	}, nil
}

func (s *DeviceService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	return s.deviceRepo.GetByID(ctx, id)
}

func (s *DeviceService) List(ctx context.Context) ([]domain.Device, error) {
	return s.deviceRepo.List(ctx)
}

// ListByCustomer returns devices owned by or shared with the given customer.
func (s *DeviceService) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Device, error) {
	return s.deviceRepo.ListByCustomer(ctx, customerID)
}

// GetByIDForCustomer returns a device only if the customer owns it or has shared access.
func (s *DeviceService) GetByIDForCustomer(ctx context.Context, deviceID uuid.UUID, customerID uuid.UUID) (*domain.Device, error) {
	return s.deviceRepo.GetByIDForCustomer(ctx, deviceID, customerID)
}

func (s *DeviceService) SendCommand(ctx context.Context, deviceID uuid.UUID, req *domain.CommandRequest) (*domain.DeviceCommand, error) {
	cmd := &domain.DeviceCommand{
		ID:       uuid.New(),
		DeviceID: deviceID,
		Type:     req.Type,
		Status:   domain.CommandStatusPending,
		Payload:  req.Payload,
	}

	if err := s.commandRepo.Create(ctx, cmd); err != nil {
		return nil, fmt.Errorf("create command: %w", err)
	}

	// Push command to device via tunnel for instant delivery (fire-and-forget)
	tunnelURL := s.getTunnelPushURL(ctx, deviceID)
	if tunnelURL != "" {
		go s.pushCommandToTunnel(tunnelURL, deviceID, cmd)
	}

	return cmd, nil
}

// pushCommandToTunnel sends a command to the tunnel server's push API for instant delivery.
func (s *DeviceService) pushCommandToTunnel(tunnelURL string, deviceID uuid.UUID, cmd *domain.DeviceCommand) {
	body, _ := json.Marshal(map[string]string{
		"device_id": deviceID.String(),
		"id":        cmd.ID.String(),
		"type":      string(cmd.Type),
		"payload":   cmd.Payload,
	})

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(tunnelURL+"/push-command", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Push command to tunnel failed (device=%s): %v", deviceID, err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		log.Printf("Command %s pushed to device %s via tunnel (%s)", cmd.ID, deviceID, tunnelURL)
	} else {
		log.Printf("Push command returned %d for device %s (device may be offline, will deliver via heartbeat)", resp.StatusCode, deviceID)
	}
}

func (s *DeviceService) UpdateCommandStatus(ctx context.Context, commandID uuid.UUID, status domain.CommandStatus, result string) error {
	return s.commandRepo.UpdateStatus(ctx, commandID, status, result)
}

func (s *DeviceService) GetIPHistory(ctx context.Context, deviceID uuid.UUID, limit int) ([]domain.IPHistory, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.ipHistRepo.ListByDevice(ctx, deviceID, limit)
}

func (s *DeviceService) GetCommandHistory(ctx context.Context, deviceID uuid.UUID, limit int) ([]domain.DeviceCommand, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.commandRepo.GetByDevice(ctx, deviceID, limit)
}

func (s *DeviceService) UpdateNameDescription(ctx context.Context, id uuid.UUID, name, description string) error {
	return s.deviceRepo.UpdateNameDescription(ctx, id, name, description)
}

func (s *DeviceService) GetByName(ctx context.Context, name string) (*domain.Device, error) {
	return s.deviceRepo.GetByName(ctx, name)
}

func (s *DeviceService) SetVpnIP(ctx context.Context, id uuid.UUID, vpnIP string) error {
	return s.deviceRepo.SetVpnIP(ctx, id, vpnIP)
}

func (s *DeviceService) UpdateAutoRotate(ctx context.Context, id uuid.UUID, minutes int) error {
	return s.deviceRepo.UpdateAutoRotate(ctx, id, minutes)
}

// RunAutoRotations checks all online devices with auto_rotate_minutes > 0 and sends
// a rotate command if enough time has passed since the last rotation.
func (s *DeviceService) RunAutoRotations(ctx context.Context) error {
	devices, err := s.deviceRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list devices: %w", err)
	}

	now := time.Now()
	for _, d := range devices {
		if d.Status != domain.DeviceStatusOnline || d.AutoRotateMinutes <= 0 {
			continue
		}

		lastCmd, err := s.commandRepo.GetLastRotationCommand(ctx, d.ID)
		if err == nil && lastCmd != nil {
			nextRotation := lastCmd.CreatedAt.Add(time.Duration(d.AutoRotateMinutes) * time.Minute)
			if now.Before(nextRotation) {
				continue
			}
		}
		// No previous command or interval has elapsed — send rotation
		_, err = s.SendCommand(ctx, d.ID, &domain.CommandRequest{
			Type:    "rotate_ip_airplane",
			Payload: "{}",
		})
		if err != nil {
			log.Printf("Auto-rotate failed for device %s: %v", d.ID, err)
		} else {
			log.Printf("Auto-rotate triggered for device %s (interval: %dm)", d.ID, d.AutoRotateMinutes)
		}
	}
	return nil
}

func (s *DeviceService) MarkStaleOffline(ctx context.Context) (int64, error) {
	return s.deviceRepo.MarkStaleOffline(ctx, 2*time.Minute)
}

// MarkStaleOfflineWithLogs marks stale devices offline and logs the status transitions.
func (s *DeviceService) MarkStaleOfflineWithLogs(ctx context.Context) (int64, error) {
	if s.statusLogRepo == nil {
		return s.MarkStaleOffline(ctx)
	}

	// Get online devices that are stale
	devices, err := s.deviceRepo.List(ctx)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	threshold := now.Add(-2 * time.Minute)
	var count int64
	for _, d := range devices {
		if d.Status == domain.DeviceStatusOnline && d.LastHeartbeat != nil && d.LastHeartbeat.Before(threshold) {
			if err := s.deviceRepo.UpdateStatus(ctx, d.ID, domain.DeviceStatusOffline); err != nil {
				log.Printf("Error updating device %s status: %v", d.ID, err)
				continue
			}
			_ = s.statusLogRepo.Insert(ctx, d.ID, string(domain.DeviceStatusOffline), string(d.Status), now)
			go s.sendOfflineWebhook(ctx, d)
			count++
		}
	}
	return count, nil
}

func (s *DeviceService) GetUptimeSegments(ctx context.Context, deviceID uuid.UUID, date time.Time, tzOffsetMinutes int) ([]domain.UptimeSegment, error) {
	if s.statusLogRepo == nil {
		return nil, nil
	}

	// Compute day boundaries in UTC adjusted for timezone
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(time.Duration(tzOffsetMinutes) * time.Minute)
	dayEnd := dayStart.Add(24 * time.Hour)

	// Cap dayEnd to now if the date is today
	now := time.Now().UTC()
	if dayEnd.After(now) {
		dayEnd = now
	}

	logs, err := s.statusLogRepo.GetByDeviceAndRange(ctx, deviceID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		// No transitions in this range — try to find the last known status before this range
		initialStatus := "unknown"
		if lastLog, err := s.statusLogRepo.GetLastStatusBefore(ctx, deviceID, dayStart); err == nil && lastLog != nil {
			initialStatus = lastLog.Status
		}
		return []domain.UptimeSegment{
			{Status: initialStatus, StartTime: dayStart, EndTime: dayEnd},
		}, nil
	}

	var segments []domain.UptimeSegment

	// First segment: from day start to first transition
	if logs[0].ChangedAt.After(dayStart) {
		segments = append(segments, domain.UptimeSegment{
			Status:    logs[0].PreviousStatus,
			StartTime: dayStart,
			EndTime:   logs[0].ChangedAt,
		})
	}

	// Build segments between transitions
	for i, l := range logs {
		end := dayEnd
		if i+1 < len(logs) {
			end = logs[i+1].ChangedAt
		}
		segments = append(segments, domain.UptimeSegment{
			Status:    l.Status,
			StartTime: l.ChangedAt,
			EndTime:   end,
		})
	}

	return segments, nil
}

// sendOfflineWebhook dispatches a webhook notification when a device goes offline.
// Respects a 5-minute cooldown per device to avoid duplicate alerts.
func (s *DeviceService) sendOfflineWebhook(ctx context.Context, d domain.Device) {
	if s.userRepo == nil {
		return
	}

	// Check 5-minute cooldown
	lastAlert, err := s.deviceRepo.GetLastOfflineAlertAt(ctx, d.ID)
	if err == nil && lastAlert != nil && time.Since(*lastAlert) < 5*time.Minute {
		return // cooldown active
	}

	webhookURL, err := s.userRepo.GetWebhookURLForDevice(ctx, d.ID)
	if err != nil || webhookURL == "" {
		return
	}

	var lastSeen string
	if d.LastHeartbeat != nil {
		lastSeen = d.LastHeartbeat.Format(time.RFC3339)
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"event":       "device.offline",
		"device_id":   d.ID.String(),
		"device_name": d.Name,
		"last_seen":   lastSeen,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("[webhook] offline alert failed for device %s: %v", d.ID, err)
		return
	}
	resp.Body.Close()

	// Only update cooldown on successful delivery
	_ = s.deviceRepo.SetLastOfflineAlertAt(ctx, d.ID, time.Now().UTC())
	log.Printf("[webhook] offline alert sent for device %s (status=%d)", d.Name, resp.StatusCode)
}

// sendRecoveryWebhook dispatches a webhook notification when a device comes back online.
func (s *DeviceService) sendRecoveryWebhook(ctx context.Context, d domain.Device) {
	if s.userRepo == nil {
		return
	}

	webhookURL, err := s.userRepo.GetWebhookURLForDevice(ctx, d.ID)
	if err != nil || webhookURL == "" {
		return
	}

	now := time.Now().UTC()
	payload, _ := json.Marshal(map[string]interface{}{
		"event":          "device.online",
		"device_id":      d.ID.String(),
		"device_name":    d.Name,
		"reconnected_at": now.Format(time.RFC3339),
		"timestamp":      now.Format(time.RFC3339),
	})

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("[webhook] recovery alert failed for device %s: %v", d.ID, err)
		return
	}
	resp.Body.Close()
	log.Printf("[webhook] recovery alert sent for device %s (status=%d)", d.Name, resp.StatusCode)
}
