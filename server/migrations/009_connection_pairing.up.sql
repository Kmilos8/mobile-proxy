ALTER TABLE pairing_codes ADD COLUMN IF NOT EXISTS reassign_device_id UUID REFERENCES devices(id) ON DELETE SET NULL;
ALTER TABLE pairing_codes DROP COLUMN IF EXISTS connection_id;
