-- 048_remove_teacher_role.sql
-- Purpose: Remove 'teacher' role from users table CHECK constraint
-- Created: 2025-01-18
--
-- This migration updates the users role CHECK constraint to remove the 'teacher' role
-- Previously allowed roles: 'student', 'teacher', 'admin', 'methodologist'
-- New allowed roles: 'student', 'admin', 'methodologist'
-- Also removes the validate_teacher_role trigger and function that are no longer needed

BEGIN;

-- Drop trigger and function that validate teacher role
DROP TRIGGER IF EXISTS validate_teacher_role_before_insert ON template_lessons;
DROP FUNCTION IF EXISTS validate_teacher_role();

-- Update role constraint
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    -- Find and drop the existing role constraint
    SELECT con.conname INTO constraint_name
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    WHERE rel.relname = 'users'
      AND con.contype = 'c'
      AND pg_get_constraintdef(con.oid) LIKE '%role%';

    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE users DROP CONSTRAINT ' || quote_ident(constraint_name);
    END IF;

    -- Add new constraint without 'teacher' role
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint con
        JOIN pg_class rel ON rel.oid = con.conrelid
        WHERE rel.relname = 'users' AND con.conname = 'users_role_check'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_role_check
            CHECK (role IN ('student', 'admin', 'methodologist'));
    END IF;
END $$;

COMMIT;
