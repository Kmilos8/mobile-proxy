ALTER TABLE proxy_connections ADD COLUMN base_port INTEGER UNIQUE;
ALTER TABLE proxy_connections ADD COLUMN http_port INTEGER;
ALTER TABLE proxy_connections ADD COLUMN socks5_port INTEGER;
ALTER TABLE devices ADD COLUMN description TEXT NOT NULL DEFAULT '';
