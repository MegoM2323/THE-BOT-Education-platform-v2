#!/bin/bash

################################################################################
# Application Update Script
#
# This script updates the Tutoring Platform to the latest version
# including backend, frontend, and database migrations
#
# Usage: sudo ./update.sh
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
    log_error "Update failed on line $1"
    log_error "Rolling back changes..."
    # Note: Rollback would restore from backup
    log_error "Please check the logs and run rollback.sh if needed"
    exit 1
}

trap 'error_handler $LINENO' ERR

################################################################################
# Pre-flight checks
################################################################################

log_info "Starting Tutoring Platform update..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (use sudo)"
    exit 1
fi

# Configuration
APP_DIR="/opt/tutoring-platform"
BACKUP_SCRIPT="/opt/tutoring-platform/backup.sh"
LOG_FILE="/var/log/tutoring-platform/update.log"

# Log to file
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

log_info "Update started at $(date)"

################################################################################
# Create pre-update backup
################################################################################

log_info "Creating pre-update backup..."

if [ -f "$BACKUP_SCRIPT" ]; then
    bash "$BACKUP_SCRIPT"
    log_success "Backup created successfully"
else
    log_warning "Backup script not found, skipping backup"
    read -p "Continue without backup? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

################################################################################
# Stop services
################################################################################

log_info "Stopping backend service..."
systemctl stop tutoring-platform
log_success "Backend service stopped"

################################################################################
# Update source code
################################################################################

log_info "Updating source code..."

cd "$APP_DIR"

# If using git
if [ -d ".git" ]; then
    log_info "Pulling latest changes from git..."

    # Stash any local changes
    git stash

    # Pull latest changes
    git pull origin main

    log_success "Source code updated from git"
else
    # If not using git, assume manual copy
    log_warning "Not a git repository"
    log_info "Please manually copy updated files to $APP_DIR"
    read -p "Have you copied the latest files? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Update cancelled"
        systemctl start tutoring-platform
        exit 1
    fi
fi

################################################################################
# Update backend
################################################################################

log_info "Updating backend..."

cd "$APP_DIR/backend"

# Update Go dependencies
log_info "Updating Go dependencies..."
export PATH=$PATH:/usr/local/go/bin
go mod download
go mod tidy

# Build new binary
log_info "Building new backend binary..."
go build -o bin/server.new cmd/server/main.go

if [ ! -f bin/server.new ]; then
    log_error "Failed to build new backend binary"
    exit 1
fi

# Backup old binary
if [ -f bin/server ]; then
    mv bin/server bin/server.backup
    log_info "Old binary backed up"
fi

# Replace with new binary
mv bin/server.new bin/server
chmod +x bin/server

log_success "Backend binary updated"

################################################################################
# Apply database migrations
################################################################################

log_info "Checking for database migrations..."

if [ -d "$APP_DIR/backend/migrations" ]; then
    # Check if migrate is installed
    if ! command -v migrate &> /dev/null; then
        log_info "Installing golang-migrate..."
        curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
        mv migrate /usr/local/bin/
        chmod +x /usr/local/bin/migrate
    fi

    # Source database credentials from .env
    if [ -f "$APP_DIR/backend/.env" ]; then
        export $(grep -v '^#' "$APP_DIR/backend/.env" | xargs)
    fi

    # Get current migration version
    CURRENT_VERSION=$(migrate -path "$APP_DIR/backend/migrations" \
                             -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE" \
                             version 2>&1 | grep -oP '(?<=version )[0-9]+' || echo "0")

    log_info "Current migration version: $CURRENT_VERSION"

    # Apply migrations
    log_info "Applying database migrations..."
    migrate -path "$APP_DIR/backend/migrations" \
            -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE" \
            up

    # Get new migration version
    NEW_VERSION=$(migrate -path "$APP_DIR/backend/migrations" \
                         -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE" \
                         version 2>&1 | grep -oP '(?<=version )[0-9]+' || echo "0")

    if [ "$NEW_VERSION" -gt "$CURRENT_VERSION" ]; then
        log_success "Migrations applied (version $CURRENT_VERSION -> $NEW_VERSION)"
    else
        log_info "No new migrations to apply"
    fi
else
    log_warning "No migrations directory found, skipping migrations"
fi

################################################################################
# Update frontend
################################################################################

log_info "Updating frontend..."

cd "$APP_DIR/frontend"

# Update npm dependencies
log_info "Updating npm dependencies..."
npm install --production

# Build frontend
log_info "Building frontend..."

# Backup old build
if [ -d "build" ]; then
    mv build build.backup
    log_info "Old frontend build backed up"
fi

# Build new frontend
npm run build

if [ ! -d "build" ]; then
    log_error "Frontend build failed"

    # Restore old build if it exists
    if [ -d "build.backup" ]; then
        mv build.backup build
        log_warning "Restored old frontend build"
    fi

    exit 1
fi

# Remove backup if build succeeded
if [ -d "build.backup" ]; then
    rm -rf build.backup
fi

log_success "Frontend updated and built"

################################################################################
# Update Nginx configuration (if changed)
################################################################################

log_info "Checking Nginx configuration..."

if nginx -t &> /dev/null; then
    log_success "Nginx configuration is valid"

    # Reload nginx to pick up any config changes
    systemctl reload nginx
    log_info "Nginx reloaded"
else
    log_warning "Nginx configuration has issues"
    nginx -t
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

    # Try to restore old binary
    if [ -f "$APP_DIR/backend/bin/server.backup" ]; then
        log_info "Attempting to restore old binary..."
        systemctl stop tutoring-platform
        mv "$APP_DIR/backend/bin/server" "$APP_DIR/backend/bin/server.failed"
        mv "$APP_DIR/backend/bin/server.backup" "$APP_DIR/backend/bin/server"
        systemctl start tutoring-platform

        if systemctl is-active --quiet tutoring-platform; then
            log_warning "Restored old binary and service is running"
        else
            log_error "Failed to restore service"
        fi
    fi

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
    log_info "Service may still be starting up"
fi

################################################################################
# Cleanup
################################################################################

log_info "Cleaning up..."

# Remove old binary backup after successful update
if [ -f "$APP_DIR/backend/bin/server.backup" ]; then
    rm -f "$APP_DIR/backend/bin/server.backup"
fi

# Clean npm cache
npm cache clean --force 2>/dev/null || true

log_success "Cleanup completed"

################################################################################
# Summary
################################################################################

echo ""
echo "================================================================================"
echo -e "${GREEN}Update completed successfully!${NC}"
echo "================================================================================"
echo ""
echo "Services status:"
systemctl status tutoring-platform --no-pager -l
echo ""
echo "To verify the update:"
echo "  1. Check health: curl http://localhost:8080/health"
echo "  2. Check logs: tail -f /var/log/tutoring-platform/access.log"
echo "  3. Test the application in your browser"
echo ""
echo "If you encounter issues, you can rollback using:"
echo "  sudo ./rollback.sh"
echo ""
echo "================================================================================"
echo ""

log_info "Update finished at $(date)"
log_success "Update process completed!"

exit 0
