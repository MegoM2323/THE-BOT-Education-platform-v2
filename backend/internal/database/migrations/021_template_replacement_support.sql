-- Migration 021: Support for template week replacement feature
-- Добавляет поле updated_at и статус 'replaced' для template_applications

-- Добавляем updated_at колонку если её нет
ALTER TABLE template_applications
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Обновляем constraint для поддержки статуса 'replaced'
-- Сначала удаляем старый constraint
ALTER TABLE template_applications
DROP CONSTRAINT IF EXISTS template_applications_valid_status;

-- Добавляем новый constraint с поддержкой 'replaced'
ALTER TABLE template_applications
ADD CONSTRAINT template_applications_valid_status
CHECK (status IN ('applied', 'rolled_back', 'replaced'));

-- Обновляем rollback check constraint
-- Status 'replaced' работает так же как 'rolled_back' - имеет rolled_back_at timestamp
ALTER TABLE template_applications
DROP CONSTRAINT IF EXISTS template_applications_rollback_check;

ALTER TABLE template_applications
ADD CONSTRAINT template_applications_rollback_check
CHECK (
    (status = 'applied' AND rolled_back_at IS NULL) OR
    (status IN ('rolled_back', 'replaced') AND rolled_back_at IS NOT NULL)
);

-- Создаём индекс для быстрого поиска replaced applications
CREATE INDEX IF NOT EXISTS idx_template_applications_replaced
ON template_applications(id)
WHERE status = 'replaced';

-- Комментарии для документации
COMMENT ON COLUMN template_applications.updated_at IS 'Timestamp when application record was last updated (status change, etc.)';
COMMENT ON COLUMN template_applications.status IS 'Application status: applied (active), rolled_back (manually reverted), replaced (replaced by another template)';

-- Создаём trigger для автообновления updated_at
CREATE OR REPLACE FUNCTION update_template_applications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS template_applications_updated_at_trigger ON template_applications;
CREATE TRIGGER template_applications_updated_at_trigger
    BEFORE UPDATE ON template_applications
    FOR EACH ROW
    EXECUTE FUNCTION update_template_applications_updated_at();

-- Rollback script (для отката миграции)
-- DROP TRIGGER IF EXISTS template_applications_updated_at_trigger ON template_applications;
-- DROP FUNCTION IF EXISTS update_template_applications_updated_at();
-- DROP INDEX IF EXISTS idx_template_applications_replaced;
-- ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS template_applications_rollback_check;
-- ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS template_applications_valid_status;
-- ALTER TABLE template_applications ADD CONSTRAINT template_applications_valid_status CHECK (status IN ('applied', 'rolled_back'));
-- ALTER TABLE template_applications ADD CONSTRAINT template_applications_rollback_check CHECK ((status = 'rolled_back' AND rolled_back_at IS NOT NULL) OR (status = 'applied' AND rolled_back_at IS NULL));
-- ALTER TABLE template_applications DROP COLUMN IF EXISTS updated_at;
