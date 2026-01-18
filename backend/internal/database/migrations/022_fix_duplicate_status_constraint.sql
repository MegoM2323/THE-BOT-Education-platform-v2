-- Migration 022: Удаление дублирующегося status constraint
-- Исправляет проблему когда старый constraint template_applications_status_check
-- конфликтует с новым template_applications_valid_status

-- Удаляем старый constraint если он существует
-- (мог остаться после применения миграции 021 на уже существующей БД)
ALTER TABLE template_applications
DROP CONSTRAINT IF EXISTS template_applications_status_check;

-- Проверяем что новый constraint существует
-- Если нет - создаём его (на случай если эта миграция запущена до 021)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'template_applications_valid_status'
        AND conrelid = 'template_applications'::regclass
    ) THEN
        ALTER TABLE template_applications
        ADD CONSTRAINT template_applications_valid_status
        CHECK (status IN ('applied', 'rolled_back', 'replaced'));
    END IF;
END $$;

-- Проверяем что rollback_check constraint существует
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'template_applications_rollback_check'
        AND conrelid = 'template_applications'::regclass
    ) THEN
        ALTER TABLE template_applications
        ADD CONSTRAINT template_applications_rollback_check
        CHECK (
            (status = 'applied' AND rolled_back_at IS NULL) OR
            (status IN ('rolled_back', 'replaced') AND rolled_back_at IS NOT NULL)
        );
    END IF;
END $$;

-- Комментарий для документации
COMMENT ON CONSTRAINT template_applications_valid_status ON template_applications
IS 'Ensures status is one of: applied, rolled_back, replaced. Replaces old template_applications_status_check.';
