#!/bin/bash

# Setup Chat Creation System for Production
# Purpose: Install and configure chat creation automation
# Usage: ./setup-chat-creation.sh [--timer|--cron] [--yes]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

SCHEDULER_TYPE="timer"
AUTO_YES=false
DRY_RUN=false

# Helper functions
print_header() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}\n"
}

print_section() {
    echo -e "\n${YELLOW}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

confirm() {
    if [ "$AUTO_YES" = true ]; then
        return 0
    fi

    local prompt="$1"
    local response

    read -p "$(echo -e ${YELLOW}$prompt${NC}) (yes/no): " -r response
    [[ "$response" =~ ^[Yy]$ ]]
}

show_help() {
    cat << 'EOF'
Setup Chat Creation System for Production

Usage: ./setup-chat-creation.sh [OPTIONS]

Options:
  --timer          Use systemd timer (recommended, default)
  --cron           Use crontab instead of systemd timer
  --yes            Skip confirmation prompts
  --dry-run        Show what would be done without making changes
  --help, -h       Show this help message

Examples:
  # Install with systemd timer (default)
  sudo ./setup-chat-creation.sh

  # Install with crontab
  sudo ./setup-chat-creation.sh --cron

  # Preview changes
  ./setup-chat-creation.sh --dry-run

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --timer)
                SCHEDULER_TYPE="timer"
                shift
                ;;
            --cron)
                SCHEDULER_TYPE="cron"
                shift
                ;;
            --yes)
                AUTO_YES=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

check_requirements() {
    print_section "Checking requirements..."

    local all_ok=true

    # Check if running as root
    if [ "$EUID" -ne 0 ]; then
        print_error "This script must be run as root (use: sudo ./setup-chat-creation.sh)"
        all_ok=false
    else
        print_success "Running as root"
    fi

    # Check psql
    if command -v psql &> /dev/null; then
        print_success "PostgreSQL client found"
    else
        print_error "PostgreSQL client not found (install with: apt-get install postgresql-client)"
        all_ok=false
    fi

    # Check bash version (need 4.0+)
    if [ "${BASH_VERSINFO[0]}" -ge 4 ]; then
        print_success "Bash version: ${BASH_VERSION}"
    else
        print_error "Bash 4.0+ required (current: $BASH_VERSION)"
        all_ok=false
    fi

    # Check if systemd is available (for timer)
    if [ "$SCHEDULER_TYPE" = "timer" ]; then
        if command -v systemctl &> /dev/null; then
            print_success "systemd found"
        else
            print_error "systemd not found (use --cron for crontab-based scheduling)"
            all_ok=false
        fi
    fi

    if [ "$all_ok" = false ]; then
        exit 1
    fi
}

check_scripts() {
    print_section "Checking script files..."

    local chat_creation_cron="$SCRIPT_DIR/chat-creation-cron.sh"
    local verify_script="$SCRIPT_DIR/verify-chat-creation.sh"

    if [ -f "$chat_creation_cron" ]; then
        print_success "chat-creation-cron.sh found"
    else
        print_error "chat-creation-cron.sh not found"
        exit 1
    fi

    if [ -f "$verify_script" ]; then
        print_success "verify-chat-creation.sh found"
    else
        print_error "verify-chat-creation.sh not found"
        exit 1
    fi

    # Check permissions
    if [ -x "$chat_creation_cron" ]; then
        print_success "chat-creation-cron.sh is executable"
    else
        print_error "chat-creation-cron.sh is not executable"
        if [ "$DRY_RUN" = false ]; then
            chmod +x "$chat_creation_cron"
            print_success "Fixed permissions on chat-creation-cron.sh"
        fi
    fi
}

install_timer() {
    print_section "Installing systemd timer..."

    local service_file="/etc/systemd/system/thebot-chat-creation.service"
    local timer_file="/etc/systemd/system/thebot-chat-creation.timer"

    local local_service_file="$SCRIPT_DIR/thebot-chat-creation.service"
    local local_timer_file="$SCRIPT_DIR/thebot-chat-creation.timer"

    if [ ! -f "$local_service_file" ]; then
        print_error "Service file not found: $local_service_file"
        exit 1
    fi

    if [ ! -f "$local_timer_file" ]; then
        print_error "Timer file not found: $local_timer_file"
        exit 1
    fi

    print_info "Installing service file: $service_file"
    if [ "$DRY_RUN" = false ]; then
        cp "$local_service_file" "$service_file"
        chmod 644 "$service_file"
        print_success "Installed service file"
    fi

    print_info "Installing timer file: $timer_file"
    if [ "$DRY_RUN" = false ]; then
        cp "$local_timer_file" "$timer_file"
        chmod 644 "$timer_file"
        print_success "Installed timer file"
    fi

    print_info "Reloading systemd configuration"
    if [ "$DRY_RUN" = false ]; then
        systemctl daemon-reload
        print_success "Daemon reloaded"
    fi

    print_info "Enabling timer"
    if [ "$DRY_RUN" = false ]; then
        systemctl enable thebot-chat-creation.timer
        print_success "Timer enabled"
    fi

    print_info "Starting timer"
    if [ "$DRY_RUN" = false ]; then
        systemctl start thebot-chat-creation.timer
        print_success "Timer started"
    fi

    echo ""
    echo "Timer status:"
    systemctl status thebot-chat-creation.timer --no-pager || true
}

install_cron() {
    print_section "Installing crontab entry..."

    local cron_job="*/30 * * * * export DB_HOST=localhost DB_NAME=tutoring_platform DB_USER=postgres && $SCRIPT_DIR/chat-creation-cron.sh"
    local cron_file="/var/spool/cron/crontabs/root"

    print_info "Adding crontab entry for postgres user"
    print_info "Schedule: Every 30 minutes"
    print_info "Command: $cron_job"

    if [ "$DRY_RUN" = false ]; then
        # Check if entry already exists
        if crontab -l 2>/dev/null | grep -q "chat-creation-cron.sh"; then
            print_info "Crontab entry already exists, skipping"
        else
            (crontab -l 2>/dev/null || true; echo "$cron_job") | crontab -
            print_success "Crontab entry added"
        fi
    fi
}

create_log_directory() {
    print_section "Setting up logging..."

    local log_dir="/var/log/thebot"

    print_info "Creating log directory: $log_dir"
    if [ "$DRY_RUN" = false ]; then
        mkdir -p "$log_dir"
        chown postgres:postgres "$log_dir"
        chmod 755 "$log_dir"
        print_success "Log directory created"
    fi
}

test_setup() {
    print_section "Testing setup..."

    local verify_script="$SCRIPT_DIR/verify-chat-creation.sh"

    print_info "Running verification script..."
    if [ "$DRY_RUN" = false ]; then
        if "$verify_script" --check-only; then
            print_success "Verification passed"
        else
            print_error "Verification failed"
            return 1
        fi
    fi
}

show_summary() {
    print_section "Setup Summary"

    echo ""
    echo "Configuration:"
    echo "  Project Root: $PROJECT_ROOT"
    echo "  Script Dir: $SCRIPT_DIR"
    echo "  Scheduler: $([ "$SCHEDULER_TYPE" = "timer" ] && echo "systemd timer" || echo "crontab")"
    echo "  Dry Run: $([ "$DRY_RUN" = true ] && echo "Yes" || echo "No")"

    echo ""
    if [ "$SCHEDULER_TYPE" = "timer" ]; then
        echo "Systemd Timer:"
        echo "  Service: /etc/systemd/system/thebot-chat-creation.service"
        echo "  Timer: /etc/systemd/system/thebot-chat-creation.timer"
        echo "  Schedule: Every 30 minutes (with 2 minute random delay)"
        echo ""
        echo "Useful commands:"
        echo "  systemctl status thebot-chat-creation.timer"
        echo "  systemctl list-timers thebot-chat-creation.timer"
        echo "  systemctl start thebot-chat-creation.service"
        echo "  journalctl -u thebot-chat-creation.service -f"
    else
        echo "Crontab:"
        echo "  Schedule: Every 30 minutes"
        echo "  Log File: /var/log/thebot/chat-creation-cron.log"
        echo ""
        echo "Useful commands:"
        echo "  crontab -l"
        echo "  tail -f /var/log/thebot/chat-creation-cron.log"
    fi

    echo ""
    echo "Documentation:"
    echo "  $SCRIPT_DIR/CHAT_CREATION_README.md"
}

main() {
    print_header "Chat Creation System Setup"

    parse_args "$@"

    echo "Configuration:"
    echo "  Scheduler Type: $SCHEDULER_TYPE"
    echo "  Auto Confirm: $([ "$AUTO_YES" = true ] && echo "Yes" || echo "No")"
    echo "  Dry Run: $([ "$DRY_RUN" = true ] && echo "Yes" || echo "No")"

    check_requirements
    check_scripts
    create_log_directory

    echo ""
    if [ "$SCHEDULER_TYPE" = "timer" ]; then
        if ! confirm "Install systemd timer for chat creation?"; then
            print_error "Installation cancelled"
            exit 1
        fi
        install_timer
    else
        if ! confirm "Install crontab entry for chat creation?"; then
            print_error "Installation cancelled"
            exit 1
        fi
        install_cron
    fi

    test_setup

    show_summary

    echo ""
    if [ "$DRY_RUN" = false ]; then
        print_success "Chat creation system setup complete!"
    else
        print_info "Dry-run completed. No changes were made."
    fi
}

main "$@"
