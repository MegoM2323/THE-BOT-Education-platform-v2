-- 019_add_chat_payment_indexes.sql
-- Purpose: Add missing foreign key indexes for chat and payment tables
-- Created: 2025-12-06
--
-- This migration adds indexes for foreign key columns that don't have them:
-- 1. messages.room_id - JOIN optimization for chat history queries
-- 2. messages.sender_id - filter messages by sender
-- 3. blocked_messages.message_id - find blocked message details
--
-- Note: file_attachments.message_id and payments.user_id already have indexes
-- from previous migrations (014_forum_system.sql, 018_payments_system.sql)
--
-- IMPORTANT: CREATE INDEX CONCURRENTLY cannot run inside a transaction block
-- This migration must run without BEGIN/COMMIT wrapper

-- ============================================================================
-- CHAT INDEXES
-- ============================================================================

-- Messages: room_id foreign key index
-- Covers JOIN queries: SELECT messages.* FROM messages JOIN chat_rooms ON ...
-- Partial index excludes soft-deleted messages (deleted_at IS NULL)
-- CONCURRENTLY prevents table locks during index creation
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_room_id
ON messages(room_id)
WHERE deleted_at IS NULL;

-- Messages: sender_id foreign key index
-- Covers queries: SELECT * FROM messages WHERE sender_id = $1
-- No partial index (need to query all messages regardless of deletion status)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_sender_id
ON messages(sender_id);

-- Messages: status partial index for pending moderation queue
-- Optimization for AI moderation worker: WHERE status = 'pending_moderation'
-- Already exists in 014_forum_system.sql as idx_messages_pending_moderation
-- but we add this for completeness (IF NOT EXISTS prevents duplicate)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_status
ON messages(status)
WHERE status = 'pending_moderation';

-- Blocked messages: message_id foreign key index
-- Covers queries: SELECT * FROM blocked_messages WHERE message_id = $1
-- Used when loading blocked message details for admin review
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocked_messages_message_id
ON blocked_messages(message_id);

-- ============================================================================
-- PAYMENT INDEXES
-- ============================================================================

-- Payments: user_id foreign key index
-- Already exists in 018_payments_system.sql as idx_payments_user (composite)
-- The composite index (user_id, created_at DESC) efficiently covers both:
-- - Queries: WHERE user_id = $1 (uses leftmost column of composite)
-- - Queries: WHERE user_id = $1 ORDER BY created_at DESC
-- No additional index needed for user_id alone

-- Payments: status index for filtering by payment status
-- Already exists in 018_payments_system.sql as idx_payments_pending_status (partial)
-- But we add a general status index for all statuses (not just pending)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_status
ON payments(status);

-- Payments: yookassa_payment_id unique index
-- Already exists in 018_payments_system.sql as idx_payments_yookassa_id
-- Used for webhook processing: WHERE yookassa_payment_id = $1
-- No additional index needed

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================
-- Run these to verify indexes are being used:
--
-- Chat room message history (should use idx_messages_room_id):
-- EXPLAIN ANALYZE SELECT * FROM messages
-- WHERE room_id = (SELECT id FROM chat_rooms LIMIT 1);
--
-- User payment history (should use idx_payments_user):
-- EXPLAIN ANALYZE SELECT * FROM payments
-- WHERE user_id = (SELECT id FROM users WHERE role = 'student' LIMIT 1);
--
-- Sender's messages (should use idx_messages_sender_id):
-- EXPLAIN ANALYZE SELECT * FROM messages
-- WHERE sender_id = (SELECT id FROM users LIMIT 1);
--
-- Blocked message details (should use idx_blocked_messages_message_id):
-- EXPLAIN ANALYZE SELECT * FROM blocked_messages
-- WHERE message_id = (SELECT id FROM messages LIMIT 1);

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- DROP INDEX CONCURRENTLY IF EXISTS idx_blocked_messages_message_id;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_messages_status;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_messages_sender_id;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_messages_room_id;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_payments_status;
