#!/bin/bash

################################################################################
# System Status Script
#
# This script displays the current status of all Tutoring Platform
# services and resources
#
# Usage: ./status.sh
################################################################################

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo ""
    echo -e "${CYAN}================================================================================"
    echo -e "  $1"
    echo -e "================================================================================${NC}"
    echo ""
}

print_status() {
    STATUS=$1
    MESSAGE=$2

    if [ "$STATUS" = "ok" ]; then
        echo -e "${GREEN}✓${NC} $MESSAGE"
    elif [ "$STATUS" = "warning" ]; then
        echo -e "${YELLOW}⚠${NC} $MESSAGE"
    elif [ "$STATUS" = "error" ]; then
        echo -e "${RED}✗${NC} $MESSAGE"
    else
        echo -e "${BLUE}ℹ${NC} $MESSAGE"
    fi
}

################################################################################
# System Information
################################################################################

print_header "System Information"

echo -e "${BLUE}Hostname:${NC} $(hostname)"
echo -e "${BLUE}OS:${NC} $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo -e "${BLUE}Kernel:${NC} $(uname -r)"
echo -e "${BLUE}Uptime:${NC} $(uptime -p)"
echo -e "${BLUE}Date:${NC} $(date)"

################################################################################
# Resource Usage
################################################################################

print_header "Resource Usage"

# CPU
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
if (( $(echo "$CPU_USAGE > 80" | bc -l) )); then
    print_status "warning" "CPU Usage: ${CPU_USAGE}%"
else
    print_status "ok" "CPU Usage: ${CPU_USAGE}%"
fi

# Memory
MEMORY_USAGE=$(free | awk '/Mem/ {printf "%.0f", $3/$2 * 100}')
MEMORY_TOTAL=$(free -h | awk '/Mem/ {print $2}')
MEMORY_USED=$(free -h | awk '/Mem/ {print $3}')

if [ "$MEMORY_USAGE" -gt 80 ]; then
    print_status "warning" "Memory: ${MEMORY_USED}/${MEMORY_TOTAL} (${MEMORY_USAGE}%)"
else
    print_status "ok" "Memory: ${MEMORY_USED}/${MEMORY_TOTAL} (${MEMORY_USAGE}%)"
fi

# Disk
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
DISK_TOTAL=$(df -h / | awk 'NR==2 {print $2}')
DISK_USED=$(df -h / | awk 'NR==2 {print $3}')
DISK_AVAIL=$(df -h / | awk 'NR==2 {print $4}')

if [ "$DISK_USAGE" -gt 80 ]; then
    print_status "warning" "Disk: ${DISK_USED}/${DISK_TOTAL} used, ${DISK_AVAIL} available (${DISK_USAGE}%)"
else
    print_status "ok" "Disk: ${DISK_USED}/${DISK_TOTAL} used, ${DISK_AVAIL} available (${DISK_USAGE}%)"
fi

################################################################################
# Services Status
################################################################################

print_header "Services Status"

check_service() {
    SERVICE=$1
    LABEL=$2

    if systemctl is-active --quiet "$SERVICE"; then
        UPTIME=$(systemctl show "$SERVICE" --property=ActiveEnterTimestamp | cut -d'=' -f2)
        print_status "ok" "$LABEL: Running (since $UPTIME)"
    else
        print_status "error" "$LABEL: Not running"
    fi
}

check_service "tutoring-platform" "Backend Service"
check_service "nginx" "Nginx"
check_service "postgresql" "PostgreSQL"

# Check SSL certificate status
if systemctl is-active --quiet certbot.timer; then
    print_status "ok" "Certbot Timer: Active"
else
    print_status "warning" "Certbot Timer: Not active"
fi

################################################################################
# Application Health
################################################################################

print_header "Application Health"

# Backend health check
if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
    RESPONSE=$(curl -s http://localhost:8080/health)
    print_status "ok" "Backend Health: $RESPONSE"
else
    print_status "error" "Backend Health: Not responding"
fi

# Nginx health check
if curl -f -s http://localhost > /dev/null 2>&1; then
    print_status "ok" "Nginx: Responding"
else
    print_status "error" "Nginx: Not responding"
fi

# Database health check
if sudo -u postgres psql -c "SELECT 1" tutoring_platform > /dev/null 2>&1; then
    DB_SIZE=$(sudo -u postgres psql -t -c "SELECT pg_size_pretty(pg_database_size('tutoring_platform'));" | xargs)
    print_status "ok" "Database: Connected (Size: $DB_SIZE)"
else
    print_status "error" "Database: Connection failed"
fi

################################################################################
# SSL Certificate
################################################################################

print_header "SSL Certificate"

if command -v certbot &> /dev/null; then
    CERT_INFO=$(certbot certificates 2>/dev/null | grep -A 3 "Certificate Name")

    if [ -n "$CERT_INFO" ]; then
        DOMAIN=$(echo "$CERT_INFO" | grep "Domains:" | awk '{print $2}')
        EXPIRY=$(echo "$CERT_INFO" | grep "Expiry Date:" | cut -d':' -f2-)

        if [ -n "$DOMAIN" ]; then
            print_status "ok" "Domain: $DOMAIN"
            print_status "info" "Expires: $EXPIRY"
        else
            print_status "warning" "SSL Certificate: Not configured"
        fi
    else
        print_status "warning" "No SSL certificates found"
    fi
else
    print_status "warning" "Certbot not installed"
fi

################################################################################
# Network & Ports
################################################################################

print_header "Network & Ports"

check_port() {
    PORT=$1
    SERVICE=$2

    if ss -tuln | grep -q ":$PORT "; then
        print_status "ok" "Port $PORT ($SERVICE): Listening"
    else
        print_status "error" "Port $PORT ($SERVICE): Not listening"
    fi
}

check_port 80 "HTTP"
check_port 443 "HTTPS"
check_port 8080 "Backend"
check_port 5432 "PostgreSQL"

# Firewall status
if command -v ufw &> /dev/null; then
    UFW_STATUS=$(ufw status | head -1 | awk '{print $2}')
    if [ "$UFW_STATUS" = "active" ]; then
        print_status "ok" "Firewall: Active"
    else
        print_status "warning" "Firewall: Inactive"
    fi
fi

################################################################################
# Backups
################################################################################

print_header "Backups"

BACKUP_DIR="/var/backups/tutoring-platform"

if [ -d "$BACKUP_DIR" ]; then
    BACKUP_COUNT=$(find "$BACKUP_DIR" -name "*.sql.gz" | wc -l)
    LATEST_BACKUP=$(find "$BACKUP_DIR" -name "*.sql.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)

    if [ -n "$LATEST_BACKUP" ]; then
        BACKUP_DATE=$(stat -c %y "$LATEST_BACKUP" | cut -d' ' -f1,2 | cut -d'.' -f1)
        BACKUP_SIZE=$(du -h "$LATEST_BACKUP" | cut -f1)
        print_status "ok" "Total Backups: $BACKUP_COUNT"
        print_status "ok" "Latest Backup: $BACKUP_DATE ($BACKUP_SIZE)"
    else
        print_status "warning" "No backups found"
    fi

    TOTAL_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
    print_status "info" "Total Backup Size: $TOTAL_SIZE"
else
    print_status "warning" "Backup directory not found"
fi

################################################################################
# Monitoring
################################################################################

print_header "Monitoring Status"

# Check cron jobs
CRON_COUNT=$(crontab -l 2>/dev/null | grep -c "tutoring-platform" || echo "0")
if [ "$CRON_COUNT" -gt 0 ]; then
    print_status "ok" "Cron Jobs: $CRON_COUNT configured"
else
    print_status "warning" "No cron jobs found"
fi

# Check log files
LOG_DIR="/var/log/tutoring-platform"

if [ -d "$LOG_DIR" ]; then
    print_status "ok" "Log Directory: $LOG_DIR"

    if [ -f "$LOG_DIR/healthcheck.log" ]; then
        LAST_CHECK=$(tail -1 "$LOG_DIR/healthcheck.log" | grep -oP '^\[\K[^\]]+')
        if [ -n "$LAST_CHECK" ]; then
            print_status "ok" "Last Health Check: $LAST_CHECK"
        fi
    fi

    if [ -f "$LOG_DIR/backup.log" ]; then
        LAST_BACKUP=$(grep "Backup completed" "$LOG_DIR/backup.log" | tail -1 | grep -oP '^\[\K[^\]]+')
        if [ -n "$LAST_BACKUP" ]; then
            print_status "ok" "Last Backup Run: $LAST_BACKUP"
        fi
    fi
fi

################################################################################
# Recent Logs
################################################################################

print_header "Recent Logs (Last 5 Lines)"

if [ -f "/var/log/tutoring-platform/error.log" ]; then
    echo -e "${BLUE}Error Log:${NC}"
    tail -5 /var/log/tutoring-platform/error.log 2>/dev/null | sed 's/^/  /' || echo "  (empty)"
    echo ""
fi

if [ -f "/var/log/nginx/error.log" ]; then
    echo -e "${BLUE}Nginx Error Log:${NC}"
    tail -5 /var/log/nginx/error.log 2>/dev/null | sed 's/^/  /' || echo "  (empty)"
    echo ""
fi

################################################################################
# Process Information
################################################################################

print_header "Process Information"

# Backend process
if pgrep -f "tutoring-platform" > /dev/null; then
    PID=$(pgrep -f "tutoring-platform" | head -1)
    MEM=$(ps -p "$PID" -o %mem --no-headers | xargs)
    CPU=$(ps -p "$PID" -o %cpu --no-headers | xargs)
    print_status "ok" "Backend Process (PID: $PID, CPU: ${CPU}%, MEM: ${MEM}%)"
else
    print_status "error" "Backend Process: Not found"
fi

# PostgreSQL connections
if systemctl is-active --quiet postgresql; then
    CONN_COUNT=$(sudo -u postgres psql -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname='tutoring_platform';" | xargs)
    print_status "info" "Database Connections: $CONN_COUNT"
fi

################################################################################
# Summary
################################################################################

print_header "Summary"

# Count issues
SERVICES_OK=0
SERVICES_ERROR=0

systemctl is-active --quiet tutoring-platform && ((SERVICES_OK++)) || ((SERVICES_ERROR++))
systemctl is-active --quiet nginx && ((SERVICES_OK++)) || ((SERVICES_ERROR++))
systemctl is-active --quiet postgresql && ((SERVICES_OK++)) || ((SERVICES_ERROR++))

if [ "$SERVICES_ERROR" -eq 0 ]; then
    print_status "ok" "All critical services running ($SERVICES_OK/3)"
    echo ""
    echo -e "${GREEN}System Status: HEALTHY${NC}"
else
    print_status "error" "Some services not running ($SERVICES_OK/3)"
    echo ""
    echo -e "${RED}System Status: DEGRADED${NC}"
fi

echo ""
echo "For detailed logs, use:"
echo "  - Application: tail -f /var/log/tutoring-platform/access.log"
echo "  - Errors: tail -f /var/log/tutoring-platform/error.log"
echo "  - Health checks: tail -f /var/log/tutoring-platform/healthcheck.log"
echo ""
echo "================================================================================"
echo ""
