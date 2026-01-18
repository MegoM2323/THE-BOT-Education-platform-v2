-- Migration: Add credit balance constraints
-- Purpose: Prevent negative credit balances
-- Date: 2025-11-27

-- First, update any existing negative balances to 0
UPDATE credits SET balance = 0 WHERE balance < 0;

-- Add CHECK constraint to prevent negative balances
ALTER TABLE credits
ADD CONSTRAINT check_balance_non_negative
CHECK (balance >= 0);

-- Add comment to document the constraint
COMMENT ON CONSTRAINT check_balance_non_negative ON credits IS 'Ensures credit balance cannot be negative';