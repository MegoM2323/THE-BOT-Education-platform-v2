-- Migration 032: Add Teacher Lesson Overlap Prevention at DB Level
--
-- Purpose: Prevent creating overlapping lessons for the same teacher using PostgreSQL
--          exclusion constraints. This provides database-level enforcement of the constraint,
--          preventing race conditions and accidental overlaps at the application layer.
--
-- Constraints Added:
--   1. GiST index on (teacher_id, tstzrange(start_time, end_time))
--   2. Exclusion constraint to prevent overlapping lessons for same teacher
--   3. Soft-delete handling: constraint only applies to non-deleted lessons
--
-- PostgreSQL Features Used:
--   - tstzrange: Range type for timestamps with timezone
--   - GiST (Generalized Search Tree): Index type supporting range queries
--   - EXCLUDE: Constraint type for range-based conflicts
--
-- Reference: https://www.postgresql.org/docs/current/indexes-types.html#INDEXES-TYPES-GIST
--
-- Error Handling:
--   - PostgreSQL error code 23P01: exclusion_violation
--   - Repository layer will catch and convert to ErrLessonOverlapConflict

BEGIN;

-- Drop existing indexes/constraints if they exist (for idempotency)
DROP INDEX IF EXISTS idx_teacher_lesson_overlap_gist CASCADE;
ALTER TABLE lessons DROP CONSTRAINT IF EXISTS teacher_lessons_no_overlap;

-- Create GiST index for efficient range queries
-- This index is required for the exclusion constraint to work
CREATE INDEX idx_teacher_lesson_overlap_gist ON lessons
  USING GIST (teacher_id, tstzrange(start_time, end_time, '[)'))
  WHERE deleted_at IS NULL;

-- Add exclusion constraint to prevent overlapping lessons for same teacher
-- The constraint says: for two rows with the same teacher_id, their time ranges must not overlap
-- The '&&' operator means "overlaps" for range types
--
-- Note: tstzrange(start_time, end_time, '[)') creates a range that includes start_time
--       but excludes end_time. This allows back-to-back lessons without overlap.
--
-- WHERE deleted_at IS NULL ensures soft-deleted lessons don't affect the constraint
ALTER TABLE lessons
  ADD CONSTRAINT teacher_lessons_no_overlap
  EXCLUDE USING GIST (
    teacher_id WITH =,
    tstzrange(start_time, end_time, '[)') WITH &&
  )
  WHERE (deleted_at IS NULL);

-- Add comments for documentation
COMMENT ON CONSTRAINT teacher_lessons_no_overlap ON lessons IS
  'Prevents overlapping lessons for the same teacher. Uses GiST index with tstzrange to efficiently detect conflicts. Soft-deleted lessons are excluded from the constraint.';

COMMENT ON INDEX idx_teacher_lesson_overlap_gist IS
  'GiST index supporting teacher lesson overlap detection. Indexes teacher_id and time ranges (tstzrange) for efficient range queries.';

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration, run:
/*
BEGIN;

ALTER TABLE lessons DROP CONSTRAINT IF EXISTS teacher_lessons_no_overlap;
DROP INDEX IF EXISTS idx_teacher_lesson_overlap_gist;

COMMIT;
*/

COMMIT;
