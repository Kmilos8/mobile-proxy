package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type RelayServerRepository struct {
	db *DB
}

func NewRelayServerRepository(db *DB) *RelayServerRepository {
	return &RelayServerRepository{db: db}
}

func (r *RelayServerRepository) Create(ctx context.Context, rs *domain.RelayServer) error {
	query := `INSERT INTO relay_servers (id, name, ip, location, active)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Pool.Exec(ctx, query, rs.ID, rs.Name, rs.IP, rs.Location, rs.Active)
	return err
}

func (r *RelayServerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RelayServer, error) {
	query := `SELECT id, name, ip, location, active, created_at, updated_at
		FROM relay_servers WHERE id = $1`
	var rs domain.RelayServer
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&rs.ID, &rs.Name, &rs.IP, &rs.Location, &rs.Active, &rs.CreatedAt, &rs.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get relay server: %w", err)
	}
	return &rs, nil
}

func (r *RelayServerRepository) ListActive(ctx context.Context) ([]domain.RelayServer, error) {
	query := `SELECT id, name, ip, location, active, created_at, updated_at
		FROM relay_servers WHERE active = TRUE ORDER BY name ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []domain.RelayServer
	for rows.Next() {
		var rs domain.RelayServer
		if err := rows.Scan(&rs.ID, &rs.Name, &rs.IP, &rs.Location, &rs.Active, &rs.CreatedAt, &rs.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan relay server: %w", err)
		}
		servers = append(servers, rs)
	}
	return servers, nil
}

func (r *RelayServerRepository) List(ctx context.Context) ([]domain.RelayServer, error) {
	query := `SELECT id, name, ip, location, active, created_at, updated_at
		FROM relay_servers ORDER BY name ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []domain.RelayServer
	for rows.Next() {
		var rs domain.RelayServer
		if err := rows.Scan(&rs.ID, &rs.Name, &rs.IP, &rs.Location, &rs.Active, &rs.CreatedAt, &rs.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan relay server: %w", err)
		}
		servers = append(servers, rs)
	}
	return servers, nil
}
