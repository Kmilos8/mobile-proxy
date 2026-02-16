package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type ConnectionService struct {
	connRepo   *repository.ConnectionRepository
	deviceRepo *repository.DeviceRepository
}

func NewConnectionService(connRepo *repository.ConnectionRepository, deviceRepo *repository.DeviceRepository) *ConnectionService {
	return &ConnectionService{
		connRepo:   connRepo,
		deviceRepo: deviceRepo,
	}
}

func (s *ConnectionService) Create(ctx context.Context, req *domain.CreateConnectionRequest) (*domain.ProxyConnection, error) {
	// Verify device exists
	_, err := s.deviceRepo.GetByID(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
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
	}
	if conn.IPWhitelist == nil {
		conn.IPWhitelist = []string{}
	}

	if err := s.connRepo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	return conn, nil
}

func (s *ConnectionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProxyConnection, error) {
	return s.connRepo.GetByID(ctx, id)
}

func (s *ConnectionService) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.ProxyConnection, error) {
	return s.connRepo.ListByDevice(ctx, deviceID)
}

func (s *ConnectionService) List(ctx context.Context) ([]domain.ProxyConnection, error) {
	return s.connRepo.List(ctx)
}

func (s *ConnectionService) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.connRepo.UpdateActive(ctx, id, active)
}

func (s *ConnectionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.connRepo.Delete(ctx, id)
}
