-- Add customer ownership to devices
ALTER TABLE devices ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

-- Add customer assignment to pairing_codes
ALTER TABLE pairing_codes ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

-- Index for filtering devices by customer
CREATE INDEX IF NOT EXISTS idx_devices_customer ON devices(customer_id);

-- Device sharing table: owner shares device with another customer, with per-permission granularity
CREATE TABLE IF NOT EXISTS device_shares (
    id                   UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    device_id            UUID        NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    owner_id             UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    shared_with          UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    can_rename           BOOLEAN     NOT NULL DEFAULT FALSE,
    can_manage_ports     BOOLEAN     NOT NULL DEFAULT FALSE,
    can_download_configs BOOLEAN     NOT NULL DEFAULT FALSE,
    can_rotate_ip        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (device_id, shared_with)
);

CREATE INDEX IF NOT EXISTS idx_device_shares_shared_with ON device_shares(shared_with);
CREATE INDEX IF NOT EXISTS idx_device_shares_device      ON device_shares(device_id);

-- Seed: create a customer record for the existing operator (from first admin user)
INSERT INTO customers (id, name, email, active, email_verified)
SELECT uuid_generate_v4(), u.name, u.email, true, true
FROM users u
WHERE u.role = 'admin'
LIMIT 1
ON CONFLICT (email) DO NOTHING;

-- Stamp all existing devices to the operator's customer account
UPDATE devices
SET customer_id = (
    SELECT c.id
    FROM customers c
    JOIN users u ON LOWER(c.email) = LOWER(u.email)
    WHERE u.role = 'admin'
    LIMIT 1
)
WHERE customer_id IS NULL;

-- Backfill proxy_connections.customer_id from device ownership
UPDATE proxy_connections pc
SET customer_id = d.customer_id
FROM devices d
WHERE pc.device_id = d.id
  AND pc.customer_id IS NULL;
