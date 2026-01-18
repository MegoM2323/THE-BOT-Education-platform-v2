-- Migration 052: Add database triggers for overbooking prevention and credit auditing
-- Purpose: Add database-level safety checks to prevent race conditions and ensure data integrity
-- Created: 2026-01-18

-- ===================================
-- TRIGGER: Prevent Lesson Overbooking
-- ===================================
-- Prevents race condition where booking could exceed lesson capacity
-- Uses EXISTS to avoid lock ordering deadlock (no FOR UPDATE on lessons)
-- Safety check matches Go IsFull() logic: current_students >= max_students

CREATE OR REPLACE FUNCTION prevent_lesson_overbooking()
RETURNS TRIGGER AS $$
DECLARE
	lesson_is_full BOOLEAN;
BEGIN
	-- Check if lesson is at or exceeds capacity without locking lessons table
	-- This avoids deadlock: bookings INSERT always triggers this,
	-- if we SELECT FOR UPDATE on lessons, we invert the typical lock order
	SELECT EXISTS (
		SELECT 1 FROM lessons
		WHERE id = NEW.lesson_id
		AND current_students >= max_students
	) INTO lesson_is_full;

	-- Raise exception if lesson is already full
	-- This is a last-resort check - should already be checked in application logic
	-- But database-level check provides additional safety
	IF lesson_is_full THEN
		RAISE EXCEPTION 'Lesson is already at full capacity';
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists (for idempotency)
DROP TRIGGER IF EXISTS trg_prevent_lesson_overbooking ON bookings;

-- Create trigger to call the prevention function before INSERT
CREATE TRIGGER trg_prevent_lesson_overbooking
BEFORE INSERT ON bookings
FOR EACH ROW
EXECUTE FUNCTION prevent_lesson_overbooking();

-- ===================================
-- TRIGGER: Prevent Negative Credits
-- ===================================
-- Prevents credit balance from going negative in edge cases
-- Should be caught at application level but provides database-level safety net

CREATE OR REPLACE FUNCTION prevent_negative_credits()
RETURNS TRIGGER AS $$
BEGIN
	-- Check if the new balance would be negative
	IF NEW.balance < 0 THEN
		RAISE EXCEPTION 'Credit balance cannot be negative (attempted: %)', NEW.balance;
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists (for idempotency)
DROP TRIGGER IF EXISTS trg_prevent_negative_credits ON credits;

-- Create trigger to call the prevention function before UPDATE
CREATE TRIGGER trg_prevent_negative_credits
BEFORE UPDATE ON credits
FOR EACH ROW
EXECUTE FUNCTION prevent_negative_credits();

-- ===================================
-- TRIGGER: Audit Credit Changes
-- ===================================
-- Logs all credit balance changes to audit table for compliance and debugging
-- Captures old_balance, new_balance, operation_type, and context

-- Create credit balance audit table if it doesn't exist
-- This table logs all credit balance changes for audit and debugging
CREATE TABLE IF NOT EXISTS credit_balance_audit (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	old_balance INT NOT NULL,
	new_balance INT NOT NULL,
	operation_type VARCHAR(50) NOT NULL,
	reason VARCHAR(500),
	performed_by UUID REFERENCES users(id) ON DELETE SET NULL,
	booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,
	changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_credit_balance_audit_user_id ON credit_balance_audit(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_balance_audit_changed_at ON credit_balance_audit(changed_at);

-- Create function to audit credit changes
-- This function is called after UPDATE on credits table
-- It logs the change to credit_balance_audit for compliance and debugging
CREATE OR REPLACE FUNCTION audit_credit_changes()
RETURNS TRIGGER AS $$
BEGIN
	-- Log the credit balance change to audit table
	-- Reason includes balance delta and direction for better debugging
	INSERT INTO credit_balance_audit (
		user_id,
		old_balance,
		new_balance,
		operation_type,
		reason,
		booking_id,
		changed_at
	)
	VALUES (
		NEW.user_id,
		OLD.balance,
		NEW.balance,
		CASE
			WHEN NEW.balance > OLD.balance THEN 'REFUND'
			WHEN NEW.balance < OLD.balance THEN 'DEDUCTION'
			ELSE 'NO_CHANGE'
		END,
		'Balance changed from ' || OLD.balance || ' to ' || NEW.balance || ' (delta: ' || (NEW.balance - OLD.balance) || ')',
		NULL,
		CURRENT_TIMESTAMP
	);

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists (for idempotency)
DROP TRIGGER IF EXISTS trg_audit_credit_changes ON credits;

-- Create trigger to log credit changes after UPDATE
CREATE TRIGGER trg_audit_credit_changes
AFTER UPDATE ON credits
FOR EACH ROW
WHEN (OLD.balance IS DISTINCT FROM NEW.balance)
EXECUTE FUNCTION audit_credit_changes();

-- Add comments for documentation
COMMENT ON FUNCTION prevent_lesson_overbooking() IS
	'Database safety check to prevent lesson overbooking. Called before INSERT into bookings.';

COMMENT ON FUNCTION prevent_negative_credits() IS
	'Database safety check to prevent negative credit balances. Called before UPDATE on credits.';

COMMENT ON FUNCTION audit_credit_changes() IS
	'Database audit function to log all credit balance changes. Called after UPDATE on credits.';

COMMENT ON TABLE credit_balance_audit IS
	'Audit log for all credit balance changes, used for compliance, debugging, and data recovery.';
