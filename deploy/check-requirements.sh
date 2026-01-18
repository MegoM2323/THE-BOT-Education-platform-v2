#!/bin/bash

################################################################################
# Requirements Check Script
#
# This script verifies that your server meets all requirements
# before running the deployment
#
# Usage: ./check-requirements.sh
################################################################################

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

################################################################################
# Start checks
################################################################################

echo ""
echo "================================================================================"
echo "  Tutoring Platform - Requirements Check"
echo "================================================================================"
echo ""

PASSED=0
WARNINGS=0
FAILED=0

################################################################################
# Check OS
################################################################################

log_info "Checking operating system..."

if [ -f /etc/os-release ]; then
    . /etc/os-release
    if [[ "$ID" == "ubuntu" ]]; then
        VERSION_NUMBER=$(echo "$VERSION_ID" | cut -d. -f1)
        if [ "$VERSION_NUMBER" -ge 22 ]; then
            log_success "OS: $PRETTY_NAME"
            ((PASSED++))
        else
            log_warning "OS: $PRETTY_NAME (Ubuntu 22.04+ recommended)"
            ((WARNINGS++))
        fi
    else
        log_warning "OS: $PRETTY_NAME (Ubuntu recommended)"
        ((WARNINGS++))
    fi
else
    log_error "Unable to determine OS version"
    ((FAILED++))
fi

################################################################################
# Check root/sudo access
################################################################################

log_info "Checking privileges..."

if [ "$EUID" -eq 0 ]; then
    log_success "Running as root"
    ((PASSED++))
elif groups | grep -q sudo; then
    log_success "User has sudo privileges"
    ((PASSED++))
else
    log_error "No root or sudo access"
    ((FAILED++))
fi

################################################################################
# Check RAM
################################################################################

log_info "Checking system memory..."

TOTAL_RAM=$(free -g | awk '/^Mem:/{print $2}')
if [ "$TOTAL_RAM" -ge 4 ]; then
    log_success "RAM: ${TOTAL_RAM}GB (Excellent)"
    ((PASSED++))
elif [ "$TOTAL_RAM" -ge 2 ]; then
    log_success "RAM: ${TOTAL_RAM}GB (Acceptable)"
    ((PASSED++))
else
    log_warning "RAM: ${TOTAL_RAM}GB (Minimum 2GB recommended)"
    ((WARNINGS++))
fi

################################################################################
# Check disk space
################################################################################

log_info "Checking disk space..."

AVAILABLE_SPACE=$(df -BG / | awk 'NR==2 {print $4}' | sed 's/G//')
if [ "$AVAILABLE_SPACE" -ge 50 ]; then
    log_success "Disk space: ${AVAILABLE_SPACE}GB available (Excellent)"
    ((PASSED++))
elif [ "$AVAILABLE_SPACE" -ge 20 ]; then
    log_success "Disk space: ${AVAILABLE_SPACE}GB available (Acceptable)"
    ((PASSED++))
else
    log_warning "Disk space: ${AVAILABLE_SPACE}GB available (20GB+ recommended)"
    ((WARNINGS++))
fi

################################################################################
# Check CPU cores
################################################################################

log_info "Checking CPU cores..."

CPU_CORES=$(nproc)
if [ "$CPU_CORES" -ge 2 ]; then
    log_success "CPU cores: $CPU_CORES"
    ((PASSED++))
else
    log_warning "CPU cores: $CPU_CORES (2+ recommended)"
    ((WARNINGS++))
fi

################################################################################
# Check internet connectivity
################################################################################

log_info "Checking internet connectivity..."

if ping -c 1 google.com &> /dev/null; then
    log_success "Internet connectivity: OK"
    ((PASSED++))
else
    log_error "Internet connectivity: FAILED"
    ((FAILED++))
fi

################################################################################
# Check DNS resolution
################################################################################

log_info "Checking DNS resolution..."

if nslookup google.com &> /dev/null; then
    log_success "DNS resolution: OK"
    ((PASSED++))
else
    log_error "DNS resolution: FAILED"
    ((FAILED++))
fi

################################################################################
# Check required ports
################################################################################

log_info "Checking required ports availability..."

PORTS_OK=true

check_port() {
    PORT=$1
    SERVICE=$2

    if ss -tuln | grep -q ":$PORT "; then
        log_warning "Port $PORT ($SERVICE): Already in use"
        ((WARNINGS++))
        PORTS_OK=false
    else
        log_success "Port $PORT ($SERVICE): Available"
        ((PASSED++))
    fi
}

check_port 80 "HTTP"
check_port 443 "HTTPS"
check_port 8080 "Backend"
check_port 5432 "PostgreSQL"

################################################################################
# Check if required commands exist
################################################################################

log_info "Checking for conflicting installations..."

check_command() {
    CMD=$1
    SERVICE=$2

    if command -v "$CMD" &> /dev/null; then
        log_warning "$SERVICE already installed"
        ((WARNINGS++))
    else
        log_success "$SERVICE: Not installed (will be installed)"
        ((PASSED++))
    fi
}

# These will be installed by setup script
check_command "psql" "PostgreSQL"
check_command "nginx" "Nginx"
check_command "go" "Go"
check_command "node" "Node.js"

################################################################################
# Check for required system tools
################################################################################

log_info "Checking system tools..."

check_tool() {
    CMD=$1

    if command -v "$CMD" &> /dev/null; then
        log_success "$CMD: Available"
        ((PASSED++))
    else
        log_warning "$CMD: Not found (will be installed)"
        ((WARNINGS++))
    fi
}

check_tool "curl"
check_tool "wget"
check_tool "git"

################################################################################
# Check systemd
################################################################################

log_info "Checking systemd..."

if command -v systemctl &> /dev/null; then
    log_success "systemd: Available"
    ((PASSED++))
else
    log_error "systemd: Not found (required)"
    ((FAILED++))
fi

################################################################################
# Check if project files exist
################################################################################

log_info "Checking project files..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -d "$SCRIPT_DIR/backend" ]; then
    log_success "Backend directory found"
    ((PASSED++))
else
    log_error "Backend directory not found at $SCRIPT_DIR/backend"
    ((FAILED++))
fi

if [ -d "$SCRIPT_DIR/frontend" ]; then
    log_success "Frontend directory found"
    ((PASSED++))
else
    log_error "Frontend directory not found at $SCRIPT_DIR/frontend"
    ((FAILED++))
fi

if [ -f "$SCRIPT_DIR/backend/go.mod" ]; then
    log_success "Go module file found"
    ((PASSED++))
else
    log_error "go.mod not found in backend directory"
    ((FAILED++))
fi

if [ -f "$SCRIPT_DIR/frontend/package.json" ]; then
    log_success "Package.json found"
    ((PASSED++))
else
    log_error "package.json not found in frontend directory"
    ((FAILED++))
fi

################################################################################
# Check timezone
################################################################################

log_info "Checking timezone configuration..."

TIMEZONE=$(timedatectl | grep "Time zone" | awk '{print $3}')
log_success "Timezone: $TIMEZONE"
((PASSED++))

################################################################################
# Summary
################################################################################

echo ""
echo "================================================================================"
echo "  Requirements Check Summary"
echo "================================================================================"
echo ""
echo -e "Passed:   ${GREEN}$PASSED${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
echo -e "Failed:   ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}RESULT: FAILED${NC}"
    echo ""
    echo "Please fix the failed checks before proceeding with deployment."
    echo ""
    exit 1
elif [ $WARNINGS -gt 5 ]; then
    echo -e "${YELLOW}RESULT: WARNINGS${NC}"
    echo ""
    echo "Your system meets minimum requirements but has some warnings."
    echo "Review the warnings above before proceeding."
    echo ""
    read -p "Do you want to continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}RESULT: PASSED${NC}"
    echo ""
    echo "Your system meets all requirements for deployment!"
    echo ""
fi

echo "Next steps:"
echo "  1. Review any warnings above"
echo "  2. Ensure your domain DNS is configured"
echo "  3. Run: sudo ./setup.sh"
echo ""
echo "================================================================================"
echo ""

exit 0
