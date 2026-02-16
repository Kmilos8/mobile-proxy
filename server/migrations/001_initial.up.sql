-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'operator',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Devices table
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL DEFAULT '',
    android_id VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'offline',
    cellular_ip INET,
    wifi_ip INET,
    vpn_ip INET,
    carrier VARCHAR(100) DEFAULT '',
    network_type VARCHAR(20) DEFAULT '',
    battery_level INTEGER DEFAULT 0,
    battery_charging BOOLEAN DEFAULT FALSE,
    signal_strength INTEGER DEFAULT 0,
    base_port INTEGER NOT NULL UNIQUE,
    http_port INTEGER NOT NULL,
    socks5_port INTEGER NOT NULL,
    udp_relay_port INTEGER NOT NULL,
    ovpn_port INTEGER NOT NULL,
    last_heartbeat TIMESTAMPTZ,
    app_version VARCHAR(50) DEFAULT '',
    device_model VARCHAR(100) DEFAULT '',
    android_version VARCHAR(20) DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Customers table
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Proxy connections table
CREATE TABLE proxy_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    ip_whitelist TEXT[] DEFAULT '{}',
    bandwidth_limit BIGINT DEFAULT 0,
    bandwidth_used BIGINT DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- IP history table
CREATE TABLE ip_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    ip INET NOT NULL,
    method VARCHAR(50) NOT NULL DEFAULT 'natural',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Bandwidth logs table (partitioned by month)
CREATE TABLE bandwidth_logs (
    id UUID NOT NULL DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL,
    connection_id UUID,
    bytes_in BIGINT NOT NULL DEFAULT 0,
    bytes_out BIGINT NOT NULL DEFAULT 0,
    interval_start TIMESTAMPTZ NOT NULL,
    interval_end TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, interval_start)
) PARTITION BY RANGE (interval_start);

-- Create initial partition for current month and next month
CREATE TABLE bandwidth_logs_2026_02 PARTITION OF bandwidth_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE bandwidth_logs_2026_03 PARTITION OF bandwidth_logs
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

-- Device commands table
CREATE TABLE device_commands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payload JSONB DEFAULT '{}',
    result JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_android_id ON devices(android_id);
CREATE INDEX idx_proxy_connections_device ON proxy_connections(device_id);
CREATE INDEX idx_proxy_connections_customer ON proxy_connections(customer_id);
CREATE INDEX idx_proxy_connections_username ON proxy_connections(username);
CREATE INDEX idx_ip_history_device ON ip_history(device_id);
CREATE INDEX idx_ip_history_created ON ip_history(created_at DESC);
CREATE INDEX idx_bandwidth_logs_device ON bandwidth_logs(device_id, interval_start);
CREATE INDEX idx_device_commands_device_status ON device_commands(device_id, status);
CREATE INDEX idx_device_commands_pending ON device_commands(device_id) WHERE status = 'pending';

-- Seed admin user (password: admin123 - change in production!)
-- bcrypt hash of 'admin123' generated with cost 10
INSERT INTO users (email, password_hash, name, role) VALUES
('admin@mobileproxy.local', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Admin', 'admin');
