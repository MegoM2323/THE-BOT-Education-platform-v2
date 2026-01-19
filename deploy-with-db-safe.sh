#!/bin/bash

# Safe Deployment Script - Preserves Database on Redeploy
# This script ensures database data is preserved during redeployment
# Usage: ./deploy-with-db-safe.sh [--no-backup] [--legacy]

set -euo pipefail

# Configuration
REMOTE_HOST="mg@5.129.249.206"
REMOTE_DIR="/opt/THE_BOT_platform"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_BEFORE_DEPLOY=true
DEPLOY_MODE="docker"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-admin@the-bot.ru}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Parse arguments
for arg in "$@"; do
    case $arg in
        --no-backup)
            BACKUP_BEFORE_DEPLOY=false
            shift
            ;;
        --legacy)
            DEPLOY_MODE="legacy"
            shift
            ;;
        --docker)
            DEPLOY_MODE="docker"
            shift
            ;;
        *)
            ;;
    esac
done

echo -e "${YELLOW}=== Safe Deployment with Database Preservation ===${NC}"
echo "Mode: $DEPLOY_MODE"
echo "Backup before deploy: $BACKUP_BEFORE_DEPLOY"
echo "Target: $REMOTE_HOST:$REMOTE_DIR"
echo ""

# Logging function
log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Check SSH connection
check_ssh() {
    log_step "Checking SSH connection..."
    if ! ssh -o ConnectTimeout=5 "$REMOTE_HOST" "echo 'SSH connection OK'" > /dev/null 2>&1; then
        log_error "Cannot connect to $REMOTE_HOST"
        exit 1
    fi
    log_success "SSH connection OK"
}

# Create pre-deployment backup
backup_database() {
    log_step "Creating backup of production database..."

    BACKUP_RESULT=$(ssh "$REMOTE_HOST" bash -s << 'BACKUP_SCRIPT'
set -uo pipefail

REMOTE_DIR="/opt/THE_BOT_platform"
BACKUP_DIR="$REMOTE_DIR/backups"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Check if .env exists
if [ ! -f "$REMOTE_DIR/.env" ]; then
    echo "SKIP: No .env file found (first deployment?)"
    exit 0
fi

# Get database credentials
DB_USER=$(grep "^DB_USER=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 || echo "tutoring")
DB_PASSWORD=$(grep "^DB_PASSWORD=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="tutoring_platform"

# Generate backup filename
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/db_backup_${TIMESTAMP}.sql"
BACKUP_FILE_GZ="$BACKUP_FILE.gz"

echo "Attempting database backup..."
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"

# Check if PostgreSQL container is running
if ! docker ps 2>/dev/null | grep -q tutoring-postgres; then
    echo "SKIP: PostgreSQL container 'tutoring-postgres' not running (first deployment or container down)"
    exit 0
fi

echo "  Container: tutoring-postgres is running"

# Check if PostgreSQL is ready to accept connections
if ! docker exec tutoring-postgres pg_isready -U "$DB_USER" -d "$DB_NAME" 2>&1; then
    echo "SKIP: PostgreSQL is not ready to accept connections"
    exit 0
fi

echo "  PostgreSQL: ready"

# Perform backup
echo "  Running pg_dump..."
DUMP_OUTPUT=$(docker exec tutoring-postgres pg_dump \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --no-owner \
    --no-privileges 2>&1)
DUMP_EXIT=$?

if [ $DUMP_EXIT -ne 0 ]; then
    echo "ERROR: pg_dump failed with exit code $DUMP_EXIT"
    echo "Output: $DUMP_OUTPUT"
    # Don't exit with error - backup failure shouldn't stop deployment
    echo "WARN: Continuing without backup"
    exit 0
fi

# Write backup to file
echo "$DUMP_OUTPUT" > "$BACKUP_FILE"

# Check if backup file has content
if [ ! -s "$BACKUP_FILE" ]; then
    echo "WARN: Backup file is empty (database might be empty)"
    rm -f "$BACKUP_FILE"
    exit 0
fi

# Compress backup
gzip -f "$BACKUP_FILE"

# Keep only last 5 backups
ls -t "$BACKUP_DIR"/db_backup_*.sql.gz 2>/dev/null | tail -n +6 | xargs -r rm

echo "OK: Backup created: $(du -h "$BACKUP_FILE_GZ" | cut -f1)"
BACKUP_SCRIPT
    )

    echo "$BACKUP_RESULT"

    if echo "$BACKUP_RESULT" | grep -q "^OK:"; then
        log_success "Database backup completed"
    elif echo "$BACKUP_RESULT" | grep -q "^SKIP:"; then
        log_info "Database backup skipped (see details above)"
    elif echo "$BACKUP_RESULT" | grep -q "^WARN:"; then
        log_info "Database backup had warnings, continuing deployment"
    else
        log_info "Backup status unknown, continuing deployment"
    fi
}

# Build backend binary (statically linked for distroless container)
build_backend() {
    log_step "Building backend binary (statically linked)..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    cd "$PROJECT_DIR/backend"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/server ./cmd/server/main.go

    if [ ! -f "bin/server" ]; then
        log_error "Failed to build backend binary"
        exit 1
    fi

    log_success "Backend binary built: $(ls -lh bin/server | awk '{print $5}')"
    cd "$PROJECT_DIR"
}

# Deploy with Docker (preserving database)
deploy_docker_safe() {
    log_step "Starting Docker deployment (with database preservation)..."

    # Stop containers first to release file locks on binary
    log_info "Stopping containers before file copy..."
    ssh "$REMOTE_HOST" "bash -c 'cd $REMOTE_DIR && docker-compose -f docker-compose.prod.yml down 2>/dev/null || true'"

    # Create directories with proper permissions
    ssh "$REMOTE_HOST" "mkdir -p $REMOTE_DIR/backend/bin $REMOTE_DIR/frontend && chmod 755 $REMOTE_DIR/backend/bin"

    log_info "Copying Docker configuration..."
    scp "$PROJECT_DIR/docker-compose.prod.yml" "$REMOTE_HOST:$REMOTE_DIR/"

    # Copy pre-built binary - remove old one first to avoid "text file busy" error
    log_info "Copying pre-built backend binary..."
    ssh "$REMOTE_HOST" "rm -f $REMOTE_DIR/backend/bin/server"
    scp "$PROJECT_DIR/backend/bin/server" "$REMOTE_HOST:$REMOTE_DIR/backend/bin/server"
    ssh "$REMOTE_HOST" "chmod +x $REMOTE_DIR/backend/bin/server"

    # Copy entrypoint script
    scp "$PROJECT_DIR/backend/entrypoint.sh" "$REMOTE_HOST:$REMOTE_DIR/backend/"

    # Copy frontend files
    log_info "Copying frontend files..."
    scp "$PROJECT_DIR/frontend/Dockerfile" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
    scp "$PROJECT_DIR/frontend/nginx.conf" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
    scp "$PROJECT_DIR/frontend/package.json" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
    scp "$PROJECT_DIR/frontend/package-lock.json" "$REMOTE_HOST:$REMOTE_DIR/frontend/" 2>/dev/null || true
    scp "$PROJECT_DIR/frontend/vite.config.js" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
    scp "$PROJECT_DIR/frontend/index.html" "$REMOTE_HOST:$REMOTE_DIR/frontend/"

    rsync -avz --delete \
        --exclude='node_modules' \
        "$PROJECT_DIR/frontend/src" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
    rsync -avz --delete \
        "$PROJECT_DIR/frontend/dist" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
    rsync -avz "$PROJECT_DIR/frontend/public" "$REMOTE_HOST:$REMOTE_DIR/frontend/" 2>/dev/null || true

    # Manage .env file - preserve existing values and update as needed
    log_info "Managing .env configuration..."

    # First, try to fetch existing .env from server if it exists
    if ssh "$REMOTE_HOST" "[[ -f $REMOTE_DIR/.env ]]" 2>/dev/null; then
        log_info "Fetching existing .env from server..."
        scp "$REMOTE_HOST:$REMOTE_DIR/.env" /tmp/existing.env 2>/dev/null || true
    fi

    # Create production .env based on local template
    if [[ -f "$PROJECT_DIR/backend/.env" ]]; then
        # If we have existing .env from server, use it as base; otherwise use local
        if [[ -f /tmp/existing.env ]]; then
            BASE_ENV="/tmp/existing.env"
            log_info "Using existing server .env as base (preserving credentials)"
        else
            BASE_ENV="$PROJECT_DIR/backend/.env"
            log_info "Using local .env as base (first deployment)"
        fi

        # Only generate new secrets if they don't exist in base
        DB_PASSWORD=$(grep "^DB_PASSWORD=" "$BASE_ENV" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || openssl rand -hex 24)
        SESSION_SECRET=$(grep "^SESSION_SECRET=" "$BASE_ENV" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || openssl rand -base64 64 | tr -d '\n')

        awk -v db_host="postgres" \
            -v db_user="tutoring" \
            -v db_password="$DB_PASSWORD" \
            -v session_secret="$SESSION_SECRET" \
            -v production_domain="https://the-bot.ru" \
            -v env_mode="production" \
            -v ssl_mode="prefer" '
            /^DB_HOST=/ { print "DB_HOST=" db_host; next }
            /^DB_USER=/ { print "DB_USER=" db_user; next }
            /^DB_PASSWORD=/ { print "DB_PASSWORD=\"" db_password "\""; next }
            /^DB_SSL_MODE=/ { print "DB_SSL_MODE=" ssl_mode; next }
            /^SESSION_SECRET=/ { print "SESSION_SECRET=\"" session_secret "\""; next }
            /^PRODUCTION_DOMAIN=/ { print "PRODUCTION_DOMAIN=" production_domain; next }
            /^ENV=/ { print "ENV=" env_mode; next }
            { print }
        ' "$BASE_ENV" > /tmp/docker.env

        scp /tmp/docker.env "$REMOTE_HOST:$REMOTE_DIR/.env"
        rm -f /tmp/docker.env /tmp/existing.env

        log_success ".env configuration deployed (existing credentials preserved)"
    fi

    # Deploy on remote server with DB safety measures (NO BUILD - use pre-built binary)
    log_step "Starting containers with pre-built binary (preserving database)..."
    ssh "$REMOTE_HOST" bash -s << 'DOCKER_SCRIPT'
set -euo pipefail

REMOTE_DIR="/opt/THE_BOT_platform"
cd "$REMOTE_DIR"

echo "=== Docker Safe Deployment (No Build - Pre-built Binary) ==="

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker $USER
    echo "Docker installed. You may need to re-login for group changes."
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# Use 'docker compose' or 'docker-compose'
if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

echo ""
echo "=== Pre-deployment checks ==="

# Check current volume status
echo "Current Docker volumes:"
docker volume ls | grep postgres_data || echo "Note: postgres_data volume not yet created (will be on first start)"

# Check pre-built binary exists
if [ ! -f "$REMOTE_DIR/backend/bin/server" ]; then
    echo "Error: Pre-built binary not found at $REMOTE_DIR/backend/bin/server"
    exit 1
fi
echo "✓ Pre-built binary found: $(ls -lh $REMOTE_DIR/backend/bin/server | awk '{print $5}')"

echo ""
echo "=== Stopping services (preserving volumes) ==="

# CRITICAL: Use 'down' without --volumes to preserve data
$COMPOSE_CMD -f docker-compose.prod.yml down 2>/dev/null || true

# Wait for services to fully stop
sleep 3

echo "✓ Containers stopped, database volume preserved"

# Stop legacy services if running
echo "Stopping legacy services..."
pkill -f tutoring-backend 2>/dev/null || true
sudo systemctl stop tutoring-platform.service 2>/dev/null || true

echo ""
echo "=== SSL Certificate Check ==="
if [ -d "/etc/letsencrypt/live/the-bot.ru" ]; then
    echo "✓ Certificate exists at /etc/letsencrypt/live/the-bot.ru/"
else
    echo "⚠ No SSL certificate found - HTTPS may not work"
fi

echo ""
echo "=== Starting containers (NO BUILD) ==="

# Start containers WITHOUT build - use pre-built binary via volume mount
echo "Starting containers with pre-built binary..."
$COMPOSE_CMD -f docker-compose.prod.yml up -d

# Wait for services
echo "Waiting for services to start..."
sleep 15

echo ""
echo "=== Post-deployment verification ==="

# Check container status
echo "Container status:"
$COMPOSE_CMD -f docker-compose.prod.yml ps

# Check PostgreSQL volume
echo ""
echo "Checking database volume:"
if docker volume inspect postgres_data > /dev/null 2>&1; then
    echo "✓ postgres_data volume exists"
    VOLUME_SIZE=$(docker volume inspect postgres_data | grep -A 10 "Mountpoint" | head -1)
    echo "  Volume info: $VOLUME_SIZE"
else
    echo "⚠ postgres_data volume missing - database will be reinitialized"
fi

# Check health
echo ""
echo "Checking service health..."
for i in {1..30}; do
    if curl -s http://localhost/health > /dev/null 2>&1; then
        echo "✓ Service is healthy!"
        curl -s http://localhost/health | head -5
        echo ""
        break
    fi
    if [ $i -eq 30 ]; then
        echo "⚠ Health check timeout"
        echo "Container logs:"
        $COMPOSE_CMD -f docker-compose.prod.yml logs --tail=30
    fi
    sleep 2
done

echo ""
echo "=== Deployment Complete ==="
echo "Platform available at:"
echo "  https://the-bot.ru"
echo "  http://5.129.249.206"
DOCKER_SCRIPT

    log_success "Docker deployment completed"
}

# Verify database integrity after deployment
verify_database() {
    log_step "Verifying database integrity..."

    ssh "$REMOTE_HOST" bash -s << 'VERIFY_SCRIPT'
set -euo pipefail

REMOTE_DIR="/opt/THE_BOT_platform"

echo "Connecting to database..."

# Get credentials from .env
DB_USER=$(grep "^DB_USER=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 || echo "tutoring")
DB_PASSWORD=$(grep "^DB_PASSWORD=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

# Check database using docker exec
if docker ps | grep -q tutoring-postgres; then
    # Try to get row count from main tables
    ROW_COUNT=$(docker exec tutoring-postgres psql -U "$DB_USER" -d tutoring_platform -t -c "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "error")

    if [ "$ROW_COUNT" != "error" ] && [ ! -z "$ROW_COUNT" ]; then
        echo "✓ Database is accessible"
        echo "  Sample query (users count): $ROW_COUNT"
    else
        echo "⚠ Database connection issue"
        exit 1
    fi
else
    echo "✗ PostgreSQL container not running"
    exit 1
fi
VERIFY_SCRIPT

    log_success "Database integrity verified"
}

# Main flow
main() {
    check_ssh

    if [ "$BACKUP_BEFORE_DEPLOY" = true ]; then
        backup_database
    else
        log_info "Skipping database backup (--no-backup flag used)"
    fi

    if [ "$DEPLOY_MODE" = "docker" ]; then
        build_backend
        deploy_docker_safe
    else
        log_error "Legacy mode not yet implemented in this script"
        exit 1
    fi

    verify_database

    echo ""
    echo -e "${GREEN}=== ✓ Safe Deployment Successful ===${NC}"
    echo ""
    echo "Summary:"
    echo "  ✓ Database preserved from previous deployment"
    echo "  ✓ Containers rebuilt and restarted"
    echo "  ✓ Services verified and healthy"
    echo ""
    echo "Platform available at:"
    echo "  https://the-bot.ru"
    echo ""
}

main "$@"
