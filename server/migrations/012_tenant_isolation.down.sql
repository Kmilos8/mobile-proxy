-- Reverse: tenant isolation migration

-- Remove backfilled customer_id from proxy_connections (set to NULL for connections affected by data seed)
-- Note: We cannot reliably undo the backfill — we only nullify for safety
UPDATE proxy_connections pc
SET customer_id = NULL
FROM devices d
WHERE pc.device_id = d.id;

-- Drop device_shares table
DROP TABLE IF EXISTS device_shares;

-- Drop indexes
DROP INDEX IF EXISTS idx_devices_customer;

-- Remove customer_id from devices and pairing_codes
ALTER TABLE devices DROP COLUMN IF EXISTS customer_id;
ALTER TABLE pairing_codes DROP COLUMN IF EXISTS customer_id;
