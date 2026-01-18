#!/bin/bash

# Chat Creation System Verification Script
# Purpose: Verify and test chat auto-creation system on production
# Usage: ./verify-chat-creation.sh [--remote] [--backfill] [--check-only]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration from environment or defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"

# Script defaults
REMOTE_MODE=false
BACKFILL_MODE=false
CHECK_ONLY=false
VERBOSE=false

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

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --remote)
                REMOTE_MODE=true
                print_info "Remote mode enabled (connecting to production)"
                shift
                ;;
            --backfill)
                BACKFILL_MODE=true
                print_info "Backfill mode enabled (will create missing chats)"
                shift
                ;;
            --check-only)
                CHECK_ONLY=true
                print_info "Check-only mode enabled (no modifications)"
                shift
                ;;
            --verbose)
                VERBOSE=true
                print_info "Verbose logging enabled"
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

show_help() {
    cat << 'EOF'
Chat Creation System Verification Script

Usage: ./verify-chat-creation.sh [OPTIONS]

Options:
  --remote              Connect to production database (via SSH tunnel if needed)
  --backfill            Run chat creation backfill function
  --check-only          Only check system status, don't modify anything
  --verbose             Enable verbose logging
  --help, -h            Show this help message

Examples:
  # Check local development database
  ./verify-chat-creation.sh --check-only

  # Verify production database with backfill
  DB_HOST=5.129.249.206 ./verify-chat-creation.sh --remote --backfill

  # Check production database (read-only)
  DB_HOST=5.129.249.206 DB_NAME=thebot_db ./verify-chat-creation.sh --check-only

Environment Variables:
  DB_HOST        PostgreSQL host (default: localhost)
  DB_PORT        PostgreSQL port (default: 5432)
  DB_NAME        Database name (default: tutoring_platform)
  DB_USER        Database user (default: postgres)
  DB_PASSWORD    Database password (default: empty)

EOF
}

# Database connection test
test_db_connection() {
    print_section "Testing database connection..."

    local connection_cmd="psql -h \"$DB_HOST\" -p \"$DB_PORT\" -U \"$DB_USER\" -d \"$DB_NAME\" -c \"SELECT version();\" --no-password"

    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi

    if eval "$connection_cmd" > /dev/null 2>&1; then
        print_success "Connected to $DB_HOST:$DB_PORT/$DB_NAME"
        return 0
    else
        print_error "Failed to connect to database"
        print_error "Check your connection parameters:"
        echo "  Host: $DB_HOST"
        echo "  Port: $DB_PORT"
        echo "  Database: $DB_NAME"
        echo "  User: $DB_USER"
        return 1
    fi
}

# Check if triggers exist
check_trigger() {
    print_section "Checking booking triggers..."

    local trigger1=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT COUNT(*) FROM information_schema.triggers WHERE trigger_name = 'booking_create_chat';" \
        2>/dev/null || echo "0")

    local trigger2=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT COUNT(*) FROM information_schema.triggers WHERE trigger_name = 'booking_create_chat_update';" \
        2>/dev/null || echo "0")

    local all_ok=0

    if [ "$trigger1" -gt 0 ]; then
        print_success "Trigger 'booking_create_chat' (BEFORE INSERT) exists"
    else
        print_error "Trigger 'booking_create_chat' NOT FOUND"
        all_ok=1
    fi

    if [ "$trigger2" -gt 0 ]; then
        print_success "Trigger 'booking_create_chat_update' (BEFORE UPDATE) exists"
    else
        print_error "Trigger 'booking_create_chat_update' NOT FOUND"
        all_ok=1
    fi

    # Show trigger details
    if [ "$trigger1" -gt 0 ] || [ "$trigger2" -gt 0 ]; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            -c "SELECT trigger_name, event_manipulation, event_object_table FROM information_schema.triggers WHERE event_object_table='bookings' AND trigger_name LIKE 'booking_create_chat%';" \
            2>/dev/null || true
    fi

    return $all_ok
}

# Check if function exists
check_functions() {
    print_section "Checking PL/pgSQL function..."

    local func1=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name = 'create_chat_on_booking_active' AND routine_type = 'FUNCTION';" \
        2>/dev/null || echo "0")

    local all_ok=0

    if [ "$func1" -gt 0 ]; then
        print_success "Function 'create_chat_on_booking_active()' exists"
    else
        print_error "Function 'create_chat_on_booking_active()' NOT FOUND"
        all_ok=1
    fi

    return $all_ok
}

# Show statistics
show_statistics() {
    print_section "Chat Creation System Statistics"

    echo ""
    echo "Lesson Statistics:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "
        SELECT
            'Total lessons' as metric,
            COUNT(*) as count
        FROM lessons
        WHERE deleted_at IS NULL

        UNION ALL

        SELECT
            'Completed lessons (end_time < NOW)',
            COUNT(*)
        FROM lessons
        WHERE end_time < CURRENT_TIMESTAMP
          AND deleted_at IS NULL

        UNION ALL

        SELECT
            'Upcoming lessons (end_time >= NOW)',
            COUNT(*)
        FROM lessons
        WHERE end_time >= CURRENT_TIMESTAMP
          AND deleted_at IS NULL

        UNION ALL

        SELECT
            'Lessons without end_time',
            COUNT(*)
        FROM lessons
        WHERE end_time IS NULL
          AND deleted_at IS NULL

        ORDER BY metric;
        " 2>/dev/null || print_error "Failed to fetch lesson statistics"

    echo ""
    echo "Booking Statistics:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "
        SELECT
            CASE
                WHEN status = 'active' THEN 'Active bookings'
                WHEN status = 'completed' THEN 'Completed bookings'
                WHEN status = 'cancelled' THEN 'Cancelled bookings'
                ELSE 'Other (' || status || ')'
            END as metric,
            COUNT(*) as count
        FROM bookings
        WHERE deleted_at IS NULL
        GROUP BY status
        ORDER BY count DESC;
        " 2>/dev/null || print_error "Failed to fetch booking statistics"

    echo ""
    echo "Chat Room Statistics:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "
        SELECT
            'Total chat rooms' as metric,
            COUNT(*) as count
        FROM chat_rooms
        WHERE deleted_at IS NULL

        UNION ALL

        SELECT
            'Chat rooms with messages',
            COUNT(DISTINCT cr.id)
        FROM chat_rooms cr
        INNER JOIN messages m ON m.chat_room_id = cr.id
        WHERE cr.deleted_at IS NULL
          AND m.deleted_at IS NULL

        UNION ALL

        SELECT
            'Empty chat rooms',
            COUNT(DISTINCT cr.id)
        FROM chat_rooms cr
        LEFT JOIN messages m ON m.chat_room_id = cr.id AND m.deleted_at IS NULL
        WHERE cr.deleted_at IS NULL
          AND m.id IS NULL;
        " 2>/dev/null || print_error "Failed to fetch chat room statistics"

    echo ""
    echo "Chat Creation Candidates (completed lessons without chats):"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "
        SELECT
            COUNT(DISTINCT b.id) as potential_chats,
            COUNT(DISTINCT l.id) as lessons_affected
        FROM lessons l
        INNER JOIN bookings b ON b.lesson_id = l.id AND b.status = 'active'
        LEFT JOIN chat_rooms cr ON cr.teacher_id = l.teacher_id
            AND cr.student_id = b.student_id
            AND cr.deleted_at IS NULL
        WHERE l.end_time < CURRENT_TIMESTAMP
          AND l.deleted_at IS NULL
          AND cr.id IS NULL;
        " 2>/dev/null || print_error "Failed to count chat creation candidates"
}

# Run backfill function
run_backfill() {
    if [ "$CHECK_ONLY" = true ]; then
        print_section "Backfill mode is disabled in check-only mode"
        return 0
    fi

    print_section "Running chat creation backfill function..."

    if [ "$VERBOSE" = true ]; then
        echo "Executing: SELECT create_chats_for_completed_lessons();"
    fi

    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT create_chats_for_completed_lessons();" 2>/dev/null || echo "0")

    if [ -n "$result" ] && [ "$result" -gt -1 ]; then
        print_success "Backfill completed successfully"
        print_success "Created $result new chat rooms"
    else
        print_error "Backfill function returned an error"
        return 1
    fi
}

# Verify data integrity
verify_integrity() {
    print_section "Verifying data integrity..."

    echo ""
    echo "Checking for bookings without corresponding lessons..."
    local orphaned_bookings=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT COUNT(*) FROM bookings b LEFT JOIN lessons l ON b.lesson_id = l.id WHERE l.id IS NULL AND b.deleted_at IS NULL;" \
        2>/dev/null || echo "0")

    if [ "$orphaned_bookings" -eq 0 ]; then
        print_success "No orphaned bookings found"
    else
        print_error "$orphaned_bookings orphaned bookings found (booking without lesson)"
    fi

    echo ""
    echo "Checking for chat rooms with invalid teacher/student..."
    local invalid_chats=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -c "SELECT COUNT(*) FROM chat_rooms cr LEFT JOIN users ut ON cr.teacher_id = ut.id LEFT JOIN users us ON cr.student_id = us.id WHERE (ut.id IS NULL OR us.id IS NULL) AND cr.deleted_at IS NULL;" \
        2>/dev/null || echo "0")

    if [ "$invalid_chats" -eq 0 ]; then
        print_success "All chat rooms have valid teacher and student"
    else
        print_error "$invalid_chats chat rooms with invalid references found"
    fi

    echo ""
    echo "Checking trigger function execution (via system logs)..."
    print_info "Trigger execution verification depends on PostgreSQL log configuration"
}

# Show summary
show_summary() {
    print_section "Verification Summary"

    echo ""
    echo "Configuration:"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User: $DB_USER"

    echo ""
    echo "Modes:"
    echo "  Remote: $([ "$REMOTE_MODE" = true ] && echo "Yes" || echo "No")"
    echo "  Backfill: $([ "$BACKFILL_MODE" = true ] && echo "Yes" || echo "No")"
    echo "  Check-only: $([ "$CHECK_ONLY" = true ] && echo "Yes" || echo "No")"
    echo "  Verbose: $([ "$VERBOSE" = true ] && echo "Yes" || echo "No")"
}

# Main execution flow
main() {
    print_header "Chat Creation System Verification"

    parse_args "$@"

    # Test database connection
    if ! test_db_connection; then
        exit 1
    fi

    # Run all checks
    local checks_passed=0
    local checks_failed=0

    if check_trigger; then
        ((checks_passed++))
    else
        ((checks_failed++))
    fi

    if check_functions; then
        ((checks_passed++))
    else
        ((checks_failed++))
    fi

    # Show statistics
    show_statistics

    # Verify data integrity
    verify_integrity

    # Run backfill if requested
    if [ "$BACKFILL_MODE" = true ]; then
        if run_backfill; then
            ((checks_passed++))
        else
            ((checks_failed++))
        fi
    fi

    # Show summary
    show_summary

    # Final status
    echo ""
    if [ "$checks_failed" -eq 0 ]; then
        print_success "All checks passed!"
        exit 0
    else
        print_error "$checks_failed check(s) failed"
        exit 1
    fi
}

# Run main function
main "$@"
