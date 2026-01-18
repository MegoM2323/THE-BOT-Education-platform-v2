-- Migration 007: Добавление CHECK constraints для защиты от некорректных значений
--
-- Эта миграция добавляет ограничения на уровне БД для предотвращения:
-- 1. Отрицательных значений balance в credits
-- 2. Отрицательных значений current_students в lessons
-- 3. Превышения current_students над max_students
-- 4. Отрицательных значений max_students

-- =====================================================
-- ТАБЛИЦА: lessons
-- =====================================================

-- Ограничение: current_students должен быть >= 0 и <= max_students
ALTER TABLE lessons
ADD CONSTRAINT lessons_current_students_range
CHECK (current_students >= 0 AND current_students <= max_students);

-- Ограничение: max_students должен быть > 0
ALTER TABLE lessons
ADD CONSTRAINT lessons_max_students_positive
CHECK (max_students > 0);

-- =====================================================
-- ТАБЛИЦА: credits
-- =====================================================

-- Ограничение: balance не может быть отрицательным
ALTER TABLE credits
ADD CONSTRAINT credits_balance_non_negative
CHECK (balance >= 0);

-- =====================================================
-- ТАБЛИЦА: bookings (для будущего использования)
-- =====================================================
-- Примечание: В текущей версии cost не используется в таблице bookings,
-- но если будет добавлен, этот constraint защитит от отрицательных значений
--
-- ALTER TABLE bookings
-- ADD CONSTRAINT bookings_cost_non_negative
-- CHECK (cost >= 0);

-- =====================================================
-- ИНДЕКСЫ ДЛЯ ПРОИЗВОДИТЕЛЬНОСТИ CHECK CONSTRAINTS
-- =====================================================
-- Эти индексы не обязательны для CHECK constraints, но могут улучшить производительность
-- при частых операциях с этими полями

-- Индекс для быстрой проверки заполненности уроков
CREATE INDEX IF NOT EXISTS idx_lessons_availability
ON lessons(current_students, max_students)
WHERE deleted_at IS NULL;

-- Индекс для быстрой проверки баланса кредитов
CREATE INDEX IF NOT EXISTS idx_credits_balance
ON credits(user_id, balance);
