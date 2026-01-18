# Chat Creation System - Usage Examples

## Development Environment

### 1. Check Status (Local Database)

```bash
cd /home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment
./verify-chat-creation.sh --check-only
```

**Expected Output**:
```
================================
Chat Creation System Verification
================================

▶ Testing database connection...
✓ Connected to localhost:5432/tutoring_platform

▶ Checking triggers on bookings table...
✓ Trigger 'booking_create_chat' exists (BEFORE INSERT)
✓ Trigger 'booking_create_chat_update' exists (BEFORE UPDATE)

▶ Checking PL/pgSQL functions...
✓ Function 'create_chat_on_booking_active()' exists

▶ Chat Creation System Statistics
...

✓ All checks passed!
```

### 2. Create Missing Chats (Local)

```bash
./verify-chat-creation.sh --backfill
```

**Output if chats were created**:
```
▶ Running chat creation backfill function...
✓ Backfill completed successfully
✓ Created 5 new chat rooms
```

### 3. Verbose Debugging

```bash
./verify-chat-creation.sh --check-only --verbose
```

Shows detailed connection info and query execution.

---

## Production Environment

### 1. Verify Production Database (Read-only)

```bash
DB_HOST=5.129.249.206 DB_NAME=thebot_db DB_USER=postgres \
./verify-chat-creation.sh --check-only
```

Or with password:
```bash
DB_HOST=5.129.249.206 DB_NAME=thebot_db DB_USER=postgres DB_PASSWORD="your_password" \
./verify-chat-creation.sh --check-only
```

### 2. Create Missing Chats on Production

```bash
DB_HOST=5.129.249.206 DB_NAME=thebot_db DB_USER=postgres \
./verify-chat-creation.sh --remote --backfill
```

**Safety Features**:
- Script validates database connection before making changes
- Reports how many chats will be created
- Transactional: either all succeed or all fail

### 3. Setup Automatic Backfill (Production)

#### Option A: Systemd Timer (Recommended)

```bash
# On production server
ssh mg@5.129.249.206

# Navigate to scripts
cd /home/mg/the-bot/scripts/deployment

# Install with timer (every 30 minutes)
sudo ./setup-chat-creation.sh --timer

# Verify installation
systemctl status thebot-chat-creation.timer
systemctl list-timers thebot-chat-creation.timer
```

#### Option B: Crontab

```bash
ssh mg@5.129.249.206
cd /home/mg/the-bot/scripts/deployment

# Install with crontab
sudo ./setup-chat-creation.sh --cron

# Verify installation
crontab -l | grep chat-creation

# Monitor logs
tail -f /var/log/thebot/chat-creation-cron.log
```

---

## Specific Use Cases

### Use Case 1: After Database Migration

After running migration `050_chat_on_booking_creation.sql`:

```bash
# 1. Verify migration applied
./verify-chat-creation.sh --check-only

# 2. Create chats for lessons that completed before migration
./verify-chat-creation.sh --backfill

# 3. Verify results
./verify-chat-creation.sh --check-only
```

### Use Case 2: System Recovery After Downtime

```bash
# 1. Check how many chats are missing
./verify-chat-creation.sh --check-only

# Look for "Chat Creation Candidates" count

# 2. Recreate missing chats
./verify-chat-creation.sh --backfill

# 3. Verify all systems operational
./verify-chat-creation.sh --check-only
```

### Use Case 3: Testing on Staging

```bash
# Staging server address
STAGING_HOST="staging.example.com"
STAGING_DB="tutoring_staging"

# Check staging system
DB_HOST=$STAGING_HOST DB_NAME=$STAGING_DB \
./verify-chat-creation.sh --check-only

# Create test chats on staging
DB_HOST=$STAGING_HOST DB_NAME=$STAGING_DB \
./verify-chat-creation.sh --backfill

# Verify results
DB_HOST=$STAGING_HOST DB_NAME=$STAGING_DB \
./verify-chat-creation.sh --check-only
```

### Use Case 4: Monitor Production Backfill (Systemd)

```bash
# Real-time monitoring
journalctl -u thebot-chat-creation.service -f

# Last 50 lines
journalctl -u thebot-chat-creation.service -n 50

# Last hour
journalctl -u thebot-chat-creation.service --since "1 hour ago"

# Specific date range
journalctl -u thebot-chat-creation.service --since "2025-01-18 00:00:00" --until "2025-01-18 23:59:59"

# Count successes and failures
journalctl -u thebot-chat-creation.service | grep -c "Created"
journalctl -u thebot-chat-creation.service | grep -c "ERROR"

# Export to file
journalctl -u thebot-chat-creation.service > /tmp/chat-creation-logs.txt
```

### Use Case 5: Monitor Production Backfill (Crontab)

```bash
# Watch live
tail -f /var/log/thebot/chat-creation-cron.log

# Last 20 entries
tail -20 /var/log/thebot/chat-creation-cron.log

# Statistics
grep "SUCCESS" /var/log/thebot/chat-creation-cron.log | wc -l
grep "ERROR" /var/log/thebot/chat-creation-cron.log | wc -l

# Filter by date
grep "2025-01-18" /var/log/thebot/chat-creation-cron.log
```

---

## Troubleshooting Examples

### Problem: Triggers Not Found

```bash
$ ./verify-chat-creation.sh --check-only
✗ Triggers 'booking_create_chat' NOT FOUND
```

**Solution**:
```bash
# Apply the migration
cd /home/mego/Python\ Projects/THE_BOT_V3/backend
./scripts/migrate.sh

# Verify
cd ../scripts/deployment
./verify-chat-creation.sh --check-only
```

### Problem: Connection Refused

```bash
$ DB_HOST=5.129.249.206 ./verify-chat-creation.sh --check-only
✗ Failed to connect to database
```

**Solutions**:
```bash
# Test connection with psql
psql -h 5.129.249.206 -p 5432 -U postgres -d thebot_db -c "SELECT 1;"

# Check if host is reachable
ping 5.129.249.206
nc -zv 5.129.249.206 5432

# Try with password
DB_HOST=5.129.249.206 DB_PASSWORD="your_password" \
./verify-chat-creation.sh --check-only
```

### Problem: Backfill Creates Too Many Chats

```bash
# Preview before running
./verify-chat-creation.sh --check-only

# If numbers look wrong, manually verify with SQL
psql -h localhost -U postgres -d tutoring_platform -c "
  SELECT
    COUNT(DISTINCT b.id) as total_bookings,
    COUNT(DISTINCT cr.id) as existing_chats,
    COUNT(DISTINCT b.id) - COUNT(DISTINCT cr.id) as missing
  FROM bookings b
  LEFT JOIN chat_rooms cr ON cr.student_id = b.student_id
  WHERE b.status = 'active';
"
```

### Problem: PostgreSQL Client Not Installed

```bash
# Error message
bash: psql: command not found

# Solution on Ubuntu/Debian
sudo apt-get update
sudo apt-get install postgresql-client

# Solution on CentOS/RHEL
sudo yum install postgresql

# Verify
psql --version
```

---

## Batch Operations

### Backfill All Databases

For multi-tenancy scenario:

```bash
#!/bin/bash

# Array of databases
declare -a databases=(
    "tutoring_platform"
    "thebot_db"
    "staging_db"
)

for db in "${databases[@]}"; do
    echo "Processing $db..."
    DB_NAME="$db" ./verify-chat-creation.sh --backfill
done
```

### Check Multiple Servers

```bash
#!/bin/bash

# Array of servers
declare -a servers=(
    "localhost:tutoring_platform"
    "5.129.249.206:thebot_db"
    "staging.example.com:staging_db"
)

for server_db in "${servers[@]}"; do
    host=$(echo $server_db | cut -d: -f1)
    db=$(echo $server_db | cut -d: -f2)

    echo "Checking $host:$db"
    DB_HOST="$host" DB_NAME="$db" ./verify-chat-creation.sh --check-only
    echo "---"
done
```

---

## Scheduling Examples

### Systemd Timer Variants

#### Every 15 minutes (aggressive backfill)

Edit `/etc/systemd/system/thebot-chat-creation.timer`:
```ini
[Timer]
OnBootSec=1min
OnUnitActiveSec=15min
```

#### Every hour (conservative)

```ini
[Timer]
OnBootSec=5min
OnUnitActiveSec=1h
```

### Crontab Variants

#### Every 5 minutes
```bash
*/5 * * * * /path/to/chat-creation-cron.sh
```

#### Twice daily (6 AM and 6 PM)
```bash
0 6,18 * * * /path/to/chat-creation-cron.sh
```

#### Every Monday at 2 AM
```bash
0 2 * * 1 /path/to/chat-creation-cron.sh
```

---

## Performance Benchmarks

### Test on Different Database Sizes

```bash
# On test database with various scales

# 100 lessons, 50 completed
time ./verify-chat-creation.sh --backfill
# Expected: < 100ms

# 1000 lessons, 500 completed
time ./verify-chat-creation.sh --backfill
# Expected: < 500ms

# 10000 lessons, 5000 completed
time ./verify-chat-creation.sh --backfill
# Expected: 1-2 seconds
```

---

## Health Check Integration

### Add to Monitoring System

```bash
#!/bin/bash
# health-check.sh

SCRIPT_DIR="/home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment"
LOG_FILE="/var/log/thebot/health-check.log"

# Run verification
if $SCRIPT_DIR/verify-chat-creation.sh --check-only > /dev/null 2>&1; then
    echo "$(date): Chat creation system OK" >> $LOG_FILE
    exit 0
else
    echo "$(date): Chat creation system FAILED" >> $LOG_FILE
    # Send alert (e.g., to monitoring system)
    exit 1
fi
```

### Prometheus Metrics Export

```bash
#!/bin/bash
# metrics.sh

# For Prometheus scraping
echo "# HELP thebot_chats_created Total chat rooms created"
echo "# TYPE thebot_chats_created gauge"

count=$(psql -h localhost -U postgres -d tutoring_platform \
    -t -c "SELECT COUNT(*) FROM chat_rooms WHERE deleted_at IS NULL;")

echo "thebot_chats_created{instance=\"local\"} $count"
```

---

## Documentation Links

- **Quick Start**: QUICKSTART.md
- **Full Documentation**: CHAT_CREATION_README.md
- **File Index**: INDEX.md
- **Database Migration**: `backend/internal/database/migrations/050_chat_on_booking_creation.sql`
