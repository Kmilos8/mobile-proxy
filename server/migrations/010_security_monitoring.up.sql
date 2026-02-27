-- Add webhook_url to users (per-operator, used by Plan 03-02)
ALTER TABLE users ADD COLUMN IF NOT EXISTS webhook_url TEXT;

-- Add last_offline_alert_at to devices (cooldown tracking, used by Plan 03-02)
ALTER TABLE devices ADD COLUMN IF NOT EXISTS last_offline_alert_at TIMESTAMPTZ;

-- Null out password_plain per user decision (SOCKS5 path now uses password_hash as credential)
UPDATE proxy_connections SET password_plain = NULL WHERE password_plain IS NOT NULL;

-- Make password_plain nullable
ALTER TABLE proxy_connections ALTER COLUMN password_plain DROP NOT NULL;
ALTER TABLE proxy_connections ALTER COLUMN password_plain DROP DEFAULT;
