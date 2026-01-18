-- 022_user_payment_enabled.sql
-- Purpose: Add payment_enabled flag to users table for admin control
-- Created: 2025-12-08
--
-- This migration adds:
-- 1. payment_enabled column (BOOLEAN NOT NULL DEFAULT true)
-- 2. Partial index on payment_enabled for students (most frequent filter)
-- 3. Column comment for documentation
--
-- All existing users will have payment_enabled = true by default

BEGIN;

-- ============================================================================
-- ADD COLUMN
-- ============================================================================

-- Добавляем payment_enabled для контроля доступа к оплате
-- NOT NULL с DEFAULT true гарантирует что все существующие пользователи получат true
-- Новые пользователи также будут иметь payment_enabled = true по умолчанию
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
          AND column_name = 'payment_enabled'
    ) THEN
        ALTER TABLE users
        ADD COLUMN payment_enabled BOOLEAN NOT NULL DEFAULT true;
    END IF;
END $$;

-- ============================================================================
-- ADD COMMENT
-- ============================================================================

-- Добавляем описание колонки для документации
COMMENT ON COLUMN users.payment_enabled IS
'Флаг доступа к оплате. FALSE = пользователь не может покупать кредиты через YooKassa. Контролируется администратором.';

-- ============================================================================
-- CREATE INDEX
-- ============================================================================

-- Partial индекс только для студентов с payment_enabled = true
-- Оптимизирует запросы вида: WHERE role = 'student' AND payment_enabled = true
-- Студенты - основная группа пользователей которые платят
CREATE INDEX IF NOT EXISTS idx_users_payment_enabled
ON users(payment_enabled)
WHERE role = 'student' AND deleted_at IS NULL;

COMMIT;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- BEGIN;
-- DROP INDEX IF EXISTS idx_users_payment_enabled;
-- ALTER TABLE users DROP COLUMN IF EXISTS payment_enabled;
-- COMMIT;
