-- 026_add_booking_fk_constraint.sql
-- Add Foreign Key constraint on credit_transactions.booking_id
-- Purpose: Ensure referential integrity between credit transactions and bookings
-- Date: 2025-12-29

-- Step 1: Set orphaned booking_ids to NULL (records where booking_id doesn't exist in bookings)
UPDATE credit_transactions
SET booking_id = NULL
WHERE booking_id IS NOT NULL
  AND booking_id NOT IN (SELECT id FROM bookings);

-- Step 2: Add Foreign Key constraint with ON DELETE SET NULL
-- This ensures that if a booking is deleted, the transaction's booking_id is set to NULL
ALTER TABLE credit_transactions
ADD CONSTRAINT fk_credit_transactions_booking_id
  FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE SET NULL;

-- Step 3: Add index for foreign key performance
CREATE INDEX IF NOT EXISTS idx_credit_transactions_booking_id
ON credit_transactions(booking_id);

-- Step 4: Add comment to document the constraint
COMMENT ON CONSTRAINT fk_credit_transactions_booking_id ON credit_transactions
IS 'Foreign key constraint ensuring transaction references valid booking with soft delete support via SET NULL';

COMMENT ON INDEX idx_credit_transactions_booking_id
IS 'Index to optimize queries filtering credit transactions by booking_id';
