-- Migration 027: Fix Payment Status CHECK Constraint
-- Issue: The 'failed' status is used in code (models/payment.go) but missing from database CHECK constraint
-- This migration updates the payments table to include all valid statuses: pending, succeeded, failed, cancelled, refunded

-- Drop the old CHECK constraint that only has: pending, succeeded, cancelled
ALTER TABLE payments
DROP CONSTRAINT IF EXISTS payments_status_check;

-- Add the new CHECK constraint with all valid statuses
ALTER TABLE payments
ADD CONSTRAINT payments_status_check
CHECK (status IN ('pending', 'succeeded', 'failed', 'cancelled', 'refunded'));

-- Add comment documenting the valid statuses
COMMENT ON CONSTRAINT payments_status_check ON payments
IS 'Ensures status is one of: pending (awaiting payment), succeeded (payment successful), failed (payment error), cancelled (user cancelled), refunded (money returned).';

-- Add comment on the status column to document all valid values
COMMENT ON COLUMN payments.status
IS 'Payment status: pending (awaiting payment), succeeded (successfully paid), failed (payment error occurred), cancelled (user cancelled payment), refunded (refunded to user)';
