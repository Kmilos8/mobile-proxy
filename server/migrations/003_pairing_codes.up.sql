-- Pairing codes: allow devices to self-register via QR code / manual code entry
CREATE TABLE pairing_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(8) NOT NULL UNIQUE,
    device_auth_token VARCHAR(64) NOT NULL,
    claimed_by_device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    claimed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pairing_codes_code ON pairing_codes(code);

-- Add auth_token column to devices for token-based device authentication
ALTER TABLE devices ADD COLUMN auth_token VARCHAR(64);
