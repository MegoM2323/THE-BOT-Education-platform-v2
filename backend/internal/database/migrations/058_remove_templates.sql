-- +migrate Up
-- Migration 058: Remove template functionality

-- Удаляем столбцы из lessons (сначала удаляем внешние ключи)
ALTER TABLE lessons DROP COLUMN IF EXISTS applied_from_template;
ALTER TABLE lessons DROP COLUMN IF EXISTS template_application_id;

-- Удаляем таблицы в правильном порядке (с учетом foreign keys)
DROP TABLE IF EXISTS lesson_modifications CASCADE;
DROP TABLE IF EXISTS template_applications CASCADE;
DROP TABLE IF EXISTS template_lesson_students CASCADE;
DROP TABLE IF EXISTS template_lessons CASCADE;
DROP TABLE IF EXISTS lesson_templates CASCADE;

-- +migrate Down
-- Восстановление не поддерживается - backward compatibility невозможна
-- Для восстановления нужно использовать бэкап БД
