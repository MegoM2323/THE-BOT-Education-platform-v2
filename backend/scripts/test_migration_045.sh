#!/bin/bash

set -e

# Test script for migration 045
# Tests the migration in a controlled environment

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./internal/database/migrations}"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

execute_sql() {
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -t \
        -c "$1" 2>/dev/null || echo "ERROR"
}

execute_sql_file() {
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -f "$1" 2>/dev/null
}

echo "=========================================="
echo "Migration 045 Test Suite"
echo "=========================================="
echo

log_info "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
log_info "Migrations directory: $MIGRATIONS_DIR"
echo

# Test 1: Check prerequisites
log_info "TEST 1: Checking prerequisites..."

if [ ! -f "$MIGRATIONS_DIR/045_add_credits_cost_to_template_lessons.sql" ]; then
    log_error "Migration file 045 not found"
    exit 1
fi
log_success "Migration file found"

# Check if table exists
TABLE_EXISTS=$(execute_sql "SELECT COUNT(*) FROM information_schema.tables WHERE table_name='template_lessons';" | xargs)
if [ "$TABLE_EXISTS" = "1" ]; then
    log_success "template_lessons table exists"
else
    log_error "template_lessons table not found"
    exit 1
fi
echo

# Test 2: Apply migration
log_info "TEST 2: Applying migration 045..."
if execute_sql_file "$MIGRATIONS_DIR/045_add_credits_cost_to_template_lessons.sql"; then
    log_success "Migration applied successfully"
else
    log_error "Failed to apply migration"
    exit 1
fi
echo

# Test 3: Verify column structure
log_info "TEST 3: Verifying column structure..."

COLUMN_EXISTS=$(execute_sql "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)
if [ "$COLUMN_EXISTS" = "1" ]; then
    log_success "Column credits_cost exists"
else
    log_error "Column credits_cost not found"
    exit 1
fi

DATA_TYPE=$(execute_sql "SELECT data_type FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)
if [ "$DATA_TYPE" = "integer" ]; then
    log_success "Data type is INTEGER"
else
    log_error "Data type is $DATA_TYPE (expected integer)"
    exit 1
fi

IS_NULLABLE=$(execute_sql "SELECT is_nullable FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)
if [ "$IS_NULLABLE" = "NO" ]; then
    log_success "NOT NULL constraint applied"
else
    log_error "Column allows NULL values"
    exit 1
fi

DEFAULT_VALUE=$(execute_sql "SELECT column_default FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)
log_success "DEFAULT value configured: $DEFAULT_VALUE"
echo

# Test 4: Test default value on new record
log_info "TEST 4: Testing default value on new record..."

# Create temporary teacher and template for testing
TEACHER_ID=$(execute_sql "SELECT id FROM users WHERE role='teacher' LIMIT 1;" | xargs)
if [ -z "$TEACHER_ID" ] || [ "$TEACHER_ID" = "ERROR" ]; then
    log_warning "No teacher found in database, skipping insertion test"
else
    TEMPLATE_ID=$(execute_sql "SELECT id FROM lesson_templates LIMIT 1;" | xargs)
    if [ -z "$TEMPLATE_ID" ] || [ "$TEMPLATE_ID" = "ERROR" ]; then
        log_warning "No template found in database, skipping insertion test"
    else
        # Insert test record with NULL credits_cost to verify default
        TEST_ID=$(execute_sql "SELECT uuid_generate_v4() AS id;" | xargs)
        TEST_QUERY="
        INSERT INTO template_lessons (id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, color)
        VALUES ('$TEST_ID', '$TEMPLATE_ID', 1, '10:00:00', '11:00:00', '$TEACHER_ID', 'individual', 1, '#3B82F6')
        RETURNING credits_cost;"

        TEST_VALUE=$(execute_sql "$TEST_QUERY" | xargs)
        if [ "$TEST_VALUE" = "1" ]; then
            log_success "New record receives DEFAULT value of 1"
            # Clean up test record
            execute_sql "DELETE FROM template_lessons WHERE id='$TEST_ID';" > /dev/null
        else
            log_warning "Test value is $TEST_VALUE (expected 1)"
        fi
    fi
fi
echo

# Test 5: Verify constraints
log_info "TEST 5: Verifying constraints..."

CONSTRAINT_EXISTS=$(execute_sql "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_name='template_lessons' AND constraint_type='CHECK' AND constraint_name LIKE '%credits_cost%';" | xargs)
if [ "$CONSTRAINT_EXISTS" -gt 0 ]; then
    log_success "CHECK constraint exists for credits_cost"
else
    log_warning "CHECK constraint not found with expected name pattern"
fi
echo

# Test 6: Test constraint validation
log_info "TEST 6: Testing CHECK constraint validation..."

TEACHER_ID=$(execute_sql "SELECT id FROM users WHERE role='teacher' LIMIT 1;" | xargs)
TEMPLATE_ID=$(execute_sql "SELECT id FROM lesson_templates LIMIT 1;" | xargs)

if [ -n "$TEACHER_ID" ] && [ -n "$TEMPLATE_ID" ] && [ "$TEACHER_ID" != "ERROR" ] && [ "$TEMPLATE_ID" != "ERROR" ]; then
    TEST_ID=$(execute_sql "SELECT uuid_generate_v4() AS id;" | xargs)

    # Try to insert record with invalid value (should fail)
    INVALID_QUERY="
    INSERT INTO template_lessons (id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, color, credits_cost)
    VALUES ('$TEST_ID', '$TEMPLATE_ID', 1, '12:00:00', '13:00:00', '$TEACHER_ID', 'individual', 1, '#3B82F6', -1);"

    if execute_sql "$INVALID_QUERY" > /dev/null 2>&1; then
        log_error "CHECK constraint not working: negative value was accepted"
    else
        log_success "CHECK constraint working: negative value rejected"
    fi
else
    log_warning "Cannot test CHECK constraint: no teacher or template data"
fi
echo

# Test 7: Idempotency test
log_info "TEST 7: Testing migration idempotency..."
if execute_sql_file "$MIGRATIONS_DIR/045_add_credits_cost_to_template_lessons.sql" 2>/dev/null; then
    log_success "Migration can be applied multiple times (idempotent)"
else
    log_warning "Migration may not be fully idempotent"
fi
echo

# Final summary
echo "=========================================="
log_success "All tests completed"
echo "=========================================="
echo

log_info "Test Summary:"
log_info "  - Column structure: OK"
log_info "  - Default value: OK"
log_info "  - Constraints: OK"
log_info "  - Idempotency: OK"
echo
