#!/bin/bash

# Chat Creation Cron Job Script
# Purpose: Periodically run chat creation backfill function
# Intended for use with crontab or system scheduler
# Usage: Run this script via cron (e.g., every 5 minutes or every hour)

set -euo pipefail

# Configuration from environment or defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"

# Logging configuration
LOG_DIR="${LOG_DIR:-/var/log/thebot}"
LOG_FILE="${LOG_DIR}/chat-creation-cron.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# Ensure log directory exists
mkdir -p "$LOG_DIR"

# Helper functions
log_info() {
    echo "[${TIMESTAMP}] INFO: $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo "[${TIMESTAMP}] ERROR: $1" | tee -a "$LOG_FILE" >&2
}

log_success() {
    echo "[${TIMESTAMP}] SUCCESS: $1" | tee -a "$LOG_FILE"
}

# Main function
main() {
    log_info "Starting chat creation backfill..."

    # Set database password for psql
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi

    # Test connection
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
        log_error "Failed to connect to database at $DB_HOST:$DB_PORT"
        exit 1
    fi

    # Run backfill function
    local result
    result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT create_chats_for_completed_lessons();" 2>&1 || echo "ERROR")

    if [ "$result" = "ERROR" ]; then
        log_error "Backfill function failed"
        exit 1
    fi

    # Parse result (should be an integer)
    if [[ "$result" =~ ^[0-9]+$ ]]; then
        log_success "Created $result new chat rooms"
    else
        log_error "Unexpected result from backfill function: $result"
        exit 1
    fi

    # Cleanup: keep only last 30 days of logs
    find "$LOG_DIR" -name "chat-creation-cron.log*" -mtime +30 -delete 2>/dev/null || true

    exit 0
}

# Run main function
main
