#!/bin/sh

set -e

# Simple logging
log_info() {
    echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_success() {
    echo "[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_warning() {
    echo "[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1"
}

# Configuration with defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tutoring_platform}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/migrations}"
MAX_RETRIES="${MAX_RETRIES:-30}"
RETRY_INTERVAL="${RETRY_INTERVAL:-2}"

# Wait for PostgreSQL to be ready
wait_for_postgres() {
    log_info "Waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}..."

    retries=0
    while [ "$retries" -lt "$MAX_RETRIES" ]; do
        if PGPASSWORD="$DB_PASSWORD" pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -q 2>/dev/null; then
            log_success "PostgreSQL is ready"
            return 0
        fi

        retries=$((retries + 1))
        log_info "Waiting for PostgreSQL... (attempt $retries/$MAX_RETRIES)"
        sleep "$RETRY_INTERVAL"
    done

    log_error "PostgreSQL is not available after $MAX_RETRIES attempts"
    return 1
}

# Apply database migrations
apply_migrations() {
    log_info "Applying database migrations from ${MIGRATIONS_DIR}..."

    if [ ! -d "$MIGRATIONS_DIR" ]; then
        log_warning "Migrations directory not found: ${MIGRATIONS_DIR}"
        return 0
    fi

    migration_count=0
    success_count=0
    failed_count=0

    # Sort migration files numerically by leading digits
    for migration_file in $(find "$MIGRATIONS_DIR" -maxdepth 1 -name "*.sql" -type f | sort -V); do
        # Skip if no files match
        [ -e "$migration_file" ] || continue

        filename=$(basename "$migration_file")

        # Skip seed files (apply them last)
        case "$filename" in
            seed_*) continue ;;
        esac

        migration_count=$((migration_count + 1))
        log_info "[$migration_count] Applying migration: $filename"

        if PGPASSWORD="$DB_PASSWORD" psql \
            -h "$DB_HOST" \
            -p "$DB_PORT" \
            -U "$DB_USER" \
            -d "$DB_NAME" \
            -f "$migration_file" \
            2>/dev/null; then
            log_success "Applied: $filename"
            success_count=$((success_count + 1))

            # Special logging for migration 045 (credits_cost)
            if [ "$filename" = "045_add_credits_cost_to_template_lessons.sql" ]; then
                log_info "[045] Verifying credits_cost column in template_lessons..."
                COLUMN_CHECK=$(PGPASSWORD="$DB_PASSWORD" psql \
                    -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
                    -t -c "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" 2>/dev/null)

                if [ "$COLUMN_CHECK" = "1" ]; then
                    log_success "[045] Column credits_cost successfully added to template_lessons"

                    DEFAULT_CHECK=$(PGPASSWORD="$DB_PASSWORD" psql \
                        -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
                        -t -c "SELECT column_default FROM information_schema.columns WHERE table_name='template_lessons' AND column_name='credits_cost';" 2>/dev/null)
                    log_info "[045] Default value: $DEFAULT_CHECK"

                    CONSTRAINT_CHECK=$(PGPASSWORD="$DB_PASSWORD" psql \
                        -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
                        -t -c "SELECT constraint_name FROM information_schema.table_constraints WHERE table_name='template_lessons' AND constraint_type='CHECK';" 2>/dev/null | grep -c credits_cost)
                    log_info "[045] CHECK constraints verified: $CONSTRAINT_CHECK"
                fi
            fi
        else
            log_warning "Skipped (may already exist): $filename"
            failed_count=$((failed_count + 1))
        fi
    done

    # Apply seed files last
    log_info "Applying seed files..."
    seed_count=0
    for seed_file in $(find "$MIGRATIONS_DIR" -maxdepth 1 -name "seed_*.sql" -type f | sort -V); do
        [ -e "$seed_file" ] || continue

        filename=$(basename "$seed_file")
        seed_count=$((seed_count + 1))
        log_info "[SEED-$seed_count] Applying: $filename"

        if PGPASSWORD="$DB_PASSWORD" psql \
            -h "$DB_HOST" \
            -p "$DB_PORT" \
            -U "$DB_USER" \
            -d "$DB_NAME" \
            -f "$seed_file" \
            2>/dev/null; then
            log_success "Applied seed: $filename"
        else
            log_warning "Seed file may have already run: $filename"
        fi
    done

    log_info "=========================================="
    log_info "Migrations Summary:"
    log_info "  Regular migrations: $success_count applied"
    if [ "$failed_count" -gt 0 ]; then
        log_info "  Already existed: $failed_count"
    fi
    if [ "$seed_count" -gt 0 ]; then
        log_info "  Seed files: $seed_count"
    fi
    log_info "=========================================="
    return 0
}

# Validate environment variables before starting
validate_environment() {
    log_info "Validating environment variables..."

    local validation_failed=0

    # Check SESSION_SECRET is set and not empty
    if [ -z "$SESSION_SECRET" ]; then
        log_error "SESSION_SECRET is not set or empty"
        log_error "Generate with: openssl rand -base64 48"
        validation_failed=1
    fi

    # Check ENV variable
    ENV="${ENV:-development}"

    # In production, check PRODUCTION_DOMAIN
    if [ "$ENV" = "production" ]; then
        if [ -z "$PRODUCTION_DOMAIN" ]; then
            log_error "PRODUCTION_DOMAIN is required in production mode"
            log_error "Set PRODUCTION_DOMAIN to your domain (e.g., example.com)"
            validation_failed=1
        fi

        # In production, DB_PASSWORD must not be empty
        if [ -z "$DB_PASSWORD" ]; then
            log_error "DB_PASSWORD must not be empty in production"
            log_error "Empty password allows unauthorized database access"
            validation_failed=1
        fi
    fi

    # Check DB_PASSWORD is set (warning for development, error for production)
    if [ -z "$DB_PASSWORD" ] && [ "$ENV" != "production" ]; then
        log_warning "DB_PASSWORD is not set (acceptable in development with peer auth)"
    fi

    # Check DB_HOST is accessible
    if [ -z "$DB_HOST" ]; then
        log_error "DB_HOST is not set"
        validation_failed=1
    else
        log_info "Checking DB_HOST accessibility: ${DB_HOST}:${DB_PORT}"

        # Try to resolve hostname
        if ! getent hosts "$DB_HOST" >/dev/null 2>&1; then
            log_error "DB_HOST '$DB_HOST' cannot be resolved"
            log_error "Ensure database host is accessible from this container"
            validation_failed=1
        else
            log_success "DB_HOST '$DB_HOST' is resolvable"
        fi
    fi

    if [ "$validation_failed" -eq 1 ]; then
        log_error "Environment validation failed. Fix the errors above and restart."
        exit 1
    fi

    log_success "Environment validation passed"
}

# Start the backend
start_backend() {
    log_info "Starting server..."

    if [ ! -x "./server" ]; then
        log_error "Backend binary not found or not executable: ./server"
        exit 1
    fi

    exec ./server
}

# Main function
main() {
    log_info "=== Tutoring Platform Backend Entrypoint ==="
    log_info "Database: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

    # Validate environment variables first
    validate_environment

    if ! wait_for_postgres; then
        log_error "Failed to connect to PostgreSQL"
        exit 1
    fi

    if ! apply_migrations; then
        log_error "Migration failed"
        exit 1
    fi

    start_backend
}

main "$@"
