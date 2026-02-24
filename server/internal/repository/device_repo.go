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
	query := `INSERT INTO devices (id, name, android_id, status, base_port, http_port, socks5_port, udp_relay_port, ovpn_port, device_model, android_version, app_version, relay_server_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.Pool.Exec(ctx, query,
		d.ID, d.Name, d.AndroidID, d.Status, d.BasePort,
		d.HTTPPort, d.SOCKS5Port, d.UDPRelayPort, d.OVPNPort,
		d.DeviceModel, d.AndroidVersion, d.AppVersion, d.RelayServerID)
	return err
}

const deviceSelectColumns = `d.id, d.name, d.description, d.android_id, d.status,
		COALESCE(host(d.cellular_ip),'') as cellular_ip,
		COALESCE(host(d.wifi_ip),'') as wifi_ip,
		COALESCE(host(d.vpn_ip),'') as vpn_ip,
		d.carrier, d.network_type, d.battery_level, d.battery_charging, d.signal_strength,
		d.base_port, d.http_port, d.socks5_port, d.udp_relay_port, d.ovpn_port,
		d.last_heartbeat, d.app_version, d.device_model, d.android_version,
		d.relay_server_id, COALESCE(rs.ip, '') as relay_server_ip,
		d.auto_rotate_minutes,
		d.created_at, d.updated_at`

const deviceFromJoin = `FROM devices d LEFT JOIN relay_servers rs ON d.relay_server_id = rs.id`

func (r *DeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	query := `SELECT ` + deviceSelectColumns + ` ` + deviceFromJoin + ` WHERE d.id = $1`
	return r.scanDevice(r.db.Pool.QueryRow(ctx, query, id))
}

func (r *DeviceRepository) GetByAndroidID(ctx context.Context, androidID string) (*domain.Device, error) {
	query := `SELECT ` + deviceSelectColumns + ` ` + deviceFromJoin + ` WHERE d.android_id = $1`
	return r.scanDevice(r.db.Pool.QueryRow(ctx, query, androidID))
}

func (r *DeviceRepository) GetByName(ctx context.Context, name string) (*domain.Device, error) {
	query := `SELECT ` + deviceSelectColumns + ` ` + deviceFromJoin + ` WHERE d.name = $1`
	return r.scanDevice(r.db.Pool.QueryRow(ctx, query, name))
}

func (r *DeviceRepository) List(ctx context.Context) ([]domain.Device, error) {
	query := `SELECT ` + deviceSelectColumns + ` ` + deviceFromJoin + ` ORDER BY d.name ASC`
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
	query := `SELECT COALESCE(MAX(max_port), 29996) + 4 FROM (
		SELECT MAX(base_port) AS max_port FROM devices
		UNION ALL
		SELECT MAX(base_port) AS max_port FROM proxy_connections
	) sub`
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

func (r *DeviceRepository) SetAuthToken(ctx context.Context, id uuid.UUID, token string) error {
	query := `UPDATE devices SET auth_token = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, token)
	return err
}

func (r *DeviceRepository) UpdateRelayServer(ctx context.Context, id uuid.UUID, relayServerID uuid.UUID) error {
	query := `UPDATE devices SET relay_server_id = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, relayServerID)
	return err
}

func (r *DeviceRepository) CountByStatus(ctx context.Context) (total int, online int, err error) {
	query := `SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'online') FROM devices`
	err = r.db.Pool.QueryRow(ctx, query).Scan(&total, &online)
	return
}

func (r *DeviceRepository) UpdateNameDescription(ctx context.Context, id uuid.UUID, name, description string) error {
	query := `UPDATE devices SET name = $2, description = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, name, description)
	return err
}

func (r *DeviceRepository) scanDevice(row pgx.Row) (*domain.Device, error) {
	var d domain.Device
	err := row.Scan(
		&d.ID, &d.Name, &d.Description, &d.AndroidID, &d.Status,
		&d.CellularIP, &d.WifiIP, &d.VpnIP,
		&d.Carrier, &d.NetworkType, &d.BatteryLevel, &d.BatteryCharging, &d.SignalStrength,
		&d.BasePort, &d.HTTPPort, &d.SOCKS5Port, &d.UDPRelayPort, &d.OVPNPort,
		&d.LastHeartbeat, &d.AppVersion, &d.DeviceModel, &d.AndroidVersion,
		&d.RelayServerID, &d.RelayServerIP,
		&d.AutoRotateMinutes,
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
		&d.ID, &d.Name, &d.Description, &d.AndroidID, &d.Status,
		&d.CellularIP, &d.WifiIP, &d.VpnIP,
		&d.Carrier, &d.NetworkType, &d.BatteryLevel, &d.BatteryCharging, &d.SignalStrength,
		&d.BasePort, &d.HTTPPort, &d.SOCKS5Port, &d.UDPRelayPort, &d.OVPNPort,
		&d.LastHeartbeat, &d.AppVersion, &d.DeviceModel, &d.AndroidVersion,
		&d.RelayServerID, &d.RelayServerIP,
		&d.AutoRotateMinutes,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan device: %w", err)
	}
	return &d, nil
}

func (r *DeviceRepository) UpdateAutoRotate(ctx context.Context, id uuid.UUID, minutes int) error {
	query := `UPDATE devices SET auto_rotate_minutes = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, minutes)
	return err
}

func (r *DeviceRepository) Upsert(ctx context.Context, d *domain.Device, authToken string) error {
	query := `INSERT INTO devices (id, name, description, android_id, status, base_port, http_port, socks5_port, udp_relay_port, ovpn_port,
			device_model, android_version, app_version, relay_server_id, auth_token, auto_rotate_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, description = EXCLUDED.description, android_id = EXCLUDED.android_id,
			base_port = EXCLUDED.base_port, http_port = EXCLUDED.http_port, socks5_port = EXCLUDED.socks5_port,
			udp_relay_port = EXCLUDED.udp_relay_port, ovpn_port = EXCLUDED.ovpn_port,
			device_model = EXCLUDED.device_model, android_version = EXCLUDED.android_version,
			app_version = EXCLUDED.app_version, relay_server_id = EXCLUDED.relay_server_id,
			auth_token = EXCLUDED.auth_token, auto_rotate_minutes = EXCLUDED.auto_rotate_minutes,
			updated_at = NOW()`
	_, err := r.db.Pool.Exec(ctx, query,
		d.ID, d.Name, d.Description, d.AndroidID, d.Status, d.BasePort,
		d.HTTPPort, d.SOCKS5Port, d.UDPRelayPort, d.OVPNPort,
		d.DeviceModel, d.AndroidVersion, d.AppVersion, d.RelayServerID, authToken, d.AutoRotateMinutes)
	return err
}

func (r *DeviceRepository) GetAuthToken(ctx context.Context, id uuid.UUID) (string, error) {
	query := `SELECT COALESCE(auth_token, '') FROM devices WHERE id = $1`
	var token string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("get auth token: %w", err)
	}
	return token, nil
}
