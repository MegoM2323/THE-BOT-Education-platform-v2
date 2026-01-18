-- 024_telegram_users_unique_constraint.sql
-- Добавляет UNIQUE constraint на user_id в telegram_users (если отсутствует)

-- Проверяем и добавляем constraint если его нет
DO $$
BEGIN
    -- Проверяем существование constraint
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'telegram_users_user_id_key'
        OR conname = 'telegram_users_user_id_unique'
    ) THEN
        -- Добавляем UNIQUE constraint
        ALTER TABLE telegram_users ADD CONSTRAINT telegram_users_user_id_unique UNIQUE (user_id);
        RAISE NOTICE 'UNIQUE constraint telegram_users_user_id_unique добавлен';
    ELSE
        RAISE NOTICE 'UNIQUE constraint на user_id уже существует';
    END IF;
END $$;
