-- 046_add_teacher_role.sql
-- Добавление роли teacher в CHECK constraint таблицы users

DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    -- Находим и удаляем старый constraint на role
    SELECT con.conname INTO constraint_name
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    WHERE rel.relname = 'users'
      AND con.contype = 'c'
      AND pg_get_constraintdef(con.oid) LIKE '%role%';

    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE users DROP CONSTRAINT ' || quote_ident(constraint_name);
    END IF;

    -- Добавляем новый constraint (проверяем что его нет)
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint con
        JOIN pg_class rel ON rel.oid = con.conrelid
        WHERE rel.relname = 'users' AND con.conname = 'users_role_check'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_role_check
            CHECK (role IN ('student', 'teacher', 'admin', 'teacher'));
    END IF;
END $$;
