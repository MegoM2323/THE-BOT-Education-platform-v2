-- 047_allow_zero_credits_cost.sql
-- Purpose: Allow credits_cost to be 0 (free lessons)
-- Created: 2025-01-XX
--
-- This migration changes CHECK constraints to allow credits_cost = 0
-- for both lessons and template_lessons tables

BEGIN;

-- Drop existing CHECK constraints
ALTER TABLE lessons
DROP CONSTRAINT IF EXISTS lessons_credits_cost_check;

ALTER TABLE template_lessons
DROP CONSTRAINT IF EXISTS chk_template_lessons_credits_cost;

-- Add new CHECK constraints allowing 0
ALTER TABLE lessons
ADD CONSTRAINT lessons_credits_cost_check CHECK (credits_cost >= 0);

ALTER TABLE template_lessons
ADD CONSTRAINT chk_template_lessons_credits_cost CHECK (credits_cost >= 0);

COMMIT;



