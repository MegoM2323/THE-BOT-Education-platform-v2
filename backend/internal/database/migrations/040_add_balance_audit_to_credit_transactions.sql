-- 040_add_balance_audit_to_credit_transactions.sql
-- Add balance_before and balance_after columns to credit_transactions table for audit trail

-- Add balance_before column
ALTER TABLE credit_transactions
ADD COLUMN IF NOT EXISTS balance_before INTEGER;

-- Add balance_after column
ALTER TABLE credit_transactions
ADD COLUMN IF NOT EXISTS balance_after INTEGER;

-- Add comments for documentation
COMMENT ON COLUMN credit_transactions.balance_before IS 'User credit balance before this transaction';
COMMENT ON COLUMN credit_transactions.balance_after IS 'User credit balance after this transaction';
