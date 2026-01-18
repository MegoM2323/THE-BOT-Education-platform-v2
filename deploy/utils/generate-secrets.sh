#!/bin/bash

################################################################################
# Secret Generation Utility
#
# This script generates secure random secrets for use in configuration
#
# Usage: ./generate-secrets.sh
################################################################################

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}================================================================================"
echo "  Tutoring Platform - Secret Generator"
echo "================================================================================${NC}"
echo ""

################################################################################
# Generate Database Password
################################################################################

echo -e "${BLUE}Database Password:${NC}"
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
echo "$DB_PASSWORD"
echo ""

################################################################################
# Generate Session Secret
################################################################################

echo -e "${BLUE}Session Secret (minimum 32 characters):${NC}"
SESSION_SECRET=$(openssl rand -base64 48)
echo "$SESSION_SECRET"
echo ""

################################################################################
# Generate Admin Password Hash
################################################################################

echo -e "${BLUE}Generate Admin Password Hash${NC}"
read -sp "Enter admin password: " ADMIN_PASSWORD
echo ""

if command -v htpasswd &> /dev/null; then
    PASSWORD_HASH=$(echo "$ADMIN_PASSWORD" | htpasswd -bnBC 10 "" password | tr -d ':\n' | sed 's/password://')
    echo -e "${BLUE}Bcrypt Hash:${NC}"
    echo "$PASSWORD_HASH"
    echo ""
elif command -v python3 &> /dev/null; then
    echo -e "${GREEN}Using Python to generate hash...${NC}"
    PASSWORD_HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw('$ADMIN_PASSWORD'.encode(), bcrypt.gensalt(10)).decode())")
    echo -e "${BLUE}Bcrypt Hash:${NC}"
    echo "$PASSWORD_HASH"
    echo ""
else
    echo "Neither htpasswd nor python3-bcrypt found. Install one to generate password hash."
    echo "Install: sudo apt install apache2-utils  (for htpasswd)"
    echo "    or: pip3 install bcrypt  (for python3-bcrypt)"
fi

################################################################################
# Generate .env Template
################################################################################

echo -e "${BLUE}.env File Template:${NC}"
echo ""

cat <<EOF
# Copy this to /opt/tutoring-platform/backend/.env

ENV=production

DB_HOST=localhost
DB_PORT=5432
DB_NAME=tutoring_platform
DB_USER=tutoring
DB_PASSWORD=$DB_PASSWORD
DB_SSL_MODE=disable

SERVER_PORT=8080

SESSION_SECRET=$SESSION_SECRET
SESSION_MAX_AGE=86400
SESSION_SAME_SITE=Strict
EOF

echo ""

################################################################################
# Generate SQL for Admin User
################################################################################

if [ -n "$PASSWORD_HASH" ]; then
    echo -e "${BLUE}SQL to Create Admin User:${NC}"
    echo ""
    cat <<EOF
-- Run this in PostgreSQL to create admin user:
-- sudo -u postgres psql tutoring_platform

INSERT INTO users (
    email,
    password_hash,
    first_name,
    last_name,
    role,
    is_active,
    created_at,
    updated_at
) VALUES (
    'admin@example.com',
    '$PASSWORD_HASH',
    'Admin',
    'User',
    'admin',
    true,
    NOW(),
    NOW()
);
EOF
    echo ""
fi

################################################################################
# Security Notes
################################################################################

echo -e "${CYAN}================================================================================"
echo "  Security Notes"
echo "================================================================================${NC}"
echo ""
echo "1. Store these secrets securely (password manager, encrypted file, etc.)"
echo "2. Never commit these secrets to version control"
echo "3. Use different secrets for development and production"
echo "4. Rotate secrets periodically"
echo "5. Restrict .env file permissions: chmod 600 /opt/tutoring-platform/backend/.env"
echo ""
echo "================================================================================"
echo ""
