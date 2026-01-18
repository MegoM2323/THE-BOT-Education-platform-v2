-- 049_chat_auto_create.sql
-- Purpose: Auto-create chat room after lesson completion
-- Created: 2025-01-18
--
-- This migration adds:
-- 1. Function create_chat_after_lesson_completion()
-- 2. Trigger on lessons table (when lesson end_time passes)
-- 3. Alternative: periodic function to create chats for completed lessons
--
-- Logic:
-- - When a lesson is completed (end_time < NOW()) and has active bookings
-- - Create chat_room between student and teacher (methodologist)
-- - Only if chat doesn't already exist

BEGIN;

-- ============================================================================
-- FUNCTION: Create chat room after lesson completion
-- ============================================================================

CREATE OR REPLACE FUNCTION create_chat_after_lesson_completion()
RETURNS TRIGGER AS $$
DECLARE
    v_student_id UUID;
    v_teacher_id UUID;
    v_chat_exists BOOLEAN;
BEGIN
    -- Only process if lesson has ended (end_time is in the past)
    -- This handles both UPDATE to end_time and initial INSERT
    IF NEW.end_time IS NOT NULL AND NEW.end_time < CURRENT_TIMESTAMP THEN
        -- Get teacher_id from the lesson
        v_teacher_id := NEW.teacher_id;

        -- Create chat for each active booking of this lesson
        FOR v_student_id IN
            SELECT b.student_id
            FROM bookings b
            WHERE b.lesson_id = NEW.id
              AND b.status = 'active'
        LOOP
            -- Check if chat already exists between this teacher and student
            SELECT EXISTS(
                SELECT 1 FROM chat_rooms cr
                WHERE cr.teacher_id = v_teacher_id
                  AND cr.student_id = v_student_id
                  AND cr.deleted_at IS NULL
            ) INTO v_chat_exists;

            -- Create chat room if it doesn't exist
            IF NOT v_chat_exists THEN
                INSERT INTO chat_rooms (teacher_id, student_id, created_at, updated_at)
                VALUES (v_teacher_id, v_student_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT (teacher_id, student_id) DO NOTHING;
            END IF;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_chat_after_lesson_completion() IS
'Creates chat room between teacher and student when lesson ends';

-- ============================================================================
-- TRIGGER: On lessons update (when lesson end_time changes or passes)
-- ============================================================================

DROP TRIGGER IF EXISTS lesson_completion_create_chat ON lessons;

CREATE TRIGGER lesson_completion_create_chat
AFTER UPDATE ON lessons
FOR EACH ROW
WHEN (NEW.deleted_at IS NULL AND NEW.end_time < CURRENT_TIMESTAMP)
EXECUTE FUNCTION create_chat_after_lesson_completion();

-- ============================================================================
-- FUNCTION: Batch create chats for all completed lessons (for cron/scheduler)
-- ============================================================================

CREATE OR REPLACE FUNCTION create_chats_for_completed_lessons()
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER := 0;
    v_lesson RECORD;
    v_student_id UUID;
    v_chat_exists BOOLEAN;
BEGIN
    -- Find all completed lessons (end_time passed) with active bookings
    -- that don't have corresponding chat rooms yet
    FOR v_lesson IN
        SELECT DISTINCT l.id, l.teacher_id
        FROM lessons l
        INNER JOIN bookings b ON b.lesson_id = l.id AND b.status = 'active'
        WHERE l.end_time < CURRENT_TIMESTAMP
          AND l.deleted_at IS NULL
    LOOP
        -- For each active booking in this lesson
        FOR v_student_id IN
            SELECT b.student_id
            FROM bookings b
            WHERE b.lesson_id = v_lesson.id
              AND b.status = 'active'
        LOOP
            -- Check if chat already exists
            SELECT EXISTS(
                SELECT 1 FROM chat_rooms cr
                WHERE cr.teacher_id = v_lesson.teacher_id
                  AND cr.student_id = v_student_id
                  AND cr.deleted_at IS NULL
            ) INTO v_chat_exists;

            -- Create chat room if it doesn't exist
            IF NOT v_chat_exists THEN
                INSERT INTO chat_rooms (teacher_id, student_id, created_at, updated_at)
                VALUES (v_lesson.teacher_id, v_student_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT (teacher_id, student_id) DO NOTHING;

                v_count := v_count + 1;
            END IF;
        END LOOP;
    END LOOP;

    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_chats_for_completed_lessons() IS
'Batch function to create chat rooms for all completed lessons. Returns count of created chats. Call periodically via scheduler.';

-- ============================================================================
-- INITIAL BACKFILL: Create chats for already completed lessons
-- ============================================================================

DO $$
DECLARE
    v_created INTEGER;
BEGIN
    SELECT create_chats_for_completed_lessons() INTO v_created;
    RAISE NOTICE 'Created % chat rooms for completed lessons', v_created;
END $$;

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- BEGIN;
-- DROP TRIGGER IF EXISTS lesson_completion_create_chat ON lessons;
-- DROP FUNCTION IF EXISTS create_chat_after_lesson_completion();
-- DROP FUNCTION IF EXISTS create_chats_for_completed_lessons();
-- COMMIT;
