package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

type RelayServerService struct {
	repo *repository.RelayServerRepository
}

func NewRelayServerService(repo *repository.RelayServerRepository) *RelayServerService {
	return &RelayServerService{repo: repo}
}

func (s *RelayServerService) List(ctx context.Context) ([]domain.RelayServer, error) {
	return s.repo.List(ctx)
}

func (s *RelayServerService) ListActive(ctx context.Context) ([]domain.RelayServer, error) {
	return s.repo.ListActive(ctx)
}

func (s *RelayServerService) Create(ctx context.Context, rs *domain.RelayServer) error {
	rs.ID = uuid.New()
	rs.Active = true
	return s.repo.Create(ctx, rs)
}

func (s *RelayServerService) GetByID(ctx context.Context, id uuid.UUID) (*domain.RelayServer, error) {
	return s.repo.GetByID(ctx, id)
}
