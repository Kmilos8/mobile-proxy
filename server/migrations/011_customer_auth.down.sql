DROP TABLE IF EXISTS customer_auth_tokens;

ALTER TABLE customers DROP COLUMN IF EXISTS password_hash;
ALTER TABLE customers DROP COLUMN IF EXISTS email_verified;
ALTER TABLE customers DROP COLUMN IF EXISTS google_id;
ALTER TABLE customers DROP COLUMN IF EXISTS google_email;
