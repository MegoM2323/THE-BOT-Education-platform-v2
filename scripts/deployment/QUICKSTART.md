# Chat Creation System - Quick Start Guide

## What Does This Do?

Automatically creates chat rooms between teachers and students after lessons are completed. This ensures that students and teachers can communicate about their lessons.

## 30-Second Setup

```bash
cd /home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment

# 1. Verify system works (read-only)
./verify-chat-creation.sh --check-only

# 2. Setup automatic backfill (production)
sudo ./setup-chat-creation.sh --timer
```

Done! The system will now automatically create missing chat rooms every 30 minutes.

## Common Commands

### Check Status

```bash
# Local development database
./verify-chat-creation.sh --check-only

# Production database
DB_HOST=5.129.249.206 DB_NAME=thebot_db \
./verify-chat-creation.sh --check-only --verbose
```

### Create Missing Chats Now

```bash
# Local database
./verify-chat-creation.sh --backfill

# Production database
DB_HOST=5.129.249.206 DB_NAME=thebot_db \
./verify-chat-creation.sh --remote --backfill
```

### Monitor Automatic Backfill

```bash
# If using systemd timer (recommended)
systemctl status thebot-chat-creation.timer
journalctl -u thebot-chat-creation.service -f

# If using crontab
tail -f /var/log/thebot/chat-creation-cron.log
```

## How It Works

1. **Trigger (Automatic)**: When a lesson's `end_time` passes, the database trigger fires automatically and creates a chat room between teacher and student.

2. **Backfill (Periodic)**: Every 30 minutes, the system checks for any completed lessons that don't have chat rooms yet and creates them. This handles:
   - Cases where the trigger didn't fire (e.g., during migrations)
   - Historical completed lessons that need chat rooms
   - System recovery after downtime

## Installation Options

### Option 1: Systemd Timer (Recommended)

```bash
sudo ./setup-chat-creation.sh --timer
```

**Advantages:**
- Automatic start after reboot
- Integrated with system logging
- Easy to monitor and manage
- Native systemd integration

**Monitor:**
```bash
systemctl status thebot-chat-creation.timer
journalctl -u thebot-chat-creation.service -f
```

### Option 2: Crontab

```bash
sudo ./setup-chat-creation.sh --cron
```

**Advantages:**
- Simple, traditional approach
- No additional system configuration needed

**Monitor:**
```bash
crontab -l
tail -f /var/log/thebot/chat-creation-cron.log
```

### Option 3: Manual (No Automation)

```bash
# Just run verification and backfill when needed
./verify-chat-creation.sh --backfill
```

## Verification Output

Good output:
```
✓ Connected to localhost:5432/tutoring_platform
✓ Trigger 'booking_create_chat' exists (BEFORE INSERT)
✓ Trigger 'booking_create_chat_update' exists (BEFORE UPDATE)
✓ Function 'create_chat_on_booking_active()' exists
```

If you see errors, the database migration (050_chat_on_booking_creation.sql) may not have been applied.

## Troubleshooting

### "Trigger NOT FOUND"

The migration hasn't been applied:

```bash
# Apply migrations
cd /home/mego/Python\ Projects/THE_BOT_V3/backend
./scripts/migrate.sh
```

### "psql: command not found"

Install PostgreSQL client:

```bash
# Ubuntu/Debian
sudo apt-get install postgresql-client

# CentOS/RHEL
sudo yum install postgresql
```

### "Connection refused"

Check database credentials:

```bash
psql -h 5.129.249.206 -p 5432 -U postgres -d thebot_db -c "SELECT 1;"
```

### "No chat rooms created"

This is normal if there are no completed lessons with active bookings. Check:

```bash
./verify-chat-creation.sh --check-only

# Look for "Chat Creation Candidates" section
# If it shows 0, all chats are already created
```

## For Different Environments

### Development (localhost)

```bash
./verify-chat-creation.sh --check-only
./verify-chat-creation.sh --backfill
```

### Staging (custom server)

```bash
DB_HOST=staging.example.com DB_NAME=tutoring_staging \
./verify-chat-creation.sh --check-only
```

### Production (the-bot.ru)

```bash
DB_HOST=5.129.249.206 DB_NAME=thebot_db \
./verify-chat-creation.sh --check-only

# Then setup automation
sudo ./setup-chat-creation.sh --timer
```

## What Gets Created

```
For each completed lesson with active bookings:
  - Check if chat room exists between teacher and student
  - If not, create it with timestamp
  - Return count of newly created chat rooms
```

## Performance

- Creates chat rooms: < 1 second for < 1000 lessons
- Runs every: 30 minutes
- Database impact: Low (uses indexed queries)
- CPU impact: Negligible

## Logs

If using systemd timer:
```bash
journalctl -u thebot-chat-creation.service --since "1 hour ago"
```

If using cron:
```bash
tail -100 /var/log/thebot/chat-creation-cron.log
```

## Next Steps

1. Run verification: `./verify-chat-creation.sh --check-only`
2. Setup automation: `sudo ./setup-chat-creation.sh --timer`
3. Monitor first run: `journalctl -u thebot-chat-creation.service -f`
4. Done! System will auto-create chats every 30 minutes

## Need Help?

See full documentation: `CHAT_CREATION_README.md`
