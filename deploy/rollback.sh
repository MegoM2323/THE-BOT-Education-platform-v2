#!/bin/bash

################################################################################
# Application Rollback Script
#
# This script rolls back the Tutoring Platform to a previous version
# by restoring from the latest backup
#
# Usage: sudo ./rollback.sh [backup_file]
################################################################################

set -e

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

# Error handler
error_handler() {
    log_error "Rollback failed on line $1"
    log_error "System may be in an inconsistent state"
    log_error "Please check manually and contact support if needed"
    exit 1
}

trap 'error_handler $LINENO' ERR

################################################################################
# Pre-flight checks
################################################################################

log_warning "========== ROLLBACK INITIATED =========="
log_warning "This will restore the database from a backup"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (use sudo)"
    exit 1
fi

# Configuration
BACKUP_DIR="/var/backups/tutoring-platform"
APP_DIR="/opt/tutoring-platform"
LOG_FILE="/var/log/tutoring-platform/rollback.log"

# Log to file
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

log_info "Rollback started at $(date)"

################################################################################
# Select backup file
################################################################################

BACKUP_FILE="$1"

if [ -z "$BACKUP_FILE" ]; then
    log_info "Available backups:"
    echo ""

    # List available backups
    BACKUPS=($(find "$BACKUP_DIR" -name "tutoring_platform_*.sql.gz" -type f | sort -r))

    if [ ${#BACKUPS[@]} -eq 0 ]; then
        log_error "No backups found in $BACKUP_DIR"
        exit 1
    fi

    # Display backups with numbers
    for i in "${!BACKUPS[@]}"; do
        BACKUP="${BACKUPS[$i]}"
        SIZE=$(du -h "$BACKUP" | cut -f1)
        DATE=$(stat -c %y "$BACKUP" | cut -d' ' -f1,2 | cut -d'.' -f1)
        echo "  [$i] $(basename "$BACKUP") - $SIZE - $DATE"
    done

    echo ""
    read -p "Select backup number (0-$((${#BACKUPS[@]}-1))) or 'q' to quit: " SELECTION

    if [ "$SELECTION" = "q" ]; then
        log_info "Rollback cancelled"
        exit 0
    fi

    if ! [[ "$SELECTION" =~ ^[0-9]+$ ]] || [ "$SELECTION" -ge "${#BACKUPS[@]}" ]; then
        log_error "Invalid selection"
        exit 1
    fi

    BACKUP_FILE="${BACKUPS[$SELECTION]}"
fi

# Verify backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    log_error "Backup file not found: $BACKUP_FILE"
    exit 1
fi

log_info "Selected backup: $(basename "$BACKUP_FILE")"
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log_info "Backup size: $BACKUP_SIZE"

################################################################################
# Confirmation
################################################################################

log_warning "This will:"
log_warning "  1. Stop the backend service"
log_warning "  2. Restore the database from backup"
log_warning "  3. All current data will be lost!"
echo ""
read -p "Are you sure you want to continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    log_info "Rollback cancelled"
    exit 0
fi

################################################################################
# Stop services
################################################################################

log_info "Stopping backend service..."
systemctl stop tutoring-platform
log_success "Backend service stopped"

################################################################################
# Restore database
################################################################################

log_info "Restoring database from backup..."

# Source database credentials from .env
if [ -f "$APP_DIR/backend/.env" ]; then
    export $(grep -v '^#' "$APP_DIR/backend/.env" | xargs)
else
    log_error ".env file not found"
    exit 1
fi

DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-tutoring_user}"
DB_HOST="${DB_HOST:-localhost}"

# Drop existing database (with confirmation)
log_warning "Dropping existing database..."

sudo -u postgres psql <<EOF
-- Terminate existing connections
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();

-- Drop database
DROP DATABASE IF EXISTS $DB_NAME;

-- Recreate database
CREATE DATABASE $DB_NAME OWNER $DB_USER;
EOF

log_success "Database dropped and recreated"

# Restore from backup
log_info "Restoring data from backup file..."

# Decompress and restore
gunzip -c "$BACKUP_FILE" | sudo -u postgres psql -d "$DB_NAME"

log_success "Database restored from backup"

# Verify restoration
log_info "Verifying database restoration..."

TABLE_COUNT=$(sudo -u postgres psql -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';")
log_info "Restored $TABLE_COUNT tables"

################################################################################
# Restore application code (if git is used)
################################################################################

cd "$APP_DIR"

if [ -d ".git" ]; then
    log_info "Git repository detected"
    read -p "Do you want to checkout a specific commit/tag? (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Show recent commits
        echo ""
        log_info "Recent commits:"
        git log --oneline -10
        echo ""

        read -p "Enter commit hash or tag: " GIT_REF

        if [ -n "$GIT_REF" ]; then
            log_info "Checking out $GIT_REF..."
            git checkout "$GIT_REF"

            # Rebuild backend
            log_info "Rebuilding backend..."
            cd "$APP_DIR/backend"
            export PATH=$PATH:/usr/local/go/bin
            go mod download
            go build -o bin/server cmd/server/main.go
            chmod +x bin/server
            log_success "Backend rebuilt"

            # Rebuild frontend
            log_info "Rebuilding frontend..."
            cd "$APP_DIR/frontend"
            npm install --production
            npm run build
            log_success "Frontend rebuilt"
        fi
    fi
else
    log_warning "Not a git repository, skipping code rollback"
    log_info "If you need to restore code, do it manually"
fi

################################################################################
# Start services
################################################################################

log_info "Starting backend service..."
systemctl start tutoring-platform

# Wait for service to start
sleep 3

# Check if service started successfully
if systemctl is-active --quiet tutoring-platform; then
    log_success "Backend service started"
else
    log_error "Failed to start backend service"
    systemctl status tutoring-platform
    exit 1
fi

################################################################################
# Health check
################################################################################

log_info "Running health check..."

sleep 5

if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log_success "Health check passed"
else
    log_warning "Health check failed"
    log_warning "Service may still be starting up"
fi

################################################################################
# Reload Nginx
################################################################################

log_info "Reloading Nginx..."
systemctl reload nginx
log_success "Nginx reloaded"

################################################################################
# Summary
################################################################################

echo ""
echo "================================================================================"
echo -e "${GREEN}Rollback completed!${NC}"
echo "================================================================================"
echo ""
echo "Restored from backup: $(basename "$BACKUP_FILE")"
echo ""
echo "Services status:"
systemctl status tutoring-platform --no-pager -l
echo ""
echo "Database verification:"
echo "  Tables restored: $TABLE_COUNT"
echo ""
echo "Next steps:"
echo "  1. Verify the application is working correctly"
echo "  2. Check logs: tail -f /var/log/tutoring-platform/access.log"
echo "  3. Test critical functionality"
echo ""
echo "================================================================================"
echo ""

log_info "Rollback finished at $(date)"
log_success "Rollback process completed!"

exit 0
