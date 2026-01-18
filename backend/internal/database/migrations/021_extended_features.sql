-- 021_extended_features.sql
-- Расширенные функции платформы: домашние задания, рассылки, файлы, отмененные бронирования
-- Создание: 2025-12-08
-- Зависимости: lessons, bookings, users (из 001_init_schema.sql)

-- =============================================================================
-- ТАБЛИЦА 1: Домашние задания к урокам (файловые вложения)
-- =============================================================================

CREATE TABLE IF NOT EXISTS lesson_homework (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0 AND file_size <= 10485760), -- Максимум 10MB
    mime_type VARCHAR(100) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Комментарии к таблице и столбцам
COMMENT ON TABLE lesson_homework IS 'Файловые вложения домашних заданий к урокам';
COMMENT ON COLUMN lesson_homework.id IS 'Уникальный идентификатор вложения';
COMMENT ON COLUMN lesson_homework.lesson_id IS 'Ссылка на урок, к которому прикреплено задание';
COMMENT ON COLUMN lesson_homework.file_name IS 'Оригинальное название файла';
COMMENT ON COLUMN lesson_homework.file_path IS 'Путь к файлу на сервере (UUID-based для безопасности)';
COMMENT ON COLUMN lesson_homework.file_size IS 'Размер файла в байтах (максимум 10MB)';
COMMENT ON COLUMN lesson_homework.mime_type IS 'MIME-тип файла (image/png, application/pdf, и т.д.)';
COMMENT ON COLUMN lesson_homework.created_by IS 'Пользователь (обычно преподаватель), загрузивший файл';
COMMENT ON COLUMN lesson_homework.created_at IS 'Дата и время загрузки файла';

-- =============================================================================
-- ТАБЛИЦА 2: Рассылки сообщений для урока (broadcast к студентам)
-- =============================================================================

CREATE TABLE IF NOT EXISTS lesson_broadcasts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    message TEXT NOT NULL CHECK (char_length(message) > 0 AND char_length(message) <= 4096),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sending', 'completed', 'failed')),
    sent_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Комментарии к таблице и столбцам
COMMENT ON TABLE lesson_broadcasts IS 'Рассылки сообщений всем студентам конкретного урока';
COMMENT ON COLUMN lesson_broadcasts.id IS 'Уникальный идентификатор рассылки';
COMMENT ON COLUMN lesson_broadcasts.lesson_id IS 'Урок, студентам которого отправляется сообщение';
COMMENT ON COLUMN lesson_broadcasts.sender_id IS 'Отправитель (обычно преподаватель урока)';
COMMENT ON COLUMN lesson_broadcasts.message IS 'Текст сообщения (максимум 4096 символов)';
COMMENT ON COLUMN lesson_broadcasts.status IS 'pending: ожидает отправки, sending: в процессе, completed: завершено, failed: ошибка';
COMMENT ON COLUMN lesson_broadcasts.sent_count IS 'Количество успешно доставленных сообщений';
COMMENT ON COLUMN lesson_broadcasts.failed_count IS 'Количество неудачных доставок';
COMMENT ON COLUMN lesson_broadcasts.created_at IS 'Дата и время создания рассылки';
COMMENT ON COLUMN lesson_broadcasts.completed_at IS 'Дата и время завершения отправки';

-- =============================================================================
-- ТАБЛИЦА 3: Файловые вложения к рассылкам
-- =============================================================================

CREATE TABLE IF NOT EXISTS broadcast_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    broadcast_id UUID NOT NULL REFERENCES lesson_broadcasts(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0 AND file_size <= 10485760), -- Максимум 10MB
    mime_type VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Комментарии к таблице и столбцам
COMMENT ON TABLE broadcast_files IS 'Файловые вложения к рассылкам сообщений';
COMMENT ON COLUMN broadcast_files.id IS 'Уникальный идентификатор файла';
COMMENT ON COLUMN broadcast_files.broadcast_id IS 'Ссылка на рассылку, к которой прикреплен файл';
COMMENT ON COLUMN broadcast_files.file_name IS 'Оригинальное название файла';
COMMENT ON COLUMN broadcast_files.file_path IS 'Путь к файлу на сервере (UUID-based)';
COMMENT ON COLUMN broadcast_files.file_size IS 'Размер файла в байтах (максимум 10MB)';
COMMENT ON COLUMN broadcast_files.mime_type IS 'MIME-тип файла';
COMMENT ON COLUMN broadcast_files.uploaded_at IS 'Дата и время загрузки файла';

-- =============================================================================
-- ТАБЛИЦА 4: Отмененные бронирования (предотвращение повторного бронирования)
-- =============================================================================

CREATE TABLE IF NOT EXISTS cancelled_bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    cancelled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id, lesson_id) -- Студент не может дважды отменить одно и то же занятие
);

-- Комментарии к таблице и столбцам
COMMENT ON TABLE cancelled_bookings IS 'История отмененных бронирований для предотвращения повторного бронирования';
COMMENT ON COLUMN cancelled_bookings.id IS 'Уникальный идентификатор записи об отмене';
COMMENT ON COLUMN cancelled_bookings.booking_id IS 'Ссылка на отмененное бронирование';
COMMENT ON COLUMN cancelled_bookings.student_id IS 'Студент, отменивший бронирование';
COMMENT ON COLUMN cancelled_bookings.lesson_id IS 'Урок, бронирование которого было отменено';
COMMENT ON COLUMN cancelled_bookings.cancelled_at IS 'Дата и время отмены бронирования';

-- =============================================================================
-- ИНДЕКСЫ ДЛЯ ПРОИЗВОДИТЕЛЬНОСТИ
-- =============================================================================

-- Индексы для lesson_homework (частые запросы по lesson_id)
CREATE INDEX IF NOT EXISTS idx_lesson_homework_lesson
ON lesson_homework(lesson_id);

COMMENT ON INDEX idx_lesson_homework_lesson IS 'Быстрый поиск всех домашних заданий для конкретного урока';

-- Индексы для lesson_broadcasts (частые запросы по lesson_id, фильтрация по статусу)
CREATE INDEX IF NOT EXISTS idx_lesson_broadcasts_lesson
ON lesson_broadcasts(lesson_id);

CREATE INDEX IF NOT EXISTS idx_lesson_broadcasts_status
ON lesson_broadcasts(status, created_at DESC);

COMMENT ON INDEX idx_lesson_broadcasts_lesson IS 'Быстрый поиск всех рассылок для конкретного урока';
COMMENT ON INDEX idx_lesson_broadcasts_status IS 'Поиск рассылок по статусу (для фоновых задач обработки)';

-- Индексы для broadcast_files (частые запросы по broadcast_id)
CREATE INDEX IF NOT EXISTS idx_broadcast_files_broadcast
ON broadcast_files(broadcast_id);

COMMENT ON INDEX idx_broadcast_files_broadcast IS 'Быстрый поиск файлов конкретной рассылки';

-- Индексы для cancelled_bookings (частые проверки перед бронированием)
CREATE INDEX IF NOT EXISTS idx_cancelled_bookings_student_lesson
ON cancelled_bookings(student_id, lesson_id);

CREATE INDEX IF NOT EXISTS idx_cancelled_bookings_lesson
ON cancelled_bookings(lesson_id);

COMMENT ON INDEX idx_cancelled_bookings_student_lesson IS 'Быстрая проверка был ли студент уже отменен на урок (композитный индекс)';
COMMENT ON INDEX idx_cancelled_bookings_lesson IS 'Поиск всех отмен для конкретного урока';

-- =============================================================================
-- ДОБАВЛЕНИЕ ПОЛЯ payment_enabled В ТАБЛИЦУ users (если не существует)
-- =============================================================================

-- Добавляем поле payment_enabled для студентов (включено по умолчанию)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'payment_enabled'
    ) THEN
        ALTER TABLE users ADD COLUMN payment_enabled BOOLEAN DEFAULT TRUE;
        COMMENT ON COLUMN users.payment_enabled IS 'Может ли студент покупать кредиты (только для role=student)';
    END IF;
END $$;

-- Индекс для users (частые запросы студентов с включенными платежами)
CREATE INDEX IF NOT EXISTS idx_users_payment_enabled
ON users(payment_enabled)
WHERE role = 'student';

COMMENT ON INDEX idx_users_payment_enabled IS 'Частичный индекс для поиска студентов с активированными платежами';

-- =============================================================================
-- VERIFICATION QUERIES (Проверка успешности миграции)
-- =============================================================================

-- Раскомментируйте эти запросы после применения миграции для проверки:
--
-- SELECT table_name, table_type
-- FROM information_schema.tables
-- WHERE table_name IN ('lesson_homework', 'lesson_broadcasts', 'broadcast_files', 'cancelled_bookings');
--
-- SELECT indexname, tablename
-- FROM pg_indexes
-- WHERE tablename IN ('lesson_homework', 'lesson_broadcasts', 'broadcast_files', 'cancelled_bookings');
--
-- SELECT column_name, data_type, is_nullable
-- FROM information_schema.columns
-- WHERE table_name = 'users' AND column_name = 'payment_enabled';

-- =============================================================================
-- ROLLBACK INSTRUCTIONS (Откат миграции)
-- =============================================================================

-- ВНИМАНИЕ: Выполнение отката удалит ВСЕ данные из новых таблиц!
-- Используйте только если требуется полная отмена миграции.
--
-- Шаг 1: Удалить индексы
-- DROP INDEX IF EXISTS idx_users_payment_enabled;
-- DROP INDEX IF EXISTS idx_cancelled_bookings_lesson;
-- DROP INDEX IF EXISTS idx_cancelled_bookings_student_lesson;
-- DROP INDEX IF EXISTS idx_broadcast_files_broadcast;
-- DROP INDEX IF EXISTS idx_lesson_broadcasts_status;
-- DROP INDEX IF EXISTS idx_lesson_broadcasts_lesson;
-- DROP INDEX IF EXISTS idx_lesson_homework_lesson;
--
-- Шаг 2: Удалить таблицы (CASCADE удалит все зависимые данные)
-- DROP TABLE IF EXISTS cancelled_bookings CASCADE;
-- DROP TABLE IF EXISTS broadcast_files CASCADE;
-- DROP TABLE IF EXISTS lesson_broadcasts CASCADE;
-- DROP TABLE IF EXISTS lesson_homework CASCADE;
--
-- Шаг 3: Удалить поле из users
-- ALTER TABLE users DROP COLUMN IF EXISTS payment_enabled;
--
-- КОНЕЦ ОТКАТА
