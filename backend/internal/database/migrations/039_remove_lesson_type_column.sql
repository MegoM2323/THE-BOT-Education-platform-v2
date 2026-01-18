-- Migration 039: Remove lesson_type column from lessons table
-- Since lesson_type was removed from the Lesson model and is no longer stored in the DB,
-- we need to drop the column entirely to align with the code.

BEGIN;

-- Drop the index on lesson_type
DROP INDEX IF EXISTS idx_lessons_lesson_type;

-- Remove the lesson_type column
ALTER TABLE lessons DROP COLUMN lesson_type;

-- Update the CHECK constraint if needed (already handles max_students)

COMMIT;
