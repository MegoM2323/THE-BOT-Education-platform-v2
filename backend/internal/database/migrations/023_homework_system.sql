-- 023_homework_system.sql
-- Создание системы домашних заданий для уроков

-- Таблица lesson_homework: файлы домашних заданий, привязанные к урокам
CREATE TABLE IF NOT EXISTS lesson_homework (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Валидация размера файла: > 0 и <= 10MB
    CONSTRAINT homework_file_size_valid CHECK (
        file_size > 0 AND
        file_size <= 10485760
    ),

    -- Валидация имени файла: не пустое
    CONSTRAINT homework_file_name_not_empty CHECK (
        char_length(file_name) > 0 AND
        char_length(file_name) <= 255
    ),

    -- Валидация MIME типа: не пустой
    CONSTRAINT homework_mime_type_not_empty CHECK (
        char_length(mime_type) > 0
    )
);

-- Индексы для быстрого поиска домашних заданий
CREATE INDEX idx_lesson_homework_lesson_id ON lesson_homework(lesson_id);
CREATE INDEX idx_lesson_homework_created_by ON lesson_homework(created_by);
CREATE INDEX idx_lesson_homework_created_at ON lesson_homework(created_at DESC);

-- Комментарии к таблице и полям
COMMENT ON TABLE lesson_homework IS 'Файлы домашних заданий, прикрепленные к урокам';
COMMENT ON COLUMN lesson_homework.id IS 'Уникальный идентификатор домашнего задания';
COMMENT ON COLUMN lesson_homework.lesson_id IS 'ID урока, к которому прикреплено домашнее задание';
COMMENT ON COLUMN lesson_homework.file_name IS 'Оригинальное имя файла';
COMMENT ON COLUMN lesson_homework.file_path IS 'Путь к файлу на сервере (UUID-based)';
COMMENT ON COLUMN lesson_homework.file_size IS 'Размер файла в байтах (максимум 10MB)';
COMMENT ON COLUMN lesson_homework.mime_type IS 'MIME тип файла (например, application/pdf)';
COMMENT ON COLUMN lesson_homework.created_by IS 'ID пользователя (teacher/admin), загрузившего файл';
COMMENT ON COLUMN lesson_homework.created_at IS 'Время загрузки файла';
