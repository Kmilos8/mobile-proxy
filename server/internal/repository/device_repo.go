package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mobileproxy/server/internal/domain"
)

type DeviceRepository struct {
	db *DB
}

func NewDeviceRepository(db *DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(ctx context.Context, d *domain.Device) error {
	query := `INSERT INTO devices (id, name, android_id, status, base_port, http_port, socks5_port, udp_relay_port, ovpn_port, device_model, android_version, app_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Pool.Exec(ctx, query,
		d.ID, d.Name, d.AndroidID, d.Status, d.BasePort,
		d.HTTPPort, d.SOCKS5Port, d.UDPRelayPort, d.OVPNPort,
		d.DeviceModel, d.AndroidVersion, d.AppVersion)
	return err
}

func (r *DeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	query := `SELECT id, name, android_id, status,
		COALESCE(host(cellular_ip),'') as cellular_ip,
		COALESCE(host(wifi_ip),'') as wifi_ip,
		COALESCE(host(vpn_ip),'') as vpn_ip,
		carrier, network_type, battery_level, battery_charging, signal_strength,
		base_port, http_port, socks5_port, udp_relay_port, ovpn_port,
		last_heartbeat, app_version, device_model, android_version,
		created_at, updated_at
		FROM devices WHERE id = $1`
	return r.scanDevice(r.db.Pool.QueryRow(ctx, query, id))
}

func (r *DeviceRepository) GetByAndroidID(ctx context.Context, androidID string) (*domain.Device, error) {
	query := `SELECT id, name, android_id, status,
		COALESCE(host(cellular_ip),'') as cellular_ip,
		COALESCE(host(wifi_ip),'') as wifi_ip,
		COALESCE(host(vpn_ip),'') as vpn_ip,
		carrier, network_type, battery_level, battery_charging, signal_strength,
		base_port, http_port, socks5_port, udp_relay_port, ovpn_port,
		last_heartbeat, app_version, device_model, android_version,
		created_at, updated_at
		FROM devices WHERE android_id = $1`
	return r.scanDevice(r.db.Pool.QueryRow(ctx, query, androidID))
}

func (r *DeviceRepository) GetByName(ctx context.Context, name string) (*domain.Device, error) {
	query := `SELECT id, name, android_id, status,
		COALESCE(host(cellular_ip),'') as cellular_ip,
		COALESCE(host(wifi_ip),'') as wifi_ip,
		COALESCE(host(vpn_ip),'') as vpn_ip,
		carrier, network_type, battery_level, battery_charging, signal_strength,
		base_port, http_port, socks5_port, udp_relay_port, ovpn_port,
		last_heartbeat, app_version, device_model, android_version,
		created_at, updated_at
		FROM devices WHERE name = $1`
	return r.scanDevice(r.db.Pool.QueryRow(ctx, query, name))
}

func (r *DeviceRepository) List(ctx context.Context) ([]domain.Device, error) {
	query := `SELECT id, name, android_id, status,
		COALESCE(host(cellular_ip),'') as cellular_ip,
		COALESCE(host(wifi_ip),'') as wifi_ip,
		COALESCE(host(vpn_ip),'') as vpn_ip,
		carrier, network_type, battery_level, battery_charging, signal_strength,
		base_port, http_port, socks5_port, udp_relay_port, ovpn_port,
		last_heartbeat, app_version, device_model, android_version,
		created_at, updated_at
		FROM devices ORDER BY name ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []domain.Device
	for rows.Next() {
		d, err := r.scanDeviceRow(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, *d)
	}
	return devices, nil
}

func (r *DeviceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.DeviceStatus) error {
	query := `UPDATE devices SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, status)
	return err
}

func (r *DeviceRepository) UpdateHeartbeat(ctx context.Context, id uuid.UUID, req *domain.HeartbeatRequest) error {
	query := `UPDATE devices SET
		cellular_ip = NULLIF($2, '')::inet,
		wifi_ip = NULLIF($3, '')::inet,
		carrier = $4, network_type = $5,
		battery_level = $6, battery_charging = $7, signal_strength = $8,
		app_version = $9, status = 'online',
		last_heartbeat = NOW(), updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id,
		req.CellularIP, req.WifiIP, req.Carrier, req.NetworkType,
		req.BatteryLevel, req.BatteryCharging, req.SignalStrength, req.AppVersion)
	return err
}

func (r *DeviceRepository) SetVpnIP(ctx context.Context, id uuid.UUID, vpnIP string) error {
	query := `UPDATE devices SET vpn_ip = $2::inet, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, vpnIP)
	return err
}

func (r *DeviceRepository) GetNextBasePort(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(MAX(base_port), 29996) + 4 FROM devices`
	var port int
	err := r.db.Pool.QueryRow(ctx, query).Scan(&port)
	if err != nil {
		return 0, err
	}
	if port < 30000 {
		port = 30000
	}
	return port, nil
}

func (r *DeviceRepository) MarkStaleOffline(ctx context.Context, threshold time.Duration) (int64, error) {
	query := `UPDATE devices SET status = 'offline', updated_at = NOW()
		WHERE status = 'online' AND last_heartbeat < $1`
	tag, err := r.db.Pool.Exec(ctx, query, time.Now().Add(-threshold))
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *DeviceRepository) scanDevice(row pgx.Row) (*domain.Device, error) {
	var d domain.Device
	err := row.Scan(
		&d.ID, &d.Name, &d.AndroidID, &d.Status,
		&d.CellularIP, &d.WifiIP, &d.VpnIP,
		&d.Carrier, &d.NetworkType, &d.BatteryLevel, &d.BatteryCharging, &d.SignalStrength,
		&d.BasePort, &d.HTTPPort, &d.SOCKS5Port, &d.UDPRelayPort, &d.OVPNPort,
		&d.LastHeartbeat, &d.AppVersion, &d.DeviceModel, &d.AndroidVersion,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan device: %w", err)
	}
	return &d, nil
}

func (r *DeviceRepository) scanDeviceRow(rows pgx.Rows) (*domain.Device, error) {
	var d domain.Device
	err := rows.Scan(
		&d.ID, &d.Name, &d.AndroidID, &d.Status,
		&d.CellularIP, &d.WifiIP, &d.VpnIP,
		&d.Carrier, &d.NetworkType, &d.BatteryLevel, &d.BatteryCharging, &d.SignalStrength,
		&d.BasePort, &d.HTTPPort, &d.SOCKS5Port, &d.UDPRelayPort, &d.OVPNPort,
		&d.LastHeartbeat, &d.AppVersion, &d.DeviceModel, &d.AndroidVersion,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan device: %w", err)
	}
	return &d, nil
}
