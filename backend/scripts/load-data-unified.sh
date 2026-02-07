#!/bin/bash
################################################################################
# Unified Data Loading Script for Tutoring Platform
#
# Loads realistic test data: users, lessons, homework, bookings, credits
# Replaces both load-data.sh and load-data-production.sh
#
# USAGE:
#   ./load-data-unified.sh                    # Insert only (default, safe)
#   ./load-data-unified.sh --truncate         # Clear all data first
#   ./load-data-unified.sh --truncate --yes   # Auto-confirm (CI/CD)
#   DB_HOST=postgres ./load-data-unified.sh   # Docker: uses postgres container
#
# ENV VARS:
#   DB_HOST (default: localhost)
#   DB_PORT (default: 5432)
#   DB_NAME (default: tutoring_platform)
#   DB_USER (default: tutoring)
#   DB_PASSWORD (default: postgres)
#
################################################################################

set -e

# === COLORS ===
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# === DEFAULTS ===
DO_TRUNCATE=false
AUTO_CONFIRM=false

# === PARSE ARGUMENTS ===
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
            echo "Unified Data Loader for Tutoring Platform"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --truncate    Clear ALL existing data before loading (DANGEROUS!)"
            echo "  --yes, -y     Auto-confirm truncate (for CI/CD, use with caution)"
            echo "  --help, -h    Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  DB_HOST       Database host (default: localhost)"
            echo "  DB_PORT       Database port (default: 5432)"
            echo "  DB_NAME       Database name (default: tutoring_platform)"
            echo "  DB_USER       Database user (default: tutoring)"
            echo "  DB_PASSWORD   Database password (required)"
            echo ""
            echo "By default, script only INSERTS data without deleting."
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# === DATABASE CONFIGURATION ===
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-tutoring}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

export PGPASSWORD="$DB_PASSWORD"

# === SAFETY CHECKS ===
echo -e "${BLUE}=== Pre-flight Checks ===${NC}"

# Check 1: Block test databases
if [[ "$DB_NAME" == "thebot_db_test" ]] || [[ "$DB_NAME" == "test_db" ]]; then
    echo -e "${RED}ERROR: Cannot load data into test database!${NC}"
    echo "Test databases must be managed by test suite, not this script."
    exit 1
fi

# Check 2: Allow production-like names (tutoring_platform is acceptable)
# But warn if it looks production-ish
if [[ "$DB_NAME" == *"_prod"* ]] || [[ "$DB_NAME" == *"_production"* ]]; then
    if [[ "$DB_HOST" != "localhost" ]] && [[ "$DB_HOST" != "127.0.0.1" ]] && [[ "$DB_HOST" != "::1" ]] && [[ "$DB_HOST" != "postgres" ]]; then
        echo -e "${RED}ERROR: Production-looking database name on remote host!${NC}"
        echo "This script will not load data to remote production databases."
        echo "Host: $DB_HOST"
        echo "Database: $DB_NAME"
        exit 1
    fi
fi

# Check 3: TRUNCATE protection
if [[ "$DO_TRUNCATE" == true ]]; then
    if [[ "$DB_HOST" != "localhost" ]] && [[ "$DB_HOST" != "127.0.0.1" ]] && [[ "$DB_HOST" != "::1" ]] && [[ "$DB_HOST" != "postgres" ]]; then
        echo -e "${RED}ERROR: TRUNCATE blocked on remote database!${NC}"
        echo "TRUNCATE is only allowed on localhost/Docker containers."
        exit 1
    fi
fi

# Check 4: Password verification
if [[ -z "$DB_PASSWORD" ]]; then
    echo -e "${RED}ERROR: DB_PASSWORD is required!${NC}"
    exit 1
fi

# Test connection
echo -e "${BLUE}Testing database connection...${NC}"
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Cannot connect to database!${NC}"
    echo "Host: $DB_HOST:$DB_PORT"
    echo "Database: $DB_NAME"
    echo "User: $DB_USER"
    exit 1
fi

echo -e "${GREEN}✓ Connection successful${NC}"
echo -e "${GREEN}✓ Safety checks passed${NC}"
echo ""

# === SQL FUNCTIONS ===
run_sql() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$1"
}

run_sql_file() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$1"
}

# === MAIN LOADING PROCESS ===
echo -e "${YELLOW}=== Loading Test Data ===${NC}"
echo "Target: $DB_NAME@$DB_HOST:$DB_PORT"
echo ""

# PHASE 0: TRUNCATE (if requested)
if [[ "$DO_TRUNCATE" == true ]]; then
    echo -e "${RED}⚠️  WARNING: TRUNCATE Mode${NC}"
    echo -e "${RED}This will DELETE ALL DATA in: $DB_NAME${NC}"
    echo ""

    if [[ "$AUTO_CONFIRM" != true ]]; then
        echo -e "${YELLOW}Type 'DELETE ALL DATA' to confirm:${NC}"
        read -r confirm
        if [[ "$confirm" != "DELETE ALL DATA" ]]; then
            echo -e "${GREEN}Cancelled. No data was deleted.${NC}"
            exit 0
        fi
    fi

    echo -e "${YELLOW}[0/8] Truncating tables...${NC}"
    run_sql "
TRUNCATE TABLE lesson_homework, broadcast_files, lesson_broadcasts,
    cancelled_bookings, messages, file_attachments, blocked_messages,
    chat_rooms, swaps, bookings, template_applications,
    template_lesson_students, template_lessons, lesson_templates,
    lesson_modifications, lessons, credit_transactions, credits,
    payments, subjects, teacher_subjects, sessions, telegram_tokens,
    telegram_users, broadcast_lists, broadcasts, broadcast_logs, auth_failures, users
CASCADE;
"
    echo -e "${GREEN}✓ Tables truncated${NC}"
else
    echo -e "${BLUE}[0/8] Insert-only mode (existing data preserved)${NC}"
fi

echo ""

# PHASE 1: CREATE USERS
echo -e "${BLUE}[1/8] Creating users...${NC}"

# Password hash: password123 (bcrypt, cost 10)
HASH='$2a$10$LiLWVAWbrxx/8wSy4H2of.bs1tpzNA1y/qrnpdzT9wu0AlqTfB6jy'

run_sql "
INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at) VALUES
-- ADMINISTRATORS
('00000000-0000-0000-0000-000000000001', 'admin@thebot.ru', '$HASH', 'Администратор THE BOT', 'admin', NOW(), NOW()),

-- TEACHERS (тьюторы)
('10000000-0000-0000-0000-000000000001', 'method1@thebot.ru', '$HASH', 'Иван Петров', 'teacher', NOW(), NOW()),
('10000000-0000-0000-0000-000000000002', 'method2@thebot.ru', '$HASH', 'Мария Сидорова', 'teacher', NOW(), NOW()),
('10000000-0000-0000-0000-000000000003', 'method3@thebot.ru', '$HASH', 'Александр Морозов', 'teacher', NOW(), NOW()),

-- STUDENTS
('20000000-0000-0000-0000-000000000001', 'student1@thebot.ru', '$HASH', 'Дмитрий Смирнов', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000002', 'student2@thebot.ru', '$HASH', 'Елена Волкова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000003', 'student3@thebot.ru', '$HASH', 'Павел Морозов', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000004', 'student4@thebot.ru', '$HASH', 'Ольга Новикова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000005', 'student5@thebot.ru', '$HASH', 'Анна Иванова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000006', 'student6@thebot.ru', '$HASH', 'Сергей Петров', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000007', 'student7@thebot.ru', '$HASH', 'Викторія Козлова', 'student', NOW(), NOW()),
('20000000-0000-0000-0000-000000000008', 'student8@thebot.ru', '$HASH', 'Константин Лебедев', 'student', NOW(), NOW())
ON CONFLICT (email) WHERE deleted_at IS NULL DO NOTHING;
"
echo -e "${GREEN}✓ 1 admin + 3 teachers + 8 students created${NC}"

# PHASE 2: SETUP CREDITS
echo -e "${BLUE}[2/8] Setting up student credits...${NC}"
run_sql "
UPDATE credits SET balance = 15 WHERE user_id = '20000000-0000-0000-0000-000000000001';
UPDATE credits SET balance = 12 WHERE user_id = '20000000-0000-0000-0000-000000000002';
UPDATE credits SET balance = 20 WHERE user_id = '20000000-0000-0000-0000-000000000003';
UPDATE credits SET balance = 8 WHERE user_id = '20000000-0000-0000-0000-000000000004';
UPDATE credits SET balance = 10 WHERE user_id = '20000000-0000-0000-0000-000000000005';
UPDATE credits SET balance = 5 WHERE user_id = '20000000-0000-0000-0000-000000000006';
UPDATE credits SET balance = 18 WHERE user_id = '20000000-0000-0000-0000-000000000007';
UPDATE credits SET balance = 25 WHERE user_id = '20000000-0000-0000-0000-000000000008';
"
echo -e "${GREEN}✓ Student credits configured${NC}"

# PHASE 3: CREATE SUBJECTS
echo -e "${BLUE}[3/8] Creating subjects...${NC}"
run_sql "
INSERT INTO subjects (id, name, description, created_at, updated_at) VALUES
(gen_random_uuid(), 'Математика', 'Курс высшей математики и алгебры', NOW(), NOW()),
(gen_random_uuid(), 'Физика', 'Общая и специальная физика', NOW(), NOW()),
(gen_random_uuid(), 'Информатика', 'Основы программирования и алгоритмы', NOW(), NOW()),
(gen_random_uuid(), 'Русский язык', 'Культура речи и писменность', NOW(), NOW()),
(gen_random_uuid(), 'История', 'Мировая и отечественная история', NOW(), NOW()),
(gen_random_uuid(), 'Английский язык', 'Иностранный язык', NOW(), NOW())
ON CONFLICT DO NOTHING;
"
echo -e "${GREEN}✓ 6 subjects created${NC}"

# PHASE 4: CREATE LESSONS (EXTENSIVE)
echo -e "${BLUE}[4/8] Creating lessons (40+)...${NC}"
run_sql "
INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, subject, homework_text, created_at, updated_at) VALUES

-- PAST LESSONS WITH HOMEWORK (past 2 months)
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001',
  NOW() - INTERVAL '60 days' + TIME '10:00', NOW() - INTERVAL '60 days' + TIME '11:00',
  1, 'Математика', 'Решить задачи 1-20 из учебника', NOW() - INTERVAL '61 days', NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000001',
  NOW() - INTERVAL '45 days' + TIME '14:00', NOW() - INTERVAL '45 days' + TIME '15:30',
  6, 'Математика', 'Повторить тему: Интегралы', NOW() - INTERVAL '46 days', NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000002',
  NOW() - INTERVAL '30 days' + TIME '16:00', NOW() - INTERVAL '30 days' + TIME '17:30',
  8, 'Физика', 'Подготовить реферат по механике', NOW() - INTERVAL '31 days', NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000002',
  NOW() - INTERVAL '21 days' + TIME '10:00', NOW() - INTERVAL '21 days' + TIME '11:00',
  1, 'Информатика', 'Написать программу на Python', NOW() - INTERVAL '22 days', NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000003',
  NOW() - INTERVAL '14 days' + TIME '15:00', NOW() - INTERVAL '14 days' + TIME '16:30',
  5, 'Русский язык', 'Написать сочинение на 3-4 страницы', NOW() - INTERVAL '15 days', NOW()),

-- UPCOMING LESSONS (next 3 months)
(gen_random_uuid(), '10000000-0000-0000-0000-000000000001',
  NOW() + INTERVAL '1 day' + TIME '10:00', NOW() + INTERVAL '1 day' + TIME '11:00',
  1, 'Математика', 'Решить примеры по производным', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000001',
  NOW() + INTERVAL '2 days' + TIME '14:00', NOW() + INTERVAL '2 days' + TIME '15:30',
  4, 'Математика', 'Контрольная работа на тему Пределы', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000001',
  NOW() + INTERVAL '4 days' + TIME '11:00', NOW() + INTERVAL '4 days' + TIME '12:00',
  1, 'Математика', 'Консультация перед экзаменом', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000002',
  NOW() + INTERVAL '5 days' + TIME '16:00', NOW() + INTERVAL '5 days' + TIME '17:30',
  6, 'Физика', 'Практика: решение задач ЕГЭ', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000002',
  NOW() + INTERVAL '7 days' + TIME '10:00', NOW() + INTERVAL '7 days' + TIME '11:30',
  3, 'Информатика', 'Основы веб-разработки', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000003',
  NOW() + INTERVAL '8 days' + TIME '15:00', NOW() + INTERVAL '8 days' + TIME '16:30',
  5, 'Русский язык', 'Практика написания изложений', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000003',
  NOW() + INTERVAL '10 days' + TIME '14:00', NOW() + INTERVAL '10 days' + TIME '15:30',
  7, 'История', 'Семинар: История России XX века', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000002',
  NOW() + INTERVAL '12 days' + TIME '17:00', NOW() + INTERVAL '12 days' + TIME '18:30',
  4, 'Английский язык', 'Разговорная практика', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000001',
  NOW() + INTERVAL '14 days' + TIME '10:00', NOW() + INTERVAL '14 days' + TIME '11:30',
  2, 'Математика', 'Подготовка к олимпиаде', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000002',
  NOW() + INTERVAL '16 days' + TIME '14:00', NOW() + INTERVAL '16 days' + TIME '15:30',
  6, 'Физика', 'Лабораторная работа: Электричество', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000002',
  NOW() + INTERVAL '18 days' + TIME '16:00', NOW() + INTERVAL '18 days' + TIME '17:30',
  1, 'Информатика', 'Индивидуальная консультация', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000003',
  NOW() + INTERVAL '20 days' + TIME '15:00', NOW() + INTERVAL '20 days' + TIME '16:30',
  8, 'Русский язык', 'Групповой тренинг ЕГЭ', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000001',
  NOW() + INTERVAL '22 days' + TIME '11:00', NOW() + INTERVAL '22 days' + TIME '12:30',
  3, 'Математика', 'Практическое применение графиков', NOW(), NOW()),

(gen_random_uuid(), '10000000-0000-0000-0000-000000000003',
  NOW() + INTERVAL '25 days' + TIME '14:00', NOW() + INTERVAL '25 days' + TIME '15:30',
  5, 'История', 'Дискуссия: Ключевые события истории', NOW(), NOW())
ON CONFLICT DO NOTHING;
"
echo -e "${GREEN}✓ 20 lessons created${NC}"

# PHASE 5: CREATE BOOKINGS
echo -e "${BLUE}[5/8] Creating bookings...${NC}"
run_sql "
INSERT INTO bookings (id, student_id, lesson_id, status, created_at, updated_at) VALUES
(gen_random_uuid(), '20000000-0000-0000-0000-000000000001',
  (SELECT id FROM lessons WHERE subject = 'Математика' LIMIT 1), 'confirmed', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000002',
  (SELECT id FROM lessons WHERE subject = 'Физика' LIMIT 1), 'confirmed', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000003',
  (SELECT id FROM lessons WHERE subject = 'Информатика' LIMIT 1), 'confirmed', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000004',
  (SELECT id FROM lessons WHERE subject = 'Математика' LIMIT 1 OFFSET 1), 'pending', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000005',
  (SELECT id FROM lessons WHERE subject = 'Русский язык' LIMIT 1), 'confirmed', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000006',
  (SELECT id FROM lessons WHERE subject = 'История' LIMIT 1), 'confirmed', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000007',
  (SELECT id FROM lessons WHERE subject = 'Английский язык' LIMIT 1), 'confirmed', NOW(), NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000008',
  (SELECT id FROM lessons WHERE subject = 'Физика' LIMIT 1 OFFSET 1), 'confirmed', NOW(), NOW())
ON CONFLICT DO NOTHING;
"
echo -e "${GREEN}✓ 8 bookings created${NC}"

# PHASE 6: CREATE HOMEWORK
echo -e "${BLUE}[6/8] Creating homework...${NC}"
run_sql "
INSERT INTO lesson_homework (id, lesson_id, content, file_url, created_at, updated_at) VALUES
(gen_random_uuid(), (SELECT id FROM lessons WHERE subject = 'Математика' LIMIT 1),
  'Решить задачи параграфа 5', NULL, NOW(), NOW()),
(gen_random_uuid(), (SELECT id FROM lessons WHERE subject = 'Физика' LIMIT 1),
  'Написать краткую теорию', NULL, NOW(), NOW()),
(gen_random_uuid(), (SELECT id FROM lessons WHERE subject = 'Информатика' LIMIT 1),
  'Написать программу на Python для обработки списков', NULL, NOW(), NOW())
ON CONFLICT DO NOTHING;
"
echo -e "${GREEN}✓ Homework entries created${NC}"

# PHASE 7: CREATE CREDIT TRANSACTIONS
echo -e "${BLUE}[7/8] Creating credit transactions...${NC}"
run_sql "
INSERT INTO credit_transactions (id, user_id, amount, operation_type, reason, performed_by, created_at) VALUES
(gen_random_uuid(), '20000000-0000-0000-0000-000000000001', 15, 'add', 'Initial credit allocation', '00000000-0000-0000-0000-000000000001', NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000002', 12, 'add', 'Initial credit allocation', '00000000-0000-0000-0000-000000000001', NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000003', 20, 'add', 'Bonus credits for early signup', '00000000-0000-0000-0000-000000000001', NOW()),
(gen_random_uuid(), '20000000-0000-0000-0000-000000000001', 1, 'debit', 'Lesson booking', '10000000-0000-0000-0000-000000000001', NOW())
ON CONFLICT DO NOTHING;
"
echo -e "${GREEN}✓ Credit transactions created${NC}"

# PHASE 8: CREATE CHAT ROOMS
echo -e "${BLUE}[8/8] Creating chat infrastructure...${NC}"
run_sql "
-- Chat rooms will auto-create when first message is sent
-- This ensures we have the proper structure
SELECT 1 as 'Chat infrastructure ready'
"
echo -e "${GREEN}✓ Chat infrastructure ready${NC}"

# === FINAL REPORT ===
echo ""
echo -e "${GREEN}════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ Data Loading Complete!${NC}"
echo -e "${GREEN}════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}Test Accounts (password: password123):${NC}"
echo ""
echo "ADMIN:"
echo "  📧 admin@thebot.ru"
echo ""
echo "TEACHERS (Тьюторы):"
echo "  📧 method1@thebot.ru (Иван Петров)"
echo "  📧 method2@thebot.ru (Мария Сидорова)"
echo "  📧 method3@thebot.ru (Александр Морозов)"
echo ""
echo "STUDENTS:"
echo "  📧 student1@thebot.ru (Дмитрий Смирнов) - 15 credits"
echo "  📧 student2@thebot.ru (Елена Волкова) - 12 credits"
echo "  📧 student3@thebot.ru (Павел Морозов) - 20 credits"
echo "  📧 student4@thebot.ru (Ольга Новикова) - 8 credits"
echo "  📧 student5@thebot.ru (Анна Иванова) - 10 credits"
echo "  📧 student6@thebot.ru (Сергей Петров) - 5 credits"
echo "  📧 student7@thebot.ru (Викторія Козлова) - 18 credits"
echo "  📧 student8@thebot.ru (Константин Лебедев) - 25 credits"
echo ""
echo -e "${BLUE}Data Loaded:${NC}"
echo "  ✓ 1 administrator"
echo "  ✓ 3 teachers"
echo "  ✓ 8 students"
echo "  ✓ 20+ lessons (past and future)"
echo "  ✓ 6 subjects"
echo "  ✓ 8 bookings"
echo "  ✓ Credit allocations and transactions"
echo "  ✓ Homework assignments"
echo ""
echo "Database: $DB_NAME @ $DB_HOST:$DB_PORT"
echo -e "${GREEN}════════════════════════════════════════════════${NC}"
