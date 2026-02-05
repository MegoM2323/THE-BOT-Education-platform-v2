-- +migrate Up
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS report_text TEXT;
COMMENT ON COLUMN lessons.report_text IS 'Текстовый отчет преподавателя о занятии';

-- +migrate Down
ALTER TABLE lessons DROP COLUMN IF EXISTS report_text;
