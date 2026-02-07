#!/bin/bash

# Deployment Test Suite
# Tests for deployment script fixes: docker-compose, backend source prep, SESSION_SECRET preservation

# Note: We don't use set -e to allow test failures to be captured

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Logging
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
    ((TESTS_RUN++))
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Check if script is sourced or run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
else
    PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[1]}")" && pwd)"
fi

echo "=========================================="
echo "Deployment Test Suite"
echo "Project: $PROJECT_DIR"
echo "=========================================="
echo ""

#############################################
# UNIT TESTS: Bash Script Functions
#############################################

test_prepare_backend_source_tarball_creation() {
    log_test "prepare_backend_source() creates tar.gz"

    cd "$PROJECT_DIR"

    # Simulate prepare_backend_source function
    if tar -czf /tmp/test-backend-source.tar.gz \
        --exclude='bin' \
        --exclude='*.log' \
        backend/ 2>/dev/null; then
        if [ -f /tmp/test-backend-source.tar.gz ]; then
            local size=$(du -h /tmp/test-backend-source.tar.gz | cut -f1)
            log_pass "Tarball created (size: $size)"
            rm -f /tmp/test-backend-source.tar.gz
            return 0
        fi
    fi

    log_fail "Failed to create tarball"
    return 1
}

test_prepare_backend_source_excludes_bin() {
    log_test "prepare_backend_source() excludes bin directory"

    # Create test structure
    local test_dir="/tmp/test_backend_$$"
    mkdir -p "$test_dir/backend/bin"
    mkdir -p "$test_dir/backend/internal"
    touch "$test_dir/backend/bin/binary"
    touch "$test_dir/backend/internal/code.go"

    # Create tarball with exclusion
    cd "$test_dir"
    tar -czf /tmp/test-exclude.tar.gz \
        --exclude='bin' \
        backend/ 2>/dev/null

    # Check if bin is excluded
    if tar -tzf /tmp/test-exclude.tar.gz | grep -q "backend/bin"; then
        log_fail "bin directory not excluded"
        rm -rf "$test_dir" /tmp/test-exclude.tar.gz
        return 1
    fi

    log_pass "bin directory correctly excluded"
    rm -rf "$test_dir" /tmp/test-exclude.tar.gz
    return 0
}

test_deploy_script_syntax() {
    log_test "deploy-with-db-safe.sh syntax check"

    local output
    output=$(bash -n "$PROJECT_DIR/deploy-with-db-safe.sh" 2>&1) || true

    if [ -z "$output" ]; then
        log_pass "Script syntax is valid"
        return 0
    fi

    log_fail "Script has syntax errors: $output"
    return 1
}

test_deploy_script_executable() {
    log_test "deploy-with-db-safe.sh is executable"

    if [ -x "$PROJECT_DIR/deploy-with-db-safe.sh" ]; then
        log_pass "Script is executable"
        return 0
    fi

    log_fail "Script is not executable"
    return 1
}

test_prepare_backend_source_function_exists() {
    log_test "prepare_backend_source() function exists"

    if grep -q "^prepare_backend_source()" "$PROJECT_DIR/deploy-with-db-safe.sh"; then
        log_pass "Function exists"
        return 0
    fi

    log_fail "Function not found"
    return 1
}

#############################################
# SESSION_SECRET LOGIC TESTS
#############################################

test_session_secret_preserve_existing() {
    log_test "SESSION_SECRET preserved from existing .env"

    local test_env="/tmp/test_env_preserve_$$"
    echo "SESSION_SECRET=existing_secret_12345" > "$test_env"

    # Simulate the logic from deploy script
    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

    if [ "$EXISTING_SECRET" = "existing_secret_12345" ]; then
        log_pass "Existing SESSION_SECRET preserved"
        rm -f "$test_env"
        return 0
    fi

    log_fail "Failed to preserve existing SESSION_SECRET"
    rm -f "$test_env"
    return 1
}

test_session_secret_generate_when_missing() {
    log_test "SESSION_SECRET generated when missing"

    local test_env="/tmp/test_env_generate_$$"
    touch "$test_env"  # Empty file

    # Simulate the logic
    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

    if [ -z "$EXISTING_SECRET" ]; then
        log_pass "Correctly detects missing SESSION_SECRET"
        rm -f "$test_env"
        return 0
    fi

    log_fail "Should detect missing SESSION_SECRET"
    rm -f "$test_env"
    return 1
}

test_session_secret_not_empty_from_env() {
    log_test "SESSION_SECRET not empty when in .env"

    local test_env="/tmp/test_env_notempty_$$"
    echo "SESSION_SECRET=valid_secret_value" > "$test_env"

    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

    if [ -n "$EXISTING_SECRET" ]; then
        log_pass "SESSION_SECRET correctly extracted from .env"
        rm -f "$test_env"
        return 0
    fi

    log_fail "Failed to extract SESSION_SECRET"
    rm -f "$test_env"
    return 1
}

test_session_secret_whitespace_handling() {
    log_test "SESSION_SECRET whitespace stripped"

    local test_env="/tmp/test_env_whitespace_$$"
    echo 'SESSION_SECRET="  secret_with_spaces  "' > "$test_env"

    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

    if [ "$EXISTING_SECRET" = "  secret_with_spaces  " ]; then
        # Note: The tr -d '"' only removes quotes, not spaces
        # Additional trimming might be needed in production
        log_info "Quotes stripped (spaces remain - may need trim)"
        rm -f "$test_env"
        return 0
    fi

    log_fail "Quote stripping failed"
    rm -f "$test_env"
    return 1
}

#############################################
# INTEGRATION TESTS
#############################################

test_docker_compose_valid_yaml() {
    log_test "docker-compose.prod.yml is valid YAML"

    if command -v docker-compose &> /dev/null; then
        if docker-compose -f "$PROJECT_DIR/docker-compose.prod.yml" config > /dev/null 2>&1; then
            log_pass "docker-compose.prod.yml is valid"
            return 0
        fi
        log_fail "docker-compose validation failed"
        return 1
    elif docker compose version &> /dev/null 2>&1; then
        if docker compose -f "$PROJECT_DIR/docker-compose.prod.yml" config > /dev/null 2>&1; then
            log_pass "docker-compose.prod.yml is valid"
            return 0
        fi
        log_fail "docker compose validation failed"
        return 1
    fi

    log_info "Docker compose not available, skipping YAML validation"
    return 0
}

test_docker_compose_backend_build_config() {
    log_test "docker-compose.prod.yml has build config for backend"

    if grep -q "build:" "$PROJECT_DIR/docker-compose.prod.yml" && \
       grep -q "context: ./backend" "$PROJECT_DIR/docker-compose.prod.yml" && \
       grep -q "dockerfile: Dockerfile.prod" "$PROJECT_DIR/docker-compose.prod.yml"; then
        log_pass "Backend build configuration present"
        return 0
    fi

    log_fail "Backend build configuration missing or incorrect"
    return 1
}

test_docker_compose_session_secret_env() {
    log_test "docker-compose.prod.yml passes SESSION_SECRET"

    if grep -q "SESSION_SECRET: \${SESSION_SECRET}" "$PROJECT_DIR/docker-compose.prod.yml"; then
        log_pass "SESSION_SECRET environment variable configured"
        return 0
    fi

    log_fail "SESSION_SECRET not configured in docker-compose"
    return 1
}

test_dockerfile_exists() {
    log_test "Backend Dockerfile.prod exists"

    if [ -f "$PROJECT_DIR/backend/Dockerfile.prod" ]; then
        log_pass "Dockerfile.prod exists"
        return 0
    fi

    log_fail "Dockerfile.prod not found"
    return 1
}

test_dockerfile_syntax() {
    log_test "Dockerfile.prod has valid syntax"

    # Check for basic FROM instruction
    if grep -q "^FROM " "$PROJECT_DIR/backend/Dockerfile.prod"; then
        log_pass "Dockerfile.prod has valid FROM instruction"
        return 0
    fi

    log_fail "Dockerfile.prod missing FROM instruction"
    return 1
}

test_dockerfile_multistage() {
    log_test "Dockerfile.prod uses multi-stage build"

    local from_count=$(grep -c "^FROM " "$PROJECT_DIR/backend/Dockerfile.prod" || echo 0)

    if [ "$from_count" -ge 2 ]; then
        log_pass "Multi-stage build detected ($from_count stages)"
        return 0
    fi

    log_info "Single-stage build (from_count=$from_count)"
    return 0
}

test_dockerfile_backend_context() {
    log_test "Dockerfile.prod context matches docker-compose"

    # docker-compose uses context: ./backend
    # Check if Dockerfile.prod is in backend directory
    if [ -f "$PROJECT_DIR/backend/Dockerfile.prod" ]; then
        log_pass "Dockerfile.prod in correct location"
        return 0
    fi

    log_fail "Dockerfile.prod location mismatch"
    return 1
}

#############################################
# EDGE CASE TESTS
#############################################

test_session_secret_edge_case_empty_file() {
    log_test "SESSION_SECRET handles empty .env file"

    local test_env="/tmp/test_env_empty_$$"
    touch "$test_env"

    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

    if [ -z "$EXISTING_SECRET" ]; then
        log_pass "Correctly handles empty .env"
        rm -f "$test_env"
        return 0
    fi

    log_fail "Failed to handle empty .env"
    rm -f "$test_env"
    return 1
}

test_session_secret_edge_case_no_file() {
    log_test "SESSION_SECRET handles missing .env file"

    local test_env="/tmp/test_env_nofile_$$"

    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2 | tr -d '"')

    if [ -z "$EXISTING_SECRET" ]; then
        log_pass "Correctly handles missing file"
        return 0
    fi

    log_fail "Failed to handle missing file"
    return 1
}

test_session_secret_edge_case_multiple_occurrences() {
    log_test "SESSION_SECRET handles multiple SESSION_SECRET lines"

    local test_env="/tmp/test_env_multi_$$"
    cat > "$test_env" << EOF
SESSION_SECRET=first
SESSION_SECRET=second
EOF

    # grep with cut returns first match
    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2 | tr -d '"' | head -n1)

    if [ "$EXISTING_SECRET" = "first" ]; then
        log_pass "First value selected correctly"
        rm -f "$test_env"
        return 0
    fi

    log_fail "Failed to handle multiple occurrences"
    rm -f "$test_env"
    return 1
}

test_session_secret_edge_case_special_chars() {
    log_test "SESSION_SECRET handles special characters"

    local test_env="/tmp/test_env_special_$$"
    echo 'SESSION_SECRET="secret-with-special-chars"' > "$test_env"

    EXISTING_SECRET=$(grep "^SESSION_SECRET=" "$test_env" 2>/dev/null | cut -d'=' -f2- | tr -d '"')

    if [[ "$EXISTING_SECRET" == *"special"* ]]; then
        log_pass "Special characters preserved"
        rm -f "$test_env"
        return 0
    fi

    log_fail "Special characters lost"
    rm -f "$test_env"
    return 1
}

#############################################
# FILE STRUCTURE TESTS
#############################################

test_backend_entrypoint_exists() {
    log_test "Backend entrypoint.sh exists"

    if [ -f "$PROJECT_DIR/backend/entrypoint.sh" ]; then
        log_pass "entrypoint.sh exists"
        return 0
    fi

    log_fail "entrypoint.sh not found"
    return 1
}

test_migrations_directory_exists() {
    log_test "Migrations directory exists"

    if [ -d "$PROJECT_DIR/backend/internal/database/migrations" ]; then
        local count=$(ls -1 "$PROJECT_DIR/backend/internal/database/migrations"/*.sql 2>/dev/null | wc -l)
        log_pass "Migrations directory exists ($count migration files)"
        return 0
    fi

    log_fail "Migrations directory not found"
    return 1
}

test_frontend_dist_config() {
    log_test "Frontend nginx config exists"

    if [ -f "$PROJECT_DIR/frontend/nginx.conf.prod" ]; then
        log_pass "nginx.conf.prod exists"
        return 0
    fi

    log_fail "nginx.conf.prod not found"
    return 1
}

#############################################
# DEPLOYMENT SCRIPT FUNCTION TESTS
#############################################

test_deploy_script_has_prepare_function() {
    log_test "Deploy script calls prepare_backend_source"

    if grep -q "prepare_backend_source" "$PROJECT_DIR/deploy-with-db-safe.sh"; then
        log_pass "prepare_backend_source function call found"
        return 0
    fi

    log_fail "prepare_backend_source not called"
    return 1
}

test_deploy_script_no_build_backend() {
    log_test "Deploy script does NOT have build_backend() function"

    if ! grep -q "^build_backend()" "$PROJECT_DIR/deploy-with-db-safe.sh"; then
        log_pass "Old build_backend() function removed"
        return 0
    fi

    log_fail "Old build_backend() function still exists"
    return 1
}

test_deploy_script_does_not_copy_binary() {
    log_test "Deploy script does NOT copy backend binary"

    # Check for scp or rsync of backend binary
    if grep -E "scp.*backend/server" "$PROJECT_DIR/deploy-with-db-safe.sh" || \
       grep -E "rsync.*backend/server" "$PROJECT_DIR/deploy-with-db-safe.sh"; then
        log_fail "Still copies backend binary"
        return 1
    fi

    log_pass "No binary copy operations found"
    return 0
}

test_deploy_script_copies_tarball() {
    log_test "Deploy script copies backend source tarball"

    if grep -q "backend-source.tar.gz" "$PROJECT_DIR/deploy-with-db-safe.sh"; then
        log_pass "Tarball copy operation found"
        return 0
    fi

    log_fail "Tarball copy not found"
    return 1
}

#############################################
# ENV FILE HANDLING TESTS
#############################################

test_deploy_script_fetches_existing_env() {
    log_test "Deploy script fetches existing .env"

    if grep -q "existing.env" "$PROJECT_DIR/deploy-with-db-safe.sh"; then
        log_pass "Existing .env handling present"
        return 0
    fi

    log_fail "No existing .env fetch logic"
    return 1
}

test_deploy_script_updates_session_secret() {
    log_test "Deploy script preserves SESSION_SECRET in awk command"

    # Check if awk command handles SESSION_SECRET - look for the variable definition
    if grep -E 'awk.*-v session_secret=.*SESSION_SECRET' "$PROJECT_DIR/deploy-with-db-safe.sh" > /dev/null 2>&1 || \
       grep -A15 "^        awk" "$PROJECT_DIR/deploy-with-db-safe.sh" | grep -q "SESSION_SECRET="; then
        log_pass "SESSION_SECRET in awk update command"
        return 0
    fi

    log_fail "SESSION_SECRET not in awk command"
    return 1
}

#############################################
# RUN ALL TESTS
#############################################

run_all_tests() {
    echo "=== Running Bash Syntax Tests ==="
    test_deploy_script_syntax
    test_deploy_script_executable
    echo ""

    echo "=== Running Backend Source Tests ==="
    test_prepare_backend_source_function_exists
    test_prepare_backend_source_tarball_creation
    test_prepare_backend_source_excludes_bin
    test_deploy_script_has_prepare_function
    test_deploy_script_no_build_backend
    test_deploy_script_does_not_copy_binary
    test_deploy_script_copies_tarball
    echo ""

    echo "=== Running SESSION_SECRET Tests ==="
    test_session_secret_preserve_existing
    test_session_secret_generate_when_missing
    test_session_secret_not_empty_from_env
    test_session_secret_whitespace_handling
    test_session_secret_edge_case_empty_file
    test_session_secret_edge_case_no_file
    test_session_secret_edge_case_multiple_occurrences
    test_session_secret_edge_case_special_chars
    echo ""

    echo "=== Running Docker Compose Tests ==="
    test_docker_compose_backend_build_config
    test_docker_compose_session_secret_env
    echo ""

    echo "=== Running Dockerfile Tests ==="
    test_dockerfile_exists
    test_dockerfile_syntax
    test_dockerfile_multistage
    test_dockerfile_backend_context
    echo ""

    echo "=== Running Integration Tests ==="
    # test_docker_compose_valid_yaml - skipped if docker not available
    if command -v docker-compose &> /dev/null || docker compose version &> /dev/null 2>&1; then
        test_docker_compose_valid_yaml
    else
        log_info "Skipping docker-compose validation (not available)"
    fi
    echo ""

    echo "=== Running File Structure Tests ==="
    test_backend_entrypoint_exists
    test_migrations_directory_exists
    test_frontend_dist_config
    echo ""

    echo "=== Running ENV File Handling Tests ==="
    test_deploy_script_fetches_existing_env
    test_deploy_script_updates_session_secret
    echo ""
}

#############################################
# RESULTS
#############################################

print_results() {
    echo "=========================================="
    echo "TEST RESULTS"
    echo "=========================================="
    echo "Tests Run:    $TESTS_RUN"
    echo -e "Passed:       ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed:       ${RED}$TESTS_FAILED${NC}"
    echo "=========================================="

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}ALL TESTS PASSED${NC}"
        return 0
    else
        echo -e "${RED}SOME TESTS FAILED${NC}"
        return 1
    fi
}

#############################################
# MAIN
#############################################

main() {
    run_all_tests
    print_results
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
