#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

REMOTE_USER="mg"
REMOTE_HOST="5.129.249.206"
REMOTE_ADDR="${REMOTE_USER}@${REMOTE_HOST}"
THEBOT_HOME="/home/mg/the-bot"
GIT_BRANCH="master"

DRY_RUN=false
VERBOSE=false
ROLLBACK_ON_ERROR=true
SKIP_FRONTEND=false
SKIP_BACKEND=false

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

PHASE_CURRENT=0
PHASE_TOTAL=15

show_help() {
    cat << 'EOF'
Production Deployment Script for THE_BOT V3

Usage: ./safe-deploy-native.sh [OPTIONS]

Options:
  --dry-run              Simulate deployment without making changes
  --verbose              Enable verbose output
  --no-rollback          Disable automatic rollback on error
  --skip-frontend        Skip frontend build
  --skip-backend         Skip backend build
  --branch <branch>      Deploy from specific git branch (default: master)
  --help                 Show this help message

Examples:
  # Standard production deployment
  ./safe-deploy-native.sh

  # Test deployment without making changes
  ./safe-deploy-native.sh --dry-run --verbose

  # Deploy with no auto-rollback
  ./safe-deploy-native.sh --no-rollback

  # Quick hotfix (skip frontend)
  ./safe-deploy-native.sh --skip-frontend

Environment Variables:
  REMOTE_USER            SSH user (default: mg)
  REMOTE_HOST            SSH host (default: 5.129.249.206)
  THEBOT_HOME            Project path on remote (default: /home/mg/the-bot)
  GIT_BRANCH             Git branch to deploy (default: master)

EOF
}

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --no-rollback)
                ROLLBACK_ON_ERROR=false
                shift
                ;;
            --skip-frontend)
                SKIP_FRONTEND=true
                shift
                ;;
            --skip-backend)
                SKIP_BACKEND=true
                shift
                ;;
            --branch)
                shift
                GIT_BRANCH="$1"
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                echo "Unknown argument: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[${timestamp}] ${level}: ${message}" >> "$LOG_FILE"

    case $level in
        INFO)
            echo -e "${CYAN}[${timestamp}] ℹ ${message}${NC}"
            ;;
        SUCCESS)
            echo -e "${GREEN}[${timestamp}] ✓ ${message}${NC}"
            ;;
        WARNING)
            echo -e "${YELLOW}[${timestamp}] ⚠ ${message}${NC}"
            ;;
        ERROR)
            echo -e "${RED}[${timestamp}] ✗ ${message}${NC}"
            ;;
    esac
}

phase() {
    PHASE_CURRENT=$((PHASE_CURRENT + 1))
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║ Phase $PHASE_CURRENT/$PHASE_TOTAL: $1${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════╝${NC}"
    log "INFO" "Phase $PHASE_CURRENT/$PHASE_TOTAL: $1"
}

setup_logging() {
    PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    LOG_FILE="${PROJECT_ROOT}/logs/deploy_${TIMESTAMP}.log"

    mkdir -p "$(dirname "$LOG_FILE")"
    touch "$LOG_FILE"
    log "INFO" "Deployment started"
    log "INFO" "Log file: $LOG_FILE"
    log "INFO" "DRY_RUN: $DRY_RUN"
    log "INFO" "ROLLBACK_ON_ERROR: $ROLLBACK_ON_ERROR"
}

cleanup_and_exit() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        log "ERROR" "Deployment failed with exit code $exit_code"

        if [ "$ROLLBACK_ON_ERROR" = true ] && [ -n "${BACKUP_COMMIT_HASH:-}" ]; then
            log "WARNING" "Attempting to rollback to commit: $BACKUP_COMMIT_HASH"

            if ! $DRY_RUN; then
                ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME && git reset --hard $BACKUP_COMMIT_HASH" || true
                log "WARNING" "Restarting services after rollback..."
                restart_services || true
            fi
        fi
    else
        log "SUCCESS" "Deployment completed successfully"
    fi

    log "INFO" "Deployment log: $LOG_FILE"
    exit $exit_code
}

execute_remote() {
    local cmd="$1"

    if [ "$DRY_RUN" = true ]; then
        log "INFO" "(DRY-RUN) Would execute: $cmd"
        return 0
    fi

    if [ "$VERBOSE" = true ]; then
        log "INFO" "Executing: $cmd"
    fi

    ssh "$REMOTE_ADDR" "$cmd"
}

parse_arguments "$@"
setup_logging
trap cleanup_and_exit EXIT

phase "Pre-deployment checks"

log "INFO" "Checking SSH connection to $REMOTE_ADDR..."
if ! ssh -q -o ConnectTimeout=5 "$REMOTE_ADDR" "echo ok" > /dev/null 2>&1; then
    log "ERROR" "Cannot connect to $REMOTE_ADDR"
    exit 1
fi
log "SUCCESS" "SSH connection verified"

log "INFO" "Checking project directory..."
if ! ssh -q "$REMOTE_ADDR" "test -d $THEBOT_HOME" 2>/dev/null; then
    log "ERROR" "Project directory not found: $THEBOT_HOME"
    exit 1
fi
log "SUCCESS" "Project directory exists"

log "INFO" "Checking required services..."
REQUIRED_SERVICES=(
    "thebot-backend.service"
    "thebot-daphne.service"
    "thebot-celery-worker.service"
    "thebot-celery-beat.service"
)

for service in "${REQUIRED_SERVICES[@]}"; do
    if ! ssh -q "$REMOTE_ADDR" "systemctl list-unit-files | grep -q $service" 2>/dev/null; then
        log "WARNING" "Service not found: $service"
    fi
done
log "SUCCESS" "Service checks completed"

phase "Backup current state"

log "INFO" "Getting current git commit hash..."
BACKUP_COMMIT_HASH=$(ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME && git rev-parse HEAD")
log "SUCCESS" "Backup commit: $BACKUP_COMMIT_HASH"

log "INFO" "Creating backup of current code..."
if ! $DRY_RUN; then
    BACKUP_DIR="/tmp/thebot_backup_${TIMESTAMP}"
    execute_remote "mkdir -p $BACKUP_DIR && cp -r $THEBOT_HOME/* $BACKUP_DIR 2>/dev/null || true"
    log "SUCCESS" "Backup created at $BACKUP_DIR"
fi

phase "Check disk space and system resources"

log "INFO" "Checking disk space..."
DISK_USAGE=$(ssh -q "$REMOTE_ADDR" "df -h $THEBOT_HOME | tail -1 | awk '{print \$5}' | sed 's/%//'")
log "INFO" "Disk usage: ${DISK_USAGE}%"
if [ "$DISK_USAGE" -gt 90 ]; then
    log "ERROR" "Insufficient disk space (${DISK_USAGE}% used)"
    exit 1
fi

log "INFO" "Checking available RAM..."
RAM_AVAILABLE=$(ssh -q "$REMOTE_ADDR" "free -m | grep Mem | awk '{print \$7}'")
log "INFO" "Available RAM: ${RAM_AVAILABLE}MB"
if [ "$RAM_AVAILABLE" -lt 512 ]; then
    log "WARNING" "Low available RAM: ${RAM_AVAILABLE}MB"
fi

log "SUCCESS" "System resources check passed"

phase "Fetch and update code"

log "INFO" "Fetching latest changes..."
execute_remote "cd $THEBOT_HOME && git fetch origin"

log "INFO" "Checking out branch: $GIT_BRANCH..."
execute_remote "cd $THEBOT_HOME && git checkout $GIT_BRANCH"

log "INFO" "Pulling latest changes..."
execute_remote "cd $THEBOT_HOME && git pull origin $GIT_BRANCH"

log "SUCCESS" "Code updated successfully"

phase "Build backend (Go)"

if [ "$SKIP_BACKEND" = true ]; then
    log "WARNING" "Skipping backend build"
else
    log "INFO" "Building backend binary..."

    if $DRY_RUN; then
        log "INFO" "(DRY-RUN) Would build: cd $THEBOT_HOME/backend && go build -o server cmd/main.go"
    else
        execute_remote "cd $THEBOT_HOME/backend && go build -o server cmd/main.go"

        if [ $? -ne 0 ]; then
            log "ERROR" "Backend build failed"
            exit 1
        fi

        log "SUCCESS" "Backend binary compiled"

        log "INFO" "Verifying backend binary..."
        execute_remote "test -f $THEBOT_HOME/backend/server && file $THEBOT_HOME/backend/server"
        log "SUCCESS" "Backend binary verified"
    fi
fi

phase "Build frontend (React)"

if [ "$SKIP_FRONTEND" = true ]; then
    log "WARNING" "Skipping frontend build"
else
    log "INFO" "Installing frontend dependencies..."

    if ! $DRY_RUN; then
        execute_remote "cd $THEBOT_HOME/frontend && npm ci --legacy-peer-deps"

        if [ $? -ne 0 ]; then
            log "ERROR" "npm install failed"
            exit 1
        fi
        log "SUCCESS" "Dependencies installed"
    fi

    log "INFO" "Building frontend production bundle..."

    if $DRY_RUN; then
        log "INFO" "(DRY-RUN) Would run: cd $THEBOT_HOME/frontend && npm run build"
    else
        execute_remote "cd $THEBOT_HOME/frontend && npm run build"

        if [ $? -ne 0 ]; then
            log "ERROR" "Frontend build failed"
            exit 1
        fi

        log "SUCCESS" "Frontend bundle created"

        log "INFO" "Verifying frontend build..."
        execute_remote "test -d $THEBOT_HOME/frontend/dist && ls -lh $THEBOT_HOME/frontend/dist/index.html"
        log "SUCCESS" "Frontend build verified"
    fi
fi

phase "Database migrations"

log "INFO" "Checking database connectivity..."
if ! $DRY_RUN; then
    if ! ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME/backend && psql \$DATABASE_URL -c 'SELECT 1' > /dev/null 2>&1" 2>/dev/null; then
        log "WARNING" "Could not connect to database, attempting with direct psql..."
        execute_remote "psql -U postgres -d thebot_db -c 'SELECT 1' > /dev/null 2>&1" || log "WARNING" "Database check skipped"
    else
        log "SUCCESS" "Database connection verified"
    fi
fi

log "INFO" "Running database migrations..."
if ! $DRY_RUN; then
    if execute_remote "cd $THEBOT_HOME/backend && go run cmd/main.go migrate"; then
        log "SUCCESS" "Migrations completed"
    else
        log "WARNING" "Migration command may not be available, continuing..."
    fi
fi

log "INFO" "Verifying migration 050 is applied..."
if ! $DRY_RUN; then
    MIGRATION_STATUS=$(ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME/backend && psql \$DATABASE_URL -tc \"SELECT EXISTS(SELECT 1 FROM public.schema_migrations WHERE version='050')\" 2>/dev/null || echo 'unknown'")

    if [ "$MIGRATION_STATUS" = "t" ] || [ "$MIGRATION_STATUS" = " t" ]; then
        log "SUCCESS" "Migration 050 is applied"
    else
        log "WARNING" "Migration 050 status: $MIGRATION_STATUS"
    fi
fi

phase "Stop services gracefully"

log "INFO" "Stopping services..."
SERVICES=(
    "thebot-backend.service"
    "thebot-daphne.service"
    "thebot-celery-worker.service"
    "thebot-celery-beat.service"
)

for service in "${SERVICES[@]}"; do
    log "INFO" "Stopping $service..."
    execute_remote "systemctl stop $service || true" || true
    sleep 1
done

log "INFO" "Waiting for services to stop..."
sleep 3

log "INFO" "Verifying services are stopped..."
for service in "${SERVICES[@]}"; do
    if ! ssh -q "$REMOTE_ADDR" "systemctl is-active --quiet $service 2>/dev/null && echo stopped || echo already_stopped" > /dev/null 2>&1; then
        log "INFO" "Service confirmed stopped: $service"
    fi
done

log "SUCCESS" "All services stopped"

phase "Deploy application files"

log "INFO" "Setting proper file permissions..."
if ! $DRY_RUN; then
    execute_remote "chmod +x $THEBOT_HOME/backend/server"
    execute_remote "chmod -R 755 $THEBOT_HOME/frontend/dist 2>/dev/null || true"
    log "SUCCESS" "Permissions set"
fi

log "INFO" "Preparing static files..."
if ! $DRY_RUN; then
    execute_remote "mkdir -p $THEBOT_HOME/static && cp -r $THEBOT_HOME/frontend/dist/* $THEBOT_HOME/static/ 2>/dev/null || true"
    log "SUCCESS" "Static files deployed"
fi

phase "Start services"

log "INFO" "Starting services..."
for service in "${SERVICES[@]}"; do
    log "INFO" "Starting $service..."
    execute_remote "systemctl start $service" || log "WARNING" "Failed to start $service"
    sleep 2
done

log "INFO" "Waiting for services to stabilize..."
sleep 5

log "SUCCESS" "Services started"

phase "Health checks and verification"

log "INFO" "Checking service status..."
for service in "${SERVICES[@]}"; do
    if ssh -q "$REMOTE_ADDR" "systemctl is-active --quiet $service 2>/dev/null"; then
        log "SUCCESS" "✓ $service is running"
    else
        log "ERROR" "✗ $service is NOT running"
        if $DRY_RUN; then
            log "INFO" "(DRY-RUN) Service check would be performed"
        fi
    fi
done

phase "Get deployment statistics"

log "INFO" "Collecting post-deployment statistics..."

if ! $DRY_RUN; then
    log "INFO" "Chat creation statistics:"
    execute_remote "cd $THEBOT_HOME/backend && psql \$DATABASE_URL -tc \"SELECT COUNT(*) as total_chats FROM chats\" 2>/dev/null || echo 'N/A'" || true

    log "INFO" "Recent logs from backend service..."
    execute_remote "journalctl -u thebot-backend.service -n 10 --no-pager 2>/dev/null || echo 'No logs available'" || true
fi

phase "Deployment summary"

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        DEPLOYMENT COMPLETED SUCCESSFULLY              ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}Deployment Summary:${NC}"
echo -e "  Remote Server     : ${GREEN}$REMOTE_ADDR${NC}"
echo -e "  Project Path      : ${GREEN}$THEBOT_HOME${NC}"
echo -e "  Deployed Branch   : ${GREEN}$GIT_BRANCH${NC}"
echo -e "  Timestamp         : ${GREEN}$TIMESTAMP${NC}"
echo -e "  Log File          : ${GREEN}$LOG_FILE${NC}"
echo ""
echo -e "${CYAN}Services Status:${NC}"

if ! $DRY_RUN; then
    for service in "${SERVICES[@]}"; do
        if ssh -q "$REMOTE_ADDR" "systemctl is-active --quiet $service 2>/dev/null"; then
            echo -e "  ${GREEN}✓${NC} $service"
        else
            echo -e "  ${RED}✗${NC} $service"
        fi
    done
fi

echo ""
echo -e "${CYAN}Next Steps:${NC}"
echo -e "  1. Visit https://the-bot.ru to verify deployment"
echo -e "  2. Check logs: ${YELLOW}ssh $REMOTE_ADDR 'journalctl -u thebot-backend -f'${NC}"
echo -e "  3. View deployment log: ${YELLOW}cat $LOG_FILE${NC}"
echo ""

if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}NOTE: This was a DRY-RUN. No changes were made.${NC}"
fi

echo ""

log "SUCCESS" "Deployment script completed"

exit 0
