-- 051_improve_chat_trigger_logging.sql
-- Purpose: Add logging to chat creation trigger for debugging
-- Created: 2026-01-18
--
-- Changes:
-- 1. Recreate function with RAISE NOTICE for debugging
-- 2. Add explicit ID generation (gen_random_uuid())
-- 3. Backfill with self-chat prevention check

BEGIN;

-- ============================================================================
-- FUNCTION: Create chat room when booking becomes active (with logging)
-- ============================================================================

CREATE OR REPLACE FUNCTION create_chat_on_booking_active()
RETURNS TRIGGER AS $$
DECLARE
    v_teacher_id UUID;
    v_chat_exists BOOLEAN;
BEGIN
    -- Only process if booking status is 'active'
    IF NEW.status != 'active' THEN
        RETURN NEW;
    END IF;

    -- Get teacher_id from lesson
    SELECT l.teacher_id INTO v_teacher_id
    FROM lessons l
    WHERE l.id = NEW.lesson_id
      AND l.deleted_at IS NULL;

    -- Log the attempt
    RAISE NOTICE 'Chat trigger: booking=%, student=%, lesson=%, teacher=%',
        NEW.id, NEW.student_id, NEW.lesson_id, v_teacher_id;

    -- Exit if teacher not found
    IF v_teacher_id IS NULL THEN
        RAISE NOTICE 'Chat not created: teacher_id is NULL for lesson %', NEW.lesson_id;
        RETURN NEW;
    END IF;

    -- Prevent self-chat
    IF v_teacher_id = NEW.student_id THEN
        RAISE NOTICE 'Chat not created: teacher=student (self-chat prevention)';
        RETURN NEW;
    END IF;

    -- Check if chat already exists
    SELECT EXISTS(
        SELECT 1 FROM chat_rooms
        WHERE teacher_id = v_teacher_id
          AND student_id = NEW.student_id
          AND deleted_at IS NULL
    ) INTO v_chat_exists;

    IF v_chat_exists THEN
        RAISE NOTICE 'Chat already exists between teacher % and student %', v_teacher_id, NEW.student_id;
        RETURN NEW;
    END IF;

    -- Create chat room
    INSERT INTO chat_rooms (id, teacher_id, student_id, created_at, updated_at)
    VALUES (gen_random_uuid(), v_teacher_id, NEW.student_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    ON CONFLICT (teacher_id, student_id) DO NOTHING;

    RAISE NOTICE 'Chat created successfully for booking %', NEW.id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_chat_on_booking_active() IS
'Creates chat room between teacher and student when booking becomes active. Includes logging for debugging.';

-- ============================================================================
-- BACKFILL: Create chats for existing active bookings that don't have chat rooms
-- ============================================================================

DO $$
DECLARE
    v_created INTEGER := 0;
    v_booking RECORD;
BEGIN
    FOR v_booking IN
        SELECT DISTINCT b.id, b.student_id, l.teacher_id
        FROM bookings b
        INNER JOIN lessons l ON l.id = b.lesson_id
        WHERE b.status = 'active'
          AND l.deleted_at IS NULL
          AND l.teacher_id IS NOT NULL
          AND l.teacher_id != b.student_id
          AND NOT EXISTS (
              SELECT 1 FROM chat_rooms cr
              WHERE cr.teacher_id = l.teacher_id
                AND cr.student_id = b.student_id
                AND cr.deleted_at IS NULL
          )
    LOOP
        INSERT INTO chat_rooms (id, teacher_id, student_id, created_at, updated_at)
        VALUES (gen_random_uuid(), v_booking.teacher_id, v_booking.student_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        ON CONFLICT (teacher_id, student_id) DO NOTHING;

        v_created := v_created + 1;
    END LOOP;

    RAISE NOTICE 'Backfill: created % chat rooms for existing bookings', v_created;
END $$;

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration (restore original function without logging):
-- Run migration 050 again to restore original function
