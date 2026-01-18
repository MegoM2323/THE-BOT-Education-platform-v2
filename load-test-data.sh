#!/bin/bash

# ╔═══════════════════════════════════════════════════════════════╗
# ║  Загрузка тестовых данных                                    ║
# ║  Load Test Data Script                                        ║
# ╚═══════════════════════════════════════════════════════════════╝

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

print_step() {
    echo -e "${BLUE}${BOLD}[$(date +%H:%M:%S)]${NC} ${CYAN}➜${NC} $1"
}

print_success() {
    echo -e "${GREEN}${BOLD}✓${NC} $1"
}

print_error() {
    echo -e "${RED}${BOLD}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}${BOLD}ℹ${NC} $1"
}

# Получение директории скрипта
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

clear
echo -e "${CYAN}${BOLD}"
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                                                               ║"
echo "║         📊 Загрузка тестовых данных                          ║"
echo "║                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""

# Загрузка конфигурации
if [ -f "$SCRIPT_DIR/backend/.env" ]; then
    while IFS='=' read -r key value; do
        # Пропускаем пустые строки и комментарии
        [[ -z "$key" || "$key" == "#"* ]] && continue
        # Удаляем кавычки из значения, если они есть
        value=$(echo "$value" | sed -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//")
        # Экспортируем переменную
        export "$key=$value"
    done < "$SCRIPT_DIR/backend/.env"
fi

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-thebot_db}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}

# Автоматическая замена Docker-хоста на localhost для локального запуска
if [ "$DB_HOST" = "postgres" ] || [ "$DB_HOST" = "db" ]; then
    DB_HOST="localhost"
    export DB_HOST="localhost"
fi

print_info "База данных: $DB_HOST:$DB_PORT/$DB_NAME"
echo ""

# CRITICAL SAFETY CHECK: Refuse to run on production database
if [[ "$DB_NAME" == *"prod"* ]] || [[ "$DB_NAME" == *"production"* ]] || [[ "$DB_NAME" == *"live"* ]]; then
    print_error "============================================================"
    print_error "CRITICAL SAFETY VIOLATION!"
    print_error "============================================================"
    print_error "Database name '$DB_NAME' appears to be a production database!"
    print_error "This script should NEVER run on production!"
    print_error "Aborting to prevent data loss."
    print_error "============================================================"
    exit 1
fi

# CRITICAL SAFETY CHECK: Block on remote databases
if [[ "$DB_HOST" != "localhost" ]] && [[ "$DB_HOST" != "127.0.0.1" ]] && [[ "$DB_HOST" != "::1" ]]; then
    print_error "============================================================"
    print_error "BLOCKED: Remote database detected!"
    print_error "============================================================"
    print_error "This script is only allowed on localhost databases."
    print_error "Host: $DB_HOST"
    print_error "Aborting to prevent data loss."
    print_error "============================================================"
    exit 1
fi

# Проверка подключения к БД
print_step "Проверка подключения к базе данных..."
if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; then
    print_error "Не удалось подключиться к базе данных"
    print_info "Убедитесь что PostgreSQL запущен и база данных создана"
    exit 1
fi
print_success "Подключение установлено"
echo ""

# Выбор режима
echo -e "${YELLOW}${BOLD}Выберите режим загрузки:${NC}"
echo "  1) Только добавить данные (безопасно, существующие данные сохранятся)"
echo "  2) Очистить ВСЕ данные и загрузить заново (ОПАСНО!)"
echo ""
read -p "Ваш выбор (1/2): " mode_choice

TRUNCATE_FLAG=""
if [[ "$mode_choice" == "2" ]]; then
    echo ""
    print_error "============================================================"
    print_error "ВНИМАНИЕ: Все данные в базе $DB_NAME будут УДАЛЕНЫ!"
    print_error "============================================================"
    echo ""
    echo -e "${YELLOW}Введите 'DELETE ALL DATA' для подтверждения:${NC}"
    read -r confirm
    if [[ "$confirm" != "DELETE ALL DATA" ]]; then
        print_success "Отменено. Данные не были удалены."
        exit 0
    fi
    TRUNCATE_FLAG="--truncate --yes"
elif [[ "$mode_choice" != "1" ]]; then
    print_error "Неверный выбор. Отменено."
    exit 1
fi

echo ""
print_step "Загрузка тестовых данных..."
echo ""

# Загрузка данных через load-data.sh
export DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD
bash "$SCRIPT_DIR/backend/scripts/load-data.sh" $TRUNCATE_FLAG

if [ $? -eq 0 ]; then
    echo ""
    print_success "Тестовые данные успешно загружены!"
    echo ""

    # Вывод информации об учетных данных
    echo -e "${GREEN}${BOLD}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║                                                               ║"
    echo "║              УЧЕТНЫЕ ДАННЫЕ ДЛЯ ВХОДА                         ║"
    echo "║                                                               ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo ""

    echo -e "${YELLOW}${BOLD}Все пароли:${NC} password123"
    echo ""

    echo -e "${BLUE}${BOLD}АДМИНИСТРАТОР:${NC}"
    echo "  Email: admin@thebot.com"
    echo "  Имя:   Администратор системы"
    echo ""

    echo -e "${BLUE}${BOLD}МЕТОДИСТЫ (Наставники):${NC}"
    echo "  1. methodologist1@thebot.com - Елена Петрова"
    echo "  2. methodologist2@thebot.com - Иван Сидоров"
    echo "  3. methodologist3@thebot.com - Мария Волкова"
    echo ""

    echo -e "${BLUE}${BOLD}СТУДЕНТЫ:${NC}"
    echo "  1. student1@thebot.com       - Алексей Иванов"
    echo "  2. student2@thebot.com       - Анна Смирнова"
    echo "  3. student3@thebot.com       - Дмитрий Козлов"
    echo "  4. student4@thebot.com       - Елена Морозова"
    echo "  5. student5@thebot.com       - Павел Новиков"
    echo ""

    echo -e "${GREEN}${BOLD}СОЗДАНО:${NC}"
    echo "  • 1 администратор"
    echo "  • 3 методиста (наставника)"
    echo "  • 5 студентов"
    echo "  • Занятия и расписания"
    echo "  • История платежей и бронирований"
    echo "  • Чаты и коммуникация"
    echo ""

    print_info "Теперь вы можете войти на http://localhost:3000"

else
    echo ""
    print_error "Ошибка при загрузке данных"
    exit 1
fi
