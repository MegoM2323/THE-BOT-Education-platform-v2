#!/bin/bash

################################################################################
# Database Backup Script
#
# This script creates automated backups of the Tutoring Platform database
# and manages backup retention
#
# Usage: ./backup.sh
# Recommended: Run via cron daily at 2:00 AM
################################################################################

set -e

# Configuration
BACKUP_DIR="/var/backups/tutoring-platform"
DB_NAME="tutoring_platform"
DB_USER="tutoring_user"
RETENTION_DAYS=30
LOG_FILE="/var/log/tutoring-platform/backup.log"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

# Error handler
error_handler() {
    log_error "Backup failed on line $1"
    exit 1
}

trap 'error_handler $LINENO' ERR

################################################################################
# Pre-flight checks
################################################################################

log "========== Backup Started =========="

# Create backup directory if it doesn't exist
if [ ! -d "$BACKUP_DIR" ]; then
    log_info "Creating backup directory: $BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
    chmod 700 "$BACKUP_DIR"
fi

# Check if PostgreSQL is running
if ! systemctl is-active --quiet postgresql; then
    log_error "PostgreSQL is not running"
    exit 1
fi

################################################################################
# Create backup
################################################################################

# Generate backup filename with timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/tutoring_platform_$TIMESTAMP.sql"
BACKUP_FILE_GZ="$BACKUP_FILE.gz"

log_info "Creating database backup: $BACKUP_FILE"

# Perform database dump
sudo -u postgres pg_dump \
    --dbname="$DB_NAME" \
    --username="$DB_USER" \
    --format=plain \
    --no-owner \
    --no-privileges \
    --verbose \
    --file="$BACKUP_FILE" \
    2>> "$LOG_FILE"

if [ ! -f "$BACKUP_FILE" ]; then
    log_error "Backup file was not created"
    exit 1
fi

# Get backup size
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log_info "Backup created: $BACKUP_SIZE"

################################################################################
# Compress backup
################################################################################

log_info "Compressing backup..."

gzip -f "$BACKUP_FILE"

if [ ! -f "$BACKUP_FILE_GZ" ]; then
    log_error "Failed to compress backup"
    exit 1
fi

# Get compressed size
COMPRESSED_SIZE=$(du -h "$BACKUP_FILE_GZ" | cut -f1)
log_info "Backup compressed: $COMPRESSED_SIZE"

# Set permissions
chmod 600 "$BACKUP_FILE_GZ"

################################################################################
# Verify backup
################################################################################

log_info "Verifying backup integrity..."

# Check if file is a valid gzip archive
if gzip -t "$BACKUP_FILE_GZ" 2>/dev/null; then
    log_success "Backup integrity verified"
else
    log_error "Backup file is corrupted"
    exit 1
fi

################################################################################
# Create backup manifest
################################################################################

MANIFEST_FILE="$BACKUP_DIR/backup_manifest.txt"

cat >> "$MANIFEST_FILE" <<EOF
Timestamp: $(date '+%Y-%m-%d %H:%M:%S')
Backup File: $(basename "$BACKUP_FILE_GZ")
Database: $DB_NAME
Original Size: $BACKUP_SIZE
Compressed Size: $COMPRESSED_SIZE
---
EOF

################################################################################
# Clean up old backups
################################################################################

log_info "Cleaning up old backups (retention: $RETENTION_DAYS days)..."

# Count backups before cleanup
BEFORE_COUNT=$(find "$BACKUP_DIR" -name "tutoring_platform_*.sql.gz" | wc -l)

# Delete backups older than retention period
find "$BACKUP_DIR" -name "tutoring_platform_*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete

# Count backups after cleanup
AFTER_COUNT=$(find "$BACKUP_DIR" -name "tutoring_platform_*.sql.gz" | wc -l)
DELETED_COUNT=$((BEFORE_COUNT - AFTER_COUNT))

if [ $DELETED_COUNT -gt 0 ]; then
    log_info "Deleted $DELETED_COUNT old backup(s)"
else
    log_info "No old backups to delete"
fi

################################################################################
# Display backup statistics
################################################################################

log_info "Current backup statistics:"
log_info "  Total backups: $AFTER_COUNT"
log_info "  Latest backup: $(basename "$BACKUP_FILE_GZ")"
log_info "  Backup size: $COMPRESSED_SIZE"

# Calculate total backup size
TOTAL_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
log_info "  Total backup directory size: $TOTAL_SIZE"

################################################################################
# Optional: Upload to remote storage
################################################################################

# Uncomment and configure to enable remote backup uploads
# Example: AWS S3
# if command -v aws &> /dev/null; then
#     log_info "Uploading backup to S3..."
#     aws s3 cp "$BACKUP_FILE_GZ" "s3://your-bucket/tutoring-platform/backups/"
#     log_success "Backup uploaded to S3"
# fi

# Example: rsync to remote server
# if command -v rsync &> /dev/null; then
#     log_info "Syncing backup to remote server..."
#     rsync -avz "$BACKUP_FILE_GZ" user@remote-server:/path/to/backups/
#     log_success "Backup synced to remote server"
# fi

################################################################################
# Final summary
################################################################################

log_success "Backup completed successfully"
log "========== Backup Finished =========="

# Exit successfully
exit 0
