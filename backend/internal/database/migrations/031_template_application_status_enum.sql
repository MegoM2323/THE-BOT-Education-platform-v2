-- Migration 031: Add TemplateApplication Status Enum Constraint
--
-- Purpose: Add CHECK constraint to enforce valid status values for template_applications table
--
-- Valid status values (database):
--   - applied: Template has been applied to a week (default state)
--   - replaced: Application has been replaced by another template application
--   - rolled_back: Application has been rolled back (lessons deleted)
--
-- Note: "preview" is also used in the codebase but only in API responses (dry-run mode),
-- it is NOT stored in the database, so it is excluded from the database constraint.
--
-- The original migration (011) had a constraint that only allowed 'applied' and 'rolled_back',
-- but the code also uses 'replaced'. This migration updates the constraint to allow all three values.

BEGIN;

-- Drop the old constraint if it exists
ALTER TABLE template_applications
DROP CONSTRAINT IF EXISTS template_applications_valid_status;

-- Drop the old rollback check constraint if it exists
ALTER TABLE template_applications
DROP CONSTRAINT IF EXISTS template_applications_rollback_check;

-- Add new CHECK constraint for valid status values
ALTER TABLE template_applications
ADD CONSTRAINT template_applications_status_enum
CHECK (status IN ('applied', 'replaced', 'rolled_back'));

-- Add constraint documentation
COMMENT ON CONSTRAINT template_applications_status_enum ON template_applications IS
  'Enforces that status can only be one of: applied, replaced, or rolled_back';

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration, run:
/*
BEGIN;

ALTER TABLE template_applications
DROP CONSTRAINT IF EXISTS template_applications_status_enum;

-- Restore original constraints if needed:
-- ALTER TABLE template_applications
-- ADD CONSTRAINT template_applications_valid_status
-- CHECK (status IN ('applied', 'rolled_back'));
--
-- ALTER TABLE template_applications
-- ADD CONSTRAINT template_applications_rollback_check
-- CHECK ((status = 'rolled_back' AND rolled_back_at IS NOT NULL) OR (status = 'applied' AND rolled_back_at IS NULL));

COMMIT;
*/
