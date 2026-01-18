#!/bin/bash

# Скрипт настройки Telegram webhook для production deployment
# Использование: ./setup_telegram_webhook.sh [OPTIONS]

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции вывода
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

# Проверка наличия curl
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        print_error "curl не установлен. Установите curl для продолжения."
        exit 1
    fi
    print_success "Зависимости проверены"
}

# Загрузка переменных окружения
load_env() {
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
        print_success "Переменные окружения загружены из .env"
    else
        print_warning "Файл .env не найден, используются системные переменные"
    fi
}

# Проверка обязательных переменных
check_env_vars() {
    local missing=0

    if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
        print_error "TELEGRAM_BOT_TOKEN не задан"
        missing=1
    fi

    if [ "$1" != "--delete" ] && [ "$1" != "--check" ] && [ -z "$TELEGRAM_WEBHOOK_URL" ]; then
        print_error "TELEGRAM_WEBHOOK_URL не задан"
        missing=1
    fi

    if [ $missing -eq 1 ]; then
        echo ""
        print_info "Убедитесь, что в файле .env или переменных окружения заданы:"
        print_info "  - TELEGRAM_BOT_TOKEN"
        print_info "  - TELEGRAM_WEBHOOK_URL (для установки webhook)"
        exit 1
    fi

    print_success "Переменные окружения проверены"
}

# Получение информации о боте
get_bot_info() {
    print_header "Информация о боте"

    local response=$(curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getMe")
    local ok=$(echo $response | grep -o '"ok":true' || true)

    if [ -z "$ok" ]; then
        print_error "Не удалось получить информацию о боте"
        print_error "Ответ: $response"
        exit 1
    fi

    local username=$(echo $response | grep -oP '(?<="username":")[^"]*')
    local first_name=$(echo $response | grep -oP '(?<="first_name":")[^"]*')

    print_success "Бот найден:"
    echo "  Имя: $first_name"
    echo "  Username: @$username"
}

# Установка webhook
set_webhook() {
    print_header "Установка webhook"

    print_info "URL: $TELEGRAM_WEBHOOK_URL"

    local response=$(curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" \
        -H "Content-Type: application/json" \
        -d "{\"url\":\"${TELEGRAM_WEBHOOK_URL}\",\"allowed_updates\":[\"message\",\"callback_query\"]}")

    local ok=$(echo $response | grep -o '"ok":true' || true)

    if [ -z "$ok" ]; then
        print_error "Не удалось установить webhook"
        print_error "Ответ: $response"
        exit 1
    fi

    local description=$(echo $response | grep -oP '(?<="description":")[^"]*' || echo "Webhook установлен")

    print_success "$description"
}

# Получение информации о webhook
get_webhook_info() {
    print_header "Статус webhook"

    local response=$(curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getWebhookInfo")
    local ok=$(echo $response | grep -o '"ok":true' || true)

    if [ -z "$ok" ]; then
        print_error "Не удалось получить информацию о webhook"
        print_error "Ответ: $response"
        exit 1
    fi

    local url=$(echo $response | grep -oP '(?<="url":")[^"]*' || echo "")
    local pending_update_count=$(echo $response | grep -oP '(?<="pending_update_count":)\d+' || echo "0")
    local last_error_date=$(echo $response | grep -oP '(?<="last_error_date":)\d+' || echo "")
    local last_error_message=$(echo $response | grep -oP '(?<="last_error_message":")[^"]*' || echo "")

    if [ -z "$url" ]; then
        print_warning "Webhook не установлен (используется polling режим)"
    else
        print_success "Webhook активен:"
        echo "  URL: $url"
        echo "  Ожидающих обновлений: $pending_update_count"

        if [ -n "$last_error_date" ]; then
            print_warning "Последняя ошибка:"
            echo "  Дата: $(date -d @$last_error_date '+%Y-%m-%d %H:%M:%S')"
            echo "  Сообщение: $last_error_message"
        else
            print_success "Ошибок нет"
        fi
    fi
}

# Удаление webhook
delete_webhook() {
    print_header "Удаление webhook"

    local response=$(curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/deleteWebhook")
    local ok=$(echo $response | grep -o '"ok":true' || true)

    if [ -z "$ok" ]; then
        print_error "Не удалось удалить webhook"
        print_error "Ответ: $response"
        exit 1
    fi

    print_success "Webhook удален (бот переключен в polling режим)"
}

# Справка
show_help() {
    echo "Скрипт настройки Telegram webhook"
    echo ""
    echo "Использование:"
    echo "  $0                  Установить webhook"
    echo "  $0 --check          Проверить статус webhook"
    echo "  $0 --delete         Удалить webhook (переключить на polling)"
    echo "  $0 --help           Показать эту справку"
    echo ""
    echo "Переменные окружения (.env или system):"
    echo "  TELEGRAM_BOT_TOKEN      Токен бота от @BotFather (обязательно)"
    echo "  TELEGRAM_WEBHOOK_URL    URL для webhook (обязательно для установки)"
    echo ""
    echo "Примеры:"
    echo "  ./setup_telegram_webhook.sh"
    echo "  ./setup_telegram_webhook.sh --check"
    echo "  ./setup_telegram_webhook.sh --delete"
}

# Основная логика
main() {
    print_header "Telegram Webhook Setup"

    # Проверка зависимостей
    check_dependencies

    # Обработка команд
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        --check)
            load_env
            check_env_vars "--check"
            get_bot_info
            get_webhook_info
            ;;
        --delete)
            load_env
            check_env_vars "--delete"
            get_bot_info
            delete_webhook
            ;;
        "")
            load_env
            check_env_vars
            get_bot_info
            set_webhook
            get_webhook_info
            ;;
        *)
            print_error "Неизвестная команда: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac

    echo ""
    print_success "Готово!"
}

# Запуск
main "$@"
