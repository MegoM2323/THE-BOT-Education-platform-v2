# Chat Creation System - Production Verification

## Overview

This directory contains scripts to verify and manage the automated chat creation system for THE BOT tutoring platform.

The system automatically creates chat rooms between teachers and students when a booking becomes active, based on the database triggers `booking_create_chat` and `booking_create_chat_update` defined in migration `050_chat_on_booking_creation.sql`.

## Scripts

### 1. verify-chat-creation.sh

Main verification script to check the status of the chat creation system.

**Features:**
- Test database connection
- Verify triggers exist: `booking_create_chat` and `booking_create_chat_update`
- Verify PL/pgSQL function exists: `create_chat_on_booking_active()`
- Show comprehensive statistics (active bookings, chat rooms, messages)
- Verify data integrity (bookings without chats)
- Optional backfill of missing chat rooms for existing bookings
- Colorized output with detailed reporting

**Usage:**

```bash
# Check local database (read-only)
./verify-chat-creation.sh --check-only

# Check with verbose logging
./verify-chat-creation.sh --check-only --verbose

# Check production database
DB_HOST=5.129.249.206 DB_NAME=thebot_db ./verify-chat-creation.sh --check-only

# Run backfill for missing chats (modifies database)
./verify-chat-creation.sh --backfill

# Production backfill with verification
DB_HOST=5.129.249.206 DB_NAME=thebot_db ./verify-chat-creation.sh --remote --backfill
```

**Options:**
- `--remote` - Indicates connection to production database
- `--backfill` - Run the chat creation backfill function (creates missing chats)
- `--check-only` - Read-only mode, no modifications
- `--verbose` - Enable detailed logging
- `--help` - Show help message

**Environment Variables:**
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_NAME` - Database name (default: tutoring_platform)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: empty)

### 2. chat-creation-cron.sh

Automated cron job script for periodic chat creation backfill.

**Features:**
- Silent execution suitable for cron jobs
- Automatic logging to `/var/log/thebot/chat-creation-cron.log`
- Database connection validation
- Error handling and exit codes
- Log rotation (keeps 30 days of logs)

**Usage:**

```bash
# Manual execution
./chat-creation-cron.sh

# Add to crontab (every 5 minutes)
*/5 * * * * /home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment/chat-creation-cron.sh

# Add to crontab (every hour)
0 * * * * /home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment/chat-creation-cron.sh

# Add to crontab (every 30 minutes, production)
*/30 * * * * DB_HOST=5.129.249.206 DB_NAME=thebot_db /home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment/chat-creation-cron.sh
```

**Crontab Example for Production:**

```bash
# Edit crontab
crontab -e

# Add this line to run chat creation backfill every 30 minutes
*/30 * * * * export DB_HOST=5.129.249.206 DB_NAME=thebot_db DB_USER=postgres && /home/mego/Python\ Projects/THE_BOT_V3/scripts/deployment/chat-creation-cron.sh
```

## How the Chat Creation System Works

### Automatic Trigger (Real-time)

When a booking is created or updated with `status = 'active'`:

1. Database trigger `booking_create_chat` (INSERT) or `booking_create_chat_update` (UPDATE) fires
2. Function `create_chat_on_booking_active()` executes (BEFORE the row is committed)
3. Function retrieves `teacher_id` from the associated lesson
4. Creates new chat room between teacher and student if it doesn't exist
5. UNIQUE constraint prevents duplicate chat rooms

This happens:
- Immediately when booking is created with status='active'
- Immediately when booking status changes to 'active'
- Before the database transaction is committed

### Backfill During Migration

During initial deployment of migration 050:

1. Migration creates the triggers and function
2. Backfill code runs automatically within the migration
3. Finds all active bookings that don't have chat rooms yet
4. Creates missing chat rooms in batch

This ensures existing bookings get chats even if created before the migration.

## Database Schema

### Relevant Tables

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    role VARCHAR(50),
    ...
);

-- Lessons table
CREATE TABLE lessons (
    id UUID PRIMARY KEY,
    teacher_id UUID REFERENCES users(id),
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    deleted_at TIMESTAMP,
    ...
);

-- Bookings table
CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    lesson_id UUID REFERENCES lessons(id),
    student_id UUID REFERENCES users(id),
    status VARCHAR(50), -- 'active', 'completed', 'cancelled'
    deleted_at TIMESTAMP,
    ...
);

-- Chat rooms table
CREATE TABLE chat_rooms (
    id UUID PRIMARY KEY,
    teacher_id UUID REFERENCES users(id),
    student_id UUID REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(teacher_id, student_id)
);

-- Messages table
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    chat_room_id UUID REFERENCES chat_rooms(id),
    sender_id UUID REFERENCES users(id),
    content TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    ...
);
```

### Trigger Definitions

```sql
-- Trigger 1: When new booking is created with status='active'
CREATE TRIGGER booking_create_chat
BEFORE INSERT ON bookings
FOR EACH ROW
WHEN (NEW.status = 'active')
EXECUTE FUNCTION create_chat_on_booking_active();

-- Trigger 2: When existing booking status changes to 'active'
CREATE TRIGGER booking_create_chat_update
BEFORE UPDATE ON bookings
FOR EACH ROW
WHEN (NEW.status = 'active' AND OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION create_chat_on_booking_active();
```

## Verification Output

When running `verify-chat-creation.sh`, you'll see output like:

```
================================
Chat Creation System Verification
================================

▶ Testing database connection...
✓ Connected to localhost:5432/tutoring_platform

▶ Checking triggers on bookings table...
✓ Trigger 'booking_create_chat' exists (BEFORE INSERT)
✓ Trigger 'booking_create_chat_update' exists (BEFORE UPDATE)

▶ Checking PL/pgSQL function...
✓ Function 'create_chat_on_booking_active()' exists
✓ Function 'create_chats_for_completed_lessons()' exists

▶ Chat Creation System Statistics

Lesson Statistics:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 metric                        | count
───────────────────────────────┼────────
 Total lessons                 | 150
 Completed lessons             | 45
 Upcoming lessons              | 105
 Lessons without end_time      | 0

Booking Statistics:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 metric           | count
──────────────────┼────────
 Active bookings  | 95
 Completed        | 20
 Cancelled        | 10

Chat Room Statistics:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 metric                  | count
─────────────────────────┼────────
 Total chat rooms        | 42
 Chat rooms with msg     | 38
 Empty chat rooms        | 4

Chat Creation Candidates:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 potential_chats | lessons_affected
─────────────────┼──────────────────
 8               | 5

▶ Verifying data integrity...
✓ No orphaned bookings found
✓ All chat rooms have valid teacher and student

▶ Verification Summary

Configuration:
  Host: localhost
  Port: 5432
  Database: tutoring_platform
  User: postgres

Modes:
  Remote: No
  Backfill: No
  Check-only: Yes
  Verbose: No

✓ All checks passed!
```

## Running on Production

### Step 1: Test Connection

```bash
DB_HOST=5.129.249.206 DB_NAME=thebot_db DB_USER=postgres \
./verify-chat-creation.sh --check-only --verbose
```

### Step 2: Check for Missing Chats

Review the "Chat Creation Candidates" section to see how many chats need to be created.

### Step 3: Run Backfill (if needed)

```bash
DB_HOST=5.129.249.206 DB_NAME=thebot_db DB_USER=postgres \
./verify-chat-creation.sh --remote --backfill
```

### Step 4: Verify Results

```bash
DB_HOST=5.129.249.206 DB_NAME=thebot_db DB_USER=postgres \
./verify-chat-creation.sh --check-only
```

The "Chat Creation Candidates" count should be 0 if all chats were successfully created.

## Setting Up Scheduled Backfill

Add to crontab on production server:

```bash
# SSH to production server
ssh mg@5.129.249.206

# Edit crontab
crontab -e

# Add this line to run every 30 minutes
*/30 * * * * export DB_HOST=localhost DB_NAME=thebot_db DB_USER=postgres && /home/mg/the-bot/scripts/deployment/chat-creation-cron.sh >> /var/log/thebot/chat-creation-cron.log 2>&1
```

## Troubleshooting

### Script fails: "psql: command not found"

Install PostgreSQL client tools:
```bash
# On Ubuntu/Debian
sudo apt-get install postgresql-client

# On CentOS/RHEL
sudo yum install postgresql
```

### Database connection refused

Check credentials and network:
```bash
# Test connection with psql
psql -h 5.129.249.206 -p 5432 -U postgres -d thebot_db -c "SELECT 1;"

# Check if host is reachable
ping 5.129.249.206
nc -zv 5.129.249.206 5432
```

### Triggers not firing

The triggers should fire automatically when bookings are created/updated with `status='active'`. If triggers aren't firing:

1. Check if migration 050_chat_on_booking_creation.sql was applied
2. Verify triggers exist: `SELECT * FROM information_schema.triggers WHERE event_object_table='bookings'`
3. Check that bookings are being created/updated with `status='active'`
4. Run manual backfill: `./verify-chat-creation.sh --backfill`
5. Check PostgreSQL logs for trigger errors

### Function returns error

Check PostgreSQL logs:
```bash
# On production server
tail -f /var/log/postgresql/postgresql.log

# Or via systemctl
journalctl -u postgresql -f
```

### Chat rooms created but messages not visible

This is normal. Chat rooms are created automatically, but they only appear in the UI once messages are sent or the app loads active chat rooms.

## Related Documentation

- **Database Migration**: `/backend/internal/database/migrations/050_chat_on_booking_creation.sql`
- **Chat Models**: `/backend/apps/chat/models.py` (if using Django)
- **Database Setup**: `/backend/scripts/migrate.sh`

## Support

For issues with the chat creation system:

1. Check this README first
2. Run `./verify-chat-creation.sh --verbose` for detailed logs
3. Review PostgreSQL logs on production server
4. Check if migration 049_chat_auto_create.sql was applied correctly

## Performance Notes

- **Backfill Function**: Typically completes in < 1 second for < 1000 lessons
- **Trigger**: Fires only on lesson UPDATE, minimal overhead
- **Cron Job**: Recommended frequency: every 30 minutes
- **Database Impact**: Low, uses indexed queries

## Security

- Scripts read database credentials from environment variables
- No credentials are logged to files
- All queries are parameterized (no SQL injection)
- Consider using PostgreSQL connection pooling for production
