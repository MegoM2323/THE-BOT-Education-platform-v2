-- 035_broadcast_idempotency.sql
-- Add idempotency tracking for broadcast message delivery
-- Prevents duplicate messages when retrying failed deliveries

-- Add composite unique index on (broadcast_id, user_id) for idempotency
-- This ensures only one delivery record per user per broadcast
ALTER TABLE broadcast_logs
ADD CONSTRAINT broadcast_logs_broadcast_user_unique
UNIQUE (broadcast_id, user_id);

-- Add index for checking if user already received message (for idempotency check)
CREATE INDEX idx_broadcast_logs_broadcast_user_success
ON broadcast_logs(broadcast_id, user_id, status)
WHERE status = 'success';

-- Add comment explaining idempotency mechanism
-- Note: COMMENT ON CONSTRAINT does not support multi-line string literals with line breaks
COMMENT ON CONSTRAINT broadcast_logs_broadcast_user_unique ON broadcast_logs IS 'Ensures idempotency: only one delivery record per user per broadcast. Used to prevent duplicate messages on retry.';
