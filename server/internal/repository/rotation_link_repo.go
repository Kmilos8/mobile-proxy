package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type RotationLinkRepository struct {
	db *DB
}

func NewRotationLinkRepository(db *DB) *RotationLinkRepository {
	return &RotationLinkRepository{db: db}
}

func (r *RotationLinkRepository) Create(ctx context.Context, link *domain.RotationLink) error {
	query := `INSERT INTO rotation_links (id, device_id, token, name) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Pool.Exec(ctx, query, link.ID, link.DeviceID, link.Token, link.Name)
	return err
}

func (r *RotationLinkRepository) GetByToken(ctx context.Context, token string) (*domain.RotationLink, error) {
	query := `SELECT id, device_id, token, name, created_at, last_used_at FROM rotation_links WHERE token = $1`
	var link domain.RotationLink
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(
		&link.ID, &link.DeviceID, &link.Token, &link.Name, &link.CreatedAt, &link.LastUsedAt)
	if err != nil {
		return nil, fmt.Errorf("get rotation link by token: %w", err)
	}
	return &link, nil
}

func (r *RotationLinkRepository) ListByDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.RotationLink, error) {
	query := `SELECT id, device_id, token, name, created_at, last_used_at FROM rotation_links WHERE device_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Pool.Query(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.RotationLink
	for rows.Next() {
		var link domain.RotationLink
		err := rows.Scan(&link.ID, &link.DeviceID, &link.Token, &link.Name, &link.CreatedAt, &link.LastUsedAt)
		if err != nil {
			return nil, fmt.Errorf("scan rotation link: %w", err)
		}
		links = append(links, link)
	}
	return links, nil
}

func (r *RotationLinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM rotation_links WHERE id = $1`, id)
	return err
}

func (r *RotationLinkRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE rotation_links SET last_used_at = NOW() WHERE id = $1`, id)
	return err
}
