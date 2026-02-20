ALTER TABLE proxy_connections ADD COLUMN proxy_type TEXT NOT NULL DEFAULT 'http';
ALTER TABLE proxy_connections ADD COLUMN password_plain TEXT NOT NULL DEFAULT '';
