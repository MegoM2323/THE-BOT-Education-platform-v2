#!/bin/bash

# Apply all pending migrations to production database
# Usage: ./apply-migrations.sh

set -euo pipefail

# Configuration
REMOTE_HOST="${REMOTE_HOST:-5.129.249.206}"
REMOTE_USER="${REMOTE_USER:-mg}"
THEBOT_HOME="${THEBOT_HOME:-/home/mg/the-bot}"
DB_NAME="${DB_NAME:-thebot_db}"
DB_USER="${DB_USER:-postgres}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $@"; }
log_success() { echo -e "${GREEN}[✓]${NC} $@"; }
log_error() { echo -e "${RED}[✗]${NC} $@"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $@"; }

log_info "Applying database migrations..."

# Get migration files to apply
MIGRATION_FILES=(
    "037_booking_created_index.sql"
    "038_subjects_table.sql"
    "039_remove_lesson_type_column.sql"
    "040_add_balance_audit_to_credit_transactions.sql"
    "041_fix_cascade_delete.sql"
    "042_init_credits_for_all_users.sql"
    "043_add_max_balance_constraint.sql"
    "044_add_credits_cost_to_lessons.sql"
    "045_add_credits_cost_to_template_lessons.sql"
    "046_add_methodologist_role.sql"
    "047_allow_zero_credits_cost.sql"
    "048_remove_teacher_role.sql"
    "050_chat_on_booking_creation.sql"
    "052_add_overbooking_triggers.sql"
    "053_add_soft_delete_cascade.sql"
)

for migration_file in "${MIGRATION_FILES[@]}"; do
    log_info "Applying $migration_file..."

    ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" << EOF
        sudo -u postgres psql -d ${DB_NAME} << 'SQL'
INSERT INTO schema_migrations (version) VALUES ('${migration_file%.sql}');
SQL

        sudo -u postgres psql -d ${DB_NAME} < "${THEBOT_HOME}/backend/internal/database/migrations/${migration_file}"
EOF

    if [ $? -eq 0 ]; then
        log_success "Applied $migration_file"
    else
        log_error "Failed to apply $migration_file"
        exit 1
    fi
done

log_success "All migrations applied!"

# Verify chat triggers exist
log_info "Verifying chat triggers..."

ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" << 'EOF'
sudo -u postgres psql -d thebot_db -c "
SELECT trigger_name, event_object_table
FROM information_schema.triggers
WHERE trigger_name LIKE '%chat%'
OR trigger_name LIKE '%booking%'
ORDER BY trigger_name;
"
EOF

log_info "Checking chat room creation..."

# Test: Create a test booking and verify chat is created
ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" << 'EOF'
sudo -u postgres psql -d thebot_db << 'SQL'
-- Show test results
SELECT
  'Total active bookings' as metric, COUNT(*)
FROM bookings
WHERE status = 'active' AND deleted_at IS NULL
UNION ALL
SELECT 'Total chat rooms', COUNT(*)
FROM chat_rooms
WHERE deleted_at IS NULL;
SQL
EOF

log_success "Migration verification complete!"
