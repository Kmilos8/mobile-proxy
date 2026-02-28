package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type DeviceShareRepository struct {
	db *DB
}

func NewDeviceShareRepository(db *DB) *DeviceShareRepository {
	return &DeviceShareRepository{db: db}
}

const shareSelectCols = `id, device_id, owner_id, shared_with, can_rename, can_manage_ports, can_download_configs, can_rotate_ip, created_at, updated_at`

func (r *DeviceShareRepository) Create(ctx context.Context, share *domain.DeviceShare) error {
	query := `INSERT INTO device_shares (id, device_id, owner_id, shared_with, can_rename, can_manage_ports, can_download_configs, can_rotate_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Pool.Exec(ctx, query,
		share.ID, share.DeviceID, share.OwnerID, share.SharedWith,
		share.CanRename, share.CanManagePorts, share.CanDownloadConfigs, share.CanRotateIP)
	return err
}

func (r *DeviceShareRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DeviceShare, error) {
	query := `SELECT ` + shareSelectCols + ` FROM device_shares WHERE id = $1`
	return r.scanShare(r.db.Pool.QueryRow(ctx, query, id))
}

// GetByDeviceAndCustomer returns the share record where the customer has been given access to the device.
func (r *DeviceShareRepository) GetByDeviceAndCustomer(ctx context.Context, deviceID uuid.UUID, customerID uuid.UUID) (*domain.DeviceShare, error) {
	query := `SELECT ` + shareSelectCols + ` FROM device_shares WHERE device_id = $1 AND shared_with = $2`
	return r.scanShare(r.db.Pool.QueryRow(ctx, query, deviceID, customerID))
}

// ListByDevice returns all shares for a device (visible to the device owner).
func (r *DeviceShareRepository) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.DeviceShare, error) {
	query := `SELECT ` + shareSelectCols + ` FROM device_shares WHERE device_id = $1 ORDER BY created_at ASC`
	return r.scanShares(ctx, query, deviceID)
}

// ListByCustomer returns all shares where the given customer has been granted access.
func (r *DeviceShareRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.DeviceShare, error) {
	query := `SELECT ` + shareSelectCols + ` FROM device_shares WHERE shared_with = $1 ORDER BY created_at ASC`
	return r.scanShares(ctx, query, customerID)
}

// Update saves changes to permission booleans for an existing share record.
func (r *DeviceShareRepository) Update(ctx context.Context, share *domain.DeviceShare) error {
	query := `UPDATE device_shares SET
		can_rename = $2, can_manage_ports = $3, can_download_configs = $4, can_rotate_ip = $5,
		updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query,
		share.ID, share.CanRename, share.CanManagePorts, share.CanDownloadConfigs, share.CanRotateIP)
	return err
}

// Delete removes a share by its ID.
func (r *DeviceShareRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM device_shares WHERE id = $1`, id)
	return err
}

func (r *DeviceShareRepository) scanShare(row interface{ Scan(dest ...interface{}) error }) (*domain.DeviceShare, error) {
	var s domain.DeviceShare
	err := row.Scan(
		&s.ID, &s.DeviceID, &s.OwnerID, &s.SharedWith,
		&s.CanRename, &s.CanManagePorts, &s.CanDownloadConfigs, &s.CanRotateIP,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan device share: %w", err)
	}
	return &s, nil
}

func (r *DeviceShareRepository) scanShares(ctx context.Context, query string, args ...interface{}) ([]domain.DeviceShare, error) {
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []domain.DeviceShare
	for rows.Next() {
		s, err := r.scanShare(rows)
		if err != nil {
			return nil, err
		}
		shares = append(shares, *s)
	}
	return shares, nil
}
