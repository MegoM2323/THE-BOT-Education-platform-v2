-- 035_template_capacity_check.sql
-- Purpose: Fix template lesson capacity validation on INSERT operations
-- Issue: FB016 - Capacity check only on UPDATE trigger, not INSERT
--
-- The validate_template_lesson_capacity() function in migration 011 only validates
-- capacity when max_students is updated, but does not validate when students are
-- added to the template lesson. This allows initial student additions to exceed
-- max_students, violating data integrity constraints.
--
-- Fix: Add INSERT trigger on template_lesson_students to validate capacity

BEGIN;

-- Create new function to validate student count against max_students on INSERT
CREATE OR REPLACE FUNCTION validate_student_capacity_on_insert()
RETURNS TRIGGER AS $$
DECLARE
    max_students INTEGER;
    current_count INTEGER;
BEGIN
    -- Get max_students from the template_lesson
    SELECT tl.max_students INTO max_students
    FROM template_lessons tl
    WHERE tl.id = NEW.template_lesson_id;

    IF max_students IS NULL THEN
        RAISE EXCEPTION 'Template lesson % not found', NEW.template_lesson_id;
    END IF;

    -- Count current students (this row is not yet inserted, so COUNT(*) gives pre-insert count)
    SELECT COUNT(*) INTO current_count
    FROM template_lesson_students
    WHERE template_lesson_id = NEW.template_lesson_id;

    -- Validate that adding this student would not exceed capacity
    IF (current_count + 1) > max_students THEN
        RAISE EXCEPTION 'Cannot add student: lesson capacity (%) exceeded. Current: %, Max: %',
            NEW.student_id, current_count, max_students;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create INSERT trigger on template_lesson_students
CREATE TRIGGER validate_student_capacity_on_insert
    BEFORE INSERT ON template_lesson_students
    FOR EACH ROW
    EXECUTE FUNCTION validate_student_capacity_on_insert();

-- Add comment
COMMENT ON FUNCTION validate_student_capacity_on_insert() IS 'Validates that adding a student to a template lesson does not exceed max_students capacity';
COMMENT ON TRIGGER validate_student_capacity_on_insert ON template_lesson_students IS 'Ensures template lesson capacity is enforced on student insertion';

-- Verify existing data does not violate constraint
-- Report any template_lessons that have more students than max_students
DO $$
DECLARE
    v_violation RECORD;
    v_violation_count INTEGER := 0;
BEGIN
    FOR v_violation IN
        SELECT
            tls.id,
            tls.max_students,
            COUNT(tlstud.id) as student_count
        FROM template_lessons tls
        LEFT JOIN template_lesson_students tlstud ON tlstud.template_lesson_id = tls.id
        GROUP BY tls.id, tls.max_students
        HAVING COUNT(tlstud.id) > tls.max_students
    LOOP
        v_violation_count := v_violation_count + 1;
        RAISE WARNING 'Capacity violation found: template_lesson % has % students but max_students is %',
            v_violation.id, v_violation.student_count, v_violation.max_students;
    END LOOP;

    IF v_violation_count > 0 THEN
        RAISE WARNING 'Found % template lessons with student count exceeding max_students', v_violation_count;
    ELSE
        RAISE NOTICE 'No capacity violations found in existing data';
    END IF;
END;
$$;

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration, run:
/*
BEGIN;

DROP TRIGGER IF EXISTS validate_student_capacity_on_insert ON template_lesson_students;
DROP FUNCTION IF EXISTS validate_student_capacity_on_insert();

COMMIT;
*/
