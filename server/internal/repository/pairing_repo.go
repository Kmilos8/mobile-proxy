package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type PairingCodeRepository struct {
	db *DB
}

func NewPairingCodeRepository(db *DB) *PairingCodeRepository {
	return &PairingCodeRepository{db: db}
}

func (r *PairingCodeRepository) Create(ctx context.Context, pc *domain.PairingCode) error {
	query := `INSERT INTO pairing_codes (id, code, device_auth_token, expires_at, created_by, relay_server_id)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Pool.Exec(ctx, query, pc.ID, pc.Code, pc.DeviceAuthToken, pc.ExpiresAt, pc.CreatedBy, pc.RelayServerID)
	return err
}

func (r *PairingCodeRepository) GetByCode(ctx context.Context, code string) (*domain.PairingCode, error) {
	query := `SELECT id, code, device_auth_token, claimed_by_device_id, claimed_at, expires_at, created_by, relay_server_id, created_at
		FROM pairing_codes
		WHERE code = $1 AND claimed_at IS NULL AND expires_at > NOW()`
	var pc domain.PairingCode
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&pc.ID, &pc.Code, &pc.DeviceAuthToken, &pc.ClaimedByDeviceID,
		&pc.ClaimedAt, &pc.ExpiresAt, &pc.CreatedBy, &pc.RelayServerID, &pc.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get pairing code: %w", err)
	}
	return &pc, nil
}

func (r *PairingCodeRepository) Claim(ctx context.Context, id uuid.UUID, deviceID uuid.UUID) error {
	query := `UPDATE pairing_codes SET claimed_by_device_id = $2, claimed_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, deviceID)
	return err
}

func (r *PairingCodeRepository) List(ctx context.Context) ([]domain.PairingCode, error) {
	query := `SELECT id, code, device_auth_token, claimed_by_device_id, claimed_at, expires_at, created_by, relay_server_id, created_at
		FROM pairing_codes ORDER BY created_at DESC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []domain.PairingCode
	for rows.Next() {
		var pc domain.PairingCode
		err := rows.Scan(&pc.ID, &pc.Code, &pc.DeviceAuthToken, &pc.ClaimedByDeviceID,
			&pc.ClaimedAt, &pc.ExpiresAt, &pc.CreatedBy, &pc.RelayServerID, &pc.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan pairing code: %w", err)
		}
		codes = append(codes, pc)
	}
	return codes, nil
}

func (r *PairingCodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM pairing_codes WHERE id = $1`, id)
	return err
}

func (r *PairingCodeRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM pairing_codes WHERE expires_at < $1 AND claimed_at IS NULL`, time.Now())
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
