package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

type DeviceService struct {
	deviceRepo  *repository.DeviceRepository
	ipHistRepo  *repository.IPHistoryRepository
	commandRepo *repository.CommandRepository
	portService *PortService
}

func NewDeviceService(
	deviceRepo *repository.DeviceRepository,
	ipHistRepo *repository.IPHistoryRepository,
	commandRepo *repository.CommandRepository,
	portService *PortService,
) *DeviceService {
	return &DeviceService{
		deviceRepo:  deviceRepo,
		ipHistRepo:  ipHistRepo,
		commandRepo: commandRepo,
		portService: portService,
	}
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

	return &domain.DeviceRegistrationResponse{
		DeviceID: device.ID,
		BasePort: basePort,
	}, nil
}

func (s *DeviceService) Heartbeat(ctx context.Context, deviceID uuid.UUID, req *domain.HeartbeatRequest) (*domain.HeartbeatResponse, error) {
	// Get current device to check for IP change
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
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

	return cmd, nil
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

func (s *DeviceService) MarkStaleOffline(ctx context.Context) (int64, error) {
	return s.deviceRepo.MarkStaleOffline(ctx, 2*time.Minute)
}
