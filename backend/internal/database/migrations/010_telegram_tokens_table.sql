-- 010_telegram_tokens_table.sql
-- Telegram token persistence for account linking flow
-- Replaces in-memory token storage with persistent PostgreSQL storage

-- Telegram tokens table - stores temporary linking tokens with TTL
CREATE TABLE telegram_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CHECK (expires_at > created_at)
);

-- Indexes for efficient queries
CREATE INDEX idx_telegram_tokens_token ON telegram_tokens(token);
CREATE INDEX idx_telegram_tokens_user_id ON telegram_tokens(user_id);
CREATE INDEX idx_telegram_tokens_expires_at ON telegram_tokens(expires_at);

-- Note: Removed partial index with CURRENT_TIMESTAMP as it requires IMMUTABLE functions
-- Application will filter expired tokens in code instead of at database level

-- Comments
COMMENT ON TABLE telegram_tokens IS 'Temporary tokens for linking Telegram accounts to platform users';
COMMENT ON COLUMN telegram_tokens.token IS 'Unique token string sent to Telegram bot by user';
COMMENT ON COLUMN telegram_tokens.user_id IS 'Platform user this token belongs to';
COMMENT ON COLUMN telegram_tokens.expires_at IS 'Token expiration time (typically 15 minutes from creation)';
COMMENT ON COLUMN telegram_tokens.created_at IS 'Timestamp when token was generated';
