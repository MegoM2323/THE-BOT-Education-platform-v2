-- 014_forum_system.sql
-- Purpose: Add forum/chat system with AI moderation
-- Created: 2025-12-05
--
-- This migration adds:
-- 1. Chat rooms (teacher ↔ student, one-to-one)
-- 2. Messages with AI moderation status
-- 3. File attachments for messages
-- 4. Blocked messages audit trail
-- 5. Triggers for auto-update chat room timestamps

BEGIN;

-- ============================================================================
-- TABLE DEFINITIONS
-- ============================================================================

-- Chat rooms: One-to-one chat between teacher and student
-- Each pair (teacher, student) can have only one active chat room
CREATE TABLE chat_rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    -- Unique constraint: one room per (teacher, student) pair
    -- Partial index excludes soft-deleted rooms
    CONSTRAINT chat_rooms_unique_pair UNIQUE(teacher_id, student_id),
    -- Prevent self-chat (teacher cannot chat with themselves)
    CONSTRAINT chat_rooms_no_self_chat CHECK (teacher_id != student_id)
);

-- Messages: Chat messages with AI moderation
-- Status workflow: pending_moderation → delivered | blocked
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending_moderation',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    -- Validate message length: 1-4096 characters (Telegram limit)
    CONSTRAINT messages_text_length CHECK (
        char_length(message_text) > 0 AND
        char_length(message_text) <= 4096
    ),
    -- Validate status: only allowed values
    CONSTRAINT messages_status_valid CHECK (
        status IN ('pending_moderation', 'delivered', 'blocked')
    )
);

-- File attachments: Files attached to messages
-- Max file size: 10MB (enforced at application layer and DB)
CREATE TABLE file_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Validate file size: > 0 and <= 10MB
    CONSTRAINT file_attachments_size_valid CHECK (
        file_size > 0 AND
        file_size <= 10485760
    ),
    -- Validate file name not empty
    CONSTRAINT file_attachments_name_not_empty CHECK (
        char_length(file_name) > 0
    )
);

-- Blocked messages: Audit trail for AI-blocked messages
-- Used for admin review and AI training feedback
CREATE TABLE blocked_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    ai_response JSONB,
    blocked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    admin_notified BOOLEAN NOT NULL DEFAULT FALSE,
    -- Validate reason not empty
    CONSTRAINT blocked_messages_reason_not_empty CHECK (
        char_length(reason) > 0
    )
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Chat rooms: Find all chats for a teacher (active only)
CREATE INDEX idx_chat_rooms_teacher
ON chat_rooms(teacher_id)
WHERE deleted_at IS NULL;

-- Chat rooms: Find all chats for a student (active only)
CREATE INDEX idx_chat_rooms_student
ON chat_rooms(student_id)
WHERE deleted_at IS NULL;

-- Messages: Get messages for a room ordered by time (active only)
-- Most recent messages first for pagination
CREATE INDEX idx_messages_room_time
ON messages(room_id, created_at DESC)
WHERE deleted_at IS NULL;

-- Messages: Find pending moderation messages (for AI worker)
-- Partial index for efficiency (only pending messages)
CREATE INDEX idx_messages_pending_moderation
ON messages(status, created_at)
WHERE status = 'pending_moderation';

-- File attachments: Find all attachments for a message
CREATE INDEX idx_file_attachments_message
ON file_attachments(message_id);

-- Blocked messages: Find unnotified blocks (for admin alerts)
-- Partial index for efficiency (only unnotified)
CREATE INDEX idx_blocked_messages_admin_notified
ON blocked_messages(admin_notified, blocked_at DESC)
WHERE admin_notified = FALSE;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Trigger function: Update chat_rooms.updated_at when new message added
-- Keeps updated_at in sync with latest message time
CREATE OR REPLACE FUNCTION update_chat_room_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chat_rooms
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.room_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Execute update_chat_room_timestamp after INSERT on messages
CREATE TRIGGER messages_update_room_timestamp
AFTER INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION update_chat_room_timestamp();

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- BEGIN;
-- DROP TRIGGER IF EXISTS messages_update_room_timestamp ON messages;
-- DROP FUNCTION IF EXISTS update_chat_room_timestamp();
-- DROP INDEX IF EXISTS idx_blocked_messages_admin_notified;
-- DROP INDEX IF EXISTS idx_file_attachments_message;
-- DROP INDEX IF EXISTS idx_messages_pending_moderation;
-- DROP INDEX IF EXISTS idx_messages_room_time;
-- DROP INDEX IF EXISTS idx_chat_rooms_student;
-- DROP INDEX IF EXISTS idx_chat_rooms_teacher;
-- DROP TABLE IF EXISTS blocked_messages;
-- DROP TABLE IF EXISTS file_attachments;
-- DROP TABLE IF EXISTS messages;
-- DROP TABLE IF EXISTS chat_rooms;
-- COMMIT;
