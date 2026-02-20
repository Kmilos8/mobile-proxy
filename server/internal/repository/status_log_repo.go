package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type StatusLogRepository struct {
	db *DB
}

func NewStatusLogRepository(db *DB) *StatusLogRepository {
	return &StatusLogRepository{db: db}
}

func (r *StatusLogRepository) Insert(ctx context.Context, deviceID uuid.UUID, status, previousStatus string, changedAt time.Time) error {
	query := `INSERT INTO device_status_logs (id, device_id, status, previous_status, changed_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Pool.Exec(ctx, query, uuid.New(), deviceID, status, previousStatus, changedAt)
	return err
}

func (r *StatusLogRepository) HasLogsForDate(ctx context.Context, deviceID uuid.UUID, date time.Time) (bool, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	query := `SELECT EXISTS(SELECT 1 FROM device_status_logs WHERE device_id = $1 AND changed_at >= $2 AND changed_at < $3)`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, deviceID, start, end).Scan(&exists)
	return exists, err
}

func (r *StatusLogRepository) GetByDeviceAndDate(ctx context.Context, deviceID uuid.UUID, date time.Time) ([]domain.DeviceStatusLog, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	query := `SELECT id, device_id, status, previous_status, changed_at
		FROM device_status_logs
		WHERE device_id = $1 AND changed_at >= $2 AND changed_at < $3
		ORDER BY changed_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, deviceID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.DeviceStatusLog
	for rows.Next() {
		var l domain.DeviceStatusLog
		if err := rows.Scan(&l.ID, &l.DeviceID, &l.Status, &l.PreviousStatus, &l.ChangedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
