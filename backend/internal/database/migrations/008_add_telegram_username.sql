-- 008_add_telegram_username.sql
-- Add telegram_username field to users table

ALTER TABLE users
ADD COLUMN IF NOT EXISTS telegram_username VARCHAR(255);

-- Create index for faster lookups of telegram usernames
CREATE INDEX IF NOT EXISTS idx_users_telegram_username ON users(telegram_username) WHERE telegram_username IS NOT NULL;

-- Comment
COMMENT ON COLUMN users.telegram_username IS 'Telegram username for the user (unique identifier in Telegram)';
