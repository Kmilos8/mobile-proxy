CREATE TABLE relay_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    ip VARCHAR(45) NOT NULL UNIQUE,
    location VARCHAR(100) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO relay_servers (name, ip, location) VALUES ('Ashburn VA', '178.156.210.156', 'Ashburn, VA, US');

ALTER TABLE devices ADD COLUMN relay_server_id UUID REFERENCES relay_servers(id);
UPDATE devices SET relay_server_id = (SELECT id FROM relay_servers LIMIT 1);

ALTER TABLE pairing_codes ADD COLUMN relay_server_id UUID REFERENCES relay_servers(id);
