-- Rotation links: public URLs that trigger IP rotation on a device
CREATE TABLE rotation_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX idx_rotation_links_device ON rotation_links(device_id);
CREATE INDEX idx_rotation_links_token ON rotation_links(token);
