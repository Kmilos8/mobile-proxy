-- Add auth columns to customers table
ALTER TABLE customers ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS google_id VARCHAR(255) UNIQUE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS google_email VARCHAR(255);

-- Create customer_auth_tokens table for email verification and password reset tokens
CREATE TABLE IF NOT EXISTS customer_auth_tokens (
    id          UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    customer_id UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    type        VARCHAR(50) NOT NULL CHECK (type IN ('email_verify', 'password_reset')),
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customer_auth_tokens_hash     ON customer_auth_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_customer_auth_tokens_customer ON customer_auth_tokens(customer_id, type);
