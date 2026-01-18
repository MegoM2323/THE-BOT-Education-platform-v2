-- 006_add_missing_indexes.sql
-- Add missing indexes for improved query performance

-- Index on bookings(booked_at) for sorting booking history
CREATE INDEX idx_bookings_booked_at ON bookings(booked_at DESC);

-- Index on credit_transactions(performed_by) for admin operations lookup
CREATE INDEX idx_credit_transactions_performed_by ON credit_transactions(performed_by) WHERE performed_by IS NOT NULL;

-- Composite index on lessons(schedule_date, teacher_id) for fast schedule lookups
-- Using start_time as schedule_date
CREATE INDEX idx_lessons_schedule_teacher ON lessons(start_time, teacher_id) WHERE deleted_at IS NULL;

-- Comments for clarity
COMMENT ON INDEX idx_bookings_booked_at IS 'Speeds up booking history queries sorted by booking time';
COMMENT ON INDEX idx_credit_transactions_performed_by IS 'Improves performance when searching for admin operations';
COMMENT ON INDEX idx_lessons_schedule_teacher IS 'Optimizes teacher schedule queries by date';
