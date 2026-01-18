# Chat Creation System - Deployment Scripts

Production-ready scripts for managing the automated chat room creation system.

## Quick Start (30 seconds)

```bash
# Check if system is working
./verify-chat-creation.sh --check-only

# Setup automatic backfill (production)
sudo ./setup-chat-creation.sh --timer
```

## Files

| File | Purpose | Usage |
|------|---------|-------|
| **verify-chat-creation.sh** | Main verification & backfill | `./verify-chat-creation.sh --check-only` |
| **chat-creation-cron.sh** | Cron job script | Use with crontab or systemd |
| **setup-chat-creation.sh** | Installation script | `sudo ./setup-chat-creation.sh --timer` |
| **thebot-chat-creation.service** | Systemd service unit | Install to `/etc/systemd/system/` |
| **thebot-chat-creation.timer** | Systemd timer unit | Install to `/etc/systemd/system/` |

## Documentation

| File | For |
|------|-----|
| **QUICKSTART.md** | Quick reference (start here) |
| **CHAT_CREATION_README.md** | Full documentation & troubleshooting |
| **EXAMPLES.md** | Usage examples & scenarios |
| **INDEX.md** | File reference & structure |
| **MANIFEST.txt** | Complete file inventory |

## What It Does

Automatically creates chat rooms between teachers and students after lessons are completed.

**Real-time**: Database trigger fires when lesson ends
**Periodic**: Backfill function runs every 30 minutes for recovery/catch-up

## Installation

### Option 1: Systemd Timer (Recommended)
```bash
sudo ./setup-chat-creation.sh --timer
systemctl status thebot-chat-creation.timer
```

### Option 2: Crontab
```bash
sudo ./setup-chat-creation.sh --cron
crontab -l
```

### Option 3: Manual
```bash
./verify-chat-creation.sh --backfill
```

## Common Commands

```bash
# Check status (local)
./verify-chat-creation.sh --check-only

# Check production
DB_HOST=5.129.249.206 DB_NAME=thebot_db \
./verify-chat-creation.sh --check-only

# Create missing chats
./verify-chat-creation.sh --backfill

# Monitor (systemd)
journalctl -u thebot-chat-creation.service -f

# Monitor (cron)
tail -f /var/log/thebot/chat-creation-cron.log
```

## Requirements

- PostgreSQL 12+
- Bash 4.0+
- `psql` client
- Database migration: `050_chat_on_booking_creation.sql`

## Documentation Links

- Start here: **QUICKSTART.md**
- Full guide: **CHAT_CREATION_README.md**
- Examples: **EXAMPLES.md**
- Reference: **INDEX.md** or **MANIFEST.txt**

## Environment Variables

```bash
DB_HOST=localhost              # PostgreSQL server
DB_PORT=5432                   # PostgreSQL port
DB_NAME=tutoring_platform      # Database name
DB_USER=postgres               # Database user
DB_PASSWORD=                   # Database password (if needed)
LOG_DIR=/var/log/thebot        # Log directory for cron jobs
```

## Troubleshooting

### Trigger not found
```bash
cd /home/mego/Python\ Projects/THE_BOT_V3/backend
./scripts/migrate.sh
```

### Connection refused
```bash
# Test connection
psql -h 5.129.249.206 -U postgres -d thebot_db -c "SELECT 1;"
```

### psql command not found
```bash
# Ubuntu/Debian
sudo apt-get install postgresql-client

# CentOS/RHEL
sudo yum install postgresql
```

See **CHAT_CREATION_README.md** for detailed troubleshooting.

## Performance

- Backfill: < 1 second for 1000 lessons
- Frequency: Every 30 minutes
- Database impact: Low
- CPU/Memory: Minimal

## Size

Total: 92 KB
- Scripts: 26 KB (3 files)
- Config: 1.5 KB (2 files)
- Documentation: 64 KB (5 files)

## Created

2025-01-18 for THE BOT V3 tutoring platform
