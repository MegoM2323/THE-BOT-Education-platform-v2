-- 025_add_homework_text_to_lessons.sql
-- Add homework_text field to lessons table

-- Add homework_text column to lessons table
ALTER TABLE lessons
ADD COLUMN IF NOT EXISTS homework_text TEXT;

-- Add comment for the column
COMMENT ON COLUMN lessons.homework_text IS 'Text description/instructions for homework assigned to this lesson';
