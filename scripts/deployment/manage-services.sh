#!/bin/bash

set -euo pipefail

REMOTE_USER="mg"
REMOTE_HOST="5.129.249.206"
REMOTE_ADDR="${REMOTE_USER}@${REMOTE_HOST}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

SERVICES=(
    "thebot-backend.service"
    "thebot-daphne.service"
    "thebot-celery-worker.service"
    "thebot-celery-beat.service"
)

show_help() {
    cat << 'EOF'
Service Management for THE_BOT V3

Usage: ./manage-services.sh <COMMAND> [OPTIONS]

Commands:
  status              Show status of all services
  start               Start all services
  stop                Stop all services
  restart             Restart all services
  logs <service>      Show logs for specific service
  tail                Tail logs for all services
  enable              Enable services on system boot
  disable             Disable services on system boot
  help                Show this help message

Examples:
  # Check service status
  ./manage-services.sh status

  # Restart all services
  ./manage-services.sh restart

  # View backend logs
  ./manage-services.sh logs thebot-backend

  # Tail all service logs (live)
  ./manage-services.sh tail

  # Start services on boot
  ./manage-services.sh enable

Service Names:
  - thebot-backend.service
  - thebot-daphne.service
  - thebot-celery-worker.service
  - thebot-celery-beat.service

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
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}════════════════════════════════════════${NC}"
}

status() {
    print_header "Service Status"

    for service in "${SERVICES[@]}"; do
        if ssh -q "$REMOTE_ADDR" "systemctl is-active --quiet $service 2>/dev/null"; then
            UPTIME=$(ssh -q "$REMOTE_ADDR" "systemctl show -p ActiveEnterTimestamp --value $service")
            echo -e "  ${GREEN}●${NC} $service (running since $UPTIME)"
        elif ssh -q "$REMOTE_ADDR" "systemctl is-enabled --quiet $service 2>/dev/null"; then
            echo -e "  ${YELLOW}●${NC} $service (enabled but not running)"
        else
            echo -e "  ${RED}●${NC} $service (stopped)"
        fi
    done

    echo ""
    echo -e "${CYAN}System Resources:${NC}"
    ssh -q "$REMOTE_ADDR" "ps aux | grep -E 'thebot-backend|thebot-daphne|thebot-celery' | grep -v grep | awk '{printf \"  %-40s CPU: %5.1f%% MEM: %5.1f%%\n\", \$11, \$3, \$4}'" || true
    echo ""
}

start_services() {
    print_header "Starting Services"

    for service in "${SERVICES[@]}"; do
        echo -n "  Starting $service ... "
        if ssh -q "$REMOTE_ADDR" "systemctl start $service"; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
    done

    echo ""
    sleep 3
    status
}

stop_services() {
    print_header "Stopping Services"

    for service in "${SERVICES[@]}"; do
        echo -n "  Stopping $service ... "
        if ssh -q "$REMOTE_ADDR" "systemctl stop $service"; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
    done

    echo ""
}

restart_services() {
    print_header "Restarting Services"

    for service in "${SERVICES[@]}"; do
        echo -n "  Restarting $service ... "
        if ssh -q "$REMOTE_ADDR" "systemctl restart $service"; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
    done

    echo ""
    sleep 3
    status
}

enable_services() {
    print_header "Enabling Services on Boot"

    for service in "${SERVICES[@]}"; do
        echo -n "  Enabling $service ... "
        if ssh -q "$REMOTE_ADDR" "systemctl enable $service"; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
    done

    echo ""
}

disable_services() {
    print_header "Disabling Services on Boot"

    for service in "${SERVICES[@]}"; do
        echo -n "  Disabling $service ... "
        if ssh -q "$REMOTE_ADDR" "systemctl disable $service"; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
    done

    echo ""
}

show_logs() {
    local service=$1
    local lines=${2:-50}

    print_header "Logs for $service (last $lines lines)"
    echo ""

    ssh -q "$REMOTE_ADDR" "journalctl -u $service -n $lines --no-pager" || echo "No logs available"

    echo ""
}

tail_logs() {
    print_header "Tailing Service Logs (Press Ctrl+C to exit)"
    echo ""

    ssh "$REMOTE_ADDR" "journalctl -u thebot-backend.service -u thebot-daphne.service -u thebot-celery-worker.service -f --no-pager" || echo "Failed to tail logs"

    echo ""
}

if [ $# -eq 0 ]; then
    show_help
    exit 0
fi

COMMAND=$1
shift || true

case "$COMMAND" in
    status)
        verify_ssh
        status
        ;;
    start)
        verify_ssh
        start_services
        ;;
    stop)
        verify_ssh
        stop_services
        ;;
    restart)
        verify_ssh
        restart_services
        ;;
    logs)
        verify_ssh
        SERVICE=${1:-thebot-backend.service}
        LINES=${2:-50}
        show_logs "$SERVICE" "$LINES"
        ;;
    tail)
        verify_ssh
        tail_logs
        ;;
    enable)
        verify_ssh
        enable_services
        ;;
    disable)
        verify_ssh
        disable_services
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "Unknown command: $COMMAND"
        echo ""
        show_help
        exit 1
        ;;
esac

exit 0
