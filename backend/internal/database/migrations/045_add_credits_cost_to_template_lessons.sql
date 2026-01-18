-- 045_add_credits_cost_to_template_lessons.sql
-- Purpose: Add credits_cost field to template_lessons table
-- Created: 2025-01-07
--
-- This migration adds credits_cost column to template_lessons table
-- to support setting lesson price when creating lessons from templates

BEGIN;

-- Add credits_cost column if it doesn't exist (all existing records will get DEFAULT 1)
ALTER TABLE template_lessons
ADD COLUMN IF NOT EXISTS credits_cost INTEGER NOT NULL DEFAULT 1;

-- Update any NULL values to 1 (for backward compatibility)
UPDATE template_lessons SET credits_cost = 1 WHERE credits_cost IS NULL;

-- Add CHECK constraint for positive values (will be ignored if already exists)
ALTER TABLE template_lessons
ADD CONSTRAINT chk_template_lessons_credits_cost CHECK (credits_cost > 0);

-- Add comment for documentation
COMMENT ON COLUMN template_lessons.credits_cost IS 'Cost in credits for lessons created from this template entry (default: 1)';

COMMIT;

