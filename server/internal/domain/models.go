package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeviceStatus string

const (
	DeviceStatusOnline   DeviceStatus = "online"
	DeviceStatusOffline  DeviceStatus = "offline"
	DeviceStatusRotating DeviceStatus = "rotating"
	DeviceStatusError    DeviceStatus = "error"
)

type CommandType string

const (
	CommandRotateIP     CommandType = "rotate_ip"
	CommandReboot       CommandType = "reboot"
	CommandFindPhone    CommandType = "find_phone"
	CommandWifiOn       CommandType = "wifi_on"
	CommandWifiOff      CommandType = "wifi_off"
	CommandUpdateConfig CommandType = "update_config"
)

type CommandStatus string

const (
	CommandStatusPending   CommandStatus = "pending"
	CommandStatusSent      CommandStatus = "sent"
	CommandStatusCompleted CommandStatus = "completed"
	CommandStatusFailed    CommandStatus = "failed"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	Role         string    `json:"role" db:"role"` // admin, operator
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type RelayServer struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	IP        string    `json:"ip" db:"ip"`
	Location  string    `json:"location" db:"location"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Device struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	Name            string       `json:"name" db:"name"`
	Description     string       `json:"description" db:"description"`
	AndroidID       string       `json:"android_id" db:"android_id"`
	Status          DeviceStatus `json:"status" db:"status"`
	CellularIP      string       `json:"cellular_ip" db:"cellular_ip"`
	WifiIP          string       `json:"wifi_ip" db:"wifi_ip"`
	VpnIP           string       `json:"vpn_ip" db:"vpn_ip"`
	Carrier         string       `json:"carrier" db:"carrier"`
	NetworkType     string       `json:"network_type" db:"network_type"` // 4G, 5G
	BatteryLevel    int          `json:"battery_level" db:"battery_level"`
	BatteryCharging bool         `json:"battery_charging" db:"battery_charging"`
	SignalStrength  int          `json:"signal_strength" db:"signal_strength"`
	BasePort        int          `json:"base_port" db:"base_port"`
	HTTPPort        int          `json:"http_port" db:"http_port"`
	SOCKS5Port      int          `json:"socks5_port" db:"socks5_port"`
	UDPRelayPort    int          `json:"udp_relay_port" db:"udp_relay_port"`
	OVPNPort        int          `json:"ovpn_port" db:"ovpn_port"`
	LastHeartbeat   *time.Time   `json:"last_heartbeat" db:"last_heartbeat"`
	AppVersion      string       `json:"app_version" db:"app_version"`
	DeviceModel     string       `json:"device_model" db:"device_model"`
	AndroidVersion  string       `json:"android_version" db:"android_version"`
	RelayServerID   *uuid.UUID   `json:"relay_server_id" db:"relay_server_id"`
	RelayServerIP     string       `json:"relay_server_ip" db:"-"`
	AutoRotateMinutes int          `json:"auto_rotate_minutes" db:"auto_rotate_minutes"`
	CreatedAt         time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at" db:"updated_at"`
}

type Customer struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ProxyConnection struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	DeviceID       uuid.UUID  `json:"device_id" db:"device_id"`
	CustomerID     *uuid.UUID `json:"customer_id" db:"customer_id"`
	Username       string     `json:"username" db:"username"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	PasswordPlain  string     `json:"-" db:"password_plain"`
	Password       string     `json:"password,omitempty" db:"-"` // plaintext only on creation response
	IPWhitelist    []string   `json:"ip_whitelist" db:"ip_whitelist"`
	BandwidthLimit int64      `json:"bandwidth_limit" db:"bandwidth_limit"` // bytes, 0 = unlimited
	BandwidthUsed  int64      `json:"bandwidth_used" db:"bandwidth_used"`
	Active         bool       `json:"active" db:"active"`
	ProxyType      string     `json:"proxy_type" db:"proxy_type"` // "http" or "socks5"
	BasePort       *int       `json:"base_port" db:"base_port"`
	HTTPPort       *int       `json:"http_port" db:"http_port"`
	SOCKS5Port     *int       `json:"socks5_port" db:"socks5_port"`
	ExpiresAt      *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type IPHistory struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DeviceID  uuid.UUID `json:"device_id" db:"device_id"`
	IP        string    `json:"ip" db:"ip"`
	Method    string    `json:"method" db:"method"` // airplane_mode, cellular_reconnect, natural
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type BandwidthLog struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	DeviceID      uuid.UUID  `json:"device_id" db:"device_id"`
	ConnectionID  *uuid.UUID `json:"connection_id" db:"connection_id"`
	BytesIn       int64      `json:"bytes_in" db:"bytes_in"`
	BytesOut      int64      `json:"bytes_out" db:"bytes_out"`
	IntervalStart time.Time  `json:"interval_start" db:"interval_start"`
	IntervalEnd   time.Time  `json:"interval_end" db:"interval_end"`
}

type DeviceCommand struct {
	ID         uuid.UUID     `json:"id" db:"id"`
	DeviceID   uuid.UUID     `json:"device_id" db:"device_id"`
	Type       CommandType   `json:"type" db:"type"`
	Status     CommandStatus `json:"status" db:"status"`
	Payload    string        `json:"payload" db:"payload"` // JSON
	Result     string        `json:"result" db:"result"`   // JSON
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	ExecutedAt *time.Time    `json:"executed_at" db:"executed_at"`
}

type RotationLink struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	DeviceID   uuid.UUID  `json:"device_id" db:"device_id"`
	Token      string     `json:"token" db:"token"`
	Name       string     `json:"name" db:"name"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at" db:"last_used_at"`
}

type PairingCode struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Code              string     `json:"code" db:"code"`
	DeviceAuthToken   string     `json:"-" db:"device_auth_token"`
	ClaimedByDeviceID *uuid.UUID `json:"claimed_by_device_id" db:"claimed_by_device_id"`
	ClaimedAt         *time.Time `json:"claimed_at" db:"claimed_at"`
	ExpiresAt         time.Time  `json:"expires_at" db:"expires_at"`
	CreatedBy         *uuid.UUID `json:"created_by" db:"created_by"`
	RelayServerID     *uuid.UUID `json:"relay_server_id" db:"relay_server_id"`
	ReassignDeviceID  *uuid.UUID `json:"reassign_device_id" db:"reassign_device_id"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type DeviceStatusLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	DeviceID       uuid.UUID `json:"device_id" db:"device_id"`
	Status         string    `json:"status" db:"status"`
	PreviousStatus string    `json:"previous_status" db:"previous_status"`
	ChangedAt      time.Time `json:"changed_at" db:"changed_at"`
}

type BandwidthHourly struct {
	Hour          int   `json:"hour"`
	DownloadBytes int64 `json:"download_bytes"`
	UploadBytes   int64 `json:"upload_bytes"`
}

type UptimeSegment struct {
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// API request/response types

type DeviceRegistrationRequest struct {
	AndroidID      string `json:"android_id" binding:"required"`
	DeviceModel    string `json:"device_model" binding:"required"`
	AndroidVersion string `json:"android_version" binding:"required"`
	AppVersion     string `json:"app_version" binding:"required"`
	Name           string `json:"name"`
}

type DeviceRegistrationResponse struct {
	DeviceID  uuid.UUID `json:"device_id"`
	VpnConfig string    `json:"vpn_config"`
	BasePort  int       `json:"base_port"`
}

type HeartbeatRequest struct {
	CellularIP      string `json:"cellular_ip"`
	WifiIP          string `json:"wifi_ip"`
	Carrier         string `json:"carrier"`
	NetworkType     string `json:"network_type"`
	BatteryLevel    int    `json:"battery_level"`
	BatteryCharging bool   `json:"battery_charging"`
	SignalStrength  int    `json:"signal_strength"`
	AppVersion      string `json:"app_version"`
	BytesIn         int64  `json:"bytes_in"`
	BytesOut        int64  `json:"bytes_out"`
}

type ProxyCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HeartbeatResponse struct {
	Commands    []DeviceCommand  `json:"commands"`
	Credentials []ProxyCredential `json:"credentials,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateConnectionRequest struct {
	DeviceID       uuid.UUID  `json:"device_id" binding:"required"`
	CustomerID     *uuid.UUID `json:"customer_id"`
	Username       string     `json:"username" binding:"required"`
	Password       string     `json:"password" binding:"required"`
	ProxyType      string     `json:"proxy_type"` // "http" or "socks5", defaults to "http"
	IPWhitelist    []string   `json:"ip_whitelist"`
	BandwidthLimit int64      `json:"bandwidth_limit"`
}

type CommandRequest struct {
	Type    CommandType `json:"type" binding:"required"`
	Payload string      `json:"payload"`
}

type CreatePairingCodeRequest struct {
	ExpiresInMinutes int        `json:"expires_in_minutes"`
	RelayServerID    *uuid.UUID `json:"relay_server_id"`
	ReassignDeviceID *uuid.UUID `json:"reassign_device_id"`
}

type CreatePairingCodeResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ClaimPairingCodeRequest struct {
	Code           string `json:"code" binding:"required"`
	AndroidID      string `json:"android_id" binding:"required"`
	DeviceModel    string `json:"device_model" binding:"required"`
	AndroidVersion string `json:"android_version" binding:"required"`
	AppVersion     string `json:"app_version" binding:"required"`
}

type ClaimPairingCodeResponse struct {
	DeviceID      uuid.UUID `json:"device_id"`
	AuthToken     string    `json:"auth_token"`
	ServerURL     string    `json:"server_url"`
	VpnConfig     string    `json:"vpn_config"`
	BasePort      int       `json:"base_port"`
	RelayServerIP string    `json:"relay_server_ip"`
}

// WebSocket message types

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
