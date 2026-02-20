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
)

type DeviceService struct {
	deviceRepo     *repository.DeviceRepository
	ipHistRepo     *repository.IPHistoryRepository
	commandRepo    *repository.CommandRepository
	statusLogRepo  *repository.StatusLogRepository
	portService    *PortService
	vpnService     *VPNService
	tunnelPushURL  string // URL of tunnel server's push API (e.g. http://178.156.210.156:8081)
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

	// Log status transition
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
	if s.tunnelPushURL != "" {
		go s.pushCommandToTunnel(deviceID, cmd)
	}

	return cmd, nil
}

// pushCommandToTunnel sends a command to the tunnel server's push API for instant delivery.
func (s *DeviceService) pushCommandToTunnel(deviceID uuid.UUID, cmd *domain.DeviceCommand) {
	body, _ := json.Marshal(map[string]string{
		"device_id": deviceID.String(),
		"id":        cmd.ID.String(),
		"type":      string(cmd.Type),
		"payload":   cmd.Payload,
	})

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(s.tunnelPushURL+"/push-command", "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Push command to tunnel failed (device=%s): %v", deviceID, err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		log.Printf("Command %s pushed to device %s via tunnel", cmd.ID, deviceID)
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

func (s *DeviceService) GetByName(ctx context.Context, name string) (*domain.Device, error) {
	return s.deviceRepo.GetByName(ctx, name)
}

func (s *DeviceService) SetVpnIP(ctx context.Context, id uuid.UUID, vpnIP string) error {
	return s.deviceRepo.SetVpnIP(ctx, id, vpnIP)
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
