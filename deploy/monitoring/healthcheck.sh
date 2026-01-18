#!/bin/bash

################################################################################
# Health Check Script
#
# This script monitors the health of all Tutoring Platform services
# and automatically restarts them if they're not responding
#
# Usage: ./healthcheck.sh
# Recommended: Run via cron every 5 minutes
################################################################################

# Configuration
BACKEND_URL="http://localhost:8080/health"
LOG_FILE="/var/log/tutoring-platform/healthcheck.log"
MAX_RETRIES=3
RETRY_DELAY=5

# Color codes (for terminal output)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

################################################################################
# Check Backend Service
################################################################################

check_backend() {
    log "Checking backend service..."

    # Check if service is running
    if ! systemctl is-active --quiet tutoring-platform; then
        log_error "Backend service is not running"
        return 1
    fi

    # Check health endpoint
    for i in $(seq 1 $MAX_RETRIES); do
        if curl -f -s -o /dev/null -w "%{http_code}" "$BACKEND_URL" | grep -q "200"; then
            log_success "Backend health check passed"
            return 0
        fi

        if [ $i -lt $MAX_RETRIES ]; then
            log_warning "Backend health check failed (attempt $i/$MAX_RETRIES), retrying..."
            sleep $RETRY_DELAY
        fi
    done

    log_error "Backend health check failed after $MAX_RETRIES attempts"
    return 1
}

restart_backend() {
    log "Restarting backend service..."
    systemctl restart tutoring-platform

    # Wait for service to start
    sleep 5

    # Verify service started
    if systemctl is-active --quiet tutoring-platform; then
        log_success "Backend service restarted successfully"

        # Check health endpoint after restart
        sleep 3
        if check_backend; then
            return 0
        else
            log_error "Backend service started but health check still failing"
            return 1
        fi
    else
        log_error "Failed to restart backend service"
        systemctl status tutoring-platform | tee -a "$LOG_FILE"
        return 1
    fi
}

################################################################################
# Check PostgreSQL
################################################################################

check_postgresql() {
    log "Checking PostgreSQL..."

    if ! systemctl is-active --quiet postgresql; then
        log_error "PostgreSQL is not running"
        return 1
    fi

    # Test database connection
    if sudo -u postgres psql -c "SELECT 1" tutoring_platform &> /dev/null; then
        log_success "PostgreSQL health check passed"
        return 0
    else
        log_error "PostgreSQL connection failed"
        return 1
    fi
}

restart_postgresql() {
    log "Restarting PostgreSQL..."
    systemctl restart postgresql

    # Wait for service to start
    sleep 5

    if check_postgresql; then
        log_success "PostgreSQL restarted successfully"
        return 0
    else
        log_error "Failed to restart PostgreSQL"
        systemctl status postgresql | tee -a "$LOG_FILE"
        return 1
    fi
}

################################################################################
# Check Nginx
################################################################################

check_nginx() {
    log "Checking Nginx..."

    if ! systemctl is-active --quiet nginx; then
        log_error "Nginx is not running"
        return 1
    fi

    # Test nginx configuration
    if nginx -t &> /dev/null; then
        log_success "Nginx health check passed"
        return 0
    else
        log_error "Nginx configuration is invalid"
        nginx -t 2>&1 | tee -a "$LOG_FILE"
        return 1
    fi
}

restart_nginx() {
    log "Restarting Nginx..."

    # Test configuration before restart
    if ! nginx -t &> /dev/null; then
        log_error "Cannot restart Nginx: configuration is invalid"
        nginx -t 2>&1 | tee -a "$LOG_FILE"
        return 1
    fi

    systemctl restart nginx

    # Wait for service to start
    sleep 2

    if check_nginx; then
        log_success "Nginx restarted successfully"
        return 0
    else
        log_error "Failed to restart Nginx"
        systemctl status nginx | tee -a "$LOG_FILE"
        return 1
    fi
}

################################################################################
# Check disk space
################################################################################

check_disk_space() {
    log "Checking disk space..."

    # Get disk usage percentage (excluding %)
    USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')

    if [ "$USAGE" -gt 90 ]; then
        log_error "Disk space critical: ${USAGE}% used"
        return 1
    elif [ "$USAGE" -gt 80 ]; then
        log_warning "Disk space warning: ${USAGE}% used"
        return 0
    else
        log_success "Disk space OK: ${USAGE}% used"
        return 0
    fi
}

################################################################################
# Check memory usage
################################################################################

check_memory() {
    log "Checking memory usage..."

    # Get memory usage percentage
    USAGE=$(free | awk 'NR==2 {printf "%.0f", $3*100/$2}')

    if [ "$USAGE" -gt 90 ]; then
        log_error "Memory usage critical: ${USAGE}%"
        return 1
    elif [ "$USAGE" -gt 80 ]; then
        log_warning "Memory usage warning: ${USAGE}%"
        return 0
    else
        log_success "Memory usage OK: ${USAGE}%"
        return 0
    fi
}

################################################################################
# Main execution
################################################################################

log "========== Health Check Started =========="

ALL_OK=true

# Check PostgreSQL
if ! check_postgresql; then
    ALL_OK=false
    restart_postgresql
fi

# Check Backend
if ! check_backend; then
    ALL_OK=false
    restart_backend
fi

# Check Nginx
if ! check_nginx; then
    ALL_OK=false
    restart_nginx
fi

# Check system resources
check_disk_space
check_memory

if [ "$ALL_OK" = true ]; then
    log_success "All health checks passed"
    log "========== Health Check Completed =========="
    exit 0
else
    log_warning "Some health checks failed or services were restarted"
    log "========== Health Check Completed with Issues =========="
    exit 1
fi
