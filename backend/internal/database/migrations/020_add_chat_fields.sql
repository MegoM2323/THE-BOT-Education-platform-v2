-- 020_add_chat_fields.sql
-- Purpose: Add missing fields to chat_rooms and messages tables
-- Created: 2025-12-08
--
-- This migration adds:
-- 1. last_message_at to chat_rooms (for sorting by activity)
-- 2. moderation_completed_at to messages (for tracking moderation timing)
-- 3. Index on last_message_at for efficient sorting
-- 4. Updates trigger to set last_message_at on message insert

BEGIN;

-- ============================================================================
-- ADD COLUMNS
-- ============================================================================

-- Добавляем last_message_at в chat_rooms
-- Используется для сортировки комнат по последней активности
ALTER TABLE chat_rooms
ADD COLUMN last_message_at TIMESTAMP WITH TIME ZONE;

-- Добавляем moderation_completed_at в messages
-- Время завершения модерации (успешной или блокировки)
ALTER TABLE messages
ADD COLUMN moderation_completed_at TIMESTAMP WITH TIME ZONE;

-- ============================================================================
-- CREATE INDEX
-- ============================================================================

-- Индекс для сортировки комнат по последней активности (DESC)
-- Используется в списке чатов для отображения самых активных сверху
CREATE INDEX idx_chat_rooms_last_message
ON chat_rooms(last_message_at DESC)
WHERE deleted_at IS NULL;

-- ============================================================================
-- UPDATE TRIGGER
-- ============================================================================

-- Обновляем функцию триггера: теперь обновляет и last_message_at
DROP FUNCTION IF EXISTS update_chat_room_timestamp() CASCADE;

CREATE OR REPLACE FUNCTION update_chat_room_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chat_rooms
    SET
        updated_at = CURRENT_TIMESTAMP,
        last_message_at = CURRENT_TIMESTAMP
    WHERE id = NEW.room_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Пересоздаём триггер (DROP CASCADE выше уже удалил старый)
CREATE TRIGGER messages_update_room_timestamp
AFTER INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION update_chat_room_timestamp();

-- ============================================================================
-- BACKFILL DATA (OPTIONAL)
-- ============================================================================

-- Заполняем last_message_at для существующих комнат
-- Берём время последнего сообщения в каждой комнате
UPDATE chat_rooms cr
SET last_message_at = (
    SELECT MAX(m.created_at)
    FROM messages m
    WHERE m.room_id = cr.id
      AND m.deleted_at IS NULL
)
WHERE EXISTS (
    SELECT 1
    FROM messages m
    WHERE m.room_id = cr.id
      AND m.deleted_at IS NULL
);

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- BEGIN;
-- DROP TRIGGER IF EXISTS messages_update_room_timestamp ON messages;
-- DROP FUNCTION IF EXISTS update_chat_room_timestamp() CASCADE;
--
-- -- Restore old trigger function (without last_message_at update)
-- CREATE OR REPLACE FUNCTION update_chat_room_timestamp()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     UPDATE chat_rooms
--     SET updated_at = CURRENT_TIMESTAMP
--     WHERE id = NEW.room_id;
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE TRIGGER messages_update_room_timestamp
-- AFTER INSERT ON messages
-- FOR EACH ROW
-- EXECUTE FUNCTION update_chat_room_timestamp();
--
-- DROP INDEX IF EXISTS idx_chat_rooms_last_message;
-- ALTER TABLE messages DROP COLUMN IF EXISTS moderation_completed_at;
-- ALTER TABLE chat_rooms DROP COLUMN IF EXISTS last_message_at;
-- COMMIT;
