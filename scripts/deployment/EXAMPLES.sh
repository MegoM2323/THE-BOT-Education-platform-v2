#!/bin/bash

# EXAMPLES OF COMMON DEPLOYMENT SCENARIOS
# ========================================

# NOTE: These are documentation examples. Do not execute this file directly.
# Copy individual commands to use them.

# ============================================================================
# EXAMPLE 1: STANDARD PRODUCTION DEPLOYMENT
# ============================================================================

echo "=== EXAMPLE 1: Standard Production Deployment ==="

# Step 1: Pre-check system
./pre-deploy-check.sh

# Step 2: Deploy (takes 5-25 minutes)
./safe-deploy-native.sh

# Step 3: Verify deployment
./verify-deployment.sh --full


# ============================================================================
# EXAMPLE 2: SAFE DEPLOYMENT WITH DRY-RUN
# ============================================================================

echo "=== EXAMPLE 2: Test Deployment (Dry-Run, No Changes) ==="

# Simulate deployment without making any changes
./safe-deploy-native.sh --dry-run --verbose

# This shows exactly what WILL happen without affecting production


# ============================================================================
# EXAMPLE 3: EMERGENCY HOTFIX (SKIP FRONTEND)
# ============================================================================

echo "=== EXAMPLE 3: Quick Hotfix (Skip Frontend Build) ==="

# Deploy only backend (Go) - useful for critical bug fixes
./safe-deploy-native.sh --skip-frontend

# Verify
./verify-deployment.sh --services


# ============================================================================
# EXAMPLE 4: DEPLOY FROM DIFFERENT BRANCH
# ============================================================================

echo "=== EXAMPLE 4: Deploy from Develop Branch ==="

# Deploy from develop instead of master
./safe-deploy-native.sh --branch develop

# Verify
./verify-deployment.sh --full


# ============================================================================
# EXAMPLE 5: CHECK SYSTEM BEFORE DEPLOYMENT
# ============================================================================

echo "=== EXAMPLE 5: Pre-Deployment Validation ==="

# Full system check
./pre-deploy-check.sh

# If exit code is 0, safe to deploy
if [ $? -eq 0 ]; then
    echo "System is ready. Proceeding with deployment..."
    ./safe-deploy-native.sh
else
    echo "System issues found. Please fix them first."
fi


# ============================================================================
# EXAMPLE 6: MONITOR DEPLOYMENT STATUS
# ============================================================================

echo "=== EXAMPLE 6: Monitor Services During Deployment ==="

# Watch service status
./manage-services.sh status

# Watch live logs
./manage-services.sh tail

# Or check specific service
./manage-services.sh logs thebot-backend.service


# ============================================================================
# EXAMPLE 7: POST-DEPLOYMENT VERIFICATION
# ============================================================================

echo "=== EXAMPLE 7: Verify Deployment Success ==="

# Quick status
./verify-deployment.sh

# Full health check
./verify-deployment.sh --full

# Check database
./verify-deployment.sh --database

# Check Redis
./verify-deployment.sh --redis

# View logs
./verify-deployment.sh --logs


# ============================================================================
# EXAMPLE 8: SERVICE MANAGEMENT
# ============================================================================

echo "=== EXAMPLE 8: Service Control ==="

# Check all services
./manage-services.sh status

# Restart all services
./manage-services.sh restart

# View backend logs
./manage-services.sh logs thebot-backend.service

# Stream all logs
./manage-services.sh tail

# Start services on boot
./manage-services.sh enable


# ============================================================================
# EXAMPLE 9: CONTINUOUS MONITORING
# ============================================================================

echo "=== EXAMPLE 9: Monitor Deployment Over Time ==="

# Check status every 10 seconds (Ctrl+C to stop)
while true; do
    clear
    echo "=== Service Status at $(date) ==="
    ./manage-services.sh status
    sleep 10
done


# ============================================================================
# EXAMPLE 10: COMPLETE DEPLOYMENT WORKFLOW
# ============================================================================

echo "=== EXAMPLE 10: Complete Safe Workflow ==="

set -e  # Exit on error

echo "Step 1: Validate system..."
./pre-deploy-check.sh || exit 1

echo "Step 2: Test deployment (dry-run)..."
./safe-deploy-native.sh --dry-run --verbose || exit 1

read -p "Ready to deploy? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Step 3: Deploy to production..."
    ./safe-deploy-native.sh || exit 1
    
    echo "Step 4: Verify deployment..."
    ./verify-deployment.sh --full || exit 1
    
    echo "Step 5: Monitor for issues..."
    for i in {1..5}; do
        echo "Check $i/5..."
        ./manage-services.sh status
        sleep 5
    done
    
    echo "Deployment complete and verified!"
else
    echo "Deployment cancelled."
fi


# ============================================================================
# EXAMPLE 11: DEPLOYMENT WITH CUSTOM BRANCH
# ============================================================================

echo "=== EXAMPLE 11: Deploy Custom Branch with Verification ==="

BRANCH="feature/new-feature"

echo "Deploying branch: $BRANCH"

# Validate first
./pre-deploy-check.sh || exit 1

# Dry-run with verbose output
echo "Testing deployment (dry-run)..."
./safe-deploy-native.sh --branch "$BRANCH" --dry-run --verbose || exit 1

# Actually deploy
echo "Deploying..."
./safe-deploy-native.sh --branch "$BRANCH" || exit 1

# Verify
echo "Verifying..."
./verify-deployment.sh --full


# ============================================================================
# EXAMPLE 12: ROLLBACK AFTER FAILED DEPLOYMENT
# ============================================================================

echo "=== EXAMPLE 12: Manual Rollback (If Needed) ==="

# If auto-rollback didn't work, do it manually:

ssh mg@5.129.249.206 << 'SSH_COMMANDS'
    echo "Checking git log..."
    cd /home/mg/the-bot
    git log --oneline -n 5
    
    echo "Rolling back..."
    git reset --hard HEAD~1
    
    echo "Restarting services..."
    systemctl restart thebot-backend.service
    systemctl restart thebot-daphne.service
    systemctl restart thebot-celery-worker.service
    systemctl restart thebot-celery-beat.service
    
    echo "Verifying rollback..."
    systemctl status thebot-*.service
SSH_COMMANDS


# ============================================================================
# EXAMPLE 13: DEPLOYMENT WITH LOGGING
# ============================================================================

echo "=== EXAMPLE 13: Deployment with Detailed Logging ==="

# Deploy with verbose output and save log
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOGFILE="deployment_${TIMESTAMP}.log"

./safe-deploy-native.sh --verbose 2>&1 | tee "$LOGFILE"

echo "Deployment log saved to: $LOGFILE"

# Search for errors
if grep -q ERROR "$LOGFILE"; then
    echo "Errors found in deployment!"
    grep ERROR "$LOGFILE"
else
    echo "No errors in deployment."
fi


# ============================================================================
# EXAMPLE 14: HEALTH CHECK AFTER DEPLOYMENT
# ============================================================================

echo "=== EXAMPLE 14: Comprehensive Health Check ==="

echo "1. Service status..."
./manage-services.sh status

echo "2. Health endpoints..."
./verify-deployment.sh --full

echo "3. Database status..."
./verify-deployment.sh --database

echo "4. Redis status..."
./verify-deployment.sh --redis

echo "5. Deployment statistics..."
./verify-deployment.sh --stats


# ============================================================================
# EXAMPLE 15: CRON JOB FOR PERIODIC DEPLOYMENT CHECK
# ============================================================================

echo "=== EXAMPLE 15: Periodic Health Check (via cron) ==="

# Add to crontab: crontab -e
# Run health check every hour:
# 0 * * * * cd /home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment && ./verify-deployment.sh --stats >> health-check.log 2>&1

# View logs:
# tail -100 health-check.log


# ============================================================================
# COMMON ISSUES AND SOLUTIONS
# ============================================================================

# Issue: "Cannot connect to production server"
# Solution: Check SSH key and connectivity
ssh mg@5.129.249.206 "echo ok"

# Issue: "Disk space insufficient"
# Solution: Clean up old backups
ssh mg@5.129.249.206 "rm -rf /tmp/thebot_backup_* && df -h /home/mg"

# Issue: "Services won't start"
# Solution: Check logs
./manage-services.sh logs thebot-backend.service

# Issue: "Database migration fails"
# Solution: Check database
./verify-deployment.sh --database

# Issue: "Redis connection error"
# Solution: Check Redis
./verify-deployment.sh --redis


