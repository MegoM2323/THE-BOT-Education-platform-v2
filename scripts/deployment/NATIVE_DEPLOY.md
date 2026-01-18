# Production Deployment Guide (Native Systemd)

THE_BOT V3 production deployment использует **native systemd** вместо Docker для минимизации overhead и максимальной производительности.

## Architecture

```
Production Server: mg@5.129.249.206 (the-bot.ru)
├── Backend      : Go binary (thebot-backend.service)
├── Daphne       : WebSocket/ASGI server (thebot-daphne.service)
├── Celery Worker: Task processing (thebot-celery-worker.service)
├── Celery Beat  : Scheduler (thebot-celery-beat.service)
├── PostgreSQL   : Database (localhost:5432)
└── Redis        : Cache/Broker (localhost:6379)
```

## Quick Start

### 1. Standard Deployment

```bash
./safe-deploy-native.sh
```

This will:
- ✓ Check SSH connection to production server
- ✓ Backup current code state
- ✓ Pull latest code from git (master branch)
- ✓ Build Go backend binary
- ✓ Build React frontend bundle
- ✓ Run database migrations
- ✓ Restart all systemd services
- ✓ Verify services are running
- ✓ Show deployment statistics

### 2. Dry-Run Mode (Recommended First Step)

```bash
./safe-deploy-native.sh --dry-run --verbose
```

Simulates deployment without making any changes. Safe to run anytime.

### 3. Verify Deployment

After deployment completes, verify everything is working:

```bash
./verify-deployment.sh --full
```

## Deployment Options

### Standard Deployment
```bash
./safe-deploy-native.sh
```

### Dry-Run (Test Without Changes)
```bash
./safe-deploy-native.sh --dry-run --verbose
```

### Skip Frontend Build (Hotfixes)
```bash
./safe-deploy-native.sh --skip-frontend
```

### Skip Backend Build
```bash
./safe-deploy-native.sh --skip-backend
```

### Deploy Specific Branch
```bash
./safe-deploy-native.sh --branch develop
```

### Disable Auto-Rollback (Not Recommended)
```bash
./safe-deploy-native.sh --no-rollback
```

### Verbose Output
```bash
./safe-deploy-native.sh --verbose
```

## 15-Phase Deployment Process

### Phase 1: Pre-deployment checks
- SSH connectivity
- Project directory existence
- Service availability
- System resources

### Phase 2: Backup current state
- Git commit hash saved
- Code directory backed up to `/tmp/thebot_backup_*`

### Phase 3: Check disk space and RAM
- Verify sufficient disk space (must have >10% free)
- Check available RAM (warning if <512MB)

### Phase 4: Fetch and update code
- `git fetch origin`
- `git checkout master`
- `git pull origin master`

### Phase 5: Build backend (Go)
- `cd backend && go build -o server cmd/main.go`
- Verify binary is executable
- File type verification

### Phase 6: Build frontend (React)
- `npm ci --legacy-peer-deps`
- `npm run build`
- Verify dist/ directory created

### Phase 7: Database migrations
- PostgreSQL connectivity check
- Run migrations
- Verify migration 050 is applied

### Phase 8: Stop services gracefully
- systemctl stop thebot-backend
- systemctl stop thebot-daphne
- systemctl stop thebot-celery-worker
- systemctl stop thebot-celery-beat

### Phase 9: Deploy application files
- Set executable permissions
- Copy static files to `/home/mg/the-bot/static/`

### Phase 10: Start services
- systemctl start all services
- Wait for stabilization (5 seconds)

### Phase 11: Health checks
- Verify each service is running
- Check process status via systemctl

### Phase 12: Get statistics
- Chat count from database
- Recent backend service logs

### Phase 13: Deployment summary
- Show service status
- Provide next steps

### Phase 14-15: Logging and cleanup
- All actions logged to `logs/deploy_YYYYMMDD_HHMMSS.log`
- Automatic rollback on error (if enabled)

## Service Management

### View Service Status

```bash
# Quick status
./verify-deployment.sh

# Detailed status
./verify-deployment.sh --full

# Service-only check
./verify-deployment.sh --services
```

### Manual Service Management (on Production Server)

```bash
# Check all services
systemctl status thebot-*.service

# Restart backend
systemctl restart thebot-backend.service

# Restart all services
systemctl restart thebot-backend.service thebot-daphne.service thebot-celery-worker.service thebot-celery-beat.service

# View live logs
journalctl -u thebot-backend.service -f

# View recent logs (last 50 lines)
journalctl -u thebot-backend.service -n 50
```

### View Logs

```bash
# Recent deployment logs (on local machine)
cat logs/deploy_*.log | tail -100

# Recent service logs (on production)
ssh mg@5.129.249.206 'journalctl -u thebot-backend -n 50 --no-pager'

# All service logs in deployment verification
./verify-deployment.sh --logs
```

## Verification

### Check Health Endpoints

```bash
# Backend health check
ssh mg@5.129.249.206 'curl -s http://localhost:8000/health | jq'

# WebSocket endpoint
ssh mg@5.129.249.206 'curl -I -s http://localhost:8001/'
```

### Check Database

```bash
# View chat statistics
./verify-deployment.sh --database

# Or manually:
ssh mg@5.129.249.206 'cd /home/mg/the-bot/backend && psql $DATABASE_URL -c "SELECT COUNT(*) as total_chats FROM chats;"'
```

### Check Redis

```bash
# Redis health
./verify-deployment.sh --redis

# Or manually:
ssh mg@5.129.249.206 'redis-cli ping'
```

## Error Handling

### Automatic Rollback (Default Behavior)

If deployment fails, script automatically:
1. Rolls back git to previous commit: `git reset --hard <backup_commit>`
2. Restarts all services from previous version
3. Logs error details to `logs/deploy_*.log`

**To disable auto-rollback:**
```bash
./safe-deploy-native.sh --no-rollback
```

### Manual Rollback

If you need to manually rollback a failed deployment:

```bash
# On production server
cd /home/mg/the-bot
git log --oneline -n 5
git reset --hard <previous-commit-hash>
systemctl restart thebot-*.service

# Verify rollback
systemctl status thebot-*.service
```

### Common Issues

#### Build fails with "go: command not found"
- Go is not installed on production server
- Solution: Install Go on production server
- `ssh mg@5.129.249.206 'go version'`

#### npm build fails
- Node.js or npm not installed
- Solution: Install Node.js and npm on production server
- `ssh mg@5.129.249.206 'node -v && npm -v'`

#### Services won't start
- Check systemd service files exist
- View error logs: `journalctl -u thebot-backend -n 50`
- Check .env file permissions (must be 600)

#### Database migration fails
- Check PostgreSQL is running: `systemctl status postgresql`
- Verify DATABASE_URL in .env
- Check database user permissions

#### Insufficient disk space
- Free up space before deployment
- Check disk usage: `df -h /home/mg`
- Remove old backups: `rm -rf /tmp/thebot_backup_*`

## Environment Variables

Deployment script reads from production server's `/home/mg/the-bot/.env`

**Required variables for deployment:**
```bash
DATABASE_URL=postgresql://user:pass@localhost:5432/thebot_db
REDIS_URL=redis://localhost:6379
ENVIRONMENT=production
DJANGO_SECRET_KEY=<secret>
ALLOWED_HOSTS=5.129.249.206,the-bot.ru,www.the-bot.ru
```

## Deployment Timeline

**First deployment:**
- Pre-checks: 1-2 min
- Code build: 10-15 min (Go compilation + npm install)
- Migrations: 2-3 min
- Services restart: 1-2 min
- **Total: 15-25 minutes**

**Subsequent deployments:**
- All phases: 5-10 minutes
- (Most time spent on npm install for frontend)

## Systemd Service Files

Services are configured in `/etc/systemd/system/` on production server:

```
/etc/systemd/system/thebot-backend.service
/etc/systemd/system/thebot-daphne.service
/etc/systemd/system/thebot-celery-worker.service
/etc/systemd/system/thebot-celery-beat.service
```

**Note:** Service files should already be installed. If not, contact system administrator.

## Monitoring

### System Resources

```bash
# View system stats during deployment
ssh mg@5.129.249.206 'watch -n 1 "free -h && echo && ps aux | grep thebot"'
```

### Service Monitoring

```bash
# Monitor all services
watch -n 5 './verify-deployment.sh --services'
```

## Post-Deployment Checklist

- [ ] Run `./verify-deployment.sh --full`
- [ ] Check https://the-bot.ru loads correctly
- [ ] Verify chat creation works
- [ ] Check database statistics via `./verify-deployment.sh --database`
- [ ] Review recent logs: `./verify-deployment.sh --logs`
- [ ] Monitor for errors for 5 minutes after deployment

## Deployment Logs

All deployments are logged to `logs/deploy_YYYYMMDD_HHMMSS.log`

```bash
# View latest deployment log
tail -100 logs/deploy_*.log | sort | tail -100

# Search for errors in logs
grep "ERROR" logs/deploy_*.log
```

## Emergency Rollback

If something goes wrong in production:

```bash
# 1. SSH to production
ssh mg@5.129.249.206

# 2. Check git history
cd /home/mg/the-bot
git log --oneline -n 10

# 3. Rollback to previous commit
git reset --hard <previous-commit-hash>

# 4. Restart services
systemctl restart thebot-backend thebot-daphne thebot-celery-worker thebot-celery-beat

# 5. Verify
systemctl status thebot-*.service
journalctl -u thebot-backend -n 20 --no-pager
```

## Help

```bash
./safe-deploy-native.sh --help
./verify-deployment.sh --help
```

## Related Documentation

- **Chat Creation System:** See `CHAT_CREATION_README.md`
- **Monitoring Setup:** See `deploy/monitoring/healthcheck.sh`
- **SSL/TLS Management:** See `deploy/ssl/certbot-setup.sh`

---

**Last Updated:** 2026-01-18
**Deployment Method:** Native Systemd (No Docker)
**Target:** Production (mg@5.129.249.206)
