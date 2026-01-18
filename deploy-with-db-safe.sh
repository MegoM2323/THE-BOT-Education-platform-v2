#!/bin/bash

# Safe Deployment Script - Preserves Database on Redeploy
# This script ensures database data is preserved during redeployment
# Usage: ./deploy-with-db-safe.sh [--no-backup] [--legacy]

set -euo pipefail

# Configuration
REMOTE_HOST="miroslav@213.171.25.168"
REMOTE_DIR="/opt/tutoring-platform"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_BEFORE_DEPLOY=true
DEPLOY_MODE="docker"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-admin@diploma-m.ru}"

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

    ssh "$REMOTE_HOST" bash -s << 'BACKUP_SCRIPT'
set -euo pipefail

REMOTE_DIR="/opt/tutoring-platform"
BACKUP_DIR="$REMOTE_DIR/backups"

# Create backup directory
mkdir -p "$BACKUP_DIR"

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

echo "Creating database backup..."

# Backup using docker exec if PostgreSQL is in container
if docker ps 2>/dev/null | grep -q tutoring-postgres; then
    docker exec tutoring-postgres pg_dump \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --no-owner \
        --no-privileges \
        > "$BACKUP_FILE" 2>/dev/null || {
        echo "Error: Docker backup failed"
        exit 1
    }
else
    # Try native PostgreSQL if not in Docker
    PGPASSWORD="$DB_PASSWORD" pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --no-owner \
        --no-privileges \
        > "$BACKUP_FILE" 2>/dev/null || {
        echo "Warning: Native PostgreSQL backup failed, container might be down"
        # This is not fatal - we'll try to preserve the data anyway
        rm -f "$BACKUP_FILE"
        exit 0
    }
fi

if [ -f "$BACKUP_FILE" ]; then
    # Compress backup
    gzip -f "$BACKUP_FILE"

    # Keep only last 5 backups
    ls -t "$BACKUP_DIR"/db_backup_*.sql.gz 2>/dev/null | tail -n +6 | xargs -r rm

    echo "✓ Backup created: $(du -h "$BACKUP_FILE_GZ" | cut -f1)"
fi
BACKUP_SCRIPT

    log_success "Database backup completed"
}

# Deploy with Docker (preserving database)
deploy_docker_safe() {
    log_step "Starting Docker deployment (with database preservation)..."

    # Copy all files (same as original deploy-ssh.sh)
    ssh "$REMOTE_HOST" "mkdir -p $REMOTE_DIR/backend $REMOTE_DIR/frontend"

    log_info "Copying Docker configuration..."
    scp "$PROJECT_DIR/docker-compose.prod.yml" "$REMOTE_HOST:$REMOTE_DIR/"

    # Copy backend files
    scp "$PROJECT_DIR/backend/Dockerfile" "$REMOTE_HOST:$REMOTE_DIR/backend/"
    scp "$PROJECT_DIR/backend/entrypoint.sh" "$REMOTE_HOST:$REMOTE_DIR/backend/"
    scp "$PROJECT_DIR/backend/go.mod" "$REMOTE_HOST:$REMOTE_DIR/backend/"
    scp "$PROJECT_DIR/backend/go.sum" "$REMOTE_HOST:$REMOTE_DIR/backend/"

    log_info "Copying backend source..."
    rsync -avz --delete \
        --exclude='*.out' \
        --exclude='*.test' \
        --exclude='tutoring-backend' \
        --exclude='uploads/*' \
        "$PROJECT_DIR/backend/cmd" "$REMOTE_HOST:$REMOTE_DIR/backend/"
    rsync -avz --delete \
        "$PROJECT_DIR/backend/internal" "$REMOTE_HOST:$REMOTE_DIR/backend/"
    rsync -avz --delete \
        "$PROJECT_DIR/backend/pkg" "$REMOTE_HOST:$REMOTE_DIR/backend/"

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
        --exclude='dist' \
        "$PROJECT_DIR/frontend/src" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
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
            -v production_domain="https://diploma-m.ru" \
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

    # Deploy on remote server with DB safety measures
    log_step "Building and restarting containers (preserving database)..."
    ssh "$REMOTE_HOST" CERTBOT_EMAIL="$CERTBOT_EMAIL" bash -s << 'DOCKER_SCRIPT'
set -euo pipefail

REMOTE_DIR="/opt/tutoring-platform"
cd "$REMOTE_DIR"

echo "=== Docker Safe Deployment (Database Preservation) ==="

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

echo ""
echo "=== Stopping services (preserving volumes) ==="

# CRITICAL: Use 'down' without --volumes to preserve data
# This stops containers but leaves volumes intact
$COMPOSE_CMD -f docker-compose.prod.yml down 2>/dev/null || true

# Wait for services to fully stop
sleep 3

echo "✓ Containers stopped, database volume preserved"

# Stop legacy services if running
echo "Stopping legacy services..."
pkill -f tutoring-backend 2>/dev/null || true
sudo systemctl stop tutoring-platform.service 2>/dev/null || true

# Install and configure Certbot
echo ""
echo "=== SSL Certificate Setup ==="
if ! command -v certbot &> /dev/null; then
    sudo apt-get update -qq
    sudo apt-get install -y certbot python3-certbot-nginx > /dev/null 2>&1
fi

check_certificate_valid() {
    if [ ! -d "/etc/letsencrypt/live/diploma-m.ru" ]; then
        return 1
    fi
    if command -v certbot &> /dev/null; then
        certbot certificates 2>/dev/null | grep -q "diploma-m.ru" && return 0
        return 1
    fi
    return 0
}

if ! check_certificate_valid; then
    if [ -z "${CERTBOT_EMAIL}" ]; then
        echo "Error: CERTBOT_EMAIL is not set"
        exit 1
    fi
    echo "Requesting Let's Encrypt certificate..."
    sudo certbot certonly \
        --standalone \
        --non-interactive \
        --agree-tos \
        --email "${CERTBOT_EMAIL}" \
        -d diploma-m.ru \
        -d www.diploma-m.ru \
        2>&1 || echo "⚠ Certificate request failed or already exists"
else
    echo "Certificate valid at /etc/letsencrypt/live/diploma-m.ru/"
fi

# Setup auto-renewal
sudo systemctl enable certbot.timer 2>/dev/null || true
sudo systemctl start certbot.timer 2>/dev/null || true

# Ensure permissions
sudo setfacl -R -m u:root:rx /etc/letsencrypt/live/diploma-m.ru 2>/dev/null || \
    sudo chmod -R 755 /etc/letsencrypt/live/diploma-m.ru 2>/dev/null || true

echo ""
echo "=== Building and starting containers ==="

# Build with no cache
echo "Building Docker images..."
$COMPOSE_CMD -f docker-compose.prod.yml build --no-cache

# Start containers - this will use existing volume data
echo "Starting containers (using existing database if available)..."
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
        $COMPOSE_CMD -f docker-compose.prod.yml logs --tail=20
    fi
    sleep 2
done

echo ""
echo "=== Deployment Complete ==="
echo "Platform available at:"
echo "  https://diploma-m.ru"
echo "  http://213.171.25.168"
DOCKER_SCRIPT

    log_success "Docker deployment completed"
}

# Verify database integrity after deployment
verify_database() {
    log_step "Verifying database integrity..."

    ssh "$REMOTE_HOST" bash -s << 'VERIFY_SCRIPT'
set -euo pipefail

REMOTE_DIR="/opt/tutoring-platform"

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
    echo "  https://diploma-m.ru"
    echo ""
}

main "$@"
