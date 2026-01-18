#!/bin/bash

set -euo pipefail

REMOTE_USER="mg"
REMOTE_HOST="5.129.249.206"
REMOTE_ADDR="${REMOTE_USER}@${REMOTE_HOST}"
THEBOT_HOME="/home/mg/the-bot"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0

check_result() {
    local name=$1
    local result=$2
    local details=${3:-}

    if [ "$result" = "pass" ]; then
        echo -e "  ${GREEN}✓${NC} $name"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    elif [ "$result" = "fail" ]; then
        echo -e "  ${RED}✗${NC} $name"
        [ -n "$details" ] && echo -e "    ${RED}$details${NC}"
        CHECKS_FAILED=$((CHECKS_FAILED + 1))
    elif [ "$result" = "warn" ]; then
        echo -e "  ${YELLOW}⚠${NC} $name"
        [ -n "$details" ] && echo -e "    ${YELLOW}$details${NC}"
        CHECKS_WARNING=$((CHECKS_WARNING + 1))
    fi
}

echo ""
echo -e "${CYAN}Pre-Deployment Checks for THE_BOT V3${NC}"
echo -e "${CYAN}════════════════════════════════════════${NC}"
echo ""

echo "Connectivity Checks:"
if ssh -q -o ConnectTimeout=5 "$REMOTE_ADDR" "echo ok" > /dev/null 2>&1; then
    check_result "SSH connection to $REMOTE_ADDR" "pass"
else
    check_result "SSH connection to $REMOTE_ADDR" "fail" "Cannot reach production server"
    exit 1
fi

if ssh -q "$REMOTE_ADDR" "test -d $THEBOT_HOME" 2>/dev/null; then
    check_result "Project directory $THEBOT_HOME exists" "pass"
else
    check_result "Project directory $THEBOT_HOME exists" "fail" "Directory not found"
    exit 1
fi

echo ""
echo "System Requirements:"

DISK_USAGE=$(ssh -q "$REMOTE_ADDR" "df $THEBOT_HOME | tail -1 | awk '{print \$5}' | sed 's/%//'")
if [ "$DISK_USAGE" -lt 90 ]; then
    check_result "Disk space (${DISK_USAGE}% used)" "pass"
else
    check_result "Disk space (${DISK_USAGE}% used)" "fail" "Insufficient disk space"
fi

RAM_AVAILABLE=$(ssh -q "$REMOTE_ADDR" "free -m | grep Mem | awk '{print \$7}'")
if [ "$RAM_AVAILABLE" -gt 1024 ]; then
    check_result "Available RAM (${RAM_AVAILABLE}MB)" "pass"
elif [ "$RAM_AVAILABLE" -gt 512 ]; then
    check_result "Available RAM (${RAM_AVAILABLE}MB)" "warn" "Low RAM for compilation"
else
    check_result "Available RAM (${RAM_AVAILABLE}MB)" "fail" "Insufficient RAM"
fi

echo ""
echo "Build Tools:"

if ssh -q "$REMOTE_ADDR" "go version > /dev/null 2>&1"; then
    GO_VERSION=$(ssh -q "$REMOTE_ADDR" "go version | awk '{print \$3}'")
    check_result "Go installed ($GO_VERSION)" "pass"
else
    check_result "Go installed" "fail" "Go not found on server"
fi

if ssh -q "$REMOTE_ADDR" "node --version > /dev/null 2>&1"; then
    NODE_VERSION=$(ssh -q "$REMOTE_ADDR" "node --version")
    check_result "Node.js installed ($NODE_VERSION)" "pass"
else
    check_result "Node.js installed" "fail" "Node.js not found on server"
fi

if ssh -q "$REMOTE_ADDR" "npm --version > /dev/null 2>&1"; then
    NPM_VERSION=$(ssh -q "$REMOTE_ADDR" "npm --version")
    check_result "npm installed ($NPM_VERSION)" "pass"
else
    check_result "npm installed" "fail" "npm not found on server"
fi

echo ""
echo "Services:"

SERVICES=(
    "thebot-backend.service"
    "thebot-daphne.service"
    "thebot-celery-worker.service"
    "thebot-celery-beat.service"
)

SERVICES_OK=true
for service in "${SERVICES[@]}"; do
    if ssh -q "$REMOTE_ADDR" "systemctl list-unit-files 2>/dev/null | grep -q $service"; then
        check_result "$service registered" "pass"
    else
        check_result "$service registered" "warn" "Service file not found"
        SERVICES_OK=false
    fi
done

echo ""
echo "Database:"

if ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME/backend && psql \$DATABASE_URL -tc 'SELECT 1' > /dev/null 2>&1"; then
    check_result "PostgreSQL connectivity" "pass"
    CHAT_COUNT=$(ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME/backend && psql \$DATABASE_URL -tc \"SELECT COUNT(*) FROM chats\" 2>/dev/null")
    echo -e "    Current chats: ${CYAN}$CHAT_COUNT${NC}"
else
    check_result "PostgreSQL connectivity" "warn" "Cannot connect (may need env setup)"
fi

echo ""
echo "Cache:"

if ssh -q "$REMOTE_ADDR" "redis-cli ping > /dev/null 2>&1"; then
    check_result "Redis connectivity" "pass"
else
    check_result "Redis connectivity" "warn" "Cannot connect (may impact performance)"
fi

echo ""
echo "Git Status:"

if ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME && git status > /dev/null 2>&1"; then
    check_result "Git repository" "pass"

    CURRENT_BRANCH=$(ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME && git rev-parse --abbrev-ref HEAD")
    echo -e "    Current branch: ${CYAN}$CURRENT_BRANCH${NC}"

    COMMITS_AHEAD=$(ssh -q "$REMOTE_ADDR" "cd $THEBOT_HOME && git rev-list --count origin/master..master 2>/dev/null || echo 0")
    if [ "$COMMITS_AHEAD" -gt 0 ]; then
        check_result "Git up to date" "warn" "$COMMITS_AHEAD commits ahead of origin/master"
    else
        check_result "Git up to date" "pass"
    fi
else
    check_result "Git repository" "fail" "Not a git repository"
fi

echo ""
echo "Environment:"

if ssh -q "$REMOTE_ADDR" "test -f $THEBOT_HOME/backend/.env"; then
    check_result "Backend .env exists" "pass"
else
    check_result "Backend .env exists" "warn" ".env file not found"
fi

if ssh -q "$REMOTE_ADDR" "test -f $THEBOT_HOME/.env"; then
    check_result "Project .env exists" "pass"
else
    check_result "Project .env exists" "warn" ".env file not found"
fi

echo ""
echo "Recent Deployment History:"

LAST_LOGS=$(ls -t logs/deploy_*.log 2>/dev/null | head -3 || echo "No deployment logs found")
if [ "$LAST_LOGS" != "No deployment logs found" ]; then
    echo "$LAST_LOGS" | while read log; do
        TIMESTAMP=$(basename "$log" | sed 's/deploy_//;s/.log//')
        echo -e "  ${CYAN}$TIMESTAMP${NC}"
    done
else
    echo "  No previous deployments recorded"
fi

echo ""
echo -e "${CYAN}════════════════════════════════════════${NC}"
echo "Summary:"
echo -e "  ${GREEN}Passed: $CHECKS_PASSED${NC}"
echo -e "  ${YELLOW}Warnings: $CHECKS_WARNING${NC}"
echo -e "  ${RED}Failed: $CHECKS_FAILED${NC}"
echo ""

if [ $CHECKS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ Pre-deployment checks passed!${NC}"
    echo ""
    echo "Ready for deployment. Run:"
    echo -e "  ${CYAN}./safe-deploy-native.sh${NC}"
    echo ""
    exit 0
elif [ $CHECKS_FAILED -gt 0 ]; then
    echo -e "${RED}✗ Pre-deployment checks failed!${NC}"
    echo ""
    echo "Please fix the errors above before deploying."
    echo ""
    exit 1
else
    echo -e "${YELLOW}⚠ Pre-deployment checks passed with warnings${NC}"
    echo ""
    echo "You can proceed with deployment, but please review the warnings."
    echo -e "Run: ${CYAN}./safe-deploy-native.sh${NC}"
    echo ""
    exit 0
fi
