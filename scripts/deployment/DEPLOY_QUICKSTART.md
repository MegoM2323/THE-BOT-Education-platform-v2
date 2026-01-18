# Quick Start: Deploy THE_BOT V3 to Production

## The 3-Step Deployment

### Step 1: Pre-Check (1 minute)
```bash
./pre-deploy-check.sh
```
✓ SSH connection
✓ Disk space
✓ Build tools (Go, Node.js)
✓ Services configured

Exit status 0 = safe to deploy

### Step 2: Deploy (5-25 minutes)
```bash
./safe-deploy-native.sh
```

For first deployment or if unsure, use dry-run first:
```bash
./safe-deploy-native.sh --dry-run --verbose
```

### Step 3: Verify (2 minutes)
```bash
./verify-deployment.sh --full
```

✓ All services running
✓ Database accessible
✓ Redis accessible
✓ Endpoints responding

## Common Scenarios

### Scenario 1: Full Production Deployment
```bash
./pre-deploy-check.sh
./safe-deploy-native.sh --dry-run --verbose  # Test first
./safe-deploy-native.sh                      # Go live
./verify-deployment.sh --full                # Verify
```

### Scenario 2: Emergency Hotfix (Skip Frontend)
```bash
./safe-deploy-native.sh --skip-frontend
./verify-deployment.sh --services
```

### Scenario 3: Test Deployment (Dry-Run)
```bash
./safe-deploy-native.sh --dry-run --verbose
```

### Scenario 4: Deploy Different Branch
```bash
./safe-deploy-native.sh --branch develop
```

## What Gets Deployed

```
✓ Backend      : Go binary (thebot-backend.service)
✓ Frontend     : React bundle (thebot-daphne.service)
✓ Workers      : Celery workers (thebot-celery-worker.service)
✓ Scheduler    : Celery beat (thebot-celery-beat.service)
✓ Database     : Migrations applied
✓ Static files : Deployed to /home/mg/the-bot/static/
```

## Deployment Phases

| Phase | What | Time |
|-------|------|------|
| 1 | Pre-checks | 1-2 min |
| 2 | Backup state | 1-2 min |
| 3 | System checks | 1 min |
| 4 | Git pull | 1 min |
| 5 | Build backend | 5-10 min |
| 6 | Build frontend | 5-10 min |
| 7 | Migrations | 2-3 min |
| 8 | Stop services | 1 min |
| 9 | Deploy files | 1 min |
| 10 | Start services | 2 min |
| 11 | Health checks | 1 min |
| 12 | Statistics | 1 min |
| 13 | Summary | 1 min |

**Total: 5-25 minutes** (first deployment takes longer)

## Rollback on Error

If deployment fails:
1. Script automatically rolls back to previous commit
2. Previous version of services restarted
3. Error details logged to `logs/deploy_*.log`

To disable auto-rollback (not recommended):
```bash
./safe-deploy-native.sh --no-rollback
```

## Logs

### During Deployment
Logs are printed to terminal in real-time.

### After Deployment
```bash
# View deployment log
tail -50 logs/deploy_*.log

# Search for errors
grep ERROR logs/deploy_*.log

# View specific deployment
cat logs/deploy_20260118_145417.log
```

### Service Logs (on production)
```bash
# View recent logs
./manage-services.sh logs thebot-backend.service

# Stream live logs
./manage-services.sh tail

# Or manually
ssh mg@5.129.249.206 'journalctl -u thebot-backend -f'
```

## Status Commands

```bash
# Quick status
./verify-deployment.sh

# Full health check
./verify-deployment.sh --full

# Service status
./manage-services.sh status

# View logs
./verify-deployment.sh --logs

# Database stats
./verify-deployment.sh --database

# Redis health
./verify-deployment.sh --redis
```

## Options

### safe-deploy-native.sh
```bash
--dry-run              # Test without changes
--verbose              # Show all details
--no-rollback          # Disable auto-rollback
--skip-frontend        # Skip React build (hotfixes)
--skip-backend         # Skip Go build
--branch <name>        # Deploy different branch
--help                 # Show full help
```

### verify-deployment.sh
```bash
--full                 # Comprehensive checks
--services             # Service status only
--logs                 # Show service logs
--stats                # Deployment statistics
--database             # Database check
--redis                # Redis check
--help                 # Show full help
```

### manage-services.sh
```bash
status                 # Show service status
start                  # Start all services
stop                   # Stop all services
restart                # Restart all services
logs <service>         # View service logs
tail                   # Stream all logs
enable                 # Enable on boot
disable                # Disable on boot
help                   # Show full help
```

## Troubleshooting

### Pre-check Fails
```bash
./pre-deploy-check.sh
# Review failures and fix them before deploying
```

### Deployment Fails
1. Check logs: `tail logs/deploy_*.log`
2. Automatic rollback already happened
3. Check what went wrong: `./manage-services.sh logs thebot-backend.service`
4. Fix issue and retry

### Services Won't Start
```bash
./manage-services.sh status
./manage-services.sh logs thebot-backend.service
```

### Database Migration Error
```bash
./verify-deployment.sh --database
ssh mg@5.129.249.206 'cd /home/mg/the-bot/backend && psql $DATABASE_URL'
```

### Manual Rollback (Emergency Only)
```bash
ssh mg@5.129.249.206
cd /home/mg/the-bot
git log --oneline -n 5
git reset --hard <previous-commit-hash>
systemctl restart thebot-*.service
systemctl status thebot-*.service
```

## Safety Features

✓ **Pre-checks** - Validates system before deploying
✓ **Dry-run mode** - Test without making changes
✓ **Backups** - Code backed up before deployment
✓ **Auto-rollback** - Automatic rollback on error
✓ **Health checks** - Verifies services are running
✓ **Logging** - All actions logged to files
✓ **SSH security** - Key-based authentication only

## Performance Tips

1. **For speed:** Use `--skip-frontend` if only backend changed
2. **For safety:** Always use `--dry-run` first
3. **For debugging:** Use `--verbose` to see details
4. **For testing:** Use different branch with `--branch`

## Next Steps

1. Read full documentation: `DEPLOYMENT_INDEX.md`
2. Read detailed guide: `NATIVE_DEPLOY.md`
3. Check service management: `manage-services.sh --help`
4. Review chat creation: `CHAT_CREATION_README.md`

---

**Production Server:** mg@5.129.249.206 (the-bot.ru)
**Deployment Method:** Native Systemd (No Docker)
**Last Updated:** 2026-01-18
**Ready:** Yes
