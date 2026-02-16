package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type IPHistoryRepository struct {
	db *DB
}

func NewIPHistoryRepository(db *DB) *IPHistoryRepository {
	return &IPHistoryRepository{db: db}
}

func (r *IPHistoryRepository) Create(ctx context.Context, h *domain.IPHistory) error {
	query := `INSERT INTO ip_history (id, device_id, ip, method) VALUES ($1, $2, $3::inet, $4)`
	_, err := r.db.Pool.Exec(ctx, query, h.ID, h.DeviceID, h.IP, h.Method)
	return err
}

func (r *IPHistoryRepository) ListByDevice(ctx context.Context, deviceID uuid.UUID, limit int) ([]domain.IPHistory, error) {
	query := `SELECT id, device_id, host(ip) as ip, method, created_at
		FROM ip_history WHERE device_id = $1
		ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.Pool.Query(ctx, query, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []domain.IPHistory
	for rows.Next() {
		var h domain.IPHistory
		err := rows.Scan(&h.ID, &h.DeviceID, &h.IP, &h.Method, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan ip history: %w", err)
		}
		history = append(history, h)
	}
	return history, nil
}
