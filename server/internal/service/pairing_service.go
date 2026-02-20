package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

const pairingAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

type PairingService struct {
	pairingRepo   *repository.PairingCodeRepository
	deviceService *DeviceService
	deviceRepo    *repository.DeviceRepository
	serverURL     string // e.g. "http://178.156.240.184:8080"
}

func NewPairingService(
	pairingRepo *repository.PairingCodeRepository,
	deviceService *DeviceService,
	deviceRepo *repository.DeviceRepository,
	serverURL string,
) *PairingService {
	return &PairingService{
		pairingRepo:   pairingRepo,
		deviceService: deviceService,
		deviceRepo:    deviceRepo,
		serverURL:     serverURL,
	}
}

func generatePairingCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	code := make([]byte, 8)
	for i := range code {
		code[i] = pairingAlphabet[int(b[i])%len(pairingAlphabet)]
	}
	return string(code)
}

func generateAuthToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *PairingService) CreateCode(ctx context.Context, expiresInHours int, createdBy *uuid.UUID) (*domain.CreatePairingCodeResponse, error) {
	if expiresInHours <= 0 {
		expiresInHours = 24
	}

	pc := &domain.PairingCode{
		ID:              uuid.New(),
		Code:            generatePairingCode(),
		DeviceAuthToken: generateAuthToken(),
		ExpiresAt:       time.Now().Add(time.Duration(expiresInHours) * time.Hour),
		CreatedBy:       createdBy,
	}

	if err := s.pairingRepo.Create(ctx, pc); err != nil {
		return nil, fmt.Errorf("create pairing code: %w", err)
	}

	return &domain.CreatePairingCodeResponse{
		ID:        pc.ID,
		Code:      pc.Code,
		ExpiresAt: pc.ExpiresAt,
	}, nil
}

func (s *PairingService) ClaimCode(ctx context.Context, req *domain.ClaimPairingCodeRequest) (*domain.ClaimPairingCodeResponse, error) {
	// Normalize code: strip dashes, uppercase
	code := strings.ToUpper(strings.ReplaceAll(req.Code, "-", ""))

	pc, err := s.pairingRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired pairing code")
	}

	// Register the device via existing DeviceService
	regResp, err := s.deviceService.Register(ctx, &domain.DeviceRegistrationRequest{
		AndroidID:      req.AndroidID,
		DeviceModel:    req.DeviceModel,
		AndroidVersion: req.AndroidVersion,
		AppVersion:     req.AppVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("register device: %w", err)
	}

	// Set the auth token on the device
	if err := s.deviceRepo.SetAuthToken(ctx, regResp.DeviceID, pc.DeviceAuthToken); err != nil {
		return nil, fmt.Errorf("set auth token: %w", err)
	}

	// Mark the pairing code as claimed
	if err := s.pairingRepo.Claim(ctx, pc.ID, regResp.DeviceID); err != nil {
		return nil, fmt.Errorf("claim pairing code: %w", err)
	}

	return &domain.ClaimPairingCodeResponse{
		DeviceID:  regResp.DeviceID,
		AuthToken: pc.DeviceAuthToken,
		ServerURL: s.serverURL,
		VpnConfig: regResp.VpnConfig,
		BasePort:  regResp.BasePort,
	}, nil
}

func (s *PairingService) List(ctx context.Context) ([]domain.PairingCode, error) {
	return s.pairingRepo.List(ctx)
}

func (s *PairingService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.pairingRepo.Delete(ctx, id)
}
