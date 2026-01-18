-- 050_chat_on_booking_creation.sql
-- Purpose: Auto-create chat room when booking becomes active
-- Created: 2026-01-18
--
-- This migration adds:
-- 1. Function create_chat_on_booking_active()
-- 2. Trigger on bookings table (when status = 'active')
-- 3. Backfill existing active bookings to create missing chats
--
-- Logic:
-- - When a booking is inserted or updated with status = 'active'
-- - Create chat_room between student and teacher (if not exists)
-- - Use ON CONFLICT to handle duplicates gracefully
-- - Replaces the old lesson-completion trigger for earlier chat creation

BEGIN;

-- ============================================================================
-- FUNCTION: Create chat room when booking becomes active
-- ============================================================================

CREATE OR REPLACE FUNCTION create_chat_on_booking_active()
RETURNS TRIGGER AS $$
DECLARE
    v_teacher_id UUID;
BEGIN
    -- Obtain teacher_id from the associated lesson
    SELECT l.teacher_id INTO v_teacher_id
    FROM lessons l
    WHERE l.id = NEW.lesson_id
      AND l.deleted_at IS NULL;

    -- Exit early if lesson not found or teacher is null
    IF v_teacher_id IS NULL THEN
        RETURN NEW;
    END IF;

    -- Prevent self-chat (teacher cannot chat with themselves)
    IF v_teacher_id = NEW.student_id THEN
        RETURN NEW;
    END IF;

    -- Create chat between teacher and student
    -- ON CONFLICT DO NOTHING: handle case where chat already exists
    INSERT INTO chat_rooms (teacher_id, student_id, created_at, updated_at)
    VALUES (v_teacher_id, NEW.student_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    ON CONFLICT (teacher_id, student_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_chat_on_booking_active() IS
'Creates chat room between teacher and student when booking becomes active. Called before INSERT/UPDATE on bookings.';

-- ============================================================================
-- TRIGGER: On bookings insert or update (when status = 'active')
-- ============================================================================

DROP TRIGGER IF EXISTS booking_create_chat ON bookings;

CREATE TRIGGER booking_create_chat
BEFORE INSERT ON bookings
FOR EACH ROW
WHEN (NEW.status = 'active')
EXECUTE FUNCTION create_chat_on_booking_active();

-- Note: Also handle UPDATE case if booking status changes to 'active'
DROP TRIGGER IF EXISTS booking_create_chat_update ON bookings;

CREATE TRIGGER booking_create_chat_update
BEFORE UPDATE ON bookings
FOR EACH ROW
WHEN (NEW.status = 'active' AND OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION create_chat_on_booking_active();

-- ============================================================================
-- INITIAL BACKFILL: Create chats for existing active bookings
-- ============================================================================

DO $$
DECLARE
    v_created INTEGER := 0;
    v_booking RECORD;
    v_teacher_id UUID;
BEGIN
    -- Find all active bookings that don't have a corresponding chat room
    FOR v_booking IN
        SELECT DISTINCT b.id, b.student_id, b.lesson_id, l.teacher_id
        FROM bookings b
        INNER JOIN lessons l ON l.id = b.lesson_id
        WHERE b.status = 'active'
          AND l.deleted_at IS NULL
          AND NOT EXISTS (
              SELECT 1 FROM chat_rooms cr
              WHERE cr.teacher_id = l.teacher_id
                AND cr.student_id = b.student_id
                AND cr.deleted_at IS NULL
          )
    LOOP
        -- Create chat room (with conflict handling)
        INSERT INTO chat_rooms (teacher_id, student_id, created_at, updated_at)
        VALUES (v_booking.teacher_id, v_booking.student_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        ON CONFLICT (teacher_id, student_id) DO NOTHING;

        v_created := v_created + 1;
    END LOOP;

    RAISE NOTICE 'Created % chat rooms for existing active bookings', v_created;
END $$;

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- BEGIN;
-- DROP TRIGGER IF EXISTS booking_create_chat_update ON bookings;
-- DROP TRIGGER IF EXISTS booking_create_chat ON bookings;
-- DROP FUNCTION IF EXISTS create_chat_on_booking_active();
-- COMMIT;
