#!/bin/bash

# Migration script for Tutoring Platform

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-tutoring_platform}
DB_USER=${DB_USER:-postgres}

MIGRATIONS_DIR="internal/database/migrations"

echo "=== Tutoring Platform Migration Script ==="
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo ""

# Function to run migration
run_migration() {
    local file=$1
    echo "Running: $file"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$file"
    if [ $? -eq 0 ]; then
        echo "✓ Success: $file"
    else
        echo "✗ Failed: $file"
        exit 1
    fi
    echo ""
}

# Function to get all migration files sorted by name
get_all_migrations() {
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        echo "ERROR: Migrations directory not found: $MIGRATIONS_DIR"
        exit 1
    fi

    # Find all SQL files, sort them numerically by leading digits
    find "$MIGRATIONS_DIR" -maxdepth 1 -name "*.sql" -type f | sort -V
}

# Check command
case "$1" in
    up)
        echo "Applying migrations..."
        migration_count=0
        success_count=0

        while IFS= read -r migration_file; do
            if [ -z "$migration_file" ]; then
                continue
            fi
            migration_count=$((migration_count + 1))
            run_migration "$migration_file"
            success_count=$((success_count + 1))
        done < <(get_all_migrations)

        echo ""
        echo "=========================================="
        echo "Migration Summary:"
        echo "  Total: $migration_count"
        echo "  Applied: $success_count"
        echo "=========================================="
        echo "All migrations applied successfully!"
        ;;

    down)
        echo "WARNING: This will drop all data!"

        # CRITICAL SAFETY CHECK: Verify this is NOT production database
        if [[ "$DB_NAME" != "tutoring_platform_test" ]]; then
            echo "ERROR: down command should NOT be run on production database!"
            echo "Target database: $DB_NAME"
            echo ""
            echo "down command is designed for cleaning test database only."
            echo "Use this only with TEST_DATABASE_NAME environment variable."
            exit 1
        fi

        read -p "Are you sure you want to DROP SCHEMA in $DB_NAME? (yes/no): " confirm
        if [ "$confirm" = "yes" ]; then
            echo "Rolling back all migrations..."
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
            echo "All migrations rolled back!"
        else
            echo "Cancelled."
        fi
        ;;

    reset)
        echo "WARNING: This will drop and recreate all data!"

        # CRITICAL SAFETY CHECK: Verify this is NOT production database
        if [[ "$DB_NAME" != "tutoring_platform_test" ]]; then
            echo "ERROR: reset command is ONLY for test database!"
            echo "You are about to delete all data in database: $DB_NAME"
            echo ""
            echo "If this is the test database, run: migrate.sh reset --force"
            echo "WARNING: --force requires EXPLICIT confirmation!"
            exit 1
        fi

        read -p "Are you sure you want to DROP SCHEMA in $DB_NAME? (yes/no): " confirm
        if [ "$confirm" = "yes" ]; then
            echo "Resetting database..."
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

            migration_count=0
            success_count=0

            while IFS= read -r migration_file; do
                if [ -z "$migration_file" ]; then
                    continue
                fi
                migration_count=$((migration_count + 1))
                run_migration "$migration_file"
                success_count=$((success_count + 1))
            done < <(get_all_migrations)

            echo ""
            echo "=========================================="
            echo "Reset Summary:"
            echo "  Total migrations: $migration_count"
            echo "  Applied: $success_count"
            echo "=========================================="
            echo "Database reset successfully!"
        else
            echo "Cancelled."
        fi
        ;;

    *)
        echo "Usage: $0 {up|down|reset}"
        echo ""
        echo "Commands:"
        echo "  up    - Apply all migrations from $MIGRATIONS_DIR"
        echo "  down  - Rollback all migrations (test DB only)"
        echo "  reset - Drop and reapply all migrations (test DB only)"
        exit 1
        ;;
esac
