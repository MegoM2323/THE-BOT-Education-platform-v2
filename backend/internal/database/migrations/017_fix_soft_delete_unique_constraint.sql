-- Migration: Fix soft delete unique constraint on users.email
-- Problem: users_email_key constraint prevents reusing email after soft delete
-- Solution: Replace with partial unique constraint that excludes deleted_at IS NOT NULL

-- Drop old UNIQUE constraint that doesn't account for soft delete
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

-- Create new partial UNIQUE constraint that only applies to active users
-- This allows email reuse after soft delete (deleted_at IS NOT NULL)
CREATE UNIQUE INDEX users_email_unique_active ON users (email) WHERE deleted_at IS NULL;

-- Note: We use CREATE UNIQUE INDEX instead of ADD CONSTRAINT because
-- PostgreSQL partial unique constraints must be created as indexes
