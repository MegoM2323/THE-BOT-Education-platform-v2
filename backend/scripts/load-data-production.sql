-- Unified Data Loading for THE BOT Production
-- Creates realistic test data with correct schema constraints
-- Booking status: 'active' or 'cancelled' (NOT 'confirmed'/'pending')
-- Credit operation_type: 'add', 'deduct', or 'refund' (NOT 'debit')

TRUNCATE TABLE lesson_homework, broadcast_files, lesson_broadcasts,
    cancelled_bookings, messages, file_attachments, blocked_messages,
    chat_rooms, swaps, bookings, template_applications,
    template_lesson_students, template_lessons, lesson_templates,
    lesson_modifications, lessons, credit_transactions, credits,
    payments, subjects, teacher_subjects, sessions, telegram_tokens,
    telegram_users, broadcast_lists, broadcasts, broadcast_logs, auth_failures, users
CASCADE;

-- Password: password123 (bcrypt hash)
INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000001', 'admin@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Администратор THE BOT', 'admin', NOW(), NOW()),
('10000000-0000-0000-0000-000000000001', 'method1@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Иван Петров', 'methodologist', NOW(), NOW()),
('10000000-0000-0000-0000-000000000002', 'method2@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Мария Сидорова', 'methodologist', NOW(), NOW()),
('10000000-0000-0000-0000-000000000003', 'method3@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Александр Морозов', 'methodologist', NOW(), NOW()),
('20000000-0000-0000-0000-000000000001', 'student1@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Дмитрий Смирнов', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000002', 'student2@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Елена Волкова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000003', 'student3@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Павел Морозов', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000004', 'student4@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Ольга Новикова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000005', 'student5@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Анна Иванова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000006', 'student6@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Сергей Петров', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000007', 'student7@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Викторія Козлова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000008', 'student8@thebot.ru', '$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Константин Лебедев', 'student', NOW(), NOW());

-- Setup credits
UPDATE credits SET balance = 15 WHERE user_id = '20000000-0000-0000-0000-000000000001';
UPDATE credits SET balance = 12 WHERE user_id = '20000000-0000-0000-0000-000000000002';
UPDATE credits SET balance = 20 WHERE user_id = '20000000-0000-0000-0000-000000000003';
UPDATE credits SET balance = 8 WHERE user_id = '20000000-0000-0000-0000-000000000004';
UPDATE credits SET balance = 10 WHERE user_id = '20000000-0000-0000-0000-000000000005';
UPDATE credits SET balance = 5 WHERE user_id = '20000000-0000-0000-0000-000000000006';
UPDATE credits SET balance = 18 WHERE user_id = '20000000-0000-0000-0000-000000000007';
UPDATE credits SET balance = 25 WHERE user_id = '20000000-0000-0000-0000-000000000008';

-- Create subjects
INSERT INTO subjects (id, name, description, created_at, updated_at) VALUES
(gen_random_uuid(), 'Математика', 'Курс высшей математики и алгебры', NOW(), NOW()),
(gen_random_uuid(), 'Физика', 'Общая и специальная физика', NOW(), NOW()),
(gen_random_uuid(), 'Информатика', 'Основы программирования и алгоритмы', NOW(), NOW()),
(gen_random_uuid(), 'Русский язык', 'Культура речи и писменность', NOW(), NOW()),
(gen_random_uuid(), 'История', 'Мировая и отечественная история', NOW(), NOW()),
(gen_random_uuid(), 'Английский язык', 'Иностранный язык', NOW(), NOW());

-- Create lessons (20+)
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, subject, homework_text, created_at, updated_at) VALUES
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() - INTERVAL '60 days' + TIME '10:00', NOW() - INTERVAL '60 days' + TIME '11:00', 1, 'Математика', 'Решить задачи 1-20', NOW() - INTERVAL '61 days', NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() - INTERVAL '45 days' + TIME '14:00', NOW() - INTERVAL '45 days' + TIME '15:30', 6, 'Математика', 'Повторить: Интегралы', NOW() - INTERVAL '46 days', NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000002', NOW() - INTERVAL '30 days' + TIME '16:00', NOW() - INTERVAL '30 days' + TIME '17:30', 8, 'Физика', 'Реферат по механике', NOW() - INTERVAL '31 days', NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000002', NOW() - INTERVAL '21 days' + TIME '10:00', NOW() - INTERVAL '21 days' + TIME '11:00', 1, 'Информатика', 'Python программа', NOW() - INTERVAL '22 days', NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000003', NOW() - INTERVAL '14 days' + TIME '15:00', NOW() - INTERVAL '14 days' + TIME '16:30', 5, 'Русский язык', 'Сочинение 3-4 стр', NOW() - INTERVAL '15 days', NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '1 day' + TIME '10:00', NOW() + INTERVAL '1 day' + TIME '11:00', 1, 'Математика', 'Производные', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '2 days' + TIME '14:00', NOW() + INTERVAL '2 days' + TIME '15:30', 4, 'Математика', 'Контрольная: Пределы', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '4 days' + TIME '11:00', NOW() + INTERVAL '4 days' + TIME '12:00', 1, 'Математика', 'Консультация', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '5 days' + TIME '16:00', NOW() + INTERVAL '5 days' + TIME '17:30', 6, 'Физика', 'Задачи ЕГЭ', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '7 days' + TIME '10:00', NOW() + INTERVAL '7 days' + TIME '11:30', 3, 'Информатика', 'Веб-разработка', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '8 days' + TIME '15:00', NOW() + INTERVAL '8 days' + TIME '16:30', 5, 'Русский язык', 'Изложения', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '10 days' + TIME '14:00', NOW() + INTERVAL '10 days' + TIME '15:30', 7, 'История', 'Россия XX века', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '12 days' + TIME '17:00', NOW() + INTERVAL '12 days' + TIME '18:30', 4, 'Английский язык', 'Разговорная практика', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '14 days' + TIME '10:00', NOW() + INTERVAL '14 days' + TIME '11:30', 2, 'Математика', 'Олимпиада', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '16 days' + TIME '14:00', NOW() + INTERVAL '16 days' + TIME '15:30', 6, 'Физика', 'Лаб. работа', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000002', NOW() + INTERVAL '18 days' + TIME '16:00', NOW() + INTERVAL '18 days' + TIME '17:30', 1, 'Информатика', 'Консультация', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '20 days' + TIME '15:00', NOW() + INTERVAL '20 days' + TIME '16:30', 8, 'Русский язык', 'Тренинг ЕГЭ', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '22 days' + TIME '11:00', NOW() + INTERVAL '22 days' + TIME '12:30', 3, 'Математика', 'Графики', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000003', NOW() + INTERVAL '25 days' + TIME '14:00', NOW() + INTERVAL '25 days' + TIME '15:30', 5, 'История', 'Ключевые события', NOW(), NOW()),
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001', NOW() + INTERVAL '28 days' + TIME '10:00', NOW() + INTERVAL '28 days' + TIME '11:00', 6, 'Математика', 'Финальный тест', NOW(), NOW());

-- Create bookings (using 'active' status only - no 'confirmed' or 'pending')
INSERT INTO bookings (id, student_id, lesson_id, status, created_at, updated_at) VALUES
(gen_random_uuid(), '20000000-0000-0000-0000-000000000001', (SELECT id FROM lessons WHERE subject = 'Математика' LIMIT 1), 'active', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000002', (SELECT id FROM lessons WHERE subject = 'Физика' LIMIT 1), 'active', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000003', (SELECT id FROM lessons WHERE subject = 'Информатика' LIMIT 1), 'active', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000004', (SELECT id FROM lessons WHERE subject = 'Математика' LIMIT 1 OFFSET 1), 'active', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000005', (SELECT id FROM lessons WHERE subject = 'Русский язык' LIMIT 1), 'active', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000006', (SELECT id FROM lessons WHERE subject = 'История' LIMIT 1), 'active', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000007', (SELECT id FROM lessons WHERE subject = 'Английский язык' LIMIT 1), 'active', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000008', (SELECT id FROM lessons WHERE subject = 'Физика' LIMIT 1 OFFSET 1), 'active', NOW(), NOW());

-- Create credit transactions (using 'add', 'deduct', 'refund' - NOT 'debit')
INSERT INTO credit_transactions (id, user_id, amount, operation_type, reason, performed_by, created_at) VALUES
(gen_random_uuid(), '20000000-0000-0000-0000-000000000001', 15, 'add', 'Initial allocation', '00000000-0000-0000-0000-000000000001', NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000002', 12, 'add', 'Initial allocation', '00000000-0000-0000-0000-000000000001', NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000003', 20, 'add', 'Bonus', '00000000-0000-0000-0000-000000000001', NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000001', 1, 'deduct', 'Lesson booking', '10000000-0000-0000-0000-000000000001', NOW());

SELECT '✓ Data loaded successfully!' as result;
