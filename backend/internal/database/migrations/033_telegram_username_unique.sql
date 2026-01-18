-- Migration 033: Add UNIQUE Constraint on telegram_username
--
-- Purpose: Ensure that each user can have at most one telegram_username
-- This prevents issues with Telegram login and notification targeting
--
-- Approach:
-- 1. Resolve existing duplicates by keeping only the most recently updated user
--    for each telegram_username, setting other duplicates to NULL
-- 2. Add UNIQUE constraint on telegram_username (allows multiple NULLs)
-- 3. Replace the existing index with a partial unique index
--
-- Why partial index (WHERE telegram_username IS NOT NULL)?
--  - PostgreSQL's UNIQUE constraint allows multiple NULLs
--  - This matches the requirement: "NULL usernames allowed"
--  - Users who haven't linked Telegram (NULL telegram_username) don't violate uniqueness

BEGIN;

-- Step 1: Identify and resolve duplicates
-- For each telegram_username that appears more than once, keep only the most recent
-- (by updated_at), set others to NULL
WITH duplicate_check AS (
    SELECT
        id,
        telegram_username,
        ROW_NUMBER() OVER (
            PARTITION BY telegram_username
            ORDER BY updated_at DESC NULLS LAST
        ) as rn
    FROM users
    WHERE telegram_username IS NOT NULL
        AND deleted_at IS NULL
)
UPDATE users
SET telegram_username = NULL, updated_at = CURRENT_TIMESTAMP
FROM duplicate_check
WHERE users.id = duplicate_check.id
    AND duplicate_check.rn > 1;

-- Step 2: Drop the old index (if it exists)
DROP INDEX IF EXISTS idx_users_telegram_username;

-- Step 3: Create partial UNIQUE index for telegram_username
-- PostgreSQL doesn't support WHERE clause in UNIQUE constraints via ALTER TABLE,
-- so we use CREATE UNIQUE INDEX with WHERE clause instead
CREATE UNIQUE INDEX users_telegram_username_unique
ON users(telegram_username)
WHERE telegram_username IS NOT NULL;

-- Step 4: Create a regular index for performance on lookups
-- (useful for queries like "SELECT * FROM users WHERE telegram_username = 'username'")
CREATE INDEX idx_users_telegram_username
ON users(telegram_username)
WHERE telegram_username IS NOT NULL;

COMMIT;

-- ============================================================================
-- VERIFICATION QUERIES (for testing)
-- ============================================================================
-- After migration, run these to verify:
/*
-- Check for any remaining duplicates (should return 0 rows)
SELECT telegram_username, COUNT(*) as count
FROM users
WHERE telegram_username IS NOT NULL AND deleted_at IS NULL
GROUP BY telegram_username
HAVING COUNT(*) > 1;

-- Check constraint exists
SELECT constraint_name, constraint_type
FROM information_schema.table_constraints
WHERE table_name = 'users' AND constraint_name = 'users_telegram_username_unique';

-- Check index exists
SELECT indexname FROM pg_indexes
WHERE tablename = 'users' AND indexname = 'idx_users_telegram_username';

-- Count users with NULL telegram_username (should allow multiple)
SELECT COUNT(*) as users_without_telegram
FROM users
WHERE telegram_username IS NULL AND deleted_at IS NULL;
*/

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration, run:
/*
BEGIN;

-- Drop the unique constraint
ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_telegram_username_unique;

-- Drop the unique index
DROP INDEX IF EXISTS idx_users_telegram_username;

-- Recreate the old non-unique index (from migration 008)
CREATE INDEX IF NOT EXISTS idx_users_telegram_username ON users(telegram_username)
WHERE telegram_username IS NOT NULL;

COMMIT;
*/
