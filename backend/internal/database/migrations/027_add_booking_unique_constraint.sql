-- 027_add_booking_unique_constraint.sql
-- Add UNIQUE constraint on bookings (student_id, lesson_id) to prevent double booking
-- Purpose: Prevent students from booking the same lesson multiple times
-- Date: 2025-12-29

-- Step 1: Identify and remove duplicate bookings (keep the most recent one)
-- This query finds duplicates grouped by (student_id, lesson_id) and marks older ones for deletion
DELETE FROM bookings
WHERE id IN (
    SELECT id FROM (
        SELECT
            id,
            ROW_NUMBER() OVER (PARTITION BY student_id, lesson_id ORDER BY created_at DESC) as rn
        FROM bookings
        WHERE status = 'active'
    ) numbered
    WHERE rn > 1
);

-- Step 2: Create partial UNIQUE index on (student_id, lesson_id) for active bookings
-- PostgreSQL doesn't support WHERE clause in UNIQUE constraints via ALTER TABLE,
-- so we use CREATE UNIQUE INDEX with WHERE clause instead
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_student_lesson_unique
ON bookings(student_id, lesson_id) WHERE status = 'active';

-- Note: The index name serves as the constraint enforcement mechanism
