-- 013_add_color_subject_to_template_lessons.sql
-- Add color and subject fields to template_lessons table to support lesson styling

-- Add color column with default value
ALTER TABLE template_lessons
ADD COLUMN color VARCHAR(7) NOT NULL DEFAULT '#3B82F6',
ADD CONSTRAINT template_lessons_color_format CHECK (color ~ '^#[0-9A-Fa-f]{6}$');

-- Add subject column (nullable)
ALTER TABLE template_lessons
ADD COLUMN subject VARCHAR(200);

-- Add comments
COMMENT ON COLUMN template_lessons.color IS 'Lesson color in hex format (#RRGGBB) for UI display, transferred to created lessons';
COMMENT ON COLUMN template_lessons.subject IS 'Optional lesson subject/topic (e.g., "Math", "English Grammar"), transferred to created lessons';

-- Create index on subject for filtering
CREATE INDEX idx_template_lessons_subject ON template_lessons(subject) WHERE subject IS NOT NULL;
