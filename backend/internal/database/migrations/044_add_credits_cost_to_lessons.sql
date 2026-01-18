-- 044_add_credits_cost_to_lessons.sql
-- Add credits_cost column to lessons table

-- Add credits_cost column with default value 1 (for backward compatibility)
ALTER TABLE lessons
ADD COLUMN credits_cost INTEGER NOT NULL DEFAULT 1 CHECK (credits_cost > 0);

-- Add comment
COMMENT ON COLUMN lessons.credits_cost IS 'Number of credits required to book this lesson';

