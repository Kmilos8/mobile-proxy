package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

type DeviceShareService struct {
	shareRepo  *repository.DeviceShareRepository
	deviceRepo *repository.DeviceRepository
}

func NewDeviceShareService(shareRepo *repository.DeviceShareRepository, deviceRepo *repository.DeviceRepository) *DeviceShareService {
	return &DeviceShareService{
		shareRepo:  shareRepo,
		deviceRepo: deviceRepo,
	}
}

// CanAccess returns true if the customer owns the device OR has any share on it.
func (s *DeviceShareService) CanAccess(ctx context.Context, deviceID uuid.UUID, customerID uuid.UUID) (bool, error) {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return false, err
	}
	if device.CustomerID != nil && *device.CustomerID == customerID {
		return true, nil
	}
	share, err := s.shareRepo.GetByDeviceAndCustomer(ctx, deviceID, customerID)
	if err == nil && share != nil {
		return true, nil
	}
	return false, nil
}

// CanDo returns true if the customer can perform the given action on the device.
// Owners can do everything. Shared users are limited by per-permission booleans.
// perm values: "rename", "manage_ports", "download_configs", "rotate_ip".
func (s *DeviceShareService) CanDo(ctx context.Context, deviceID uuid.UUID, customerID uuid.UUID, perm string) (bool, error) {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return false, err
	}
	// Owner can do everything.
	if device.CustomerID != nil && *device.CustomerID == customerID {
		return true, nil
	}
	// Check share record for permission.
	share, err := s.shareRepo.GetByDeviceAndCustomer(ctx, deviceID, customerID)
	if err != nil || share == nil {
		return false, nil
	}
	switch perm {
	case "rename":
		return share.CanRename, nil
	case "manage_ports":
		return share.CanManagePorts, nil
	case "download_configs":
		return share.CanDownloadConfigs, nil
	case "rotate_ip":
		return share.CanRotateIP, nil
	}
	return false, nil
}

// CreateShare creates a share after validating the caller is the device owner.
func (s *DeviceShareService) CreateShare(ctx context.Context, share *domain.DeviceShare) error {
	device, err := s.deviceRepo.GetByID(ctx, share.DeviceID)
	if err != nil {
		return err
	}
	if device.CustomerID == nil || *device.CustomerID != share.OwnerID {
		return errors.New("not device owner")
	}
	if share.SharedWith == share.OwnerID {
		return errors.New("cannot share device with yourself")
	}
	share.ID = uuid.New()
	return s.shareRepo.Create(ctx, share)
}

// UpdateShare updates share permissions after validating the caller is the device owner.
func (s *DeviceShareService) UpdateShare(ctx context.Context, share *domain.DeviceShare) error {
	existing, err := s.shareRepo.GetByID(ctx, share.ID)
	if err != nil {
		return err
	}
	device, err := s.deviceRepo.GetByID(ctx, existing.DeviceID)
	if err != nil {
		return err
	}
	if device.CustomerID == nil || *device.CustomerID != share.OwnerID {
		return errors.New("not device owner")
	}
	return s.shareRepo.Update(ctx, share)
}

// DeleteShare removes a share after validating the caller is the owner.
func (s *DeviceShareService) DeleteShare(ctx context.Context, shareID uuid.UUID, callerID uuid.UUID) error {
	share, err := s.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return err
	}
	if share.OwnerID != callerID {
		return errors.New("not share owner")
	}
	return s.shareRepo.Delete(ctx, shareID)
}

// ListSharesForDevice returns all shares for a device.
func (s *DeviceShareService) ListSharesForDevice(ctx context.Context, deviceID uuid.UUID) ([]domain.DeviceShare, error) {
	return s.shareRepo.ListByDevice(ctx, deviceID)
}
