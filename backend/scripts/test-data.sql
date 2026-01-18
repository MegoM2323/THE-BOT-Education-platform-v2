-- ╔═══════════════════════════════════════════════════════════════╗
-- ║  ТЕСТОВЫЕ ДАННЫЕ ДЛЯ ПЛАТФОРМЫ РЕПЕТИТОРСТВА                 ║
-- ║  Test Data for Tutoring Platform                             ║
-- ╚═══════════════════════════════════════════════════════════════╝

-- Очистка существующих данных (включая новые таблицы)
TRUNCATE TABLE
    -- Forum system
    blocked_messages, file_attachments, messages, chat_rooms,
    -- Lesson features
    lesson_homework, broadcast_files, lesson_broadcasts, cancelled_bookings,
    -- Payments
    payments,
    -- Core tables
    swaps, credit_transactions, bookings, lessons, sessions, credits,
    -- Templates (если есть)
    template_lesson_students, template_applications, template_lessons, lesson_templates,
    -- Users (последним из-за FK)
    telegram_users, users
RESTART IDENTITY CASCADE;

-- ═══════════════════════════════════════════════════════════════
-- 1. ПОЛЬЗОВАТЕЛИ (Users)
-- ═══════════════════════════════════════════════════════════════

-- Пароль для всех: password123
-- Bcrypt hash для "password123": $2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy

-- Администратор
INSERT INTO users (id, email, password_hash, full_name, role) VALUES
('00000000-0000-0000-0000-000000000001', 'admin@tutoring.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Администратор Системы', 'admin');

-- Преподаватели
INSERT INTO users (id, email, password_hash, full_name, role) VALUES
('10000000-0000-0000-0000-000000000001', 'ivan.petrov@tutoring.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Иван Петров', 'teacher'),
('10000000-0000-0000-0000-000000000002', 'maria.sidorova@tutoring.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Мария Сидорова', 'teacher'),
('10000000-0000-0000-0000-000000000003', 'alexey.kozlov@tutoring.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Алексей Козлов', 'teacher');

-- Ученики (с payment_enabled)
INSERT INTO users (id, email, password_hash, full_name, role, payment_enabled) VALUES
('20000000-0000-0000-0000-000000000001', 'anna.ivanova@student.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Анна Иванова', 'student', TRUE),
('20000000-0000-0000-0000-000000000002', 'dmitry.smirnov@student.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Дмитрий Смирнов', 'student', TRUE),
('20000000-0000-0000-0000-000000000003', 'elena.volkova@student.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Елена Волкова', 'student', TRUE),
('20000000-0000-0000-0000-000000000004', 'pavel.morozov@student.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Павел Морозов', 'student', FALSE),
('20000000-0000-0000-0000-000000000005', 'olga.novikova@student.com', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Ольга Новикова', 'student', TRUE);

-- ═══════════════════════════════════════════════════════════════
-- 2. КРЕДИТЫ (Credits)
-- ═══════════════════════════════════════════════════════════════

-- Обновление балансов кредитов (триггер create_credits_on_student_insert уже создал записи с balance=0)
UPDATE credits SET balance = 10 WHERE user_id = '20000000-0000-0000-0000-000000000001'; -- Анна Иванова
UPDATE credits SET balance = 8 WHERE user_id = '20000000-0000-0000-0000-000000000002';  -- Дмитрий Смирнов
UPDATE credits SET balance = 12 WHERE user_id = '20000000-0000-0000-0000-000000000003'; -- Елена Волкова
UPDATE credits SET balance = 5 WHERE user_id = '20000000-0000-0000-0000-000000000004';  -- Павел Морозов
UPDATE credits SET balance = 3 WHERE user_id = '20000000-0000-0000-0000-000000000005';  -- Ольга Новикова

-- ═══════════════════════════════════════════════════════════════
-- 3. ЗАНЯТИЯ (Lessons)
-- ═══════════════════════════════════════════════════════════════

-- Занятия на ближайшие 2 недели
-- Teachers: Иван Петров (Math), Мария Сидорова (Physics), Алексей Козлов (CS)

-- ПРОШЕДШИЕ ЗАНЯТИЯ (для истории bookings)
-- color: Иван Петров (Math) = #3B82F6 (синий), Мария Сидорова (Physics) = #10B981 (зеленый), Алексей Козлов (CS) = #8B5CF6 (фиолетовый)
-- max_students определяет тип: 1 = индивидуальное, >1 = групповое
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
-- Занятия неделю назад
('30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', NOW() - INTERVAL '7 days' + INTERVAL '10 hours', NOW() - INTERVAL '7 days' + INTERVAL '11.5 hours', 4, 3, '#3B82F6', 'Математика'),
('30000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', NOW() - INTERVAL '6 days' + INTERVAL '14 hours', NOW() - INTERVAL '6 days' + INTERVAL '15.5 hours', 4, 2, '#10B981', 'Физика'),
('30000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', NOW() - INTERVAL '5 days' + INTERVAL '16 hours', NOW() - INTERVAL '5 days' + INTERVAL '17.5 hours', 4, 4, '#8B5CF6', 'Информатика'),
('30000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', NOW() - INTERVAL '4 days' + INTERVAL '12 hours', NOW() - INTERVAL '4 days' + INTERVAL '13.5 hours', 1, 1, '#3B82F6', 'Математика (подготовка к ЕГЭ)');

-- БУДУЩИЕ ЗАНЯТИЯ (с color и subject)
-- Понедельник (+1 день)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
('31000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '1 day' + INTERVAL '10 hours', NOW() + INTERVAL '1 day' + INTERVAL '11.5 hours', 4, 0, '#3B82F6', 'Математика'),
('31000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '1 day' + INTERVAL '14 hours', NOW() + INTERVAL '1 day' + INTERVAL '15.5 hours', 4, 0, '#10B981', 'Физика'),
('31000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '1 day' + INTERVAL '16 hours', NOW() + INTERVAL '1 day' + INTERVAL '17.5 hours', 4, 0, '#8B5CF6', 'Информатика');

-- Вторник (+2 дня)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
('31000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '2 days' + INTERVAL '10 hours', NOW() + INTERVAL '2 days' + INTERVAL '11.5 hours', 4, 0, '#3B82F6', 'Алгебра'),
('31000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '2 days' + INTERVAL '14 hours', NOW() + INTERVAL '2 days' + INTERVAL '15.5 hours', 4, 0, '#10B981', 'Физика: механика'),
('31000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '2 days' + INTERVAL '16 hours', NOW() + INTERVAL '2 days' + INTERVAL '17.5 hours', 4, 0, '#8B5CF6', 'Python: основы');

-- Среда (+3 дня)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
('31000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '3 days' + INTERVAL '10 hours', NOW() + INTERVAL '3 days' + INTERVAL '11.5 hours', 4, 0, '#3B82F6', 'Геометрия'),
('31000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '3 days' + INTERVAL '14 hours', NOW() + INTERVAL '3 days' + INTERVAL '15 hours', 1, 0, '#10B981', 'Физика: ЕГЭ'),
('31000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '3 days' + INTERVAL '16 hours', NOW() + INTERVAL '3 days' + INTERVAL '17.5 hours', 4, 0, '#8B5CF6', 'Алгоритмы');

-- Четверг (+4 дня)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
('31000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '4 days' + INTERVAL '10 hours', NOW() + INTERVAL '4 days' + INTERVAL '11.5 hours', 4, 0, '#3B82F6', 'Тригонометрия'),
('31000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '4 days' + INTERVAL '14 hours', NOW() + INTERVAL '4 days' + INTERVAL '15.5 hours', 4, 0, '#10B981', 'Физика: электричество'),
('31000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '4 days' + INTERVAL '16 hours', NOW() + INTERVAL '4 days' + INTERVAL '17 hours', 1, 0, '#8B5CF6', 'Web-разработка');

-- Пятница (+5 дней)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
('31000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '5 days' + INTERVAL '10 hours', NOW() + INTERVAL '5 days' + INTERVAL '11.5 hours', 4, 0, '#3B82F6', 'Математика'),
('31000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '5 days' + INTERVAL '14 hours', NOW() + INTERVAL '5 days' + INTERVAL '15.5 hours', 4, 0, '#10B981', 'Физика'),
('31000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '5 days' + INTERVAL '16 hours', NOW() + INTERVAL '5 days' + INTERVAL '17.5 hours', 4, 0, '#8B5CF6', 'Базы данных');

-- Суббота (+6 дней)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
('31000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '6 days' + INTERVAL '10 hours', NOW() + INTERVAL '6 days' + INTERVAL '11.5 hours', 4, 0, '#3B82F6', 'Подготовка к ЕГЭ'),
('31000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '6 days' + INTERVAL '14 hours', NOW() + INTERVAL '6 days' + INTERVAL '15.5 hours', 4, 0, '#10B981', 'Физика: оптика'),
('31000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '6 days' + INTERVAL '16 hours', NOW() + INTERVAL '6 days' + INTERVAL '17.5 hours', 4, 0, '#8B5CF6', 'Информатика');

-- Следующая неделя (+8-10 дней)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject) VALUES
('31000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '8 days' + INTERVAL '10 hours', NOW() + INTERVAL '8 days' + INTERVAL '11.5 hours', 4, 0, '#3B82F6', 'Математика'),
('31000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '8 days' + INTERVAL '14 hours', NOW() + INTERVAL '8 days' + INTERVAL '15.5 hours', 4, 0, '#10B981', 'Физика'),
('31000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '9 days' + INTERVAL '16 hours', NOW() + INTERVAL '9 days' + INTERVAL '17.5 hours', 4, 0, '#8B5CF6', 'Информатика'),
('31000000-0000-0000-0000-000000000022', '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '10 days' + INTERVAL '12 hours', NOW() + INTERVAL '10 days' + INTERVAL '13.5 hours', 1, 0, '#EF4444', 'Олимпиадная математика');

-- ═══════════════════════════════════════════════════════════════
-- 4. ЗАПИСИ (Bookings)
-- ═══════════════════════════════════════════════════════════════

-- ПРОШЛЫЕ BOOKINGS (для истории)
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at) VALUES
-- Анна посетила урок неделю назад
('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '8 days'),
-- Дмитрий посетил урок 6 дней назад
('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000002', 'active', NOW() - INTERVAL '7 days'),
-- Елена посетила урок 5 дней назад
('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '6 days'),
-- Павел записался и отменил урок 4 дня назад
('40000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000004', 'cancelled', NOW() - INTERVAL '5 days');

-- АКТИВНЫЕ BOOKINGS (будущие)

-- Анна Иванова: 3 занятия (Math)
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at) VALUES
('41000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '31000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '2 days'),
('41000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', '31000000-0000-0000-0000-000000000004', 'active', NOW() - INTERVAL '2 days'),
('41000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000001', '31000000-0000-0000-0000-000000000007', 'active', NOW() - INTERVAL '1 day');

-- Дмитрий Смирнов: 2 занятия (Physics)
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at) VALUES
('41000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000002', '31000000-0000-0000-0000-000000000002', 'active', NOW() - INTERVAL '2 days'),
('41000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000002', '31000000-0000-0000-0000-000000000005', 'active', NOW() - INTERVAL '1 day');

-- Елена Волкова: 4 занятия (CS)
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at) VALUES
('41000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000003', '31000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '2 days'),
('41000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000003', '31000000-0000-0000-0000-000000000006', 'active', NOW() - INTERVAL '2 days'),
('41000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000003', '31000000-0000-0000-0000-000000000009', 'active', NOW() - INTERVAL '1 day'),
('41000000-0000-0000-0000-000000000009', '20000000-0000-0000-0000-000000000003', '31000000-0000-0000-0000-000000000012', 'active', NOW() - INTERVAL '1 hour');

-- Павел Морозов: 2 занятия (Math + индивидуальное)
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at) VALUES
('41000000-0000-0000-0000-000000000010', '20000000-0000-0000-0000-000000000004', '31000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '3 days'),
('41000000-0000-0000-0000-000000000011', '20000000-0000-0000-0000-000000000004', '31000000-0000-0000-0000-000000000022', 'active', NOW() - INTERVAL '12 hours');

-- Ольга Новикова: 1 занятие (индивидуальное Physics)
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at) VALUES
('41000000-0000-0000-0000-000000000012', '20000000-0000-0000-0000-000000000005', '31000000-0000-0000-0000-000000000008', 'active', NOW() - INTERVAL '1 day');

-- Обновляем счетчики студентов в занятиях
UPDATE lessons
SET current_students = (SELECT COUNT(*) FROM bookings WHERE lesson_id = lessons.id AND status = 'active');

-- ═══════════════════════════════════════════════════════════════
-- 5. ИСТОРИЯ ТРАНЗАКЦИЙ КРЕДИТОВ (Credit Transactions)
-- ═══════════════════════════════════════════════════════════════

-- Начисления кредитов администратором
INSERT INTO credit_transactions (user_id, amount, operation_type, reason, performed_by, created_at) VALUES
-- Анна: начислено 15, списано 4, осталось 10 (баланс соответствует)
('20000000-0000-0000-0000-000000000001', 15, 'add', 'Initial credit purchase', '00000000-0000-0000-0000-000000000001', NOW() - INTERVAL '14 days'),
-- Дмитрий: начислено 10, списано 2, осталось 8
('20000000-0000-0000-0000-000000000002', 10, 'add', 'Initial credit purchase', '00000000-0000-0000-0000-000000000001', NOW() - INTERVAL '13 days'),
-- Елена: начислено 17, списано 5, осталось 12
('20000000-0000-0000-0000-000000000003', 17, 'add', 'Initial credit purchase', '00000000-0000-0000-0000-000000000001', NOW() - INTERVAL '12 days'),
-- Павел: начислено 8, списано 2 (1 отменен с возвратом), осталось 5
('20000000-0000-0000-0000-000000000004', 8, 'add', 'Initial credit purchase', '00000000-0000-0000-0000-000000000001', NOW() - INTERVAL '11 days'),
-- Ольга: начислено 4, списано 1, осталось 3
('20000000-0000-0000-0000-000000000005', 4, 'add', 'Initial credit purchase', '00000000-0000-0000-0000-000000000001', NOW() - INTERVAL '10 days');

-- Списания за прошлые bookings
INSERT INTO credit_transactions (user_id, amount, operation_type, reason, booking_id, created_at) VALUES
-- Анна: прошлое занятие
('20000000-0000-0000-0000-000000000001', -1, 'deduct', 'Lesson booking', '40000000-0000-0000-0000-000000000001', NOW() - INTERVAL '8 days'),
-- Дмитрий: прошлое занятие
('20000000-0000-0000-0000-000000000002', -1, 'deduct', 'Lesson booking', '40000000-0000-0000-0000-000000000002', NOW() - INTERVAL '7 days'),
-- Елена: прошлое занятие
('20000000-0000-0000-0000-000000000003', -1, 'deduct', 'Lesson booking', '40000000-0000-0000-0000-000000000003', NOW() - INTERVAL '6 days'),
-- Павел: отмененное занятие (списание)
('20000000-0000-0000-0000-000000000004', -1, 'deduct', 'Lesson booking', '40000000-0000-0000-0000-000000000004', NOW() - INTERVAL '5 days'),
-- Павел: возврат за отмену
('20000000-0000-0000-0000-000000000004', 1, 'refund', 'Booking cancellation refund', '40000000-0000-0000-0000-000000000004', NOW() - INTERVAL '5 days' + INTERVAL '1 hour');

-- Списания за активные будущие bookings
INSERT INTO credit_transactions (user_id, amount, operation_type, reason, booking_id, created_at) VALUES
-- Анна: 3 активных booking
('20000000-0000-0000-0000-000000000001', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000001', NOW() - INTERVAL '2 days'),
('20000000-0000-0000-0000-000000000001', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000002', NOW() - INTERVAL '2 days' + INTERVAL '1 hour'),
('20000000-0000-0000-0000-000000000001', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000003', NOW() - INTERVAL '1 day'),
-- Дмитрий: 2 активных booking
('20000000-0000-0000-0000-000000000002', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000004', NOW() - INTERVAL '2 days'),
('20000000-0000-0000-0000-000000000002', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000005', NOW() - INTERVAL '1 day'),
-- Елена: 4 активных booking
('20000000-0000-0000-0000-000000000003', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000006', NOW() - INTERVAL '2 days'),
('20000000-0000-0000-0000-000000000003', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000007', NOW() - INTERVAL '2 days' + INTERVAL '30 minutes'),
('20000000-0000-0000-0000-000000000003', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000008', NOW() - INTERVAL '1 day'),
('20000000-0000-0000-0000-000000000003', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000009', NOW() - INTERVAL '1 hour'),
-- Павел: 2 активных booking
('20000000-0000-0000-0000-000000000004', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000010', NOW() - INTERVAL '3 days'),
('20000000-0000-0000-0000-000000000004', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000011', NOW() - INTERVAL '12 hours'),
-- Ольга: 1 активный booking
('20000000-0000-0000-0000-000000000005', -1, 'deduct', 'Lesson booking', '41000000-0000-0000-0000-000000000012', NOW() - INTERVAL '1 day');

-- ═══════════════════════════════════════════════════════════════
-- 6. ИСТОРИЯ СВАПОВ (Swaps) - опционально
-- ═══════════════════════════════════════════════════════════════

-- Пример: добавление swap если требуется для тестирования функционала обмена занятиями
-- В данной версии оставляем пустым, но структура готова для использования

-- ═══════════════════════════════════════════════════════════════
-- 7. ФОРУМ / ЧАТ (Chat Rooms & Messages)
-- ═══════════════════════════════════════════════════════════════

-- Чат-комнаты между преподавателями и учениками
INSERT INTO chat_rooms (id, teacher_id, student_id, created_at, updated_at) VALUES
-- Иван Петров (Math) с учениками
('50000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', NOW() - INTERVAL '10 days', NOW() - INTERVAL '1 day'),
('50000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000004', NOW() - INTERVAL '8 days', NOW() - INTERVAL '2 days'),
-- Мария Сидорова (Physics) с учениками
('50000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', NOW() - INTERVAL '7 days', NOW() - INTERVAL '3 hours'),
('50000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000005', NOW() - INTERVAL '5 days', NOW() - INTERVAL '1 day'),
-- Алексей Козлов (CS) с учениками
('50000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', NOW() - INTERVAL '6 days', NOW() - INTERVAL '12 hours');

-- Сообщения в чатах (статус delivered - уже прошли модерацию)
INSERT INTO messages (id, room_id, sender_id, message_text, status, created_at) VALUES
-- Чат Иван-Анна
('60000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Здравствуйте! У меня вопрос по домашнему заданию.', 'delivered', NOW() - INTERVAL '5 days'),
('60000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'Добрый день, Анна! Конечно, спрашивайте.', 'delivered', NOW() - INTERVAL '5 days' + INTERVAL '30 minutes'),
('60000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Не могу решить задачу №5 из учебника.', 'delivered', NOW() - INTERVAL '5 days' + INTERVAL '1 hour'),
('60000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'Попробуйте начать с построения графика функции. Это поможет понять решение.', 'delivered', NOW() - INTERVAL '5 days' + INTERVAL '2 hours'),
-- Чат Мария-Дмитрий
('60000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000002', 'Добрый день! Можно ли перенести занятие на час позже?', 'delivered', NOW() - INTERVAL '3 days'),
('60000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 'К сожалению, у меня после вас следующий ученик. Давайте обсудим другой день?', 'delivered', NOW() - INTERVAL '3 days' + INTERVAL '1 hour'),
-- Чат Алексей-Елена
('60000000-0000-0000-0000-000000000007', '50000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000003', 'Здравствуйте! Прикрепляю свой код для проверки.', 'delivered', NOW() - INTERVAL '2 days'),
('60000000-0000-0000-0000-000000000008', '50000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000003', 'Отлично, Елена! Посмотрю и дам обратную связь на занятии.', 'delivered', NOW() - INTERVAL '2 days' + INTERVAL '4 hours'),
-- Сообщение на модерации
('60000000-0000-0000-0000-000000000009', '50000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Спасибо за помощь!', 'pending_moderation', NOW() - INTERVAL '1 day');

-- ═══════════════════════════════════════════════════════════════
-- 8. ДОМАШНИЕ ЗАДАНИЯ (Lesson Homework)
-- ═══════════════════════════════════════════════════════════════

-- Пример домашних заданий к прошедшим занятиям
INSERT INTO lesson_homework (id, lesson_id, file_name, file_path, file_size, mime_type, created_by, created_at) VALUES
-- Математика (Иван Петров)
('70000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'homework_math_week1.pdf', 'uploads/homework/70000000-0000-0000-0000-000000000001.pdf', 245678, 'application/pdf', '10000000-0000-0000-0000-000000000001', NOW() - INTERVAL '7 days' + INTERVAL '2 hours'),
-- Физика (Мария Сидорова)
('70000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000002', 'physics_tasks.pdf', 'uploads/homework/70000000-0000-0000-0000-000000000002.pdf', 512340, 'application/pdf', '10000000-0000-0000-0000-000000000002', NOW() - INTERVAL '6 days' + INTERVAL '1 hour'),
('70000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002', 'formula_sheet.png', 'uploads/homework/70000000-0000-0000-0000-000000000003.png', 89456, 'image/png', '10000000-0000-0000-0000-000000000002', NOW() - INTERVAL '6 days' + INTERVAL '1.5 hours'),
-- Информатика (Алексей Козлов)
('70000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000003', 'python_exercises.zip', 'uploads/homework/70000000-0000-0000-0000-000000000004.zip', 1234567, 'application/zip', '10000000-0000-0000-0000-000000000003', NOW() - INTERVAL '5 days' + INTERVAL '30 minutes');

-- ═══════════════════════════════════════════════════════════════
-- 9. РАССЫЛКИ (Lesson Broadcasts)
-- ═══════════════════════════════════════════════════════════════

-- Пример рассылки сообщения всем студентам занятия
INSERT INTO lesson_broadcasts (id, lesson_id, sender_id, message, status, sent_count, failed_count, created_at, completed_at) VALUES
-- Рассылка от Ивана Петрова для будущего занятия
('80000000-0000-0000-0000-000000000001', '31000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001',
 'Напоминаю, что завтра занятие по математике. Принесите тетради для контрольной работы!',
 'completed', 2, 0, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours' + INTERVAL '5 minutes'),
-- Рассылка от Марии Сидоровой
('80000000-0000-0000-0000-000000000002', '31000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002',
 'Добрый день! На следующем занятии будем разбирать задачи повышенной сложности. Подготовьте вопросы.',
 'completed', 1, 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours' + INTERVAL '2 minutes');

-- Файлы к рассылкам
INSERT INTO broadcast_files (id, broadcast_id, file_name, file_path, file_size, mime_type, uploaded_at) VALUES
('81000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000001', 'plan_kontrolnoy.pdf', 'uploads/broadcasts/81000000-0000-0000-0000-000000000001.pdf', 156789, 'application/pdf', NOW() - INTERVAL '12 hours');

-- ═══════════════════════════════════════════════════════════════
-- 10. ОТМЕНЕННЫЕ БРОНИРОВАНИЯ (Cancelled Bookings Prevention)
-- ═══════════════════════════════════════════════════════════════

-- Запись об отмене для предотвращения повторного бронирования
INSERT INTO cancelled_bookings (id, booking_id, student_id, lesson_id, cancelled_at) VALUES
('90000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000004', NOW() - INTERVAL '5 days' + INTERVAL '1 hour');

-- ═══════════════════════════════════════════════════════════════
-- 11. ИСТОРИЯ ПЛАТЕЖЕЙ (Payments - опционально)
-- ═══════════════════════════════════════════════════════════════

-- Примеры успешных платежей (YooKassa)
INSERT INTO payments (id, user_id, amount, credits, status, yookassa_payment_id, created_at, updated_at) VALUES
-- Анна купила 5 кредитов
('A0000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 14000.00, 5, 'succeeded', 'yookassa_pay_001', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days'),
-- Дмитрий купил 3 кредита
('A0000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 8400.00, 3, 'succeeded', 'yookassa_pay_002', NOW() - INTERVAL '13 days', NOW() - INTERVAL '13 days'),
-- Елена купила 7 кредитов
('A0000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', 19600.00, 7, 'succeeded', 'yookassa_pay_003', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
-- Отмененный платеж
('A0000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000005', 2800.00, 1, 'cancelled', 'yookassa_pay_004', NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days');

-- ═══════════════════════════════════════════════════════════════
-- 11. ШАБЛОНЫ ЗАНЯТИЙ (Lesson Templates)
-- ═══════════════════════════════════════════════════════════════

-- Template A: Week A - Standard Schedule
-- 5 lessons: Mon, Mon, Wed, Fri, Fri
-- Colors: Ivan (Math) = #3B82F6, Maria (Physics) = #10B981, Alexey (CS) = #8B5CF6
INSERT INTO lesson_templates (id, admin_id, name, description, created_at, updated_at) VALUES
('70000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
 'Week A - Standard Schedule',
 'Standard weekly schedule with 5 lessons: Math, Physics, CS, Physics, Math',
 NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- Template A - Lesson 1: Monday 10:00-12:00 Individual (Ivan, Math) - assign Anna
-- max_students=1 означает индивидуальное занятие
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 1, '10:00:00'::TIME, '12:00:00'::TIME, 1, '#3B82F6', 'Математика', 'individual', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');
INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at) VALUES
('72000000-0000-0000-0000-000000000001', '71000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', NOW() - INTERVAL '30 days');

-- Template A - Lesson 2: Monday 14:00-16:00 Group (Maria, Physics) - assign Dmitry
-- max_students=4 означает групповое занятие
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 1, '14:00:00'::TIME, '16:00:00'::TIME, 4, '#10B981', 'Физика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');
INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at) VALUES
('72000000-0000-0000-0000-000000000002', '71000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', NOW() - INTERVAL '30 days');

-- Template A - Lesson 3: Wednesday 10:00-12:00 Group (Alexey, CS) - assign Elena
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', 3, '10:00:00'::TIME, '12:00:00'::TIME, 4, '#8B5CF6', 'Информатика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');
INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at) VALUES
('72000000-0000-0000-0000-000000000003', '71000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', NOW() - INTERVAL '30 days');

-- Template A - Lesson 4: Friday 10:00-12:00 Individual (Maria, Physics) - assign Dmitry
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 5, '10:00:00'::TIME, '12:00:00'::TIME, 1, '#10B981', 'Физика', 'individual', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');
INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at) VALUES
('72000000-0000-0000-0000-000000000004', '71000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000002', NOW() - INTERVAL '30 days');

-- Template A - Lesson 5: Friday 14:00-16:00 Group (Ivan, Math) - open (no students)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 5, '14:00:00'::TIME, '16:00:00'::TIME, 4, '#3B82F6', 'Математика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- Template B: Week B - Extended Schedule
-- 7 lessons: Mon, Tue, Wed, Wed, Thu, Fri, Fri
INSERT INTO lesson_templates (id, admin_id, name, description, created_at, updated_at) VALUES
('70000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
 'Week B - Extended Schedule',
 'Extended weekly schedule with 7 lessons: Math, Physics, CS, Math, Physics, Math, CS',
 NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- Template B - Lesson 1: Monday 09:00-11:00 Group (Ivan, Math)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 1, '09:00:00'::TIME, '11:00:00'::TIME, 4, '#3B82F6', 'Математика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- Template B - Lesson 2: Tuesday 10:00-12:00 Group (Maria, Physics)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 2, '10:00:00'::TIME, '12:00:00'::TIME, 4, '#10B981', 'Физика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- Template B - Lesson 3: Wednesday 10:00-12:00 Individual (Alexey, CS) - assign Elena
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000008', '70000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 3, '10:00:00'::TIME, '12:00:00'::TIME, 1, '#8B5CF6', 'Информатика', 'individual', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');
INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at) VALUES
('72000000-0000-0000-0000-000000000005', '71000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000003', NOW() - INTERVAL '30 days');

-- Template B - Lesson 4: Wednesday 14:00-16:00 Group (Ivan, Math)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000009', '70000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 3, '14:00:00'::TIME, '16:00:00'::TIME, 4, '#3B82F6', 'Математика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- Template B - Lesson 5: Thursday 10:00-12:00 Group (Maria, Physics)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000010', '70000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 4, '10:00:00'::TIME, '12:00:00'::TIME, 4, '#10B981', 'Физика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- Template B - Lesson 6: Friday 10:00-12:00 Individual (Ivan, Math) - assign Anna
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000011', '70000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 5, '10:00:00'::TIME, '12:00:00'::TIME, 1, '#3B82F6', 'Математика', 'individual', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');
INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at) VALUES
('72000000-0000-0000-0000-000000000006', '71000000-0000-0000-0000-000000000011', '20000000-0000-0000-0000-000000000001', NOW() - INTERVAL '30 days');

-- Template B - Lesson 7: Friday 15:00-17:00 Group (Alexey, CS)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, max_students, color, subject, created_at, updated_at) VALUES
('71000000-0000-0000-0000-000000000012', '70000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 5, '15:00:00'::TIME, '17:00:00'::TIME, 4, '#8B5CF6', 'Информатика', 'group', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days');

-- ═══════════════════════════════════════════════════════════════
-- 12. ИТОГОВАЯ СВОДКА И ПРОВЕРКИ
-- ═══════════════════════════════════════════════════════════════

-- Проверка данных
DO $$
DECLARE
    v_users_count INTEGER;
    v_lessons_future INTEGER;
    v_lessons_past INTEGER;
    v_bookings_active INTEGER;
    v_bookings_cancelled INTEGER;
    v_total_credits INTEGER;
    v_chat_rooms INTEGER;
    v_messages INTEGER;
    v_homework_files INTEGER;
    v_broadcasts INTEGER;
    v_payments INTEGER;
    v_lesson_templates INTEGER;
    v_template_lessons INTEGER;
    v_template_lesson_students INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_users_count FROM users;
    SELECT COUNT(*) INTO v_lessons_future FROM lessons WHERE start_time > NOW();
    SELECT COUNT(*) INTO v_lessons_past FROM lessons WHERE start_time <= NOW();
    SELECT COUNT(*) INTO v_bookings_active FROM bookings WHERE status = 'active';
    SELECT COUNT(*) INTO v_bookings_cancelled FROM bookings WHERE status = 'cancelled';
    SELECT SUM(balance) INTO v_total_credits FROM credits;
    SELECT COUNT(*) INTO v_chat_rooms FROM chat_rooms;
    SELECT COUNT(*) INTO v_messages FROM messages;
    SELECT COUNT(*) INTO v_homework_files FROM lesson_homework;
    SELECT COUNT(*) INTO v_broadcasts FROM lesson_broadcasts;
    SELECT COUNT(*) INTO v_payments FROM payments;
    SELECT COUNT(*) INTO v_lesson_templates FROM lesson_templates;
    SELECT COUNT(*) INTO v_template_lessons FROM template_lessons;
    SELECT COUNT(*) INTO v_template_lesson_students FROM template_lesson_students;

    RAISE NOTICE '═══════════════════════════════════════════════════════════════';
    RAISE NOTICE 'TEST DATA SUMMARY';
    RAISE NOTICE '═══════════════════════════════════════════════════════════════';
    RAISE NOTICE 'Users: % (1 admin, 3 teachers, 5 students)', v_users_count;
    RAISE NOTICE 'Lessons (future): %', v_lessons_future;
    RAISE NOTICE 'Lessons (past): %', v_lessons_past;
    RAISE NOTICE 'Bookings (active): %', v_bookings_active;
    RAISE NOTICE 'Bookings (cancelled): %', v_bookings_cancelled;
    RAISE NOTICE 'Total credits balance: %', v_total_credits;
    RAISE NOTICE '---';
    RAISE NOTICE 'Chat rooms: %', v_chat_rooms;
    RAISE NOTICE 'Messages: %', v_messages;
    RAISE NOTICE 'Homework files: %', v_homework_files;
    RAISE NOTICE 'Broadcasts: %', v_broadcasts;
    RAISE NOTICE 'Payments: %', v_payments;
    RAISE NOTICE '---';
    RAISE NOTICE 'Lesson Templates: %', v_lesson_templates;
    RAISE NOTICE 'Template Lessons: %', v_template_lessons;
    RAISE NOTICE 'Template Lesson Students: %', v_template_lesson_students;
    RAISE NOTICE '═══════════════════════════════════════════════════════════════';
END $$;

-- Детальная информация о пользователях
SELECT
    u.full_name,
    u.email,
    u.role,
    COALESCE(c.balance, 0) as credits,
    COUNT(b.id) FILTER (WHERE b.status = 'active') as active_bookings
FROM users u
LEFT JOIN credits c ON u.id = c.user_id
LEFT JOIN bookings b ON u.id = b.student_id
WHERE u.role = 'student'
GROUP BY u.id, u.full_name, u.email, u.role, c.balance
ORDER BY u.full_name;

-- Проверка консистентности: счетчики студентов в lessons
SELECT
    l.id,
    l.teacher_id,
    l.start_time,
    l.max_students,
    l.current_students,
    COUNT(b.id) FILTER (WHERE b.status = 'active') as actual_active_bookings
FROM lessons l
LEFT JOIN bookings b ON l.id = b.lesson_id
GROUP BY l.id
HAVING l.current_students != COUNT(b.id) FILTER (WHERE b.status = 'active');

-- Готово!
SELECT 'Test data loaded successfully!' as status;
