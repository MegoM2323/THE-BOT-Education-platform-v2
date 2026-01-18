-- 005_telegram_integration.sql
-- Telegram integration schema for notification system and broadcasts

-- Telegram users table - links platform users to their Telegram accounts
CREATE TABLE telegram_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    telegram_id BIGINT NOT NULL UNIQUE,
    chat_id BIGINT NOT NULL,
    username VARCHAR(255),
    subscribed BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Broadcast lists table - groups of users for targeted messaging
CREATE TABLE broadcast_lists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    user_ids UUID[] NOT NULL DEFAULT '{}',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Broadcasts table - history of broadcast messages sent
CREATE TABLE broadcasts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    list_id UUID REFERENCES broadcast_lists(id) ON DELETE SET NULL,
    message TEXT NOT NULL,
    sent_count INTEGER DEFAULT 0 CHECK (sent_count >= 0),
    failed_count INTEGER DEFAULT 0 CHECK (failed_count >= 0),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'cancelled')),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Broadcast logs table - detailed log of each message sent in a broadcast
CREATE TABLE broadcast_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    broadcast_id UUID NOT NULL REFERENCES broadcasts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    telegram_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('sent', 'failed', 'skipped')),
    error TEXT,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for telegram_users
CREATE INDEX idx_telegram_users_user_id ON telegram_users(user_id);
CREATE INDEX idx_telegram_users_telegram_id ON telegram_users(telegram_id);
CREATE INDEX idx_telegram_users_subscribed ON telegram_users(subscribed) WHERE subscribed = true;

-- Create indexes for broadcast_lists
CREATE INDEX idx_broadcast_lists_created_by ON broadcast_lists(created_by);
CREATE INDEX idx_broadcast_lists_deleted_at ON broadcast_lists(deleted_at) WHERE deleted_at IS NULL;

-- Create indexes for broadcasts
CREATE INDEX idx_broadcasts_list_id ON broadcasts(list_id);
CREATE INDEX idx_broadcasts_status ON broadcasts(status);
CREATE INDEX idx_broadcasts_created_at ON broadcasts(created_at DESC);
CREATE INDEX idx_broadcasts_created_by ON broadcasts(created_by);

-- Create indexes for broadcast_logs
CREATE INDEX idx_broadcast_logs_broadcast_id ON broadcast_logs(broadcast_id);
CREATE INDEX idx_broadcast_logs_user_id ON broadcast_logs(user_id);
CREATE INDEX idx_broadcast_logs_status ON broadcast_logs(status);
CREATE INDEX idx_broadcast_logs_sent_at ON broadcast_logs(sent_at DESC);

-- Comments for tables
COMMENT ON TABLE telegram_users IS 'Links platform users to their Telegram accounts for notifications';
COMMENT ON TABLE broadcast_lists IS 'Named groups of users for targeted broadcast messaging';
COMMENT ON TABLE broadcasts IS 'History of broadcast messages sent to user groups';
COMMENT ON TABLE broadcast_logs IS 'Detailed log of each individual message delivery attempt';

-- Comments for important columns
COMMENT ON COLUMN telegram_users.telegram_id IS 'Unique Telegram user ID from Telegram API';
COMMENT ON COLUMN telegram_users.chat_id IS 'Telegram chat ID for sending messages';
COMMENT ON COLUMN telegram_users.subscribed IS 'Whether user wants to receive notifications';
COMMENT ON COLUMN broadcast_lists.user_ids IS 'Array of user UUIDs included in this broadcast list';
COMMENT ON COLUMN broadcasts.status IS 'Current status of the broadcast: pending, in_progress, completed, failed, cancelled';
COMMENT ON COLUMN broadcasts.sent_count IS 'Number of messages successfully sent';
COMMENT ON COLUMN broadcasts.failed_count IS 'Number of messages that failed to send';
COMMENT ON COLUMN broadcast_logs.status IS 'Delivery status: sent, failed, or skipped';
