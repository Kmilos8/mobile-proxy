ALTER TABLE proxy_connections DROP COLUMN IF EXISTS socks5_port;
ALTER TABLE proxy_connections DROP COLUMN IF EXISTS http_port;
ALTER TABLE proxy_connections DROP COLUMN IF EXISTS base_port;
ALTER TABLE devices DROP COLUMN IF EXISTS description;
