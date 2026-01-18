#!/bin/bash

# ╔═══════════════════════════════════════════════════════════════╗
# ║  Автозапуск платформы для локальной разработки                ║
# ║  Tutoring Platform Development Starter                        ║
# ╚═══════════════════════════════════════════════════════════════╝

set -e  # Остановка при ошибке

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Функция для красивого вывода
print_step() {
    echo -e "${BLUE}${BOLD}[$(date +%H:%M:%S)]${NC} ${CYAN}➜${NC} $1"
}

print_success() {
    echo -e "${GREEN}${BOLD}✓${NC} $1"
}

print_error() {
    echo -e "${RED}${BOLD}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}${BOLD}⚠${NC} $1"
}

print_info() {
    echo -e "${MAGENTA}${BOLD}ℹ${NC} $1"
}

# Получение директории скрипта
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Логирование
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/dev-$(date +%Y%m%d-%H%M%S).log"

# Функция логирования
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Баннер
clear
echo -e "${MAGENTA}${BOLD}"
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                                                               ║"
echo "║         🎓 Tutoring Platform - Development Starter           ║"
echo "║                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""

log "===== Запуск платформы в режиме разработки ====="

# ═══════════════════════════════════════════════════════════════
# 1. ПРОВЕРКА ЗАВИСИМОСТЕЙ
# ═══════════════════════════════════════════════════════════════

print_step "Проверка установленных зависимостей..."
echo ""

MISSING_DEPS=()

# Проверка Go
if ! command -v go &> /dev/null; then
    print_error "Go не установлен"
    MISSING_DEPS+=("Go 1.21+")
else
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go установлен: $GO_VERSION"
fi

# Проверка Node.js
if ! command -v node &> /dev/null; then
    print_error "Node.js не установлен"
    MISSING_DEPS+=("Node.js 20+")
else
    NODE_VERSION=$(node --version)
    print_success "Node.js установлен: $NODE_VERSION"
fi

# Проверка npm
if ! command -v npm &> /dev/null; then
    print_error "npm не установлен"
    MISSING_DEPS+=("npm")
else
    NPM_VERSION=$(npm --version)
    print_success "npm установлен: v$NPM_VERSION"
fi

# Проверка PostgreSQL
if ! command -v psql &> /dev/null; then
    print_warning "psql не найден в PATH (PostgreSQL может быть установлен, но не в PATH)"
    print_info "Будет сделана попытка подключения..."
else
    PSQL_VERSION=$(psql --version | awk '{print $3}')
    print_success "PostgreSQL установлен: $PSQL_VERSION"
fi

echo ""

# Если есть недостающие зависимости
if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    print_error "Недостающие зависимости:"
    for dep in "${MISSING_DEPS[@]}"; do
        echo "  - $dep"
    done
    echo ""
    print_info "Установите недостающие зависимости и запустите скрипт снова."
    exit 1
fi

# ═══════════════════════════════════════════════════════════════
# 2. ЗАГРУЗКА КОНФИГУРАЦИИ
# ═══════════════════════════════════════════════════════════════

print_step "Загрузка конфигурации..."
echo ""

# Проверка .env файла
if [ ! -f "$SCRIPT_DIR/backend/.env" ]; then
    print_warning ".env файл не найден, создаю из .env.example..."

    if [ -f "$SCRIPT_DIR/backend/.env.example" ]; then
        cp "$SCRIPT_DIR/backend/.env.example" "$SCRIPT_DIR/backend/.env"
        print_success ".env файл создан"
        print_info "Отредактируйте backend/.env при необходимости"
    else
        print_error ".env.example не найден"
        exit 1
    fi
else
    print_success ".env файл найден"
fi

# Загрузка переменных из .env
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

# Установка значений по умолчанию если не заданы
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-tutoring_platform}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_SSL_MODE=${DB_SSL_MODE:-disable}
SERVER_PORT=${SERVER_PORT:-8080}

# Автоматическая замена Docker-хоста на localhost для локального запуска
if [ "$DB_HOST" = "postgres" ] || [ "$DB_HOST" = "db" ]; then
    print_info "Обнаружен Docker-хост '$DB_HOST', заменяю на 'localhost' для локального запуска"
    DB_HOST="localhost"
    export DB_HOST="localhost"
fi

# Автоматическая замена режима на development для локального запуска
if [ "$DB_HOST" = "localhost" ] || [ "$DB_HOST" = "127.0.0.1" ]; then
    if [ "$ENV" = "production" ]; then
        print_info "Обнаружен режим 'production' для локального запуска, заменяю на 'development'"
        ENV="development"
        export ENV="development"
    fi
    
    # Автоматическая замена SSL режима для локального подключения
    if [ "$DB_SSL_MODE" = "require" ] || [ "$DB_SSL_MODE" = "verify-full" ] || [ "$DB_SSL_MODE" = "verify-ca" ]; then
        print_info "Обнаружен SSL режим '$DB_SSL_MODE' для локального подключения, заменяю на 'disable'"
        DB_SSL_MODE="disable"
        export DB_SSL_MODE="disable"
    fi
fi

# Экспортируем все переменные для backend
export DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD DB_SSL_MODE ENV

print_info "База данных: $DB_HOST:$DB_PORT/$DB_NAME"
print_info "Backend порт: $SERVER_PORT"
echo ""

log "Конфигурация загружена: DB=$DB_HOST:$DB_PORT/$DB_NAME"

# ═══════════════════════════════════════════════════════════════
# 3. ПРОВЕРКА И ЗАПУСК POSTGRESQL
# ═══════════════════════════════════════════════════════════════

print_step "Проверка PostgreSQL..."
echo ""

# Функция проверки подключения к PostgreSQL с выводом ошибок
check_postgres() {
    local error_output
    error_output=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c '\q' 2>&1)
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "$error_output" >&2
    fi
    return $exit_code
}

# Проверка статуса сервиса PostgreSQL (для Linux)
check_postgres_service() {
    if command -v systemctl &> /dev/null; then
        if systemctl is-active --quiet postgresql 2>/dev/null; then
            return 0
        fi
    fi
    return 1
}

# Проверка подключения
if check_postgres; then
    print_success "PostgreSQL работает и доступен"
else
    print_warning "PostgreSQL недоступен"
    
    # Показываем детали ошибки
    print_info "Детали ошибки подключения:"
    check_postgres > /dev/null 2>&1 || true
    echo ""

    # Проверяем статус сервиса
    if check_postgres_service; then
        print_info "Сервис PostgreSQL запущен, но подключение не удалось"
        
        # Попытка подключиться с пользователем postgres для проверки (без пароля для локальных подключений)
        if psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres -c '\q' 2>/dev/null || \
           PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres -c '\q' 2>/dev/null; then
            print_info "Подключение с пользователем 'postgres' успешно"
            print_info "Проблема: пользователь '$DB_USER' не существует или неправильный пароль"
            echo ""
            print_info "Попытка создать пользователя '$DB_USER'..."
            
            # Экранирование пароля для SQL (замена одинарных кавычек)
            ESCAPED_PASSWORD=$(echo "$DB_PASSWORD" | sed "s/'/''/g")
            
            # Попытка создать пользователя (пробуем без пароля и с паролем postgres)
            CREATE_USER_SQL="DO \$\$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = '$DB_USER') THEN CREATE USER \"$DB_USER\" WITH PASSWORD '$ESCAPED_PASSWORD'; END IF; END \$\$;"
            
            if echo "$CREATE_USER_SQL" | psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres 2>/dev/null || \
               echo "$CREATE_USER_SQL" | PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres 2>/dev/null; then
                print_success "Пользователь '$DB_USER' создан или уже существует"
                # Даем права на создание баз данных
                echo "ALTER USER \"$DB_USER\" CREATEDB;" | psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres 2>/dev/null || \
                echo "ALTER USER \"$DB_USER\" CREATEDB;" | PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres 2>/dev/null || true
                
                # Повторная проверка подключения
                sleep 1
                if check_postgres; then
                    print_success "Подключение к PostgreSQL успешно восстановлено"
                else
                    print_error "Пользователь создан, но подключение все еще не удается"
                    print_info "Проверьте пароль в backend/.env"
                    exit 1
                fi
            else
                print_warning "Не удалось создать пользователя автоматически"
                echo ""
                print_info "Создайте пользователя вручную:"
                echo "  sudo -u postgres psql -c \"CREATE USER $DB_USER WITH PASSWORD '$ESCAPED_PASSWORD';\""
                echo "  sudo -u postgres psql -c \"ALTER USER $DB_USER CREATEDB;\""
                echo ""
                print_info "Или измените DB_USER в backend/.env на 'postgres'"
                exit 1
            fi
        else
            print_info "Возможные причины:"
            print_info "  - Неправильные учетные данные (пользователь: $DB_USER)"
            print_info "  - Пользователь '$DB_USER' не существует в базе данных"
            print_info "  - Неправильный пароль"
            print_info "  - База данных не настроена"
            echo ""
            print_error "Не удалось подключиться к PostgreSQL"
            print_info "Проверьте настройки в backend/.env"
            exit 1
        fi
    else
        # Попытка запустить PostgreSQL (зависит от ОС)
        print_info "Попытка запуска PostgreSQL..."

        # Для Linux
        if command -v systemctl &> /dev/null; then
            print_info "Запускаю сервис PostgreSQL через systemctl..."
            if sudo systemctl start postgresql 2>&1; then
                sleep 3
                
                if check_postgres; then
                    print_success "PostgreSQL успешно запущен"
                else
                    print_error "PostgreSQL запущен, но подключение не удалось"
                    print_info "Проверьте настройки подключения в backend/.env"
                    exit 1
                fi
            else
                print_error "Не удалось запустить PostgreSQL через systemctl"
                print_info "Попробуйте запустить вручную: sudo systemctl start postgresql"
                exit 1
            fi
        # Для macOS
        elif command -v brew &> /dev/null; then
            if brew services start postgresql 2>&1; then
                sleep 3
                
                if check_postgres; then
                    print_success "PostgreSQL успешно запущен"
                else
                    print_error "PostgreSQL запущен, но подключение не удалось"
                    print_info "Проверьте настройки подключения в backend/.env"
                    exit 1
                fi
            else
                print_error "Не удалось запустить PostgreSQL через brew"
                print_info "Попробуйте запустить вручную: brew services start postgresql"
                exit 1
            fi
        else
            print_error "Не удалось определить способ запуска PostgreSQL"
            print_info "Запустите PostgreSQL вручную и запустите скрипт снова"
            exit 1
        fi
    fi
fi

echo ""
log "PostgreSQL проверен и доступен"

# ═══════════════════════════════════════════════════════════════
# 4. СОЗДАНИЕ И НАСТРОЙКА БАЗЫ ДАННЫХ
# ═══════════════════════════════════════════════════════════════

print_step "Проверка базы данных..."
echo ""

# Проверка существования базы данных
DB_EXISTS=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" 2>/dev/null || echo "0")

if [ "$DB_EXISTS" = "1" ]; then
    print_success "База данных '$DB_NAME' существует"
else
    print_warning "База данных '$DB_NAME' не найдена"
    print_info "Создаю базу данных..."

    PGPASSWORD="$DB_PASSWORD" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" 2>/dev/null

    if [ $? -eq 0 ]; then
        print_success "База данных '$DB_NAME' создана"
    else
        print_error "Не удалось создать базу данных"
        exit 1
    fi
fi

echo ""
log "База данных '$DB_NAME' готова"

# ═══════════════════════════════════════════════════════════════
# 5. ПРИМЕНЕНИЕ МИГРАЦИЙ
# ═══════════════════════════════════════════════════════════════

print_step "Проверка миграций..."
echo ""

# Проверка существования таблицы users
TABLES_EXIST=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_name='users'" 2>/dev/null || echo "0")

if [ "$TABLES_EXIST" = "0" ]; then
    print_warning "Таблицы не найдены, применяю миграции..."

    # Применение миграций
    MIGRATION_DIR="$SCRIPT_DIR/backend/internal/database/migrations"

    if [ -d "$MIGRATION_DIR" ]; then
        for migration in "$MIGRATION_DIR"/*.sql; do
            if [ -f "$migration" ]; then
                MIGRATION_NAME=$(basename "$migration")
                print_info "Применяю миграцию: $MIGRATION_NAME"

                PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration" > /dev/null 2>&1

                if [ $? -eq 0 ]; then
                    print_success "Миграция $MIGRATION_NAME применена"
                else
                    print_error "Ошибка при применении миграции $MIGRATION_NAME"
                    exit 1
                fi
            fi
        done
    else
        print_error "Директория миграций не найдена: $MIGRATION_DIR"
        exit 1
    fi
else
    print_success "Миграции уже применены"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# 5.1 ЗАПОЛНЕНИЕ БАЗЫ ДАННЫХ (SEED)
# ═══════════════════════════════════════════════════════════════

print_step "Заполнение базы данных тестовыми данными..."
echo ""

SEED_FILE="$SCRIPT_DIR/backend/scripts/seed.sql"

if [ ! -f "$SEED_FILE" ]; then
    print_error "Файл seed.sql не найден: $SEED_FILE"
    exit 1
fi

# Выполнение seed.sql с обработкой ошибок
print_info "Выполняю seed.sql..."

PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SEED_FILE" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    print_success "Тестовые данные загружены в базу"

    # Проверка количества пользователей
    USER_COUNT=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM users" 2>/dev/null || echo "0")
    print_info "Пользователей в базе: $USER_COUNT"
else
    print_error "Ошибка при загрузке тестовых данных"
    print_info "Проверьте файл: $SEED_FILE"
    exit 1
fi

echo ""
log "Миграции проверены и применены"
log "Тестовые данные загружены. Пользователей: $USER_COUNT"

# ═══════════════════════════════════════════════════════════════
# 6. УСТАНОВКА ЗАВИСИМОСТЕЙ BACKEND
# ═══════════════════════════════════════════════════════════════

print_step "Проверка зависимостей Backend (Go)..."
echo ""

cd "$SCRIPT_DIR/backend" || exit 1

if [ ! -f "go.mod" ]; then
    print_error "go.mod не найден в backend/"
    exit 1
fi

# Проверка и установка зависимостей
print_info "Загрузка Go модулей..."
go mod download > "$LOG_FILE.backend-deps" 2>&1

if [ $? -eq 0 ]; then
    print_success "Go зависимости загружены"
else
    print_error "Ошибка загрузки Go зависимостей"
    cat "$LOG_FILE.backend-deps"
    exit 1
fi

echo ""
log "Backend зависимости установлены"

# ═══════════════════════════════════════════════════════════════
# 7. УСТАНОВКА ЗАВИСИМОСТЕЙ FRONTEND
# ═══════════════════════════════════════════════════════════════

print_step "Проверка зависимостей Frontend (npm)..."
echo ""

cd "$SCRIPT_DIR/frontend"

if [ ! -f "package.json" ]; then
    print_error "package.json не найден в frontend/"
    exit 1
fi

# Всегда делаем чистую установку для соответствия текущей версии
print_info "Удаляю старые node_modules для чистой установки..."
rm -rf node_modules

print_info "Устанавливаю зависимости (npm ci)..."
npm ci > "$LOG_FILE.frontend-deps" 2>&1

if [ $? -eq 0 ]; then
    print_success "npm зависимости установлены (чистая установка)"
else
    print_error "Ошибка установки npm зависимостей"
    cat "$LOG_FILE.frontend-deps"
    exit 1
fi

echo ""
log "Frontend зависимости установлены"

# ═══════════════════════════════════════════════════════════════
# 8. СОЗДАНИЕ PID ФАЙЛОВ
# ═══════════════════════════════════════════════════════════════

PID_DIR="$SCRIPT_DIR/.pids"
mkdir -p "$PID_DIR"

BACKEND_PID_FILE="$PID_DIR/backend.pid"
FRONTEND_PID_FILE="$PID_DIR/frontend.pid"

# Функция очистки при выходе
cleanup() {
    echo ""
    print_step "Остановка сервисов..."

    if [ -f "$BACKEND_PID_FILE" ]; then
        BACKEND_PID=$(cat "$BACKEND_PID_FILE")
        if ps -p $BACKEND_PID > /dev/null 2>&1; then
            print_info "Останавливаю Backend (PID: $BACKEND_PID)..."
            kill $BACKEND_PID 2>/dev/null || true
            wait $BACKEND_PID 2>/dev/null || true
            print_success "Backend остановлен"
        fi
        rm -f "$BACKEND_PID_FILE"
    fi

    if [ -f "$FRONTEND_PID_FILE" ]; then
        FRONTEND_PID=$(cat "$FRONTEND_PID_FILE")
        if ps -p $FRONTEND_PID > /dev/null 2>&1; then
            print_info "Останавливаю Frontend (PID: $FRONTEND_PID)..."
            kill $FRONTEND_PID 2>/dev/null || true
            wait $FRONTEND_PID 2>/dev/null || true
            print_success "Frontend остановлен"
        fi
        rm -f "$FRONTEND_PID_FILE"
    fi

    echo ""
    print_success "Все сервисы остановлены"
    echo ""
    log "Платформа остановлена"
    exit 0
}

# Установка trap для обработки Ctrl+C
trap cleanup SIGINT SIGTERM

# ═══════════════════════════════════════════════════════════════
# 9. ЗАПУСК BACKEND
# ═══════════════════════════════════════════════════════════════

print_step "Запуск Backend сервера..."
echo ""

cd "$SCRIPT_DIR/backend"

# Проверка, не запущен ли уже backend
if lsof -Pi :$SERVER_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    print_warning "Порт $SERVER_PORT уже занят"
    print_info "Останавливаю существующий процесс..."

    PID=$(lsof -ti:$SERVER_PORT)
    kill $PID 2>/dev/null || true
    sleep 2
fi

# Сборка backend
print_info "Пересборка backend..."
if ! go build -o server ./cmd/server/; then
    print_error "Ошибка сборки backend"
    exit 1
fi
print_success "Backend собран"

# Запуск backend в фоне
print_info "Запускаю Go сервер на порту $SERVER_PORT..."

# Убеждаемся, что все переменные окружения экспортированы
export DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD DB_SSL_MODE ENV SERVER_PORT

./server > "$LOG_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > "$BACKEND_PID_FILE"

# Ожидание запуска backend
print_info "Ожидание запуска backend..."
sleep 3

# Проверка, что процесс запущен
if ps -p $BACKEND_PID > /dev/null 2>&1; then
    # Проверка HTTP ответа
    for i in {1..10}; do
        if curl -s http://localhost:$SERVER_PORT/api/v1/auth/me > /dev/null 2>&1; then
            print_success "Backend запущен (PID: $BACKEND_PID)"
            print_info "Backend доступен на http://localhost:$SERVER_PORT"
            break
        fi

        if [ $i -eq 10 ]; then
            print_error "Backend не отвечает на запросы"
            print_info "Проверьте логи: $LOG_DIR/backend.log"
            kill $BACKEND_PID 2>/dev/null || true
            exit 1
        fi

        sleep 1
    done
else
    print_error "Backend не запустился"
    print_info "Проверьте логи: $LOG_DIR/backend.log"
    exit 1
fi

echo ""
log "Backend запущен (PID: $BACKEND_PID)"

# ═══════════════════════════════════════════════════════════════
# 10. ЗАПУСК FRONTEND
# ═══════════════════════════════════════════════════════════════

print_step "Запуск Frontend dev сервера..."
echo ""

cd "$SCRIPT_DIR/frontend"

# Освобождаем порты Vite (5173, 5174)
for port in 5173 5174; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "Порт $port занят, освобождаю..."
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
        sleep 1
    fi
done

# Запуск frontend в фоне
print_info "Запускаю Vite dev сервер..."

npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > "$FRONTEND_PID_FILE"

# Ожидание запуска frontend
print_info "Ожидание запуска frontend..."
sleep 3

# Определяем на каком порту запустился Vite (читаем из лога)
FRONTEND_PORT=""
for i in {1..20}; do
    # Ищем порт в логе Vite
    if grep -q "Local:" "$LOG_DIR/frontend.log" 2>/dev/null; then
        FRONTEND_PORT=$(grep "Local:" "$LOG_DIR/frontend.log" | grep -oP 'localhost:\K[0-9]+' | head -1)
        break
    fi
    sleep 0.5
done

# Если не нашли в логе, пробуем стандартные порты
if [ -z "$FRONTEND_PORT" ]; then
    for port in 5173 5174 5175; do
        if curl -s "http://localhost:$port" > /dev/null 2>&1; then
            FRONTEND_PORT=$port
            break
        fi
    done
fi

# Проверка, что процесс запущен
if ps -p $FRONTEND_PID > /dev/null 2>&1 && [ -n "$FRONTEND_PORT" ]; then
    # Проверка HTTP ответа
    for i in {1..10}; do
        if curl -s "http://localhost:$FRONTEND_PORT" > /dev/null 2>&1; then
            print_success "Frontend запущен (PID: $FRONTEND_PID)"
            print_info "Frontend доступен на http://localhost:$FRONTEND_PORT"
            break
        fi

        if [ $i -eq 10 ]; then
            print_error "Frontend не отвечает на запросы"
            print_info "Проверьте логи: $LOG_DIR/frontend.log"
            kill $FRONTEND_PID 2>/dev/null || true
            cleanup
            exit 1
        fi

        sleep 1
    done
else
    print_error "Frontend не запустился"
    print_info "Проверьте логи: $LOG_DIR/frontend.log"
    cleanup
    exit 1
fi

echo ""
log "Frontend запущен (PID: $FRONTEND_PID)"

# ═══════════════════════════════════════════════════════════════
# 11. ИТОГОВАЯ ИНФОРМАЦИЯ
# ═══════════════════════════════════════════════════════════════

echo ""
echo -e "${GREEN}${BOLD}"
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                                                               ║"
echo "║              ✓ Платформа успешно запущена!                   ║"
echo "║                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""

print_info "Сервисы:"
echo "  ┌─────────────────────────────────────────────────────────┐"
echo "  │ Frontend:  http://localhost:$FRONTEND_PORT                           │"
echo "  │ Backend:   http://localhost:$SERVER_PORT                             │"
echo "  │ Database:  $DB_HOST:$DB_PORT/$DB_NAME           │"
echo "  └─────────────────────────────────────────────────────────┘"
echo ""

print_info "PIDs:"
echo "  ┌─────────────────────────────────────────────────────────┐"
echo "  │ Backend PID:  $BACKEND_PID                                      │"
echo "  │ Frontend PID: $FRONTEND_PID                                      │"
echo "  └─────────────────────────────────────────────────────────┘"
echo ""

print_info "Логи:"
echo "  ┌─────────────────────────────────────────────────────────┐"
echo "  │ Backend:  $LOG_DIR/backend.log"
echo "  │ Frontend: $LOG_DIR/frontend.log"
echo "  │ Main:     $LOG_FILE"
echo "  └─────────────────────────────────────────────────────────┘"
echo ""

print_info "Команды:"
echo "  ┌─────────────────────────────────────────────────────────┐"
echo "  │ Остановить:   Ctrl+C                                    │"
echo "  │ Backend логи: tail -f $LOG_DIR/backend.log"
echo "  │ Frontend логи: tail -f $LOG_DIR/frontend.log"
echo "  └─────────────────────────────────────────────────────────┘"
echo ""

# Попытка открыть браузер
print_info "Открываю браузер..."
if command -v xdg-open &> /dev/null; then
    xdg-open "http://localhost:$FRONTEND_PORT" 2>/dev/null &
elif command -v open &> /dev/null; then
    open "http://localhost:$FRONTEND_PORT" 2>/dev/null &
elif command -v start &> /dev/null; then
    start "http://localhost:$FRONTEND_PORT" 2>/dev/null &
else
    print_warning "Не удалось открыть браузер автоматически"
    print_info "Откройте вручную: http://localhost:$FRONTEND_PORT"
fi

echo ""
echo -e "${YELLOW}${BOLD}Нажмите Ctrl+C для остановки всех сервисов${NC}"
echo ""

log "Платформа запущена успешно"

# ═══════════════════════════════════════════════════════════════
# 12. МОНИТОРИНГ ПРОЦЕССОВ
# ═══════════════════════════════════════════════════════════════

# Бесконечный цикл для мониторинга
while true; do
    sleep 5

    # Проверка backend
    if ! ps -p $BACKEND_PID > /dev/null 2>&1; then
        print_error "Backend процесс завершился неожиданно!"
        print_info "Проверьте логи: $LOG_DIR/backend.log"
        cleanup
        exit 1
    fi

    # Проверка frontend
    if ! ps -p $FRONTEND_PID > /dev/null 2>&1; then
        print_error "Frontend процесс завершился неожиданно!"
        print_info "Проверьте логи: $LOG_DIR/frontend.log"
        cleanup
        exit 1
    fi
done
