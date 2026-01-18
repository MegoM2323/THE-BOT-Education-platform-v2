-- 012_add_color_subject_to_lessons.sql
-- Add color and subject fields to lessons table

-- Add color column with default value
ALTER TABLE lessons
ADD COLUMN color VARCHAR(7) NOT NULL DEFAULT '#3B82F6',
ADD CONSTRAINT lessons_color_format CHECK (color ~ '^#[0-9A-Fa-f]{6}$');

-- Add subject column (nullable)
ALTER TABLE lessons
ADD COLUMN subject VARCHAR(200);

-- Add comments
COMMENT ON COLUMN lessons.color IS 'Lesson color in hex format (#RRGGBB) for UI display';
COMMENT ON COLUMN lessons.subject IS 'Optional lesson subject/topic (e.g., "Math", "English Grammar")';

-- Create index on subject for filtering
CREATE INDEX idx_lessons_subject ON lessons(subject) WHERE subject IS NOT NULL AND deleted_at IS NULL;
