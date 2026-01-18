#!/bin/bash
# Скрипт загрузки тестовых данных для Tutoring Platform
# Загружает: пользователей, занятия, шаблоны, ДЗ, рассылки, бронирования
#
# USAGE:
#   ./load-data.sh                    # Insert only (safe, default)
#   ./load-data.sh --truncate         # Clear all data first (dangerous!)
#   ./load-data.sh --truncate --yes   # Auto-confirm truncate (CI/CD only)

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Параметры по умолчанию
DO_TRUNCATE=false
AUTO_CONFIRM=false

# Парсинг аргументов
while [[ $# -gt 0 ]]; do
    case $1 in
        --truncate)
            DO_TRUNCATE=true
            shift
            ;;
        --yes|-y)
            AUTO_CONFIRM=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --truncate    Clear ALL existing data before inserting (DANGEROUS!)"
            echo "  --yes, -y     Auto-confirm truncate (for CI/CD, use with caution)"
            echo "  --help, -h    Show this help message"
            echo ""
            echo "By default, the script only INSERTS data without deleting existing records."
            echo "Use --truncate only in development environments."
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Настройки БД
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

export PGPASSWORD="$DB_PASSWORD"

# CRITICAL SAFETY CHECK 1: Verify we're not loading into test database
if [[ "$DB_NAME" == "tutoring_platform_test" ]]; then
    echo -e "${RED}ERROR: load-data.sh should NOT be run against test database!${NC}"
    echo "This script loads sample data into the DEVELOPMENT database."
    echo "To use development database, set: DB_NAME=tutoring_platform"
    echo ""
    echo "Current DB_NAME: $DB_NAME"
    exit 1
fi

# CRITICAL SAFETY CHECK 2: Refuse to run on production database
if [[ "$DB_NAME" == *"prod"* ]] || [[ "$DB_NAME" == *"production"* ]] || [[ "$DB_NAME" == *"live"* ]]; then
    echo -e "${RED}============================================================${NC}"
    echo -e "${RED}CRITICAL SAFETY VIOLATION!${NC}"
    echo -e "${RED}============================================================${NC}"
    echo -e "${RED}Database name '$DB_NAME' appears to be a production database!${NC}"
    echo -e "${RED}This script should NEVER run on production!${NC}"
    echo -e "${RED}Aborting to prevent data loss.${NC}"
    echo -e "${RED}============================================================${NC}"
    exit 1
fi

# CRITICAL SAFETY CHECK 3: Block truncate on production-like databases
if [[ "$DO_TRUNCATE" == true ]]; then
    if [[ "$DB_HOST" != "localhost" ]] && [[ "$DB_HOST" != "127.0.0.1" ]] && [[ "$DB_HOST" != "::1" ]]; then
        echo -e "${RED}============================================================${NC}"
        echo -e "${RED}TRUNCATE BLOCKED: Remote database detected!${NC}"
        echo -e "${RED}============================================================${NC}"
        echo -e "${RED}TRUNCATE is only allowed on localhost databases.${NC}"
        echo -e "${RED}Host: $DB_HOST${NC}"
        echo -e "${RED}Aborting to prevent data loss.${NC}"
        echo -e "${RED}============================================================${NC}"
        exit 1
    fi
fi

# CRITICAL SAFETY CHECK 4: Verify DB_PASSWORD is set
if [[ -z "$DB_PASSWORD" ]]; then
    echo -e "${RED}ERROR: DB_PASSWORD not set. Refusing to run.${NC}"
    echo "Set DB_PASSWORD environment variable before running this script."
    exit 1
fi

echo -e "${YELLOW}=== Загрузка тестовых данных ===${NC}"
echo "База данных: $DB_NAME@$DB_HOST:$DB_PORT"
echo -e "${GREEN}✓ Database safety checks passed${NC}"
echo ""

# Функция выполнения SQL
run_sql() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$1"
}

run_sql_file() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$1"
}

# 0. Очистка существующих данных (только если явно запрошено)
if [[ "$DO_TRUNCATE" == true ]]; then
    echo -e "${RED}============================================================${NC}"
    echo -e "${RED}WARNING: TRUNCATE requested!${NC}"
    echo -e "${RED}This will DELETE ALL DATA in database: $DB_NAME${NC}"
    echo -e "${RED}============================================================${NC}"

    if [[ "$AUTO_CONFIRM" != true ]]; then
        echo ""
        echo -e "${YELLOW}Type 'DELETE ALL DATA' to confirm:${NC}"
        read -r confirm
        if [[ "$confirm" != "DELETE ALL DATA" ]]; then
            echo -e "${GREEN}Aborted. No data was deleted.${NC}"
            exit 0
        fi
    else
        echo -e "${YELLOW}Auto-confirm enabled (--yes flag)${NC}"
    fi

    echo -e "${YELLOW}[0/8] Очистка существующих данных...${NC}"
    run_sql "
-- Очистка в правильном порядке (с учетом FK)
TRUNCATE TABLE
    lesson_homework,
    broadcast_files,
    lesson_broadcasts,
    cancelled_bookings,
    messages,
    file_attachments,
    blocked_messages,
    chat_rooms,
    swaps,
    credit_transactions,
    bookings,
    template_lesson_students,
    template_applications,
    template_lessons,
    lesson_templates,
    lessons,
    payments,
    sessions,
    credits,
    telegram_users,
    users
RESTART IDENTITY CASCADE;
"
    echo -e "${GREEN}✓ Data truncated${NC}"
else
    echo -e "${GREEN}[0/8] Skipping truncate (insert-only mode, use --truncate to clear data)${NC}"
fi

# 1. Создание пользователей
echo -e "${GREEN}[1/8] Создание пользователей...${NC}"
run_sql "
-- Пароль для всех: password123
-- Bcrypt hash: \$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy

-- Администратор
INSERT INTO users (id, email, password_hash, full_name, role) VALUES
('00000000-0000-0000-0000-000000000001', 'admin@tutoring.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Администратор Системы', 'admin');

-- Преподаватели
INSERT INTO users (id, email, password_hash, full_name, role) VALUES
('10000000-0000-0000-0000-000000000001', 'ivan.petrov@tutoring.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Иван Петров', 'teacher'),
('10000000-0000-0000-0000-000000000002', 'maria.sidorova@tutoring.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Мария Сидорова', 'teacher'),
('10000000-0000-0000-0000-000000000003', 'alexey.kozlov@tutoring.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Алексей Козлов', 'teacher');

-- Ученики (с payment_enabled)
INSERT INTO users (id, email, password_hash, full_name, role, payment_enabled) VALUES
('20000000-0000-0000-0000-000000000001', 'anna.ivanova@student.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Анна Иванова', 'student', TRUE),
('20000000-0000-0000-0000-000000000002', 'dmitry.smirnov@student.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Дмитрий Смирнов', 'student', TRUE),
('20000000-0000-0000-0000-000000000003', 'elena.volkova@student.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Елена Волкова', 'student', TRUE),
('20000000-0000-0000-0000-000000000004', 'pavel.morozov@student.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Павел Морозов', 'student', FALSE),
('20000000-0000-0000-0000-000000000005', 'olga.novikova@student.com', '\$2a\$10\$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy', 'Ольга Новикова', 'student', TRUE);
"

# 2. Создание кредитов для студентов
echo -e "${GREEN}[2/8] Настройка кредитов...${NC}"
run_sql "
-- Обновление балансов кредитов (триггер create_credits_on_student_insert уже создал записи с balance=0)
UPDATE credits SET balance = 10 WHERE user_id = '20000000-0000-0000-0000-000000000001'; -- Анна Иванова
UPDATE credits SET balance = 8 WHERE user_id = '20000000-0000-0000-0000-000000000002';  -- Дмитрий Смирнов
UPDATE credits SET balance = 12 WHERE user_id = '20000000-0000-0000-0000-000000000003'; -- Елена Волкова
UPDATE credits SET balance = 5 WHERE user_id = '20000000-0000-0000-0000-000000000004';  -- Павел Морозов
UPDATE credits SET balance = 3 WHERE user_id = '20000000-0000-0000-0000-000000000005';  -- Ольга Новикова
"

# 3. Создание занятий в календаре
echo -e "${GREEN}[3/8] Создание занятий...${NC}"
run_sql "
-- =============================================
-- ПРЕПОДАВАТЕЛЬ 1 (Иван Петров) - Математика
-- =============================================

-- ПРОШЛЫЕ ЗАНЯТИЯ С ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- Месяц назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '30 days' + TIME '10:00', NOW() - INTERVAL '30 days' + TIME '11:00', 1, 1, '#3B82F6', 'Математика', 'Решить задачи 1-10 из учебника. Учебник: https://math.ru/textbook', NOW() - INTERVAL '31 days', NOW()),
    -- 3 недели назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '21 days' + TIME '14:00', NOW() - INTERVAL '21 days' + TIME '15:30', 5, 4, '#10B981', 'Алгебра', 'Повторить формулы сокращённого умножения. Ссылка: https://math.ru/formulas', NOW() - INTERVAL '22 days', NOW()),
    -- 2 недели назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '14 days' + TIME '11:00', NOW() - INTERVAL '14 days' + TIME '12:00', 1, 1, '#F59E0B', 'Геометрия', 'Доказать теоремы 5-7 из главы 3', NOW() - INTERVAL '15 days', NOW()),
    -- Неделю назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '7 days' + TIME '15:00', NOW() - INTERVAL '7 days' + TIME '16:30', 6, 5, '#8B5CF6', 'Математика ЕГЭ', 'Решить варианты ЕГЭ 2024: https://ege.sdamgia.ru', NOW() - INTERVAL '8 days', NOW()),
    -- 5 дней назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '5 days' + TIME '09:00', NOW() - INTERVAL '5 days' + TIME '10:00', 1, 1, '#3B82F6', 'Тригонометрия', 'Выучить формулы приведения. Карточки: https://quizlet.com/trig', NOW() - INTERVAL '6 days', NOW()),
    -- 3 дня назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '3 days' + TIME '16:00', NOW() - INTERVAL '3 days' + TIME '17:30', 4, 3, '#EC4899', 'Стереометрия', 'Построить сечения многогранников (задачи 1-5)', NOW() - INTERVAL '4 days', NOW()),
    -- Вчера - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '1 day' + TIME '10:00', NOW() - INTERVAL '1 day' + TIME '11:00', 1, 1, '#10B981', 'Комбинаторика', 'Решить 15 задач на перестановки и сочетания', NOW() - INTERVAL '2 days', NOW()),
    -- 45 дней назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000030', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '45 days' + TIME '10:00', NOW() - INTERVAL '45 days' + TIME '11:30', 6, 5, '#3B82F6', 'Математика: повторение', 'Решить задачи на повторение за 10 класс', NOW() - INTERVAL '46 days', NOW()),
    -- 40 дней назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000031', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '40 days' + TIME '14:00', NOW() - INTERVAL '40 days' + TIME '15:00', 1, 1, '#F59E0B', 'Функции и графики', 'Построить графики функций 1-15', NOW() - INTERVAL '41 days', NOW()),
    -- 35 дней назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000032', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '35 days' + TIME '16:00', NOW() - INTERVAL '35 days' + TIME '17:30', 8, 6, '#8B5CF6', 'Уравнения и неравенства', 'Решить системы уравнений из сборника', NOW() - INTERVAL '36 days', NOW()),
    -- 33 дня назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000033', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '33 days' + TIME '09:00', NOW() - INTERVAL '33 days' + TIME '10:00', 1, 1, '#EC4899', 'Логарифмы', 'Вычислить логарифмы (задания 1-20)', NOW() - INTERVAL '34 days', NOW()),
    -- 27 дней назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000034', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '27 days' + TIME '11:00', NOW() - INTERVAL '27 days' + TIME '12:30', 5, 4, '#10B981', 'Производная', 'Найти производные функций из учебника', NOW() - INTERVAL '28 days', NOW()),
    -- 20 дней назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000035', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '20 days' + TIME '15:00', NOW() - INTERVAL '20 days' + TIME '16:00', 1, 1, '#3B82F6', 'Интегралы', 'Вычислить определённые интегралы', NOW() - INTERVAL '21 days', NOW()),
    -- 18 дней назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000036', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '18 days' + TIME '10:00', NOW() - INTERVAL '18 days' + TIME '11:30', 6, 5, '#F59E0B', 'Векторы', 'Решить задачи на векторы в пространстве', NOW() - INTERVAL '19 days', NOW()),
    -- 16 дней назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000037', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '16 days' + TIME '14:00', NOW() - INTERVAL '16 days' + TIME '15:00', 1, 1, '#8B5CF6', 'Планиметрия ЕГЭ', 'Решить задачи 16 из ЕГЭ', NOW() - INTERVAL '17 days', NOW()),
    -- 11 дней назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000038', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '11 days' + TIME '16:00', NOW() - INTERVAL '11 days' + TIME '17:30', 4, 3, '#EC4899', 'Стереометрия ЕГЭ', 'Решить задачи 14 из ЕГЭ', NOW() - INTERVAL '12 days', NOW()),
    -- 9 дней назад - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000039', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '9 days' + TIME '11:00', NOW() - INTERVAL '9 days' + TIME '12:00', 1, 1, '#10B981', 'Параметры', 'Решить уравнения с параметром', NOW() - INTERVAL '10 days', NOW()),
    -- 4 дня назад - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000040', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '4 days' + TIME '09:00', NOW() - INTERVAL '4 days' + TIME '10:30', 5, 4, '#3B82F6', 'Экономические задачи', 'Решить задачи на оптимизацию', NOW() - INTERVAL '5 days', NOW())
ON CONFLICT (id) DO NOTHING;

-- ПРОШЛЫЕ ЗАНЯТИЯ БЕЗ ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- Месяц назад - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '28 days' + TIME '14:00', NOW() - INTERVAL '28 days' + TIME '15:30', 8, 6, '#06B6D4', 'Введение в курс', NULL, NOW() - INTERVAL '29 days', NOW()),
    -- 2 недели назад - групповое без ДЗ (контрольная)
    ('d0000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '12 days' + TIME '10:00', NOW() - INTERVAL '12 days' + TIME '11:30', 5, 5, '#EF4444', 'Контрольная работа', NULL, NOW() - INTERVAL '13 days', NOW()),
    -- 6 дней назад - индивидуальное без ДЗ (консультация)
    ('d0000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000001', 
        NOW() - INTERVAL '6 days' + TIME '17:00', NOW() - INTERVAL '6 days' + TIME '18:00', 1, 1, '#A855F7', 'Консультация перед экзаменом', NULL, NOW() - INTERVAL '7 days', NOW()),
    -- 2 дня назад - групповое без ДЗ (разбор ошибок)
    ('d0000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '2 days' + TIME '11:00', NOW() - INTERVAL '2 days' + TIME '12:30', 6, 4, '#F97316', 'Разбор ошибок контрольной', NULL, NOW() - INTERVAL '3 days', NOW()),
    -- 50 дней назад - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000041', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '50 days' + TIME '10:00', NOW() - INTERVAL '50 days' + TIME '11:30', 10, 8, '#06B6D4', 'Вводное занятие', NULL, NOW() - INTERVAL '51 days', NOW()),
    -- 42 дня назад - индивидуальное без ДЗ
    ('d0000000-0000-0000-0000-000000000042', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '42 days' + TIME '15:00', NOW() - INTERVAL '42 days' + TIME '16:00', 1, 1, '#A855F7', 'Диагностика знаний', NULL, NOW() - INTERVAL '43 days', NOW()),
    -- 38 дней назад - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000043', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '38 days' + TIME '14:00', NOW() - INTERVAL '38 days' + TIME '15:30', 6, 5, '#EF4444', 'Пробное тестирование', NULL, NOW() - INTERVAL '39 days', NOW()),
    -- 25 дней назад - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000044', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '25 days' + TIME '16:00', NOW() - INTERVAL '25 days' + TIME '17:30', 8, 7, '#F97316', 'Промежуточное тестирование', NULL, NOW() - INTERVAL '26 days', NOW()),
    -- 22 дня назад - индивидуальное без ДЗ
    ('d0000000-0000-0000-0000-000000000045', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '22 days' + TIME '11:00', NOW() - INTERVAL '22 days' + TIME '12:00', 1, 1, '#06B6D4', 'Разбор ошибок', NULL, NOW() - INTERVAL '23 days', NOW()),
    -- 17 дней назад - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000046', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '17 days' + TIME '09:00', NOW() - INTERVAL '17 days' + TIME '10:30', 5, 4, '#A855F7', 'Мастер-класс по решению задач', NULL, NOW() - INTERVAL '18 days', NOW()),
    -- 15 дней назад - индивидуальное без ДЗ
    ('d0000000-0000-0000-0000-000000000047', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '15 days' + TIME '17:00', NOW() - INTERVAL '15 days' + TIME '18:00', 1, 1, '#EF4444', 'Консультация по олимпиаде', NULL, NOW() - INTERVAL '16 days', NOW()),
    -- 8 дней назад - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000048', '10000000-0000-0000-0000-000000000001',
        NOW() - INTERVAL '8 days' + TIME '14:00', NOW() - INTERVAL '8 days' + TIME '15:30', 6, 5, '#F97316', 'Итоговое повторение', NULL, NOW() - INTERVAL '9 days', NOW())
ON CONFLICT (id) DO NOTHING;

-- БУДУЩИЕ ЗАНЯТИЯ С ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- Завтра - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000001', 
        NOW() + INTERVAL '1 day' + TIME '09:00', NOW() + INTERVAL '1 day' + TIME '10:00', 1, 1, '#3B82F6', 'Геометрия', 'Повторить теоремы о подобии треугольников', NOW() - INTERVAL '3 days', NOW()),
    -- Через 2 дня - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000001', 
        NOW() + INTERVAL '2 days' + TIME '15:00', NOW() + INTERVAL '2 days' + TIME '16:30', 4, 2, '#8B5CF6', 'Математика ЕГЭ', 'Прорешать профильный вариант: https://ege.sdamgia.ru/test?id=54321', NOW() - INTERVAL '2 days', NOW()),
    -- Через 4 дня - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000001', 
        NOW() + INTERVAL '4 days' + TIME '11:00', NOW() + INTERVAL '4 days' + TIME '12:00', 1, 0, '#F59E0B', 'Олимпиадная математика', 'Решить задачи Всероссийской олимпиады 2023', NOW() - INTERVAL '1 day', NOW()),
    -- Через неделю - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000001', 
        NOW() + INTERVAL '7 days' + TIME '10:00', NOW() + INTERVAL '7 days' + TIME '11:30', 6, 3, '#EC4899', 'Подготовка к ОГЭ', 'Выполнить тренировочный вариант ОГЭ', NOW(), NOW()),
    -- Через 10 дней - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000001', 
        NOW() + INTERVAL '10 days' + TIME '14:00', NOW() + INTERVAL '10 days' + TIME '15:00', 1, 1, '#10B981', 'Теория вероятностей', 'Изучить главу 7 учебника', NOW(), NOW()),
    -- Через 2 недели - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '14 days' + TIME '16:00', NOW() + INTERVAL '14 days' + TIME '17:30', 5, 0, '#3B82F6', 'Математический анализ', 'Вычислить пределы функций (задачи 1-20)', NOW(), NOW()),
    -- Через 16 дней - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000050', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '16 days' + TIME '10:00', NOW() + INTERVAL '16 days' + TIME '11:00', 1, 0, '#F59E0B', 'Числовые последовательности', 'Решить задачи на арифметические и геометрические прогрессии', NOW(), NOW()),
    -- Через 18 дней - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000051', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '18 days' + TIME '15:00', NOW() + INTERVAL '18 days' + TIME '16:30', 6, 0, '#8B5CF6', 'Комплексные числа', 'Изучить операции с комплексными числами', NOW(), NOW()),
    -- Через 20 дней - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000052', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '20 days' + TIME '11:00', NOW() + INTERVAL '20 days' + TIME '12:00', 1, 0, '#EC4899', 'Теория чисел', 'Решить задачи на делимость', NOW(), NOW()),
    -- Через 22 дня - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000053', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '22 days' + TIME '14:00', NOW() + INTERVAL '22 days' + TIME '15:30', 8, 0, '#10B981', 'Тригонометрические уравнения', 'Решить уравнения разных типов', NOW(), NOW()),
    -- Через 25 дней - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000054', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '25 days' + TIME '09:00', NOW() + INTERVAL '25 days' + TIME '10:00', 1, 0, '#3B82F6', 'Показательные уравнения', 'Решить показательные уравнения и неравенства', NOW(), NOW()),
    -- Через 28 дней - групповое с ДЗ
    ('d0000000-0000-0000-0000-000000000055', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '28 days' + TIME '16:00', NOW() + INTERVAL '28 days' + TIME '17:30', 5, 0, '#F59E0B', 'Логарифмические уравнения', 'Решить логарифмические уравнения', NOW(), NOW()),
    -- Через 30 дней - индивидуальное с ДЗ
    ('d0000000-0000-0000-0000-000000000056', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '30 days' + TIME '10:00', NOW() + INTERVAL '30 days' + TIME '11:00', 1, 0, '#8B5CF6', 'Иррациональные уравнения', 'Решить иррациональные уравнения', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- БУДУЩИЕ ЗАНЯТИЯ БЕЗ ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- Через 3 дня - индивидуальное без ДЗ (пробный экзамен)
    ('d0000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '3 days' + TIME '10:00', NOW() + INTERVAL '3 days' + TIME '11:30', 1, 1, '#EF4444', 'Пробный экзамен', NULL, NOW() - INTERVAL '1 day', NOW()),
    -- Через 5 дней - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000001', 
        NOW() + INTERVAL '5 days' + TIME '14:00', NOW() + INTERVAL '5 days' + TIME '15:30', 8, 0, '#06B6D4', 'Открытый урок', NULL, NOW(), NOW()),
    -- Через 8 дней - индивидуальное без ДЗ
    ('d0000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000001', 
        NOW() + INTERVAL '8 days' + TIME '17:00', NOW() + INTERVAL '8 days' + TIME '18:00', 1, 0, '#A855F7', 'Консультация', NULL, NOW(), NOW()),
    -- Через 12 дней - групповое без ДЗ (итоговое занятие)
    ('d0000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '12 days' + TIME '11:00', NOW() + INTERVAL '12 days' + TIME '12:30', 10, 0, '#F97316', 'Итоговое занятие семестра', NULL, NOW(), NOW()),
    -- Через 15 дней - индивидуальное без ДЗ
    ('d0000000-0000-0000-0000-000000000060', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '15 days' + TIME '09:00', NOW() + INTERVAL '15 days' + TIME '10:00', 1, 0, '#06B6D4', 'Индивидуальная консультация', NULL, NOW(), NOW()),
    -- Через 17 дней - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000061', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '17 days' + TIME '14:00', NOW() + INTERVAL '17 days' + TIME '15:30', 8, 0, '#A855F7', 'Математический турнир', NULL, NOW(), NOW()),
    -- Через 19 дней - индивидуальное без ДЗ
    ('d0000000-0000-0000-0000-000000000062', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '19 days' + TIME '16:00', NOW() + INTERVAL '19 days' + TIME '17:00', 1, 0, '#EF4444', 'Разбор олимпиадных задач', NULL, NOW(), NOW()),
    -- Через 21 день - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000063', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '21 days' + TIME '10:00', NOW() + INTERVAL '21 days' + TIME '11:30', 6, 0, '#06B6D4', 'Мозговой штурм', NULL, NOW(), NOW()),
    -- Через 24 дня - индивидуальное без ДЗ
    ('d0000000-0000-0000-0000-000000000064', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '24 days' + TIME '11:00', NOW() + INTERVAL '24 days' + TIME '12:00', 1, 0, '#F97316', 'Финальная консультация', NULL, NOW(), NOW()),
    -- Через 27 дней - групповое без ДЗ
    ('d0000000-0000-0000-0000-000000000065', '10000000-0000-0000-0000-000000000001',
        NOW() + INTERVAL '27 days' + TIME '15:00', NOW() + INTERVAL '27 days' + TIME '16:30', 10, 0, '#A855F7', 'Пробный ЕГЭ полный вариант', NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================
-- ПРЕПОДАВАТЕЛЬ 2 (Мария Сидорова) - Физика/Информатика
-- =============================================

-- ПРОШЛЫЕ ЗАНЯТИЯ С ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- Месяц назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 
        NOW() - INTERVAL '30 days' + TIME '16:00', NOW() - INTERVAL '30 days' + TIME '17:30', 8, 7, '#06B6D4', 'Механика', 'Решить задачи на законы Ньютона. Видео: https://youtube.com/physics101', NOW() - INTERVAL '31 days', NOW()),
    -- 3 недели назад - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 
        NOW() - INTERVAL '21 days' + TIME '10:00', NOW() - INTERVAL '21 days' + TIME '11:00', 1, 1, '#14B8A6', 'Python основы', 'Написать программу калькулятор на Python', NOW() - INTERVAL '22 days', NOW()),
    -- 2 недели назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 
        NOW() - INTERVAL '14 days' + TIME '14:00', NOW() - INTERVAL '14 days' + TIME '15:30', 6, 5, '#22C55E', 'Термодинамика', 'Прочитать параграфы 20-25. Решить задачи 1-8', NOW() - INTERVAL '15 days', NOW()),
    -- Неделю назад - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', 
        NOW() - INTERVAL '7 days' + TIME '12:00', NOW() - INTERVAL '7 days' + TIME '13:00', 1, 1, '#14B8A6', 'Алгоритмы Python', 'Реализовать сортировку пузырьком и быструю сортировку', NOW() - INTERVAL '8 days', NOW()),
    -- 4 дня назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', 
        NOW() - INTERVAL '4 days' + TIME '16:00', NOW() - INTERVAL '4 days' + TIME '17:30', 8, 5, '#06B6D4', 'Физика ЕГЭ', 'Решить варианты ЕГЭ по физике 2024', NOW() - INTERVAL '5 days', NOW()),
    -- Вчера - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '1 day' + TIME '11:00', NOW() - INTERVAL '1 day' + TIME '12:00', 1, 1, '#22C55E', 'Веб-разработка', 'Создать HTML-страницу с формой', NOW() - INTERVAL '2 days', NOW()),
    -- 45 дней назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000030', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '45 days' + TIME '14:00', NOW() - INTERVAL '45 days' + TIME '15:30', 8, 6, '#06B6D4', 'Кинематика', 'Решить задачи на равномерное движение', NOW() - INTERVAL '46 days', NOW()),
    -- 40 дней назад - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000031', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '40 days' + TIME '10:00', NOW() - INTERVAL '40 days' + TIME '11:00', 1, 1, '#14B8A6', 'Основы Git', 'Создать репозиторий и сделать 5 коммитов', NOW() - INTERVAL '41 days', NOW()),
    -- 35 дней назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000032', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '35 days' + TIME '16:00', NOW() - INTERVAL '35 days' + TIME '17:30', 6, 5, '#22C55E', 'Динамика', 'Решить задачи на законы сохранения', NOW() - INTERVAL '36 days', NOW()),
    -- 32 дня назад - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000033', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '32 days' + TIME '12:00', NOW() - INTERVAL '32 days' + TIME '13:00', 1, 1, '#14B8A6', 'Структуры данных', 'Реализовать стек и очередь на Python', NOW() - INTERVAL '33 days', NOW()),
    -- 28 дней назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000034', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '28 days' + TIME '15:00', NOW() - INTERVAL '28 days' + TIME '16:30', 8, 7, '#06B6D4', 'Электростатика', 'Решить задачи на электрическое поле', NOW() - INTERVAL '29 days', NOW()),
    -- 24 дня назад - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000035', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '24 days' + TIME '11:00', NOW() - INTERVAL '24 days' + TIME '12:00', 1, 1, '#22C55E', 'REST API', 'Создать простой REST API на Flask', NOW() - INTERVAL '25 days', NOW()),
    -- 19 дней назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000036', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '19 days' + TIME '14:00', NOW() - INTERVAL '19 days' + TIME '15:30', 6, 5, '#14B8A6', 'Магнетизм', 'Решить задачи на магнитное поле', NOW() - INTERVAL '20 days', NOW()),
    -- 16 дней назад - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000037', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '16 days' + TIME '10:00', NOW() - INTERVAL '16 days' + TIME '11:00', 1, 1, '#06B6D4', 'React основы', 'Создать компонент TodoList', NOW() - INTERVAL '17 days', NOW()),
    -- 12 дней назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000038', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '12 days' + TIME '16:00', NOW() - INTERVAL '12 days' + TIME '17:30', 8, 6, '#22C55E', 'Колебания и волны', 'Решить задачи на гармонические колебания', NOW() - INTERVAL '13 days', NOW()),
    -- 9 дней назад - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000039', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '9 days' + TIME '12:00', NOW() - INTERVAL '9 days' + TIME '13:00', 1, 1, '#14B8A6', 'TypeScript', 'Переписать проект на TypeScript', NOW() - INTERVAL '10 days', NOW()),
    -- 6 дней назад - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000040', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '6 days' + TIME '15:00', NOW() - INTERVAL '6 days' + TIME '16:30', 6, 4, '#06B6D4', 'Квантовая физика', 'Решить задачи на фотоэффект', NOW() - INTERVAL '7 days', NOW())
ON CONFLICT (id) DO NOTHING;

-- ПРОШЛЫЕ ЗАНЯТИЯ БЕЗ ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- 25 дней назад - групповое без ДЗ (вводное)
    ('e0000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '25 days' + TIME '10:00', NOW() - INTERVAL '25 days' + TIME '11:30', 10, 8, '#A855F7', 'Введение в программирование', NULL, NOW() - INTERVAL '26 days', NOW()),
    -- 10 дней назад - индивидуальное без ДЗ (консультация)
    ('e0000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000002', 
        NOW() - INTERVAL '10 days' + TIME '17:00', NOW() - INTERVAL '10 days' + TIME '18:00', 1, 1, '#F97316', 'Консультация по проекту', NULL, NOW() - INTERVAL '11 days', NOW()),
    -- 5 дней назад - групповое без ДЗ (лабораторная)
    ('e0000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000002', 
        NOW() - INTERVAL '5 days' + TIME '14:00', NOW() - INTERVAL '5 days' + TIME '16:00', 6, 6, '#EF4444', 'Лабораторная работа', NULL, NOW() - INTERVAL '6 days', NOW()),
    -- 2 дня назад - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '2 days' + TIME '10:00', NOW() - INTERVAL '2 days' + TIME '11:30', 5, 4, '#06B6D4', 'Разбор лабораторной', NULL, NOW() - INTERVAL '3 days', NOW()),
    -- 50 дней назад - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000041', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '50 days' + TIME '10:00', NOW() - INTERVAL '50 days' + TIME '11:30', 10, 9, '#A855F7', 'Знакомство с курсом физики', NULL, NOW() - INTERVAL '51 days', NOW()),
    -- 43 дня назад - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000042', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '43 days' + TIME '16:00', NOW() - INTERVAL '43 days' + TIME '17:00', 1, 1, '#F97316', 'Входное тестирование', NULL, NOW() - INTERVAL '44 days', NOW()),
    -- 37 дней назад - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000043', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '37 days' + TIME '14:00', NOW() - INTERVAL '37 days' + TIME '15:30', 8, 7, '#EF4444', 'Демонстрационные опыты', NULL, NOW() - INTERVAL '38 days', NOW()),
    -- 30 дней назад - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000044', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '30 days' + TIME '11:00', NOW() - INTERVAL '30 days' + TIME '12:00', 1, 1, '#A855F7', 'Разбор задач', NULL, NOW() - INTERVAL '31 days', NOW()),
    -- 23 дня назад - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000045', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '23 days' + TIME '15:00', NOW() - INTERVAL '23 days' + TIME '16:30', 6, 5, '#F97316', 'Практикум по программированию', NULL, NOW() - INTERVAL '24 days', NOW()),
    -- 18 дней назад - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000046', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '18 days' + TIME '17:00', NOW() - INTERVAL '18 days' + TIME '18:00', 1, 1, '#EF4444', 'Code review', NULL, NOW() - INTERVAL '19 days', NOW()),
    -- 13 дней назад - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000047', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '13 days' + TIME '10:00', NOW() - INTERVAL '13 days' + TIME '11:30', 8, 6, '#A855F7', 'Хакатон мини', NULL, NOW() - INTERVAL '14 days', NOW()),
    -- 8 дней назад - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000048', '10000000-0000-0000-0000-000000000002',
        NOW() - INTERVAL '8 days' + TIME '14:00', NOW() - INTERVAL '8 days' + TIME '15:00', 1, 1, '#F97316', 'Презентация проекта', NULL, NOW() - INTERVAL '9 days', NOW())
ON CONFLICT (id) DO NOTHING;

-- БУДУЩИЕ ЗАНЯТИЯ С ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- Завтра - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000002', 
        NOW() + INTERVAL '1 day' + TIME '12:00', NOW() + INTERVAL '1 day' + TIME '13:00', 1, 1, '#14B8A6', 'Python ООП', 'Создать класс Student с методами', NOW() - INTERVAL '2 days', NOW()),
    -- Через 2 дня - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000002', 
        NOW() + INTERVAL '2 days' + TIME '10:00', NOW() + INTERVAL '2 days' + TIME '11:30', 10, 4, '#06B6D4', 'Электродинамика', 'Решить задачи на закон Кулона', NOW() - INTERVAL '1 day', NOW()),
    -- Через 4 дня - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000002', 
        NOW() + INTERVAL '4 days' + TIME '15:00', NOW() + INTERVAL '4 days' + TIME '16:00', 1, 1, '#22C55E', 'Базы данных SQL', 'Написать 10 SQL-запросов к учебной БД', NOW(), NOW()),
    -- Через неделю - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000002', 
        NOW() + INTERVAL '7 days' + TIME '16:00', NOW() + INTERVAL '7 days' + TIME '17:30', 8, 0, '#06B6D4', 'Оптика', 'Подготовить доклад по теме \"Дифракция света\"', NOW(), NOW()),
    -- Через 9 дней - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000002', 
        NOW() + INTERVAL '9 days' + TIME '11:00', NOW() + INTERVAL '9 days' + TIME '12:00', 1, 0, '#14B8A6', 'Django Framework', 'Создать простое веб-приложение на Django', NOW(), NOW()),
    -- Через 2 недели - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '14 days' + TIME '14:00', NOW() + INTERVAL '14 days' + TIME '15:30', 6, 0, '#22C55E', 'Физика ЕГЭ интенсив', 'Решить 3 полных варианта ЕГЭ', NOW(), NOW()),
    -- Через 16 дней - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000050', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '16 days' + TIME '10:00', NOW() + INTERVAL '16 days' + TIME '11:00', 1, 0, '#14B8A6', 'Node.js', 'Создать простой сервер на Express', NOW(), NOW()),
    -- Через 18 дней - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000051', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '18 days' + TIME '15:00', NOW() + INTERVAL '18 days' + TIME '16:30', 8, 0, '#06B6D4', 'Атомная физика', 'Решить задачи на строение атома', NOW(), NOW()),
    -- Через 20 дней - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000052', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '20 days' + TIME '12:00', NOW() + INTERVAL '20 days' + TIME '13:00', 1, 0, '#22C55E', 'Docker основы', 'Контейнеризировать простое приложение', NOW(), NOW()),
    -- Через 22 дня - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000053', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '22 days' + TIME '14:00', NOW() + INTERVAL '22 days' + TIME '15:30', 6, 0, '#14B8A6', 'Ядерная физика', 'Решить задачи на радиоактивный распад', NOW(), NOW()),
    -- Через 25 дней - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000054', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '25 days' + TIME '11:00', NOW() + INTERVAL '25 days' + TIME '12:00', 1, 0, '#06B6D4', 'CI/CD', 'Настроить GitHub Actions для проекта', NOW(), NOW()),
    -- Через 28 дней - групповое с ДЗ
    ('e0000000-0000-0000-0000-000000000055', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '28 days' + TIME '16:00', NOW() + INTERVAL '28 days' + TIME '17:30', 8, 0, '#22C55E', 'Физика: повторение', 'Решить комплексные задачи по всему курсу', NOW(), NOW()),
    -- Через 30 дней - индивидуальное с ДЗ
    ('e0000000-0000-0000-0000-000000000056', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '30 days' + TIME '10:00', NOW() + INTERVAL '30 days' + TIME '11:00', 1, 0, '#14B8A6', 'Финальный проект', 'Завершить и задокументировать проект', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- БУДУЩИЕ ЗАНЯТИЯ БЕЗ ДЗ
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, subject, homework_text, created_at, updated_at) VALUES
    -- Через 3 дня - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '3 days' + TIME '10:00', NOW() + INTERVAL '3 days' + TIME '11:30', 5, 2, '#EF4444', 'Робототехника', NULL, NOW() - INTERVAL '1 day', NOW()),
    -- Через 5 дней - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000002', 
        NOW() + INTERVAL '5 days' + TIME '17:00', NOW() + INTERVAL '5 days' + TIME '18:00', 1, 0, '#A855F7', 'Разбор проекта', NULL, NOW(), NOW()),
    -- Через 6 дней - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000002', 
        NOW() + INTERVAL '6 days' + TIME '14:00', NOW() + INTERVAL '6 days' + TIME '16:00', 8, 0, '#F97316', 'Хакатон подготовка', NULL, NOW(), NOW()),
    -- Через 11 дней - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '11 days' + TIME '12:00', NOW() + INTERVAL '11 days' + TIME '13:00', 1, 0, '#06B6D4', 'Консультация по олимпиаде', NULL, NOW(), NOW()),
    -- Через 13 дней - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000060', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '13 days' + TIME '14:00', NOW() + INTERVAL '13 days' + TIME '15:30', 8, 0, '#EF4444', 'Физический практикум', NULL, NOW(), NOW()),
    -- Через 15 дней - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000061', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '15 days' + TIME '16:00', NOW() + INTERVAL '15 days' + TIME '17:00', 1, 0, '#A855F7', 'Ревью кода', NULL, NOW(), NOW()),
    -- Через 17 дней - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000062', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '17 days' + TIME '10:00', NOW() + INTERVAL '17 days' + TIME '11:30', 6, 0, '#F97316', 'IT-викторина', NULL, NOW(), NOW()),
    -- Через 19 дней - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000063', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '19 days' + TIME '11:00', NOW() + INTERVAL '19 days' + TIME '12:00', 1, 0, '#EF4444', 'Разбор ошибок', NULL, NOW(), NOW()),
    -- Через 21 день - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000064', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '21 days' + TIME '15:00', NOW() + INTERVAL '21 days' + TIME '16:30', 10, 0, '#A855F7', 'Демо-день проектов', NULL, NOW(), NOW()),
    -- Через 24 дня - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000065', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '24 days' + TIME '14:00', NOW() + INTERVAL '24 days' + TIME '15:00', 1, 0, '#F97316', 'Карьерная консультация', NULL, NOW(), NOW()),
    -- Через 26 дней - групповое без ДЗ
    ('e0000000-0000-0000-0000-000000000066', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '26 days' + TIME '10:00', NOW() + INTERVAL '26 days' + TIME '12:00', 8, 0, '#EF4444', 'Пробный ЕГЭ физика', NULL, NOW(), NOW()),
    -- Через 29 дней - индивидуальное без ДЗ
    ('e0000000-0000-0000-0000-000000000067', '10000000-0000-0000-0000-000000000002',
        NOW() + INTERVAL '29 days' + TIME '16:00', NOW() + INTERVAL '29 days' + TIME '17:00', 1, 0, '#A855F7', 'Итоговая консультация', NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
"

# 4. Создание бронирований
echo -e "${GREEN}[4/8] Создание бронирований...${NC}"
run_sql "
-- =============================================
-- БРОНИРОВАНИЯ НА ПРОШЕДШИЕ ЗАНЯТИЯ
-- =============================================

-- Преподаватель 1 - прошлые занятия
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, created_at, updated_at) VALUES
    -- d001 - индивидуальное (месяц назад)
    ('f0000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '31 days', NOW(), NOW()),
    -- d002 - групповое (3 недели назад) - 4 студента
    ('f0000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000002', 'active', NOW() - INTERVAL '22 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000002', 'active', NOW() - INTERVAL '22 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000002', 'active', NOW() - INTERVAL '22 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000002', 'active', NOW() - INTERVAL '22 days', NOW(), NOW()),
    -- d003 - индивидуальное (2 недели назад)
    ('f0000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '15 days', NOW(), NOW()),
    -- d004 - групповое (неделю назад) - 5 студентов
    ('f0000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000004', 'active', NOW() - INTERVAL '8 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000004', 'active', NOW() - INTERVAL '8 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000009', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000004', 'active', NOW() - INTERVAL '8 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000010', '20000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000004', 'active', NOW() - INTERVAL '8 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000011', '20000000-0000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000004', 'active', NOW() - INTERVAL '8 days', NOW(), NOW()),
    -- d005 - индивидуальное (5 дней назад)
    ('f0000000-0000-0000-0000-000000000012', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000005', 'active', NOW() - INTERVAL '6 days', NOW(), NOW()),
    -- d006 - групповое (3 дня назад) - 3 студента
    ('f0000000-0000-0000-0000-000000000013', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000006', 'active', NOW() - INTERVAL '4 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000014', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000006', 'active', NOW() - INTERVAL '4 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000015', '20000000-0000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000006', 'active', NOW() - INTERVAL '4 days', NOW(), NOW()),
    -- d007 - индивидуальное (вчера)
    ('f0000000-0000-0000-0000-000000000016', '20000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000007', 'active', NOW() - INTERVAL '2 days', NOW(), NOW()),
    -- d008-d011 - прошлые без ДЗ
    ('f0000000-0000-0000-0000-000000000017', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000008', 'active', NOW() - INTERVAL '29 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000018', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000008', 'active', NOW() - INTERVAL '29 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000019', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000008', 'active', NOW() - INTERVAL '29 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000020', '20000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000008', 'active', NOW() - INTERVAL '29 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000021', '20000000-0000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000008', 'active', NOW() - INTERVAL '29 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000022', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000009', 'active', NOW() - INTERVAL '13 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000023', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000009', 'active', NOW() - INTERVAL '13 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000024', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000009', 'active', NOW() - INTERVAL '13 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000025', '20000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000009', 'active', NOW() - INTERVAL '13 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000026', '20000000-0000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000009', 'active', NOW() - INTERVAL '13 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000027', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000010', 'active', NOW() - INTERVAL '7 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000028', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000011', 'active', NOW() - INTERVAL '3 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000029', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000011', 'active', NOW() - INTERVAL '3 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000030', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000011', 'active', NOW() - INTERVAL '3 days', NOW(), NOW()),
    ('f0000000-0000-0000-0000-000000000031', '20000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000011', 'active', NOW() - INTERVAL '3 days', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Преподаватель 2 - прошлые занятия
INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, created_at, updated_at) VALUES
    -- e001 - групповое (месяц назад) - 7 студентов
    ('f0000000-0000-0000-0001-000000000001', '20000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '31 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000002', '20000000-0000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '31 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000003', '20000000-0000-0000-0000-000000000003', 'e0000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '31 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000004', '20000000-0000-0000-0000-000000000004', 'e0000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '31 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000005', '20000000-0000-0000-0000-000000000005', 'e0000000-0000-0000-0000-000000000001', 'active', NOW() - INTERVAL '31 days', NOW(), NOW()),
    -- e002 - индивидуальное
    ('f0000000-0000-0000-0001-000000000006', '20000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000002', 'active', NOW() - INTERVAL '22 days', NOW(), NOW()),
    -- e003 - групповое - 5 студентов
    ('f0000000-0000-0000-0001-000000000007', '20000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '15 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000008', '20000000-0000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '15 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000009', '20000000-0000-0000-0000-000000000003', 'e0000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '15 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000010', '20000000-0000-0000-0000-000000000004', 'e0000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '15 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000011', '20000000-0000-0000-0000-000000000005', 'e0000000-0000-0000-0000-000000000003', 'active', NOW() - INTERVAL '15 days', NOW(), NOW()),
    -- e004 - индивидуальное
    ('f0000000-0000-0000-0001-000000000012', '20000000-0000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000004', 'active', NOW() - INTERVAL '8 days', NOW(), NOW()),
    -- e005 - групповое - 5 студентов
    ('f0000000-0000-0000-0001-000000000013', '20000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000005', 'active', NOW() - INTERVAL '5 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000014', '20000000-0000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000005', 'active', NOW() - INTERVAL '5 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000015', '20000000-0000-0000-0000-000000000003', 'e0000000-0000-0000-0000-000000000005', 'active', NOW() - INTERVAL '5 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000016', '20000000-0000-0000-0000-000000000004', 'e0000000-0000-0000-0000-000000000005', 'active', NOW() - INTERVAL '5 days', NOW(), NOW()),
    ('f0000000-0000-0000-0001-000000000017', '20000000-0000-0000-0000-000000000005', 'e0000000-0000-0000-0000-000000000005', 'active', NOW() - INTERVAL '5 days', NOW(), NOW()),
    -- e006 - индивидуальное
    ('f0000000-0000-0000-0001-000000000018', '20000000-0000-0000-0000-000000000003', 'e0000000-0000-0000-0000-000000000006', 'active', NOW() - INTERVAL '2 days', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================
-- БРОНИРОВАНИЯ НА БУДУЩИЕ ЗАНЯТИЯ
-- =============================================

INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, created_at, updated_at) VALUES
    -- d012 - завтра индивидуальное
    ('f0000000-0000-0000-0002-000000000001', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000012', 'active', NOW() - INTERVAL '3 days', NOW(), NOW()),
    -- d013 - через 2 дня групповое - 2 студента
    ('f0000000-0000-0000-0002-000000000002', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000013', 'active', NOW() - INTERVAL '2 days', NOW(), NOW()),
    ('f0000000-0000-0000-0002-000000000003', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000013', 'active', NOW() - INTERVAL '2 days', NOW(), NOW()),
    -- d015 - через неделю групповое - 3 студента
    ('f0000000-0000-0000-0002-000000000004', '20000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000015', 'active', NOW() - INTERVAL '1 day', NOW(), NOW()),
    ('f0000000-0000-0000-0002-000000000005', '20000000-0000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000015', 'active', NOW() - INTERVAL '1 day', NOW(), NOW()),
    ('f0000000-0000-0000-0002-000000000006', '20000000-0000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000015', 'active', NOW(), NOW(), NOW()),
    -- d016 - через 10 дней индивидуальное
    ('f0000000-0000-0000-0002-000000000007', '20000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000016', 'active', NOW(), NOW(), NOW()),
    -- d018 - через 3 дня пробный экзамен
    ('f0000000-0000-0000-0002-000000000008', '20000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000018', 'active', NOW() - INTERVAL '1 day', NOW(), NOW()),

    -- e011 - завтра индивидуальное
    ('f0000000-0000-0000-0002-000000000009', '20000000-0000-0000-0000-000000000004', 'e0000000-0000-0000-0000-000000000011', 'active', NOW() - INTERVAL '2 days', NOW(), NOW()),
    -- e012 - через 2 дня групповое - 4 студента
    ('f0000000-0000-0000-0002-000000000010', '20000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000012', 'active', NOW() - INTERVAL '1 day', NOW(), NOW()),
    ('f0000000-0000-0000-0002-000000000011', '20000000-0000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000012', 'active', NOW() - INTERVAL '1 day', NOW(), NOW()),
    ('f0000000-0000-0000-0002-000000000012', '20000000-0000-0000-0000-000000000003', 'e0000000-0000-0000-0000-000000000012', 'active', NOW() - INTERVAL '1 day', NOW(), NOW()),
    ('f0000000-0000-0000-0002-000000000013', '20000000-0000-0000-0000-000000000005', 'e0000000-0000-0000-0000-000000000012', 'active', NOW(), NOW(), NOW()),
    -- e013 - через 4 дня индивидуальное
    ('f0000000-0000-0000-0002-000000000014', '20000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000013', 'active', NOW(), NOW(), NOW()),
    -- e017 - через 3 дня робототехника - 2 студента
    ('f0000000-0000-0000-0002-000000000015', '20000000-0000-0000-0000-000000000004', 'e0000000-0000-0000-0000-000000000017', 'active', NOW() - INTERVAL '1 day', NOW(), NOW()),
    ('f0000000-0000-0000-0002-000000000016', '20000000-0000-0000-0000-000000000005', 'e0000000-0000-0000-0000-000000000017', 'active', NOW(), NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Обновляем current_students для всех занятий
UPDATE lessons SET current_students = (
    SELECT COUNT(*) FROM bookings WHERE bookings.lesson_id = lessons.id AND status = 'active'
);
"

# 5. Создание шаблонов расписания
echo -e "${GREEN}[5/8] Создание шаблонов...${NC}"
run_sql "
-- =============================================
-- ШАБЛОНЫ РАСПИСАНИЯ (разные варианты)
-- =============================================
INSERT INTO lesson_templates (id, admin_id, name, description, created_at, updated_at) VALUES
    -- Основные шаблоны
    ('10000000-1000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Основное расписание', 'Стандартное еженедельное расписание занятий на учебный год', NOW() - INTERVAL '60 days', NOW()),
    ('10000000-1000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Летние каникулы', 'Облегчённое расписание на летний период (июнь-август)', NOW() - INTERVAL '30 days', NOW()),
    ('10000000-1000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Подготовка к ЕГЭ', 'Интенсивное расписание для подготовки к ЕГЭ (май-июнь)', NOW() - INTERVAL '20 days', NOW()),
    -- Дополнительные шаблоны
    ('10000000-1000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'Подготовка к ОГЭ', 'Расписание для 9 классов перед экзаменами', NOW() - INTERVAL '15 days', NOW()),
    ('10000000-1000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'Олимпиадная подготовка', 'Расписание для подготовки к олимпиадам (октябрь-февраль)', NOW() - INTERVAL '10 days', NOW()),
    ('10000000-1000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', 'Минимальное расписание', 'Сокращённое расписание на праздничные недели', NOW() - INTERVAL '5 days', NOW()),
    ('10000000-1000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', 'Выходной режим', 'Только индивидуальные занятия по субботам', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================
-- ЗАНЯТИЯ В ШАБЛОНАХ (day_of_week: 0=Вс, 1=Пн, 2=Вт, 3=Ср, 4=Чт, 5=Пт, 6=Сб)
-- =============================================

-- ШАБЛОН 1: Основное расписание (полная неделя)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students, color, subject, description, created_at, updated_at) VALUES
    -- Понедельник - Преподаватель 1
    ('20000000-2000-0000-0000-000000000001', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 1, '09:00', '10:00', 'individual', 1, '#3B82F6', 'Математика', 'Индивидуальные занятия по математике', NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000002', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 1, '10:30', '12:00', 'group', 5, '#10B981', 'Алгебра', 'Групповые занятия по алгебре', NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000003', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 1, '15:00', '16:00', 'individual', 1, '#F59E0B', 'Геометрия', NULL, NOW(), NOW()),
    -- Вторник - Преподаватель 2
    ('20000000-2000-0000-0000-000000000004', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 2, '10:00', '11:30', 'group', 8, '#06B6D4', 'Физика', 'Групповые занятия по физике', NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000005', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 2, '14:00', '15:00', 'individual', 1, '#14B8A6', 'Python', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000006', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 2, '16:00', '17:30', 'group', 6, '#22C55E', 'Информатика', NULL, NOW(), NOW()),
    -- Среда - Преподаватель 1
    ('20000000-2000-0000-0000-000000000007', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 3, '11:00', '12:00', 'individual', 1, '#8B5CF6', 'Тригонометрия', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000008', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 3, '14:00', '15:30', 'group', 6, '#EC4899', 'Стереометрия', NULL, NOW(), NOW()),
    -- Четверг - Преподаватель 2
    ('20000000-2000-0000-0000-000000000009', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 4, '10:00', '11:00', 'individual', 1, '#06B6D4', 'Механика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000010', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 4, '15:00', '16:30', 'group', 8, '#14B8A6', 'Веб-разработка', NULL, NOW(), NOW()),
    -- Пятница - оба преподавателя
    ('20000000-2000-0000-0000-000000000011', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 5, '10:00', '11:30', 'group', 4, '#3B82F6', 'Подготовка к ЕГЭ Математика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000012', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 5, '14:00', '15:30', 'group', 6, '#06B6D4', 'Подготовка к ЕГЭ Физика', NULL, NOW(), NOW()),
    -- Суббота - индивидуальные
    ('20000000-2000-0000-0000-000000000013', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 6, '10:00', '11:00', 'individual', 1, '#F59E0B', 'Консультация', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0000-000000000014', '10000000-1000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 6, '12:00', '13:00', 'individual', 1, '#22C55E', 'Консультация', NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ШАБЛОН 2: Летние каникулы (сокращённое расписание)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students, color, subject, description, created_at, updated_at) VALUES
    ('20000000-2000-0000-0001-000000000001', '10000000-1000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 2, '10:00', '11:00', 'individual', 1, '#3B82F6', 'Математика летний курс', 'Поддержание навыков летом', NOW(), NOW()),
    ('20000000-2000-0000-0001-000000000002', '10000000-1000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 3, '10:00', '11:00', 'individual', 1, '#06B6D4', 'Физика летний курс', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0001-000000000003', '10000000-1000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 4, '11:00', '12:30', 'group', 4, '#10B981', 'Летний лагерь математики', 'Занимательная математика', NOW(), NOW()),
    ('20000000-2000-0000-0001-000000000004', '10000000-1000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 5, '10:00', '12:00', 'group', 6, '#14B8A6', 'Летний IT-лагерь', 'Программирование для начинающих', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ШАБЛОН 3: Подготовка к ЕГЭ (интенсив)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students, color, subject, description, created_at, updated_at) VALUES
    ('20000000-2000-0000-0002-000000000001', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 1, '09:00', '11:00', 'group', 10, '#EF4444', 'Интенсив ЕГЭ Профиль', 'Часть 2 профильного ЕГЭ', NOW(), NOW()),
    ('20000000-2000-0000-0002-000000000002', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 1, '14:00', '16:00', 'group', 10, '#EF4444', 'Интенсив ЕГЭ База', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0002-000000000003', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 2, '09:00', '11:00', 'group', 10, '#F97316', 'Интенсив ЕГЭ Физика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0002-000000000004', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 3, '09:00', '11:00', 'group', 10, '#EF4444', 'Интенсив ЕГЭ Профиль', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0002-000000000005', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 4, '09:00', '11:00', 'group', 10, '#F97316', 'Интенсив ЕГЭ Физика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0002-000000000006', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 5, '09:00', '11:00', 'group', 10, '#EF4444', 'Интенсив ЕГЭ Профиль', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0002-000000000007', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 5, '14:00', '16:00', 'group', 10, '#F97316', 'Интенсив ЕГЭ Физика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0002-000000000008', '10000000-1000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 6, '10:00', '12:00', 'group', 10, '#EF4444', 'Пробный ЕГЭ', 'Полный вариант', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ШАБЛОН 4: Подготовка к ОГЭ
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students, color, subject, description, created_at, updated_at) VALUES
    ('20000000-2000-0000-0003-000000000001', '10000000-1000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', 1, '15:00', '16:30', 'group', 8, '#EC4899', 'ОГЭ Математика', 'Подготовка к ОГЭ для 9 класса', NOW(), NOW()),
    ('20000000-2000-0000-0003-000000000002', '10000000-1000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', 2, '15:00', '16:30', 'group', 8, '#A855F7', 'ОГЭ Физика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0003-000000000003', '10000000-1000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', 3, '15:00', '16:30', 'group', 8, '#EC4899', 'ОГЭ Математика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0003-000000000004', '10000000-1000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', 4, '15:00', '16:30', 'group', 8, '#A855F7', 'ОГЭ Информатика', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0003-000000000005', '10000000-1000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', 5, '15:00', '17:00', 'group', 8, '#EC4899', 'Пробный ОГЭ', NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ШАБЛОН 5: Олимпиадная подготовка
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students, color, subject, description, created_at, updated_at) VALUES
    ('20000000-2000-0000-0004-000000000001', '10000000-1000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', 6, '10:00', '12:00', 'group', 6, '#F59E0B', 'Олимпиадная математика', 'Решение нестандартных задач', NOW(), NOW()),
    ('20000000-2000-0000-0004-000000000002', '10000000-1000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', 6, '14:00', '16:00', 'group', 6, '#14B8A6', 'Олимпиадная информатика', 'Алгоритмы и структуры данных', NOW(), NOW()),
    ('20000000-2000-0000-0004-000000000003', '10000000-1000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', 0, '11:00', '13:00', 'group', 4, '#8B5CF6', 'Разбор олимпиад', 'Анализ прошлых олимпиад', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ШАБЛОН 6: Минимальное расписание (праздники)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students, color, subject, description, created_at, updated_at) VALUES
    ('20000000-2000-0000-0005-000000000001', '10000000-1000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000001', 3, '11:00', '12:00', 'individual', 1, '#3B82F6', 'Консультация', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0005-000000000002', '10000000-1000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000002', 4, '11:00', '12:00', 'individual', 1, '#06B6D4', 'Консультация', NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ШАБЛОН 7: Выходной режим (только суббота)
INSERT INTO template_lessons (id, template_id, teacher_id, day_of_week, start_time, end_time, lesson_type, max_students, color, subject, description, created_at, updated_at) VALUES
    ('20000000-2000-0000-0006-000000000001', '10000000-1000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000001', 6, '10:00', '11:00', 'individual', 1, '#3B82F6', 'Индивидуальное занятие', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0006-000000000002', '10000000-1000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000001', 6, '11:30', '12:30', 'individual', 1, '#10B981', 'Индивидуальное занятие', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0006-000000000003', '10000000-1000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000002', 6, '14:00', '15:00', 'individual', 1, '#06B6D4', 'Индивидуальное занятие', NULL, NOW(), NOW()),
    ('20000000-2000-0000-0006-000000000004', '10000000-1000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000002', 6, '15:30', '16:30', 'individual', 1, '#14B8A6', 'Индивидуальное занятие', NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================
-- СТУДЕНТЫ В ШАБЛОНАХ ЗАНЯТИЙ
-- =============================================
INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at) VALUES
    -- Основное расписание
    ('30000000-3000-0000-0000-000000000001', '20000000-2000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', NOW()),
    ('30000000-3000-0000-0000-000000000002', '20000000-2000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', NOW()),
    ('30000000-3000-0000-0000-000000000003', '20000000-2000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', NOW()),
    ('30000000-3000-0000-0000-000000000004', '20000000-2000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000003', NOW()),
    ('30000000-3000-0000-0000-000000000005', '20000000-2000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000001', NOW()),
    ('30000000-3000-0000-0000-000000000006', '20000000-2000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000002', NOW()),
    ('30000000-3000-0000-0000-000000000007', '20000000-2000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000004', NOW()),
    ('30000000-3000-0000-0000-000000000008', '20000000-2000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000005', NOW()),
    ('30000000-3000-0000-0000-000000000009', '20000000-2000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000003', NOW()),
    ('30000000-3000-0000-0000-000000000010', '20000000-2000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000004', NOW()),
    -- ЕГЭ интенсив
    ('30000000-3000-0000-0001-000000000001', '20000000-2000-0000-0002-000000000001', '20000000-0000-0000-0000-000000000001', NOW()),
    ('30000000-3000-0000-0001-000000000002', '20000000-2000-0000-0002-000000000001', '20000000-0000-0000-0000-000000000002', NOW()),
    ('30000000-3000-0000-0001-000000000003', '20000000-2000-0000-0002-000000000001', '20000000-0000-0000-0000-000000000003', NOW()),
    ('30000000-3000-0000-0001-000000000004', '20000000-2000-0000-0002-000000000003', '20000000-0000-0000-0000-000000000001', NOW()),
    ('30000000-3000-0000-0001-000000000005', '20000000-2000-0000-0002-000000000003', '20000000-0000-0000-0000-000000000004', NOW()),
    -- Олимпиадная подготовка
    ('30000000-3000-0000-0002-000000000001', '20000000-2000-0000-0004-000000000001', '20000000-0000-0000-0000-000000000001', NOW()),
    ('30000000-3000-0000-0002-000000000002', '20000000-2000-0000-0004-000000000001', '20000000-0000-0000-0000-000000000003', NOW()),
    ('30000000-3000-0000-0002-000000000003', '20000000-2000-0000-0004-000000000002', '20000000-0000-0000-0000-000000000002', NOW())
ON CONFLICT DO NOTHING;
"

# 6. Создание домашних заданий
echo -e "${GREEN}[6/8] Создание домашних заданий...${NC}"
run_sql "
-- =============================================
-- ДОМАШНИЕ ЗАДАНИЯ (файлы и текстовые)
-- =============================================

-- ПРОШЛЫЕ ЗАНЯТИЯ С ДЗ - Преподаватель 1
INSERT INTO lesson_homework (id, lesson_id, file_name, file_path, file_size, mime_type, created_by, text_content, created_at) VALUES
    -- d001 - месяц назад
    ('40000000-4000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 'math_basics.pdf', '/uploads/homework/math_basics.pdf', 524288, 'application/pdf', '10000000-0000-0000-0000-000000000001', NULL, NOW() - INTERVAL '30 days'),
    -- d002 - 3 недели назад (файл + текст)
    ('40000000-4000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000002', 'algebra_formulas.docx', '/uploads/homework/algebra_formulas.docx', 102400, 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', '10000000-0000-0000-0000-000000000001', 'Повторить формулы сокращённого умножения. Ссылка: https://math.ru/formulas', NOW() - INTERVAL '21 days'),
    -- d003 - 2 недели назад (только файл)
    ('40000000-4000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000003', 'geometry_theorems.pdf', '/uploads/homework/geometry_theorems.pdf', 256000, 'application/pdf', '10000000-0000-0000-0000-000000000001', NULL, NOW() - INTERVAL '14 days'),
    -- d004 - неделю назад (файл + текст)
    ('40000000-4000-0000-0000-000000000004', 'd0000000-0000-0000-0000-000000000004', 'ege_variant_1.pdf', '/uploads/homework/ege_variant_1.pdf', 768000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Решите полный вариант ЕГЭ. Проверка: https://ege.sdamgia.ru', NOW() - INTERVAL '7 days'),
    -- d005 - 5 дней назад (файл + текст)
    ('40000000-4000-0000-0000-000000000005', 'd0000000-0000-0000-0000-000000000005', 'trig_formulas.pdf', '/uploads/homework/trig_formulas.pdf', 128000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Выучить формулы приведения. Карточки для запоминания: https://quizlet.com/trig', NOW() - INTERVAL '5 days'),
    -- d006 - 3 дня назад (файл)
    ('40000000-4000-0000-0000-000000000006', 'd0000000-0000-0000-0000-000000000006', 'stereometry_sections.pdf', '/uploads/homework/stereometry_sections.pdf', 384000, 'application/pdf', '10000000-0000-0000-0000-000000000001', NULL, NOW() - INTERVAL '3 days'),
    -- d007 - вчера (файл + текст)
    ('40000000-4000-0000-0000-000000000007', 'd0000000-0000-0000-0000-000000000007', 'combinatorics_tasks.pdf', '/uploads/homework/combinatorics_tasks.pdf', 192000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Решить 15 задач на перестановки и сочетания из учебника', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- ПРОШЛЫЕ ЗАНЯТИЯ С ДЗ - Преподаватель 2
INSERT INTO lesson_homework (id, lesson_id, file_name, file_path, file_size, mime_type, created_by, text_content, created_at) VALUES
    -- e001 - месяц назад
    ('40000000-4000-0000-0001-000000000001', 'e0000000-0000-0000-0000-000000000001', 'mechanics_newton.pdf', '/uploads/homework/mechanics_newton.pdf', 1048576, 'application/pdf', '10000000-0000-0000-0000-000000000002', 'Законы Ньютона. Видео: https://youtube.com/physics101', NOW() - INTERVAL '30 days'),
    -- e002 - 3 недели назад
    ('40000000-4000-0000-0001-000000000002', 'e0000000-0000-0000-0000-000000000002', 'python_calculator.py', '/uploads/homework/python_calculator.py', 4096, 'text/x-python', '10000000-0000-0000-0000-000000000002', 'Написать калькулятор на Python с GUI', NOW() - INTERVAL '21 days'),
    -- e003 - 2 недели назад
    ('40000000-4000-0000-0001-000000000003', 'e0000000-0000-0000-0000-000000000003', 'thermodynamics.pdf', '/uploads/homework/thermodynamics.pdf', 512000, 'application/pdf', '10000000-0000-0000-0000-000000000002', NULL, NOW() - INTERVAL '14 days'),
    -- e004 - неделю назад
    ('40000000-4000-0000-0001-000000000004', 'e0000000-0000-0000-0000-000000000004', 'sorting_algorithms.zip', '/uploads/homework/sorting_algorithms.zip', 8192, 'application/zip', '10000000-0000-0000-0000-000000000002', 'Реализовать пузырьковую и быструю сортировку', NOW() - INTERVAL '7 days'),
    -- e005 - 4 дня назад
    ('40000000-4000-0000-0001-000000000005', 'e0000000-0000-0000-0000-000000000005', 'physics_ege_tasks.pdf', '/uploads/homework/physics_ege_tasks.pdf', 640000, 'application/pdf', '10000000-0000-0000-0000-000000000002', NULL, NOW() - INTERVAL '4 days'),
    -- e006 - вчера (файл + текст)
    ('40000000-4000-0000-0001-000000000006', 'e0000000-0000-0000-0000-000000000006', 'html_form_template.html', '/uploads/homework/html_form_template.html', 2048, 'text/html', '10000000-0000-0000-0000-000000000002', 'Создать HTML-страницу с формой регистрации', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- БУДУЩИЕ ЗАНЯТИЯ С ДЗ - Преподаватель 1
INSERT INTO lesson_homework (id, lesson_id, file_name, file_path, file_size, mime_type, created_by, text_content, created_at) VALUES
    -- d012 - завтра
    ('40000000-4000-0000-0002-000000000001', 'd0000000-0000-0000-0000-000000000012', 'similar_triangles.pdf', '/uploads/homework/similar_triangles.pdf', 320000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Повторить теоремы о подобии', NOW() - INTERVAL '3 days'),
    -- d013 - через 2 дня
    ('40000000-4000-0000-0002-000000000002', 'd0000000-0000-0000-0000-000000000013', 'ege_profile_var2.pdf', '/uploads/homework/ege_profile_var2.pdf', 896000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Прорешать профильный вариант: https://ege.sdamgia.ru/test?id=54321', NOW() - INTERVAL '2 days'),
    -- d014 - через 4 дня (файл + текст)
    ('40000000-4000-0000-0002-000000000003', 'd0000000-0000-0000-0000-000000000014', 'olympiad_2023.pdf', '/uploads/homework/olympiad_2023.pdf', 448000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Решить задачи Всероссийской олимпиады 2023 года (школьный этап)', NOW() - INTERVAL '1 day'),
    -- d015 - через неделю
    ('40000000-4000-0000-0002-000000000004', 'd0000000-0000-0000-0000-000000000015', 'oge_training.pdf', '/uploads/homework/oge_training.pdf', 480000, 'application/pdf', '10000000-0000-0000-0000-000000000001', NULL, NOW()),
    -- d016 - через 10 дней (файл + текст)
    ('40000000-4000-0000-0002-000000000005', 'd0000000-0000-0000-0000-000000000016', 'probability_chapter7.pdf', '/uploads/homework/probability_chapter7.pdf', 256000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Изучить главу 7 учебника по теории вероятностей', NOW()),
    -- d017 - через 2 недели
    ('40000000-4000-0000-0002-000000000006', 'd0000000-0000-0000-0000-000000000017', 'limits_practice.pdf', '/uploads/homework/limits_practice.pdf', 256000, 'application/pdf', '10000000-0000-0000-0000-000000000001', 'Вычислить пределы функций (задачи 1-20)', NOW())
ON CONFLICT (id) DO NOTHING;

-- БУДУЩИЕ ЗАНЯТИЯ С ДЗ - Преподаватель 2
INSERT INTO lesson_homework (id, lesson_id, file_name, file_path, file_size, mime_type, created_by, text_content, created_at) VALUES
    -- e011 - завтра
    ('40000000-4000-0000-0003-000000000001', 'e0000000-0000-0000-0000-000000000011', 'oop_basics.py', '/uploads/homework/oop_basics.py', 2048, 'text/x-python', '10000000-0000-0000-0000-000000000002', 'Создать класс Student с методами', NOW() - INTERVAL '2 days'),
    -- e012 - через 2 дня
    ('40000000-4000-0000-0003-000000000002', 'e0000000-0000-0000-0000-000000000012', 'coulomb_law.pdf', '/uploads/homework/coulomb_law.pdf', 384000, 'application/pdf', '10000000-0000-0000-0000-000000000002', NULL, NOW() - INTERVAL '1 day'),
    -- e013 - через 4 дня
    ('40000000-4000-0000-0003-000000000003', 'e0000000-0000-0000-0000-000000000013', 'sql_exercises.sql', '/uploads/homework/sql_exercises.sql', 4096, 'text/x-sql', '10000000-0000-0000-0000-000000000002', 'Написать 10 SQL-запросов к учебной БД', NOW()),
    -- e014 - через неделю (файл + текст)
    ('40000000-4000-0000-0003-000000000004', 'e0000000-0000-0000-0000-000000000014', 'diffraction_guide.pdf', '/uploads/homework/diffraction_guide.pdf', 320000, 'application/pdf', '10000000-0000-0000-0000-000000000002', 'Подготовить доклад по теме \"Дифракция света\" (5-7 минут)', NOW()),
    -- e015 - через 9 дней
    ('40000000-4000-0000-0003-000000000005', 'e0000000-0000-0000-0000-000000000015', 'django_starter.zip', '/uploads/homework/django_starter.zip', 16384, 'application/zip', '10000000-0000-0000-0000-000000000002', 'Создать простое веб-приложение на Django', NOW()),
    -- e016 - через 2 недели
    ('40000000-4000-0000-0003-000000000006', 'e0000000-0000-0000-0000-000000000016', 'physics_ege_3vars.pdf', '/uploads/homework/physics_ege_3vars.pdf', 1536000, 'application/pdf', '10000000-0000-0000-0000-000000000002', 'Решить 3 полных варианта ЕГЭ по физике', NOW())
ON CONFLICT (id) DO NOTHING;
"

# 7. Создание рассылок
echo -e "${GREEN}[7/8] Создание рассылок...${NC}"
run_sql "
-- Рассылки преподавателей
INSERT INTO lesson_broadcasts (id, lesson_id, sender_id, message, status, sent_count, failed_count, created_at, completed_at) VALUES
    ('50000000-5000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001',
        'Напоминаю, что завтра в 14:00 групповое занятие по алгебре. Не забудьте калькуляторы!',
        'completed', 3, 0, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
    ('50000000-5000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002',
        'Уважаемые студенты! Занятие по физике переносится на 30 минут позже. Начало в 16:30.',
        'completed', 5, 0, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('50000000-5000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001',
        'Добрый день! Напоминаю о предстоящем занятии по подготовке к ЕГЭ. Принесите решённые домашние задания. Полезная ссылка: https://ege.sdamgia.ru',
        'pending', 0, 0, NOW(), NULL)
ON CONFLICT (id) DO NOTHING;
"

# 8. Создание чат-комнат
echo -e "${GREEN}[8/8] Создание чат-комнат...${NC}"
run_sql "
-- Чат-комнаты между преподавателями и студентами
INSERT INTO chat_rooms (id, teacher_id, student_id, last_message_at, created_at, updated_at) VALUES
    ('60000000-6000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '7 days', NOW()),
    ('60000000-6000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002', NOW() - INTERVAL '2 days', NOW() - INTERVAL '5 days', NOW()),
    ('60000000-6000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '6 days', NOW()),
    ('60000000-6000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000004', NOW() - INTERVAL '1 day', NOW() - INTERVAL '4 days', NOW())
ON CONFLICT (id) DO NOTHING;

-- Сообщения в чатах (status: pending_moderation, delivered, blocked)
INSERT INTO messages (id, room_id, sender_id, message_text, status, created_at) VALUES
    ('70000000-7000-0000-0000-000000000001', '60000000-6000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Добрый день! Можно уточнить по домашнему заданию?', 'delivered', NOW() - INTERVAL '2 hours'),
    ('70000000-7000-0000-0000-000000000002', '60000000-6000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'Здравствуйте! Конечно, задавайте вопрос.', 'delivered', NOW() - INTERVAL '1 hour' - INTERVAL '30 minutes'),
    ('70000000-7000-0000-0000-000000000003', '60000000-6000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'В задаче 5 непонятно как применить формулу. Можете объяснить?', 'delivered', NOW() - INTERVAL '1 hour'),
    ('70000000-7000-0000-0000-000000000004', '60000000-6000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000001', 'Мария Александровна, спасибо за занятие! Очень понравилось объяснение.', 'delivered', NOW() - INTERVAL '3 hours')
ON CONFLICT (id) DO NOTHING;
"

echo ""
echo -e "${GREEN}=== Загрузка завершена! ===${NC}"
echo ""
echo "Тестовые аккаунты (пароль для всех: password123):"
echo ""
echo "  Администратор:"
echo "    admin@tutoring.com"
echo ""
echo "  Преподаватели:"
echo "    ivan.petrov@tutoring.com (Иван Петров)"
echo "    maria.sidorova@tutoring.com (Мария Сидорова)"
echo "    alexey.kozlov@tutoring.com (Алексей Козлов)"
echo ""
echo "  Студенты:"
echo "    anna.ivanova@student.com (Анна Иванова - 10 кредитов)"
echo "    dmitry.smirnov@student.com (Дмитрий Смирнов - 8 кредитов)"
echo "    elena.volkova@student.com (Елена Волкова - 12 кредитов)"
echo "    pavel.morozov@student.com (Павел Морозов - 5 кредитов)"
echo "    olga.novikova@student.com (Ольга Новикова - 3 кредита)"
echo ""
echo "Загружено:"
echo "  - 9 пользователей (1 админ, 3 препода, 5 студентов)"
echo "  - 107 занятий:"
echo "      • 53 занятия преподавателя 1 (Математика)"
echo "      • 54 занятия преподавателя 2 (Физика/IT)"
echo "      • Прошлые с ДЗ: 35"
echo "      • Прошлые без ДЗ: 24"
echo "      • Будущие с ДЗ: 26"
echo "      • Будущие без ДЗ: 22"
echo "  - Бронирования на занятия"
echo "  - 7 шаблонов расписания:"
echo "      • Основное расписание (14 занятий)"
echo "      • Летние каникулы (4 занятия)"
echo "      • Подготовка к ЕГЭ (8 занятий)"
echo "      • Подготовка к ОГЭ (5 занятий)"
echo "      • Олимпиадная подготовка (3 занятия)"
echo "      • Минимальное расписание (2 занятия)"
echo "      • Выходной режим (4 занятия)"
echo "  - 40 занятий в шаблонах"
echo "  - 19 домашних заданий (файлы и текстовые)"
echo "  - 3 рассылки"
echo "  - 4 чат-комнаты с сообщениями"
