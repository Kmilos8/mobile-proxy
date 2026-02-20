package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type BandwidthRepository struct {
	db *DB
}

func NewBandwidthRepository(db *DB) *BandwidthRepository {
	return &BandwidthRepository{db: db}
}

func (r *BandwidthRepository) Create(ctx context.Context, log *domain.BandwidthLog) error {
	query := `INSERT INTO bandwidth_logs (id, device_id, connection_id, bytes_in, bytes_out, interval_start, interval_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Pool.Exec(ctx, query,
		log.ID, log.DeviceID, log.ConnectionID, log.BytesIn, log.BytesOut,
		log.IntervalStart, log.IntervalEnd)
	return err
}

func (r *BandwidthRepository) GetDeviceTotalToday(ctx context.Context, deviceID uuid.UUID) (int64, int64, error) {
	query := `SELECT COALESCE(SUM(bytes_in), 0), COALESCE(SUM(bytes_out), 0)
		FROM bandwidth_logs WHERE device_id = $1 AND interval_start >= CURRENT_DATE`
	var bytesIn, bytesOut int64
	err := r.db.Pool.QueryRow(ctx, query, deviceID).Scan(&bytesIn, &bytesOut)
	return bytesIn, bytesOut, err
}

func (r *BandwidthRepository) GetDeviceTotalMonth(ctx context.Context, deviceID uuid.UUID, year int, month time.Month) (int64, int64, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	query := `SELECT COALESCE(SUM(bytes_in), 0), COALESCE(SUM(bytes_out), 0)
		FROM bandwidth_logs WHERE device_id = $1 AND interval_start >= $2 AND interval_start < $3`
	var bytesIn, bytesOut int64
	err := r.db.Pool.QueryRow(ctx, query, deviceID, start, end).Scan(&bytesIn, &bytesOut)
	return bytesIn, bytesOut, err
}

func (r *BandwidthRepository) GetTotalToday(ctx context.Context) (int64, int64, error) {
	query := `SELECT COALESCE(SUM(bytes_in), 0), COALESCE(SUM(bytes_out), 0)
		FROM bandwidth_logs WHERE interval_start >= CURRENT_DATE`
	var bytesIn, bytesOut int64
	err := r.db.Pool.QueryRow(ctx, query).Scan(&bytesIn, &bytesOut)
	return bytesIn, bytesOut, err
}

func (r *BandwidthRepository) GetTotalMonth(ctx context.Context, year int, month time.Month) (int64, int64, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	query := `SELECT COALESCE(SUM(bytes_in), 0), COALESCE(SUM(bytes_out), 0)
		FROM bandwidth_logs WHERE interval_start >= $1 AND interval_start < $2`
	var bytesIn, bytesOut int64
	err := r.db.Pool.QueryRow(ctx, query, start, end).Scan(&bytesIn, &bytesOut)
	return bytesIn, bytesOut, err
}

func (r *BandwidthRepository) GetDeviceHourly(ctx context.Context, deviceID uuid.UUID, date time.Time) ([]domain.BandwidthHourly, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	query := `SELECT EXTRACT(HOUR FROM interval_start)::int AS hour,
		COALESCE(SUM(bytes_in), 0) AS download_bytes,
		COALESCE(SUM(bytes_out), 0) AS upload_bytes
		FROM bandwidth_logs
		WHERE device_id = $1 AND interval_start >= $2 AND interval_start < $3
		GROUP BY hour ORDER BY hour`

	rows, err := r.db.Pool.Query(ctx, query, deviceID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hourMap := make(map[int]domain.BandwidthHourly)
	for rows.Next() {
		var h domain.BandwidthHourly
		if err := rows.Scan(&h.Hour, &h.DownloadBytes, &h.UploadBytes); err != nil {
			return nil, err
		}
		hourMap[h.Hour] = h
	}

	// Fill all 24 hours
	result := make([]domain.BandwidthHourly, 24)
	for i := 0; i < 24; i++ {
		if h, ok := hourMap[i]; ok {
			result[i] = h
		} else {
			result[i] = domain.BandwidthHourly{Hour: i}
		}
	}
	return result, nil
}

func (r *BandwidthRepository) EnsurePartition(ctx context.Context, year int, month time.Month) error {
	tableName := fmt.Sprintf("bandwidth_logs_%d_%02d", year, month)
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s PARTITION OF bandwidth_logs FOR VALUES FROM ('%s') TO ('%s')`,
		tableName, start.Format("2006-01-02"), end.Format("2006-01-02"))
	_, err := r.db.Pool.Exec(ctx, query)
	return err
}
