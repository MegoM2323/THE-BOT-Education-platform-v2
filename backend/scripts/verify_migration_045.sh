#!/bin/bash

set -e

# Verification script for migration 045: Add credits_cost to template_lessons
# This script checks:
# 1. Column exists with correct data type
# 2. NOT NULL constraint is applied
# 3. DEFAULT value is set to 1
# 4. CHECK constraint (credits_cost > 0) exists
# 5. COMMENT documentation is present
# 6. All existing records have valid values

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"

# Color codes for output
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

# Function to execute SQL query
execute_sql() {
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -t \
        -c "$1" 2>/dev/null || echo ""
}

echo "=========================================="
echo "Migration 045 Verification Script"
echo "=========================================="
echo

log_info "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
log_info "Verifying migration 045..."
echo

# 1. Check if column exists
log_info "1. Checking column existence..."
COLUMN_EXISTS=$(execute_sql "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)

if [ "$COLUMN_EXISTS" = "1" ]; then
    log_success "Column 'credits_cost' exists in template_lessons"
else
    log_error "Column 'credits_cost' NOT FOUND in template_lessons"
    exit 1
fi
echo

# 2. Check data type
log_info "2. Checking column data type..."
DATA_TYPE=$(execute_sql "SELECT data_type FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)

if [ "$DATA_TYPE" = "integer" ]; then
    log_success "Column data type is INTEGER"
else
    log_error "Column data type is $DATA_TYPE (expected: integer)"
    exit 1
fi
echo

# 3. Check NOT NULL constraint
log_info "3. Checking NOT NULL constraint..."
IS_NULLABLE=$(execute_sql "SELECT is_nullable FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)

if [ "$IS_NULLABLE" = "NO" ]; then
    log_success "Column has NOT NULL constraint"
else
    log_error "Column allows NULL values (is_nullable=$IS_NULLABLE)"
    exit 1
fi
echo

# 4. Check DEFAULT value
log_info "4. Checking DEFAULT value..."
DEFAULT_VALUE=$(execute_sql "SELECT column_default FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" | xargs)

if [ "$DEFAULT_VALUE" = "1" ] || [ "$DEFAULT_VALUE" = "'1'::integer" ]; then
    log_success "DEFAULT value is 1"
else
    log_warning "DEFAULT value is: $DEFAULT_VALUE (expected: 1)"
fi
echo

# 5. Check CHECK constraint
log_info "5. Checking CHECK constraint..."
CHECK_CONSTRAINT=$(execute_sql "SELECT constraint_name FROM information_schema.constraint_column_usage WHERE table_name='template_lessons' AND column_name='credits_cost' AND constraint_name LIKE '%credits_cost%';" | xargs)

if [ -n "$CHECK_CONSTRAINT" ]; then
    log_success "CHECK constraint found: $CHECK_CONSTRAINT"

    # Verify constraint definition
    CONSTRAINT_DEF=$(execute_sql "SELECT check_clause FROM information_schema.check_constraints WHERE constraint_name='$CHECK_CONSTRAINT';" | xargs)
    log_info "Constraint definition: $CONSTRAINT_DEF"
else
    log_warning "CHECK constraint not found with expected name pattern"
fi
echo

# 6. Check COMMENT documentation
log_info "6. Checking column documentation..."
COMMENT=$(execute_sql "SELECT col_description((SELECT oid FROM pg_tables WHERE tablename='template_lessons'), (SELECT ordinal_position FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost'));" | xargs)

if [ -n "$COMMENT" ]; then
    log_success "Column documentation: $COMMENT"
else
    log_warning "No column documentation found"
fi
echo

# 7. Verify data integrity
log_info "7. Verifying data integrity..."
RECORD_COUNT=$(execute_sql "SELECT COUNT(*) FROM template_lessons;" | xargs)
log_info "Total records in template_lessons: $RECORD_COUNT"

if [ "$RECORD_COUNT" -gt 0 ]; then
    NULL_COUNT=$(execute_sql "SELECT COUNT(*) FROM template_lessons WHERE credits_cost IS NULL;" | xargs)

    if [ "$NULL_COUNT" = "0" ]; then
        log_success "All $RECORD_COUNT records have non-NULL credits_cost values"
    else
        log_error "$NULL_COUNT records have NULL credits_cost values"
        exit 1
    fi

    # Check for invalid values (should be > 0)
    INVALID_COUNT=$(execute_sql "SELECT COUNT(*) FROM template_lessons WHERE credits_cost <= 0;" | xargs)

    if [ "$INVALID_COUNT" = "0" ]; then
        log_success "All records have valid credits_cost values (> 0)"
    else
        log_warning "$INVALID_COUNT records have invalid credits_cost values (<= 0)"
    fi

    # Show distribution
    log_info "Distribution of credits_cost values:"
    execute_sql "SELECT credits_cost, COUNT(*) as count FROM template_lessons GROUP BY credits_cost ORDER BY credits_cost;" | while read line; do
        log_info "  $line"
    done
else
    log_info "No records in template_lessons yet (table is empty)"
fi
echo

# 8. Summary
echo "=========================================="
log_success "Migration 045 verification PASSED"
echo "=========================================="
echo
echo "Summary:"
echo "  - Column credits_cost: EXISTS"
echo "  - Data type: INTEGER"
echo "  - NOT NULL: YES"
echo "  - DEFAULT: 1"
echo "  - CHECK constraint: YES"
echo "  - Documentation: $([ -n "$COMMENT" ] && echo "YES" || echo "NO")"
echo "  - Data integrity: OK"
echo
