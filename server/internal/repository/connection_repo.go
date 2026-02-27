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
	query := `INSERT INTO proxy_connections (id, device_id, customer_id, username, password_hash, password_plain, ip_whitelist, bandwidth_limit, active, proxy_type, base_port, http_port, socks5_port)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.Pool.Exec(ctx, query,
		c.ID, c.DeviceID, c.CustomerID, c.Username, c.PasswordHash, c.PasswordPlain,
		c.IPWhitelist, c.BandwidthLimit, c.Active, c.ProxyType,
		c.BasePort, c.HTTPPort, c.SOCKS5Port)
	return err
}

const connSelectCols = `id, device_id, customer_id, username, password_hash, password_plain, ip_whitelist,
		bandwidth_limit, bandwidth_used, active, proxy_type, base_port, http_port, socks5_port,
		expires_at, created_at, updated_at`

func (r *ConnectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProxyConnection, error) {
	query := `SELECT ` + connSelectCols + ` FROM proxy_connections WHERE id = $1`
	return r.scanConnection(r.db.Pool.QueryRow(ctx, query, id))
}

func (r *ConnectionRepository) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.ProxyConnection, error) {
	query := `SELECT ` + connSelectCols + ` FROM proxy_connections WHERE device_id = $1 ORDER BY created_at DESC`
	return r.scanConnections(ctx, query, deviceID)
}

func (r *ConnectionRepository) List(ctx context.Context) ([]domain.ProxyConnection, error) {
	query := `SELECT ` + connSelectCols + ` FROM proxy_connections ORDER BY created_at DESC`
	return r.scanConnections(ctx, query)
}

func (r *ConnectionRepository) ReassignAllByDeviceID(ctx context.Context, oldDeviceID, newDeviceID uuid.UUID) error {
	query := `UPDATE proxy_connections SET device_id = $2, updated_at = NOW() WHERE device_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, oldDeviceID, newDeviceID)
	return err
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

func (r *ConnectionRepository) UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string, plain string) error {
	query := `UPDATE proxy_connections SET password_hash = $2, password_plain = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, hash, plain)
	return err
}

func (r *ConnectionRepository) UpdateBandwidthUsed(ctx context.Context, username string, used int64) error {
	query := `UPDATE proxy_connections SET bandwidth_used = $2, updated_at = NOW() WHERE username = $1`
	_, err := r.db.Pool.Exec(ctx, query, username, used)
	return err
}

func (r *ConnectionRepository) ResetBandwidthUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE proxy_connections SET bandwidth_used = 0, updated_at = NOW() WHERE id = $1`
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
	query := `SELECT ` + connSelectCols + ` FROM proxy_connections WHERE username = $1 AND active = TRUE`
	return r.scanConnection(r.db.Pool.QueryRow(ctx, query, username))
}

func (r *ConnectionRepository) ExistsByDeviceAndUsername(ctx context.Context, deviceID uuid.UUID, username string) (bool, error) {
	query := `SELECT COUNT(*) FROM proxy_connections WHERE device_id = $1 AND username = $2 AND active = TRUE`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, deviceID, username).Scan(&count)
	return count > 0, err
}

func (r *ConnectionRepository) ReplaceAllByDeviceID(ctx context.Context, deviceID uuid.UUID, conns []domain.ProxyConnection) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM proxy_connections WHERE device_id = $1`, deviceID); err != nil {
		return fmt.Errorf("delete connections: %w", err)
	}

	for _, c := range conns {
		query := `INSERT INTO proxy_connections (id, device_id, customer_id, username, password_hash, password_plain, ip_whitelist, bandwidth_limit, bandwidth_used, active, proxy_type, base_port, http_port, socks5_port, expires_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
		if _, err := tx.Exec(ctx, query,
			c.ID, c.DeviceID, c.CustomerID, c.Username, c.PasswordHash, c.PasswordPlain,
			c.IPWhitelist, c.BandwidthLimit, c.BandwidthUsed, c.Active, c.ProxyType,
			c.BasePort, c.HTTPPort, c.SOCKS5Port, c.ExpiresAt, c.CreatedAt, c.UpdatedAt); err != nil {
			return fmt.Errorf("insert connection %s: %w", c.ID, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *ConnectionRepository) scanConnection(row interface{ Scan(dest ...interface{}) error }) (*domain.ProxyConnection, error) {
	var c domain.ProxyConnection
	err := row.Scan(
		&c.ID, &c.DeviceID, &c.CustomerID, &c.Username, &c.PasswordHash, &c.PasswordPlain,
		&c.IPWhitelist, &c.BandwidthLimit, &c.BandwidthUsed, &c.Active, &c.ProxyType,
		&c.BasePort, &c.HTTPPort, &c.SOCKS5Port,
		&c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan connection: %w", err)
	}
	if c.PasswordPlain != nil {
		c.Password = *c.PasswordPlain
	}
	return &c, nil
}

func (r *ConnectionRepository) scanConnections(ctx context.Context, query string, args ...interface{}) ([]domain.ProxyConnection, error) {
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conns []domain.ProxyConnection
	for rows.Next() {
		c, err := r.scanConnection(rows)
		if err != nil {
			return nil, err
		}
		conns = append(conns, *c)
	}
	return conns, nil
}
