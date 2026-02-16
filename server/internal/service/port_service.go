package service

import (
	"context"
	"fmt"

	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

type PortService struct {
	deviceRepo *repository.DeviceRepository
	config     domain.PortConfig
}

func NewPortService(deviceRepo *repository.DeviceRepository, config domain.PortConfig) *PortService {
	return &PortService{
		deviceRepo: deviceRepo,
		config:     config,
	}
}

func (s *PortService) AllocatePort(ctx context.Context) (int, error) {
	port, err := s.deviceRepo.GetNextBasePort(ctx)
	if err != nil {
		return 0, fmt.Errorf("get next port: %w", err)
	}

	if port+s.config.PortsPerDevice > s.config.MaxPort {
		return 0, fmt.Errorf("no more ports available (max: %d)", s.config.MaxPort)
	}

	return port, nil
}

// GetPortsForDevice returns the 4 ports assigned to a device
func (s *PortService) GetPortsForDevice(basePort int) (httpPort, socks5Port, udpPort, ovpnPort int) {
	return basePort, basePort + 1, basePort + 2, basePort + 3
}
