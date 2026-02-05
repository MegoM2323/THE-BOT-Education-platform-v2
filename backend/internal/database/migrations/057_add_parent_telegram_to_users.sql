-- +migrate Up
-- username для отображения в UI
ALTER TABLE users ADD COLUMN IF NOT EXISTS parent_telegram_username VARCHAR(32);
-- chat_id для отправки сообщений (получается при привязке родителя к боту)
ALTER TABLE users ADD COLUMN IF NOT EXISTS parent_chat_id BIGINT;

ALTER TABLE users ADD CONSTRAINT chk_parent_telegram_length
    CHECK (parent_telegram_username IS NULL OR LENGTH(parent_telegram_username) BETWEEN 5 AND 32);

CREATE INDEX idx_users_parent_chat_id ON users(parent_chat_id) WHERE parent_chat_id IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_users_parent_chat_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_parent_telegram_length;
ALTER TABLE users DROP COLUMN IF EXISTS parent_chat_id;
ALTER TABLE users DROP COLUMN IF EXISTS parent_telegram_username;
