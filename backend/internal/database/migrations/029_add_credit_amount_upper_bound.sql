-- Migration: Add credit amount upper bound constraint
-- Purpose: Enforce maximum credit amount of 100 per transaction
-- Date: 2025-12-29

-- Add CHECK constraint to enforce upper bound on credit transaction amounts
ALTER TABLE credit_transactions
ADD CONSTRAINT check_credit_amount_range
CHECK (amount >= -100 AND amount <= 100);

-- Add comment to document the constraint
COMMENT ON CONSTRAINT check_credit_amount_range ON credit_transactions IS 'Ensures credit transaction amount is within valid range (-100 to 100), preventing abuse';
