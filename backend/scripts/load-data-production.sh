#!/bin/bash

# Load Production Data for THE BOT
# This script loads realistic production data with proper schema constraints
# Usage: ./load-data-production.sh [--truncate] [--yes]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration from environment or defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

# Parse arguments
TRUNCATE=false
AUTO_YES=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --truncate)
            TRUNCATE=true
            shift
            ;;
        --yes)
            AUTO_YES=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--truncate] [--yes]"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}THE BOT - Production Data Loader${NC}"
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "User: $DB_USER"
echo ""

# Safety check - warn about production database
if [[ "$DB_NAME" == "thebot_db" ]] || [[ "$DB_NAME" == "the_bot" ]]; then
    if [[ "$AUTO_YES" != "true" ]]; then
        echo -e "${RED}WARNING: You are about to load data into database: $DB_NAME${NC}"
        echo "This will TRUNCATE all existing data!"
        read -p "Are you sure? (yes/no): " -r CONFIRM
        if [[ "$CONFIRM" != "yes" ]]; then
            echo "Cancelled."
            exit 1
        fi
    fi
fi

# Find the SQL file
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_FILE="$SCRIPT_DIR/load-data-production.sql"

if [[ ! -f "$SQL_FILE" ]]; then
    echo -e "${RED}Error: SQL file not found at $SQL_FILE${NC}"
    exit 1
fi

echo -e "${GREEN}Loading data from: $SQL_FILE${NC}"
echo ""

# Set up password for psql
export PGPASSWORD="$DB_PASSWORD"

# Run the SQL file
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE"

# Check exit status
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Data loaded successfully!${NC}"
    echo ""
    echo "Test credentials:"
    echo "  Admin: admin@thebot.ru / password123"
    echo "  Teacher: method1@thebot.ru / password123"
    echo "  Student: student1@thebot.ru / password123"
    echo ""
else
    echo ""
    echo -e "${RED}✗ Error loading data${NC}"
    exit 1
fi
