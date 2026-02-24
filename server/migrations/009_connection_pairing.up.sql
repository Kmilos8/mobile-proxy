ALTER TABLE pairing_codes ADD COLUMN connection_id UUID REFERENCES proxy_connections(id) ON DELETE SET NULL;
