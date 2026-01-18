#!/bin/bash

################################################################################
# Monitoring Setup Script
#
# This script configures automated monitoring and maintenance tasks
# via cron jobs
#
# Usage: sudo ./setup-monitoring.sh
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

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (use sudo)"
    exit 1
fi

log_info "Setting up monitoring and automation..."

################################################################################
# Configuration
################################################################################

APP_DIR="/opt/tutoring-platform"
HEALTHCHECK_SCRIPT="$APP_DIR/healthcheck.sh"
BACKUP_SCRIPT="$APP_DIR/backup.sh"
LOG_DIR="/var/log/tutoring-platform"

################################################################################
# Verify scripts exist
################################################################################

log_info "Verifying monitoring scripts..."

if [ ! -f "$HEALTHCHECK_SCRIPT" ]; then
    log_error "Health check script not found: $HEALTHCHECK_SCRIPT"
    exit 1
fi

if [ ! -f "$BACKUP_SCRIPT" ]; then
    log_error "Backup script not found: $BACKUP_SCRIPT"
    exit 1
fi

# Make sure scripts are executable
chmod +x "$HEALTHCHECK_SCRIPT"
chmod +x "$BACKUP_SCRIPT"

log_success "Scripts verified and made executable"

################################################################################
# Create log directory
################################################################################

if [ ! -d "$LOG_DIR" ]; then
    log_info "Creating log directory: $LOG_DIR"
    mkdir -p "$LOG_DIR"
    chown www-data:www-data "$LOG_DIR"
fi

################################################################################
# Setup cron jobs
################################################################################

log_info "Setting up cron jobs..."

# Create temporary crontab file
TEMP_CRON=$(mktemp)

# Get existing crontab (if any)
crontab -l > "$TEMP_CRON" 2>/dev/null || true

# Remove old tutoring-platform cron jobs (if any)
sed -i '/tutoring-platform/d' "$TEMP_CRON"

# Add new cron jobs
cat >> "$TEMP_CRON" <<EOF

# Tutoring Platform - Health Check (every 5 minutes)
*/5 * * * * $HEALTHCHECK_SCRIPT >> $LOG_DIR/healthcheck.log 2>&1

# Tutoring Platform - Database Backup (daily at 2:00 AM)
0 2 * * * $BACKUP_SCRIPT >> $LOG_DIR/backup.log 2>&1

# Tutoring Platform - Log Rotation (daily at 3:00 AM)
0 3 * * * find $LOG_DIR -name "*.log" -type f -size +100M -exec truncate -s 10M {} \;

# Tutoring Platform - Cleanup old logs (weekly on Sunday at 4:00 AM)
0 4 * * 0 find $LOG_DIR -name "*.log.*" -type f -mtime +30 -delete
EOF

# Install new crontab
crontab "$TEMP_CRON"
rm "$TEMP_CRON"

log_success "Cron jobs configured"

################################################################################
# Display configured jobs
################################################################################

log_info "Configured cron jobs:"
echo ""
crontab -l | grep -A 10 "Tutoring Platform"
echo ""

################################################################################
# Setup systemd timer (alternative to cron)
################################################################################

log_info "Setting up systemd timers as alternative to cron..."

# Health check timer
cat > /etc/systemd/system/tutoring-healthcheck.timer <<EOF
[Unit]
Description=Tutoring Platform Health Check Timer
Requires=tutoring-healthcheck.service

[Timer]
OnBootSec=5min
OnUnitActiveSec=5min
Unit=tutoring-healthcheck.service

[Install]
WantedBy=timers.target
EOF

# Health check service
cat > /etc/systemd/system/tutoring-healthcheck.service <<EOF
[Unit]
Description=Tutoring Platform Health Check Service
After=network.target

[Service]
Type=oneshot
ExecStart=$HEALTHCHECK_SCRIPT
StandardOutput=append:$LOG_DIR/healthcheck.log
StandardError=append:$LOG_DIR/healthcheck.log
EOF

# Backup timer
cat > /etc/systemd/system/tutoring-backup.timer <<EOF
[Unit]
Description=Tutoring Platform Backup Timer
Requires=tutoring-backup.service

[Timer]
OnCalendar=daily
OnCalendar=02:00
Persistent=true
Unit=tutoring-backup.service

[Install]
WantedBy=timers.target
EOF

# Backup service
cat > /etc/systemd/system/tutoring-backup.service <<EOF
[Unit]
Description=Tutoring Platform Backup Service
After=network.target postgresql.service

[Service]
Type=oneshot
ExecStart=$BACKUP_SCRIPT
StandardOutput=append:$LOG_DIR/backup.log
StandardError=append:$LOG_DIR/backup.log
EOF

# Reload systemd
systemctl daemon-reload

log_info "Systemd timers created but not enabled (using cron by default)"
log_info "To use systemd timers instead of cron, run:"
echo "  systemctl enable --now tutoring-healthcheck.timer"
echo "  systemctl enable --now tutoring-backup.timer"

################################################################################
# Setup logrotate
################################################################################

log_info "Setting up log rotation with logrotate..."

cat > /etc/logrotate.d/tutoring-platform <<EOF
$LOG_DIR/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    missingok
    create 0644 www-data www-data
    sharedscripts
    postrotate
        systemctl reload tutoring-platform > /dev/null 2>&1 || true
    endscript
}

/var/log/nginx/tutoring-platform*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    missingok
    create 0644 www-data www-data
    sharedscripts
    postrotate
        systemctl reload nginx > /dev/null 2>&1 || true
    endscript
}
EOF

log_success "Logrotate configured"

################################################################################
# Test monitoring scripts
################################################################################

log_info "Testing monitoring scripts..."

# Test healthcheck
log_info "Running health check..."
bash "$HEALTHCHECK_SCRIPT" || log_warning "Health check reported issues"

log_success "Health check test completed"

################################################################################
# Setup email notifications (optional)
################################################################################

log_info "Setting up email notifications..."

if command -v mail &> /dev/null; then
    log_info "Mail command found, notifications can be enabled"

    # Create notification script
    NOTIFY_SCRIPT="$APP_DIR/notify-admin.sh"

    cat > "$NOTIFY_SCRIPT" <<'EOF'
#!/bin/bash
# Email notification script
# Usage: ./notify-admin.sh "subject" "message"

ADMIN_EMAIL="${ADMIN_EMAIL:-root@localhost}"
SUBJECT="$1"
MESSAGE="$2"

if [ -z "$SUBJECT" ] || [ -z "$MESSAGE" ]; then
    echo "Usage: $0 'subject' 'message'"
    exit 1
fi

echo "$MESSAGE" | mail -s "[Tutoring Platform] $SUBJECT" "$ADMIN_EMAIL"
EOF

    chmod +x "$NOTIFY_SCRIPT"

    log_success "Email notification script created at $NOTIFY_SCRIPT"
    log_info "Set ADMIN_EMAIL environment variable to enable notifications"
else
    log_warning "Mail command not found. Install mailutils for email notifications:"
    log_info "  sudo apt install mailutils"
fi

################################################################################
# Summary
################################################################################

echo ""
echo "================================================================================"
echo -e "${GREEN}Monitoring setup completed!${NC}"
echo "================================================================================"
echo ""
echo "Configured tasks:"
echo "  - Health checks: Every 5 minutes"
echo "  - Database backups: Daily at 2:00 AM"
echo "  - Log rotation: Daily at 3:00 AM"
echo "  - Log cleanup: Weekly on Sundays"
echo ""
echo "Log files:"
echo "  - Health check: $LOG_DIR/healthcheck.log"
echo "  - Backups: $LOG_DIR/backup.log"
echo "  - Application: $LOG_DIR/access.log and $LOG_DIR/error.log"
echo ""
echo "Management commands:"
echo "  - View cron jobs: crontab -l"
echo "  - Edit cron jobs: crontab -e"
echo "  - View systemd timers: systemctl list-timers"
echo ""
echo "Manual execution:"
echo "  - Health check: sudo $HEALTHCHECK_SCRIPT"
echo "  - Backup: sudo $BACKUP_SCRIPT"
echo ""
echo "To view recent logs:"
echo "  tail -f $LOG_DIR/healthcheck.log"
echo "  tail -f $LOG_DIR/backup.log"
echo ""
echo "================================================================================"
echo ""

log_success "Monitoring setup complete!"
