# THE_BOT V3 Deployment Scripts Index

Complete guide to production deployment scripts for THE_BOT V3 (Native Systemd).

## Quick Start

```bash
# 1. Check if deployment is safe
./pre-deploy-check.sh

# 2. Deploy to production
./safe-deploy-native.sh

# 3. Verify deployment is successful
./verify-deployment.sh --full
```

## Scripts Overview

### 1. safe-deploy-native.sh (Main Deployment)
**Purpose:** Execute production deployment with 15-phase rollback capability

**Usage:**
```bash
# Standard deployment
./safe-deploy-native.sh

# Dry-run (test without making changes)
./safe-deploy-native.sh --dry-run --verbose

# Skip frontend build (for hotfixes)
./safe-deploy-native.sh --skip-frontend

# Deploy specific branch
./safe-deploy-native.sh --branch develop
```

**Options:**
- `--dry-run` - Simulate deployment without making changes
- `--verbose` - Enable verbose output
- `--no-rollback` - Disable automatic rollback on error
- `--skip-frontend` - Skip frontend build
- `--skip-backend` - Skip backend build
- `--branch <branch>` - Deploy from specific branch (default: master)
- `--help` - Show help

**What it does:**
1. Pre-deployment checks (SSH, disk, RAM)
2. Backup current state (git commit, code)
3. Check system resources (disk, RAM)
4. Fetch and update code (git pull)
5. Build backend (Go binary)
6. Build frontend (React bundle)
7. Run database migrations
8. Stop services gracefully
9. Deploy application files
10. Start services
11. Health checks
12. Get statistics
13. Deployment summary
14. Logging and cleanup

**Output:** Detailed logs saved to `logs/deploy_YYYYMMDD_HHMMSS.log`

---

### 2. verify-deployment.sh (Post-Deployment Verification)
**Purpose:** Verify deployment health and status

**Usage:**
```bash
# Quick status check
./verify-deployment.sh

# Comprehensive verification
./verify-deployment.sh --full

# Check services only
./verify-deployment.sh --services

# View service logs
./verify-deployment.sh --logs

# Check database
./verify-deployment.sh --database

# Check Redis
./verify-deployment.sh --redis

# Show statistics
./verify-deployment.sh --stats
```

**Checks:**
- Service status (running/stopped)
- HTTP endpoints
- Database connectivity
- Redis connectivity
- Deployment statistics
- Recent service logs

---

### 3. pre-deploy-check.sh (Pre-Deployment Validation)
**Purpose:** Verify system is ready for deployment

**Usage:**
```bash
# Run all checks
./pre-deploy-check.sh
```

**Checks:**
- SSH connectivity
- Project directory existence
- Disk space (must have <90% used)
- Available RAM (warning if <1GB)
- Go installation
- Node.js/npm installation
- Systemd services exist
- PostgreSQL connectivity
- Redis connectivity
- Git repository status
- Environment files (.env)
- Recent deployment history

**Exit codes:**
- `0` - All checks passed, safe to deploy
- `1` - Failed checks, do not deploy

---

### 4. manage-services.sh (Service Control)
**Purpose:** Control systemd services on production

**Usage:**
```bash
# Check status
./manage-services.sh status

# Start all services
./manage-services.sh start

# Stop all services
./manage-services.sh stop

# Restart all services
./manage-services.sh restart

# View backend logs
./manage-services.sh logs thebot-backend.service

# Tail all service logs (live)
./manage-services.sh tail

# Enable services on boot
./manage-services.sh enable

# Disable services on boot
./manage-services.sh disable
```

**Commands:**
- `status` - Show service status
- `start` - Start all services
- `stop` - Stop all services
- `restart` - Restart all services
- `logs <service>` - Show logs for service
- `tail` - Stream logs from all services
- `enable` - Enable services on boot
- `disable` - Disable services on boot

---

## Architecture

```
Production Server: mg@5.129.249.206 (the-bot.ru)
├── Backend (Go)
│   ├── Binary: /home/mg/the-bot/backend/server
│   └── Service: thebot-backend.service (port 8000)
│
├── Daphne (WebSocket/ASGI)
│   └── Service: thebot-daphne.service (port 8001)
│
├── Celery Worker (Task Processing)
│   └── Service: thebot-celery-worker.service
│
├── Celery Beat (Scheduler)
│   └── Service: thebot-celery-beat.service
│
├── PostgreSQL (Database)
│   └── localhost:5432 (thebot_db)
│
└── Redis (Cache/Broker)
    └── localhost:6379
```

## Deployment Workflow

### Phase 1: Pre-Check (Recommended)
```bash
./pre-deploy-check.sh
```
Validate all preconditions are met before deploying.

### Phase 2: Dry-Run (Recommended for First Deployment)
```bash
./safe-deploy-native.sh --dry-run --verbose
```
Test deployment without making changes. Identifies issues before actual deployment.

### Phase 3: Deployment
```bash
./safe-deploy-native.sh
```
Execute production deployment with automatic rollback on error.

### Phase 4: Verification
```bash
./verify-deployment.sh --full
```
Comprehensive health checks to ensure deployment was successful.

## Common Tasks

### Deploy Latest Changes
```bash
./pre-deploy-check.sh && ./safe-deploy-native.sh
```

### Quick Frontend Hotfix
```bash
./safe-deploy-native.sh --skip-backend
```

### Deploy from Develop Branch
```bash
./safe-deploy-native.sh --branch develop
```

### Check Service Health
```bash
./manage-services.sh status
./verify-deployment.sh --full
```

### View Recent Logs
```bash
./verify-deployment.sh --logs
./manage-services.sh logs thebot-backend.service
```

### Restart All Services
```bash
./manage-services.sh restart
```

### Emergency Rollback
```bash
# SSH to production
ssh mg@5.129.249.206

# List recent commits
cd /home/mg/the-bot
git log --oneline -n 10

# Rollback to previous commit
git reset --hard <previous-commit-hash>

# Restart services
systemctl restart thebot-backend.service thebot-daphne.service thebot-celery-worker.service thebot-celery-beat.service

# Verify
systemctl status thebot-*.service
```

## Environment Variables

### Remote SSH Configuration
```bash
export REMOTE_USER="mg"              # SSH user
export REMOTE_HOST="5.129.249.206"   # Production server IP
export THEBOT_HOME="/home/mg/the-bot" # Project path on remote
export GIT_BRANCH="master"           # Git branch to deploy
```

### Local Configuration
All scripts read from `.env` files on production server:
- `/home/mg/the-bot/.env` - Project environment
- `/home/mg/the-bot/backend/.env` - Backend environment

## Logs

### Deployment Logs
Located: `logs/deploy_YYYYMMDD_HHMMSS.log`

```bash
# View latest deployment
tail -100 logs/deploy_*.log | sort | tail -100

# Search for errors
grep "ERROR" logs/deploy_*.log

# View specific deployment
cat logs/deploy_20260118_145417.log
```

### Service Logs (on production)
```bash
# Backend
journalctl -u thebot-backend.service -f

# Daphne
journalctl -u thebot-daphne.service -f

# Celery Worker
journalctl -u thebot-celery-worker.service -f

# All services
journalctl -u thebot-*.service -f

# Recent logs
journalctl -u thebot-backend.service -n 50 --no-pager
```

## Error Handling

### Deployment Failures
1. Automatic rollback to previous commit (if enabled)
2. Services restarted from previous version
3. Error details logged to deployment log

### To Disable Auto-Rollback
```bash
./safe-deploy-native.sh --no-rollback
```

### Manual Rollback
```bash
ssh mg@5.129.249.206

cd /home/mg/the-bot
git reset --hard <previous-commit-hash>
systemctl restart thebot-*.service

# Verify
systemctl status thebot-*.service
```

## System Requirements

**Production Server:**
- OS: Linux (CachyOS)
- CPU: 4+ cores
- RAM: 2GB minimum (4GB recommended)
- Disk: 10GB+ free space
- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- Redis 6+

**Build Requirements:**
- Git
- Go compiler
- npm

## Performance Notes

**First Deployment:**
- Duration: 15-25 minutes
- Go compilation: 5-10 minutes
- npm install: 5-10 minutes
- Services startup: 1-2 minutes

**Subsequent Deployments:**
- Duration: 5-10 minutes
- Code compilation is faster
- npm install uses cache

## Security Notes

1. **SSH Key Authentication:** Uses SSH keys (no password required)
2. **Environment Variables:** .env files not committed to git
3. **Automatic Backups:** Before each deployment
4. **Rollback Capability:** Automatic on error
5. **Log Sanitization:** No secrets in logs

## Related Documentation

- **Native Deploy Guide:** NATIVE_DEPLOY.md
- **Chat Creation System:** CHAT_CREATION_README.md
- **Monitoring & Health Checks:** `deploy/monitoring/healthcheck.sh`
- **SSL/TLS Setup:** `deploy/ssl/certbot-setup.sh`

## Troubleshooting

### SSH Connection Failed
```bash
# Verify SSH access
ssh mg@5.129.249.206 "echo ok"

# Check SSH key
ssh-keygen -t ed25519 -C "deployment"
```

### Build Fails (Go not found)
```bash
# Install Go on production
ssh mg@5.129.249.206 "go version"  # Check if installed
```

### npm install Fails
```bash
# Check Node.js on production
ssh mg@5.129.249.206 "node --version && npm --version"
```

### Database Migration Fails
```bash
# Check PostgreSQL connectivity
./verify-deployment.sh --database

# View migration logs
./manage-services.sh logs thebot-backend.service
```

### Services Won't Start
```bash
# Check service status
./manage-services.sh status

# View error logs
./manage-services.sh logs thebot-backend.service

# Check .env file
ssh mg@5.129.249.206 "cat /home/mg/the-bot/.env"
```

## Deployment Checklist

Before each deployment:
- [ ] Code committed to git
- [ ] Run `./pre-deploy-check.sh`
- [ ] No critical issues found
- [ ] Backup recent database
- [ ] Have rollback plan ready

After each deployment:
- [ ] Run `./verify-deployment.sh --full`
- [ ] All services running
- [ ] Website loads correctly
- [ ] Check recent logs for errors
- [ ] Monitor for 5 minutes

## Support

For issues or questions:
1. Check deployment logs: `logs/deploy_*.log`
2. View service logs: `./manage-services.sh logs <service>`
3. Run verification: `./verify-deployment.sh --full`
4. Review error handling section above

---

**Last Updated:** 2026-01-18
**Deployment Method:** Native Systemd (No Docker)
**Target:** Production (mg@5.129.249.206)
**Status:** Ready for production use
