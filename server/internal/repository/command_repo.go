package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type CommandRepository struct {
	db *DB
}

func NewCommandRepository(db *DB) *CommandRepository {
	return &CommandRepository{db: db}
}

func (r *CommandRepository) Create(ctx context.Context, cmd *domain.DeviceCommand) error {
	query := `INSERT INTO device_commands (id, device_id, type, status, payload) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Pool.Exec(ctx, query, cmd.ID, cmd.DeviceID, cmd.Type, cmd.Status, cmd.Payload)
	return err
}

func (r *CommandRepository) GetPending(ctx context.Context, deviceID uuid.UUID) ([]domain.DeviceCommand, error) {
	query := `SELECT id, device_id, type, status, payload, result, created_at, executed_at
		FROM device_commands WHERE device_id = $1 AND status = 'pending'
		ORDER BY created_at ASC`
	rows, err := r.db.Pool.Query(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []domain.DeviceCommand
	for rows.Next() {
		var cmd domain.DeviceCommand
		err := rows.Scan(&cmd.ID, &cmd.DeviceID, &cmd.Type, &cmd.Status,
			&cmd.Payload, &cmd.Result, &cmd.CreatedAt, &cmd.ExecutedAt)
		if err != nil {
			return nil, fmt.Errorf("scan command: %w", err)
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

func (r *CommandRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CommandStatus, result string) error {
	var query string
	if status == domain.CommandStatusCompleted || status == domain.CommandStatusFailed {
		query = `UPDATE device_commands SET status = $2, result = $3, executed_at = $4 WHERE id = $1`
		_, err := r.db.Pool.Exec(ctx, query, id, status, result, time.Now())
		return err
	}
	query = `UPDATE device_commands SET status = $2, result = $3 WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, status, result)
	return err
}

func (r *CommandRepository) GetByDevice(ctx context.Context, deviceID uuid.UUID, limit int) ([]domain.DeviceCommand, error) {
	query := `SELECT id, device_id, type, status, payload, result, created_at, executed_at
		FROM device_commands WHERE device_id = $1
		ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.Pool.Query(ctx, query, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []domain.DeviceCommand
	for rows.Next() {
		var cmd domain.DeviceCommand
		err := rows.Scan(&cmd.ID, &cmd.DeviceID, &cmd.Type, &cmd.Status,
			&cmd.Payload, &cmd.Result, &cmd.CreatedAt, &cmd.ExecutedAt)
		if err != nil {
			return nil, fmt.Errorf("scan command: %w", err)
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

func (r *CommandRepository) MarkAsSent(ctx context.Context, ids []uuid.UUID) error {
	query := `UPDATE device_commands SET status = 'sent' WHERE id = ANY($1)`
	_, err := r.db.Pool.Exec(ctx, query, ids)
	return err
}
