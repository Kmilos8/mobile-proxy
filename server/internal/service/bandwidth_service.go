package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

type BandwidthService struct {
	bwRepo *repository.BandwidthRepository
}

func NewBandwidthService(bwRepo *repository.BandwidthRepository) *BandwidthService {
	return &BandwidthService{bwRepo: bwRepo}
}

func (s *BandwidthService) Record(ctx context.Context, deviceID uuid.UUID, connectionID *uuid.UUID, bytesIn, bytesOut int64) error {
	now := time.Now().UTC()
	log := &domain.BandwidthLog{
		ID:            uuid.New(),
		DeviceID:      deviceID,
		ConnectionID:  connectionID,
		BytesIn:       bytesIn,
		BytesOut:      bytesOut,
		IntervalStart: now.Truncate(time.Minute),
		IntervalEnd:   now.Truncate(time.Minute).Add(time.Minute),
	}
	return s.bwRepo.Create(ctx, log)
}

func (s *BandwidthService) GetDeviceTodayTotal(ctx context.Context, deviceID uuid.UUID) (int64, int64, error) {
	return s.bwRepo.GetDeviceTotalToday(ctx, deviceID)
}

func (s *BandwidthService) GetDeviceMonthTotal(ctx context.Context, deviceID uuid.UUID) (int64, int64, error) {
	now := time.Now()
	return s.bwRepo.GetDeviceTotalMonth(ctx, deviceID, now.Year(), now.Month())
}

func (s *BandwidthService) EnsurePartitions(ctx context.Context) error {
	now := time.Now()
	// Ensure current month and next month partitions exist
	if err := s.bwRepo.EnsurePartition(ctx, now.Year(), now.Month()); err != nil {
		return err
	}
	next := now.AddDate(0, 1, 0)
	return s.bwRepo.EnsurePartition(ctx, next.Year(), next.Month())
}
