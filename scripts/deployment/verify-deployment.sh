#!/bin/bash

set -euo pipefail

REMOTE_USER="mg"
REMOTE_HOST="5.129.249.206"
REMOTE_ADDR="${REMOTE_USER}@${REMOTE_HOST}"
THEBOT_HOME="/home/mg/the-bot"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

show_help() {
    cat << 'EOF'
Verify THE_BOT V3 Deployment

Usage: ./verify-deployment.sh [OPTIONS]

Options:
  --full           Run comprehensive health checks
  --logs           Show recent service logs
  --stats          Show deployment statistics
  --services       Check service status only
  --database       Check database connectivity
  --redis          Check Redis connectivity
  --help           Show this help message

Examples:
  # Quick status check
  ./verify-deployment.sh

  # Comprehensive verification
  ./verify-deployment.sh --full

  # View recent logs
  ./verify-deployment.sh --logs

  # Database and Redis health
  ./verify-deployment.sh --database --redis

EOF
}

verify_ssh() {
    if ! ssh -q -o ConnectTimeout=5 "$REMOTE_ADDR" "echo ok" > /dev/null 2>&1; then
        echo -e "${RED}✗ Cannot connect to $REMOTE_ADDR${NC}"
        exit 1
    fi
}

print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

check_services() {
    print_header "Service Status"

    SERVICES=(
        "thebot-backend.service"
        "thebot-daphne.service"
        "thebot-celery-worker.service"
        "thebot-celery-beat.service"
    )

    local all_running=true

    for service in "${SERVICES[@]}"; do
        if ssh -q "$REMOTE_ADDR" "systemctl is-active --quiet $service 2>/dev/null"; then
            echo -e "  ${GREEN}✓${NC} $service"
        else
            echo -e "  ${RED}✗${NC} $service"
            all_running=false
        fi
    done

    if [ "$all_running" = true ]; then
        echo -e "\n${GREEN}All services running${NC}"
        return 0
    else
        echo -e "\n${RED}Some services are not running${NC}"
        return 1
    fi
}

check_http_endpoints() {
    print_header "HTTP Endpoints"

    endpoints=(
        "http://localhost:8000/health"
        "http://localhost:8001/ws"
    )

    for endpoint in "${endpoints[@]}"; do
        echo -n "  Checking $endpoint ... "
        if ssh -q "$REMOTE_ADDR" "curl -s -m 5 $endpoint > /dev/null 2>&1"; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${YELLOW}⚠${NC} (may be expected)"
        fi
    done
}

check_database() {
    print_header "Database Connectivity"

    echo "  Attempting to connect to database..."

    if ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME/backend && psql \$DATABASE_URL -tc 'SELECT 1' > /dev/null 2>&1"; then
        echo -e "  ${GREEN}✓ Database connection successful${NC}"

        echo ""
        echo "  Chat statistics:"
        CHAT_COUNT=$(ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME/backend && psql \$DATABASE_URL -tc \"SELECT COUNT(*) FROM chats\" 2>/dev/null")
        echo -e "    Total chats: ${CYAN}$CHAT_COUNT${NC}"

        return 0
    else
        echo -e "  ${YELLOW}⚠ Could not connect to database${NC}"
        return 1
    fi
}

check_redis() {
    print_header "Redis Connectivity"

    echo "  Attempting to connect to Redis..."

    if ssh -q "$REMOTE_ADDR" "redis-cli ping > /dev/null 2>&1"; then
        echo -e "  ${GREEN}✓ Redis connection successful${NC}"

        REDIS_MEMORY=$(ssh -q "$REMOTE_ADDR" "redis-cli info memory | grep used_memory_human | cut -d: -f2")
        echo "    Memory usage: ${CYAN}$REDIS_MEMORY${NC}"

        return 0
    else
        echo -e "  ${YELLOW}⚠ Could not connect to Redis${NC}"
        return 1
    fi
}

show_logs() {
    print_header "Recent Service Logs (last 20 lines)"

    SERVICES=(
        "thebot-backend.service"
        "thebot-daphne.service"
        "thebot-celery-worker.service"
    )

    for service in "${SERVICES[@]}"; do
        echo ""
        echo -e "${CYAN}--- $service ---${NC}"
        ssh -q "$REMOTE_ADDR" "journalctl -u $service -n 20 --no-pager" || echo "No logs available"
    done
}

show_stats() {
    print_header "Deployment Statistics"

    echo "  System Resources:"
    ssh -q "$REMOTE_ADDR" "free -h | tail -2" | sed 's/^/    /'

    echo ""
    echo "  Disk Usage:"
    ssh -q "$REMOTE_ADDR" "df -h $THEBOT_HOME | tail -1" | awk '{print "    " $1 ": " $2 " total, " $3 " used, " $4 " free (" $5 " used)"}' || true

    echo ""
    echo "  Active Services:"
    ssh -q "$REMOTE_ADDR" "systemctl list-units --type=service --state=running | grep thebot | wc -l" | xargs echo "    " || true

    echo ""
    echo "  Backend Version:"
    ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME/backend && git log -1 --oneline" | sed 's/^/    /' || echo "    Not available"

    echo ""
    echo "  Frontend Build Date:"
    ssh -q "$REMOTE_ADDR" "stat -c %y $THEBOT_HOME/frontend/dist/index.html 2>/dev/null || echo 'Not available'" | sed 's/^/    /' || true
}

quick_status() {
    print_header "Deployment Status"

    verify_ssh
    check_services
    echo ""
    echo -e "For more details, run with ${CYAN}--full${NC} flag"
}

FULL_CHECK=false
SHOW_LOGS=false
SHOW_STATS=false
SERVICES_ONLY=false
CHECK_DB=false
CHECK_REDIS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --full)
            FULL_CHECK=true
            shift
            ;;
        --logs)
            SHOW_LOGS=true
            shift
            ;;
        --stats)
            SHOW_STATS=true
            shift
            ;;
        --services)
            SERVICES_ONLY=true
            shift
            ;;
        --database)
            CHECK_DB=true
            shift
            ;;
        --redis)
            CHECK_REDIS=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

if [ "$FULL_CHECK" = true ]; then
    verify_ssh
    check_services
    check_http_endpoints
    check_database
    check_redis
    show_stats
elif [ "$SERVICES_ONLY" = true ]; then
    verify_ssh
    check_services
elif [ "$SHOW_LOGS" = true ]; then
    verify_ssh
    show_logs
elif [ "$SHOW_STATS" = true ]; then
    verify_ssh
    show_stats
elif [ "$CHECK_DB" = true ] || [ "$CHECK_REDIS" = true ]; then
    verify_ssh
    [ "$CHECK_DB" = true ] && check_database
    [ "$CHECK_REDIS" = true ] && check_redis
else
    quick_status
fi

echo ""
echo -e "${CYAN}Verification complete${NC}"
echo ""
