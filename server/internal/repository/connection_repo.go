package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type ConnectionRepository struct {
	db *DB
}

func NewConnectionRepository(db *DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

func (r *ConnectionRepository) Create(ctx context.Context, c *domain.ProxyConnection) error {
	query := `INSERT INTO proxy_connections (id, device_id, customer_id, username, password_hash, ip_whitelist, bandwidth_limit, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Pool.Exec(ctx, query,
		c.ID, c.DeviceID, c.CustomerID, c.Username, c.PasswordHash,
		c.IPWhitelist, c.BandwidthLimit, c.Active)
	return err
}

func (r *ConnectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProxyConnection, error) {
	query := `SELECT id, device_id, customer_id, username, password_hash, ip_whitelist,
		bandwidth_limit, bandwidth_used, active, expires_at, created_at, updated_at
		FROM proxy_connections WHERE id = $1`
	var c domain.ProxyConnection
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.DeviceID, &c.CustomerID, &c.Username, &c.PasswordHash,
		&c.IPWhitelist, &c.BandwidthLimit, &c.BandwidthUsed, &c.Active,
		&c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	return &c, nil
}

func (r *ConnectionRepository) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.ProxyConnection, error) {
	query := `SELECT id, device_id, customer_id, username, password_hash, ip_whitelist,
		bandwidth_limit, bandwidth_used, active, expires_at, created_at, updated_at
		FROM proxy_connections WHERE device_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Pool.Query(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conns []domain.ProxyConnection
	for rows.Next() {
		var c domain.ProxyConnection
		err := rows.Scan(
			&c.ID, &c.DeviceID, &c.CustomerID, &c.Username, &c.PasswordHash,
			&c.IPWhitelist, &c.BandwidthLimit, &c.BandwidthUsed, &c.Active,
			&c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan connection: %w", err)
		}
		conns = append(conns, c)
	}
	return conns, nil
}

func (r *ConnectionRepository) List(ctx context.Context) ([]domain.ProxyConnection, error) {
	query := `SELECT id, device_id, customer_id, username, password_hash, ip_whitelist,
		bandwidth_limit, bandwidth_used, active, expires_at, created_at, updated_at
		FROM proxy_connections ORDER BY created_at DESC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conns []domain.ProxyConnection
	for rows.Next() {
		var c domain.ProxyConnection
		err := rows.Scan(
			&c.ID, &c.DeviceID, &c.CustomerID, &c.Username, &c.PasswordHash,
			&c.IPWhitelist, &c.BandwidthLimit, &c.BandwidthUsed, &c.Active,
			&c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan connection: %w", err)
		}
		conns = append(conns, c)
	}
	return conns, nil
}

func (r *ConnectionRepository) UpdateActive(ctx context.Context, id uuid.UUID, active bool) error {
	query := `UPDATE proxy_connections SET active = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, active)
	return err
}

func (r *ConnectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM proxy_connections WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *ConnectionRepository) CountActive(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM proxy_connections WHERE active = TRUE`
	var count int
	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *ConnectionRepository) GetByUsername(ctx context.Context, username string) (*domain.ProxyConnection, error) {
	query := `SELECT id, device_id, customer_id, username, password_hash, ip_whitelist,
		bandwidth_limit, bandwidth_used, active, expires_at, created_at, updated_at
		FROM proxy_connections WHERE username = $1 AND active = TRUE`
	var c domain.ProxyConnection
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&c.ID, &c.DeviceID, &c.CustomerID, &c.Username, &c.PasswordHash,
		&c.IPWhitelist, &c.BandwidthLimit, &c.BandwidthUsed, &c.Active,
		&c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get connection by username: %w", err)
	}
	return &c, nil
}
