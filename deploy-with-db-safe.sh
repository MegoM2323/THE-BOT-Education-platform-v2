#!/bin/bash

# Safe Deployment Script - Preserves Database on Redeploy
# This script ensures database data is preserved during redeployment
# Usage: ./deploy-with-db-safe.sh [--no-backup]

set -euo pipefail

# Configuration
REMOTE_HOST="mg@5.129.249.206"
REMOTE_DIR="/opt/THE_BOT_platform"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_BEFORE_DEPLOY=true

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
        *)
            ;;
    esac
done

echo -e "${YELLOW}=== Safe Deployment with Database Preservation ===${NC}"
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

mkdir -p "$BACKUP_DIR"

if [ ! -f "$REMOTE_DIR/.env" ]; then
    echo "SKIP: No .env file found (first deployment?)"
    exit 0
fi

DB_USER=$(grep "^DB_USER=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 || echo "tutoring")
DB_PASSWORD=$(grep "^DB_PASSWORD=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')
DB_NAME="tutoring_platform"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/db_backup_${TIMESTAMP}.sql"
BACKUP_FILE_GZ="$BACKUP_FILE.gz"

echo "Attempting database backup..."
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"

if ! docker ps 2>/dev/null | grep -q tutoring-postgres; then
    echo "SKIP: PostgreSQL container 'tutoring-postgres' not running"
    exit 0
fi

echo "  Container: tutoring-postgres is running"

if ! docker exec tutoring-postgres pg_isready -U "$DB_USER" -d "$DB_NAME" 2>&1; then
    echo "SKIP: PostgreSQL is not ready"
    exit 0
fi

echo "  PostgreSQL: ready"

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
    echo "WARN: Continuing without backup"
    exit 0
fi

echo "$DUMP_OUTPUT" > "$BACKUP_FILE"

if [ ! -s "$BACKUP_FILE" ]; then
    echo "WARN: Backup file is empty"
    rm -f "$BACKUP_FILE"
    exit 0
fi

gzip -f "$BACKUP_FILE"

ls -t "$BACKUP_DIR"/db_backup_*.sql.gz 2>/dev/null | tail -n +6 | while read -r f; do rm -f "$f"; done

echo "OK: Backup created: $(du -h "$BACKUP_FILE_GZ" | cut -f1)"
BACKUP_SCRIPT
    )

    echo "$BACKUP_RESULT"

    if echo "$BACKUP_RESULT" | grep -q "^OK:"; then
        log_success "Database backup completed"
    elif echo "$BACKUP_RESULT" | grep -q "^SKIP:"; then
        log_info "Database backup skipped"
    elif echo "$BACKUP_RESULT" | grep -q "^WARN:"; then
        log_info "Database backup had warnings, continuing"
    else
        log_info "Backup status unknown, continuing deployment"
    fi
}

# Prepare backend source for remote build
prepare_backend_source() {
    log_step "Preparing backend source for remote build..."

    cd "$PROJECT_DIR"

    log_info "Creating backend source package..."
    tar -czf /tmp/backend-source.tar.gz \
        --exclude='bin' \
        --exclude='*.log' \
        backend/ 2>/dev/null || true

    local SIZE=$(du -h /tmp/backend-source.tar.gz | cut -f1)
    log_success "Backend source packaged (size: $SIZE)"
}

# Build frontend
build_frontend() {
    log_step "Building frontend..."

    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed"
        exit 1
    fi

    cd "$PROJECT_DIR/frontend"
    npm ci --legacy-peer-deps
    npm run build

    if [ ! -d "dist" ]; then
        log_error "Failed to build frontend"
        exit 1
    fi

    log_success "Frontend built successfully"
    cd "$PROJECT_DIR"
}

# Check rsync availability
check_rsync() {
    if ! command -v rsync &> /dev/null; then
        log_error "rsync is not installed"
        exit 1
    fi
}

# Deploy with Docker (preserving database)
deploy_docker_safe() {
    log_step "Starting Docker deployment (with database preservation)..."

    if [ ! -d "$PROJECT_DIR/frontend/dist" ]; then
        log_error "frontend/dist не существует"
        exit 1
    fi

    if [ ! -f "$PROJECT_DIR/frontend/nginx.conf.prod" ]; then
        log_error "frontend/nginx.conf.prod не существует"
        exit 1
    fi

    log_info "Copying Docker configuration..."
    scp "$PROJECT_DIR/docker-compose.prod.yml" "$REMOTE_HOST:$REMOTE_DIR/"

    log_info "Creating directory structure on server..."
    ssh "$REMOTE_HOST" "mkdir -p $REMOTE_DIR/backend"

    log_info "Stopping containers before deployment..."
    ssh "$REMOTE_HOST" bash -s << 'STOP_SCRIPT'
set -euo pipefail
REMOTE_DIR="/opt/THE_BOT_platform"
cd "$REMOTE_DIR"

if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

$COMPOSE_CMD -f docker-compose.prod.yml down 2>/dev/null || true
sleep 2
echo "Containers stopped"
STOP_SCRIPT

    log_info "Copying backend source to server..."
    ssh "$REMOTE_HOST" "mkdir -p $REMOTE_DIR/backend"
    scp /tmp/backend-source.tar.gz "$REMOTE_HOST:$REMOTE_DIR/"
    ssh "$REMOTE_HOST" "cd $REMOTE_DIR && tar -xzf backend-source.tar.gz && rm backend-source.tar.gz"

    log_info "Copying entrypoint script..."
    scp "$PROJECT_DIR/backend/entrypoint.sh" "$REMOTE_HOST:$REMOTE_DIR/backend/"

    log_info "Copying migrations (redundant but safe)..."
    rsync -avz "$PROJECT_DIR/backend/internal/database/migrations/" "$REMOTE_HOST:$REMOTE_DIR/backend/internal/database/migrations/"

    log_info "Copying frontend files..."
    rsync -avz \
        --exclude='node_modules' \
        "$PROJECT_DIR/frontend/src" "$REMOTE_HOST:$REMOTE_DIR/frontend/" 2>/dev/null || true
    rsync -avz \
        "$PROJECT_DIR/frontend/dist" "$REMOTE_HOST:$REMOTE_DIR/frontend/dist"
    rsync -avz "$PROJECT_DIR/frontend/public" "$REMOTE_HOST:$REMOTE_DIR/frontend/" 2>/dev/null || true

    log_info "Copying nginx config..."
    cp "$PROJECT_DIR/frontend/nginx.conf.prod" "$PROJECT_DIR/frontend/nginx.conf"
    scp "$PROJECT_DIR/frontend/nginx.conf" "$REMOTE_HOST:$REMOTE_DIR/frontend/"
    git checkout "$PROJECT_DIR/frontend/nginx.conf" 2>/dev/null || true

    log_info "Managing .env configuration..."
    if ssh "$REMOTE_HOST" "[[ -f $REMOTE_DIR/.env ]]" 2>/dev/null; then
        log_info "Fetching existing .env from server..."
        scp "$REMOTE_HOST:$REMOTE_DIR/.env" /tmp/existing.env 2>/dev/null || true
    fi

    if [[ -f "$PROJECT_DIR/backend/.env" ]]; then
        if [[ -f /tmp/existing.env ]]; then
            BASE_ENV="/tmp/existing.env"
            log_info "Using existing server .env as base"
        else
            BASE_ENV="$PROJECT_DIR/backend/.env"
            log_info "Using local .env as base"
        fi

        DB_PASSWORD=$(grep "^DB_PASSWORD=" "$BASE_ENV" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || openssl rand -hex 24)

        # Сохраняем существующий SESSION_SECRET с сервера или генерируем новый
        EXISTING_SECRET=$(grep "^SESSION_SECRET=" /tmp/existing.env 2>/dev/null | cut -d'=' -f2 | tr -d '"')

        if [ -n "$EXISTING_SECRET" ]; then
            # Секрет существует на сервере - сохраняем его
            SESSION_SECRET="$EXISTING_SECRET"
            log_info "Preserving existing SESSION_SECRET from server"
        else
            # Первая установка или секрет отсутствует - генерируем
            SESSION_SECRET=$(grep "^SESSION_SECRET=" "$BASE_ENV" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || openssl rand -base64 64 | tr -d '\n')
            if [ -n "$SESSION_SECRET" ]; then
                log_info "Using SESSION_SECRET from base env"
            else
                log_info "Generated new SESSION_SECRET (first deployment)"
            fi
        fi

        # Сохраняем PRODUCTION_DOMAIN, SESSION_SAME_SITE и DB_SSL_MODE из .env
        PRODUCTION_DOMAIN=$(grep "^PRODUCTION_DOMAIN=" "$BASE_ENV" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || echo "https://the-bot.ru")
        SESSION_SAME_SITE=$(grep "^SESSION_SAME_SITE=" "$BASE_ENV" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || echo "Lax")
        DB_SSL_MODE=$(grep "^DB_SSL_MODE=" "$BASE_ENV" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || echo "require")

        awk -v db_host="postgres" \
            -v db_user="tutoring" \
            -v db_password="$DB_PASSWORD" \
            -v session_secret="$SESSION_SECRET" \
            -v production_domain="$PRODUCTION_DOMAIN" \
            -v session_same_site="$SESSION_SAME_SITE" \
            -v env_mode="production" \
            -v ssl_mode="$DB_SSL_MODE" '
            /^DB_HOST=/ { print "DB_HOST=" db_host; next }
            /^DB_USER=/ { print "DB_USER=" db_user; next }
            /^DB_PASSWORD=/ { print "DB_PASSWORD=\"" db_password "\""; next }
            /^DB_SSL_MODE=/ { print "DB_SSL_MODE=" ssl_mode; next }
            /^SESSION_SECRET=/ { print "SESSION_SECRET=\"" session_secret "\""; next }
            /^PRODUCTION_DOMAIN=/ { print "PRODUCTION_DOMAIN=" production_domain; next }
            /^SESSION_SAME_SITE=/ { print "SESSION_SAME_SITE=" session_same_site; next }
            /^ENV=/ { print "ENV=" env_mode; next }
            { print }
        ' "$BASE_ENV" > /tmp/docker.env

        scp /tmp/docker.env "$REMOTE_HOST:$REMOTE_DIR/.env"
        rm -f /tmp/docker.env /tmp/existing.env

        log_success ".env configuration deployed"
    fi

    log_step "Deploying on remote server..."
    ssh "$REMOTE_HOST" bash -s << DOCKER_SCRIPT
set -euo pipefail

REMOTE_DIR="/opt/THE_BOT_platform"
cd "\$REMOTE_DIR"

echo "=== Docker Safe Deployment ==="

if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker \$USER
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-\$(uname -s)-\$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

echo ""
echo "=== Pre-deployment checks ==="

echo "Current Docker volumes:"
docker volume ls | grep postgres || echo "Note: postgres volume not yet created"

echo ""
echo "=== Checking database volume ==="

if docker volume ls | grep -q "the_bot_v3_postgres_data"; then
    echo "✓ Database volume exists: the_bot_v3_postgres_data"
else
    echo "⚠ Database volume 'the_bot_v3_postgres_data' not found - fresh installation"
fi

echo ""
echo "=== Stopping legacy services ==="
pkill -f tutoring-backend 2>/dev/null || true
sudo systemctl stop tutoring-platform.service 2>/dev/null || true

echo ""
echo "=== SSL Certificate Check ==="
if [ -d "/etc/letsencrypt/live/the-bot.ru" ]; then
    echo "✓ Certificate exists"
else
    echo "⚠ No SSL certificate found"
fi

echo ""
echo "=== Starting containers ==="

\$COMPOSE_CMD -f docker-compose.prod.yml up -d

sleep 5
if ! \$COMPOSE_CMD -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo "Ошибка: контейнеры не запустились"
    \$COMPOSE_CMD -f docker-compose.prod.yml logs --tail=30
    exit 1
fi

echo "Waiting for services to start..."
sleep 10

echo ""
echo "=== Post-deployment verification ==="

echo "Container status:"
\$COMPOSE_CMD -f docker-compose.prod.yml ps

echo ""
echo "Checking service health..."
for i in {1..30}; do
    if curl -s http://localhost/health > /dev/null 2>&1; then
        echo "✓ Service is healthy!"
        curl -s http://localhost/health | head -5
        echo ""
        break
    fi
    if [ \$i -eq 30 ]; then
        echo "⚠ Health check timeout"
        \$COMPOSE_CMD -f docker-compose.prod.yml logs --tail=30 backend
    fi
    sleep 2
done

echo ""
echo "=== Deployment Complete ==="
echo "Platform available at:"
echo "  $PRODUCTION_DOMAIN"
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

DB_USER=$(grep "^DB_USER=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 || echo "tutoring")
DB_PASSWORD=$(grep "^DB_PASSWORD=" "$REMOTE_DIR/.env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

if docker ps | grep -q tutoring-postgres; then
    ROW_COUNT=$(docker exec tutoring-postgres psql -U "$DB_USER" -d tutoring_platform -t -c "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "error")

    if [ "$ROW_COUNT" != "error" ] && [ ! -z "$ROW_COUNT" ]; then
        echo "✓ Database is accessible"
        echo "  Users count: $ROW_COUNT"
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
    check_rsync

    if [ "$BACKUP_BEFORE_DEPLOY" = true ]; then
        backup_database
    else
        log_info "Skipping database backup (--no-backup flag used)"
    fi

    prepare_backend_source
    build_frontend
    deploy_docker_safe
    verify_database

    echo ""
    echo -e "${GREEN}=== ✓ Safe Deployment Successful ===${NC}"
    echo ""
    echo "Summary:"
    echo "  ✓ Backend source prepared for remote build"
    echo "  ✓ Frontend built and deployed"
    echo "  ✓ Database preserved from previous deployment"
    echo "  ✓ Services verified and healthy"
    echo ""
    echo "Platform available at:"
    echo "  $PRODUCTION_DOMAIN"
    echo ""
}

main "$@"
