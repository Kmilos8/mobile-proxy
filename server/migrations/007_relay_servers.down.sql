ALTER TABLE pairing_codes DROP COLUMN IF EXISTS relay_server_id;
ALTER TABLE devices DROP COLUMN IF EXISTS relay_server_id;
DROP TABLE IF EXISTS relay_servers;
