#!/bin/bash

################################################################################
# Tutoring Platform - Production Deployment Script
#
# This script automates the complete deployment of the Tutoring Platform
# on a fresh Ubuntu 22.04+ server.
#
# Usage: sudo ./setup.sh
################################################################################

set -e  # Exit immediately if a command exits with a non-zero status

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Error handler
error_handler() {
    log_error "An error occurred on line $1"
    log_error "Deployment failed. Please check the logs above."
    exit 1
}

trap 'error_handler $LINENO' ERR

################################################################################
# Pre-flight checks
################################################################################

log_info "Starting Tutoring Platform deployment..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (use sudo)"
    exit 1
fi

# Check OS version
if ! grep -q "Ubuntu" /etc/os-release; then
    log_warning "This script is designed for Ubuntu. Proceed with caution."
fi

################################################################################
# User input collection
################################################################################

log_info "Collecting deployment configuration..."

read -p "Enter domain name (e.g., tutoring.example.com): " DOMAIN_NAME
if [ -z "$DOMAIN_NAME" ]; then
    log_error "Domain name is required"
    exit 1
fi

read -sp "Enter PostgreSQL password for tutoring_user: " DB_PASSWORD
echo
if [ -z "$DB_PASSWORD" ]; then
    log_error "Database password is required"
    exit 1
fi

read -p "Enter email for SSL certificate (for Let's Encrypt): " SSL_EMAIL
if [ -z "$SSL_EMAIL" ]; then
    log_error "Email is required for SSL certificate"
    exit 1
fi

# Generate secure session secret
SESSION_SECRET=$(openssl rand -base64 48)

log_success "Configuration collected"

################################################################################
# System update
################################################################################

log_info "Updating system packages..."
apt update -qq
apt upgrade -y -qq
log_success "System updated"

################################################################################
# Install dependencies
################################################################################

log_info "Installing system dependencies..."

# Install basic tools
apt install -y -qq curl wget git ufw software-properties-common gnupg2 \
    build-essential apt-transport-https ca-certificates

log_success "Basic tools installed"

# Install PostgreSQL 15
log_info "Installing PostgreSQL 15..."
sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt update -qq
apt install -y -qq postgresql-15 postgresql-contrib-15
systemctl enable postgresql
systemctl start postgresql
log_success "PostgreSQL 15 installed"

# Install Nginx
log_info "Installing Nginx..."
apt install -y -qq nginx
systemctl enable nginx
log_success "Nginx installed"

# Install Certbot
log_info "Installing Certbot..."
apt install -y -qq certbot python3-certbot-nginx
log_success "Certbot installed"

# Install Go 1.21+
log_info "Installing Go..."
GO_VERSION="1.21.5"
wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
rm go${GO_VERSION}.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
log_success "Go $GO_VERSION installed"

# Install Node.js 20+
log_info "Installing Node.js 20..."
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt install -y -qq nodejs
log_success "Node.js $(node -v) installed"

################################################################################
# Configure PostgreSQL
################################################################################

log_info "Configuring PostgreSQL..."

# Create database and user
sudo -u postgres psql <<EOF
-- Create user
CREATE USER tutoring_user WITH PASSWORD '$DB_PASSWORD';

-- Create database
CREATE DATABASE tutoring_platform OWNER tutoring_user;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE tutoring_platform TO tutoring_user;

-- Connect to database and grant schema privileges
\c tutoring_platform
GRANT ALL ON SCHEMA public TO tutoring_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO tutoring_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO tutoring_user;
EOF

log_success "PostgreSQL configured"

################################################################################
# Deploy application
################################################################################

log_info "Deploying application to /opt/tutoring-platform..."

# Create deployment directory
mkdir -p /opt/tutoring-platform
cd /opt/tutoring-platform

# Copy project files (assuming running from project directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
log_info "Copying files from $SCRIPT_DIR..."

# Copy backend
if [ -d "$SCRIPT_DIR/backend" ]; then
    cp -r "$SCRIPT_DIR/backend" /opt/tutoring-platform/
    log_success "Backend files copied"
else
    log_error "Backend directory not found at $SCRIPT_DIR/backend"
    exit 1
fi

# Copy frontend
if [ -d "$SCRIPT_DIR/frontend" ]; then
    cp -r "$SCRIPT_DIR/frontend" /opt/tutoring-platform/
    log_success "Frontend files copied"
else
    log_error "Frontend directory not found at $SCRIPT_DIR/frontend"
    exit 1
fi

################################################################################
# Build backend
################################################################################

log_info "Building backend..."

cd /opt/tutoring-platform/backend

# Create .env file
cat > .env <<EOF
# Environment
ENV=production

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=tutoring_platform
DB_USER=tutoring_user
DB_PASSWORD=$DB_PASSWORD
DB_SSL_MODE=disable

# Server
SERVER_PORT=8080

# Session
SESSION_SECRET=$SESSION_SECRET
SESSION_MAX_AGE=86400
SESSION_SAME_SITE=Strict
EOF

chmod 600 .env
log_success "Backend .env file created"

# Download Go dependencies
log_info "Downloading Go dependencies..."
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
go mod download

# Build application
log_info "Compiling Go application..."
mkdir -p bin
go build -o bin/server cmd/server/main.go

if [ ! -f bin/server ]; then
    log_error "Failed to build backend server"
    exit 1
fi

chmod +x bin/server
log_success "Backend built successfully"

################################################################################
# Apply database migrations
################################################################################

log_info "Applying database migrations..."

# Check if migrations directory exists
if [ -d "/opt/tutoring-platform/backend/migrations" ]; then
    # Install golang-migrate if not present
    if ! command -v migrate &> /dev/null; then
        log_info "Installing golang-migrate..."
        curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
        mv migrate /usr/local/bin/
        chmod +x /usr/local/bin/migrate
    fi

    # Run migrations
    migrate -path /opt/tutoring-platform/backend/migrations \
            -database "postgres://tutoring_user:$DB_PASSWORD@localhost:5432/tutoring_platform?sslmode=disable" \
            up

    log_success "Database migrations applied"
else
    log_warning "No migrations directory found, skipping migrations"
fi

################################################################################
# Build frontend
################################################################################

log_info "Building frontend..."

cd /opt/tutoring-platform/frontend

# Install dependencies
log_info "Installing npm dependencies..."
npm install --production

# Build frontend
log_info "Building React application..."
npm run build

if [ ! -d "build" ]; then
    log_error "Frontend build failed"
    exit 1
fi

log_success "Frontend built successfully"

################################################################################
# Configure Nginx
################################################################################

log_info "Configuring Nginx..."

# Create Nginx configuration
cat > /etc/nginx/sites-available/tutoring-platform <<EOF
server {
    listen 80;
    server_name $DOMAIN_NAME;

    # Frontend static files
    location / {
        root /opt/tutoring-platform/frontend/build;
        try_files \$uri \$uri/ /index.html;

        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # Cookie settings
        proxy_cookie_path / "/; HTTPOnly; Secure; SameSite=Strict";

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://localhost:8080/health;
        access_log off;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Disable server tokens
    server_tokens off;

    # Max upload size
    client_max_body_size 10M;
}
EOF

# Enable site
ln -sf /etc/nginx/sites-available/tutoring-platform /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
nginx -t

# Reload Nginx
systemctl reload nginx

log_success "Nginx configured"

################################################################################
# Create systemd service
################################################################################

log_info "Creating systemd service..."

# Create log directory
mkdir -p /var/log/tutoring-platform
chown www-data:www-data /var/log/tutoring-platform

# Create systemd service file
cat > /etc/systemd/system/tutoring-platform.service <<EOF
[Unit]
Description=Tutoring Platform Backend Service
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/tutoring-platform/backend
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"
EnvironmentFile=/opt/tutoring-platform/backend/.env
ExecStart=/opt/tutoring-platform/backend/bin/server
Restart=always
RestartSec=10
StandardOutput=append:/var/log/tutoring-platform/access.log
StandardError=append:/var/log/tutoring-platform/error.log

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/tutoring-platform

[Install]
WantedBy=multi-user.target
EOF

# Set ownership
chown -R www-data:www-data /opt/tutoring-platform

# Reload systemd
systemctl daemon-reload

# Enable and start service
systemctl enable tutoring-platform
systemctl start tutoring-platform

# Wait for service to start
sleep 3

# Check if service is running
if systemctl is-active --quiet tutoring-platform; then
    log_success "Tutoring Platform service started"
else
    log_error "Failed to start Tutoring Platform service"
    systemctl status tutoring-platform
    exit 1
fi

################################################################################
# Configure SSL with Let's Encrypt
################################################################################

log_info "Configuring SSL certificate..."

# Ensure domain resolves to this server
log_warning "Make sure $DOMAIN_NAME points to this server's IP address"
read -p "Press Enter when DNS is configured and you're ready to obtain SSL certificate..."

# Obtain SSL certificate
certbot --nginx -d $DOMAIN_NAME --non-interactive --agree-tos -m $SSL_EMAIL --redirect

if [ $? -eq 0 ]; then
    log_success "SSL certificate obtained and configured"
else
    log_warning "SSL certificate setup failed. You can run it manually later with: certbot --nginx -d $DOMAIN_NAME"
fi

# Setup automatic renewal
systemctl enable certbot.timer
systemctl start certbot.timer

log_success "SSL auto-renewal configured"

################################################################################
# Configure firewall
################################################################################

log_info "Configuring firewall..."

# Reset UFW to default
ufw --force reset

# Default policies
ufw default deny incoming
ufw default allow outgoing

# Allow SSH (important!)
ufw allow 22/tcp comment 'SSH'

# Allow HTTP and HTTPS
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'

# Enable firewall
ufw --force enable

log_success "Firewall configured"

################################################################################
# Setup monitoring and backups
################################################################################

log_info "Setting up monitoring and backups..."

# Copy monitoring scripts
cp "$SCRIPT_DIR/deploy/monitoring/healthcheck.sh" /opt/tutoring-platform/
cp "$SCRIPT_DIR/deploy/backup.sh" /opt/tutoring-platform/
chmod +x /opt/tutoring-platform/*.sh

# Create backup directory
mkdir -p /var/backups/tutoring-platform
chown www-data:www-data /var/backups/tutoring-platform

# Setup cron jobs
(crontab -l 2>/dev/null; echo "*/5 * * * * /opt/tutoring-platform/healthcheck.sh >> /var/log/tutoring-platform/healthcheck.log 2>&1") | crontab -
(crontab -l 2>/dev/null; echo "0 2 * * * /opt/tutoring-platform/backup.sh >> /var/log/tutoring-platform/backup.log 2>&1") | crontab -

log_success "Monitoring and backups configured"

################################################################################
# Final health checks
################################################################################

log_info "Running final health checks..."

# Check PostgreSQL
if systemctl is-active --quiet postgresql; then
    log_success "PostgreSQL is running"
else
    log_error "PostgreSQL is not running"
fi

# Check Nginx
if systemctl is-active --quiet nginx; then
    log_success "Nginx is running"
else
    log_error "Nginx is not running"
fi

# Check backend service
if systemctl is-active --quiet tutoring-platform; then
    log_success "Backend service is running"
else
    log_error "Backend service is not running"
fi

# Check health endpoint
sleep 2
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log_success "Health endpoint responding"
else
    log_warning "Health endpoint not responding"
fi

################################################################################
# Display final information
################################################################################

echo ""
echo "================================================================================"
echo -e "${GREEN}Tutoring Platform deployed successfully!${NC}"
echo "================================================================================"
echo ""
echo "Application URL: https://$DOMAIN_NAME"
echo "Health check: https://$DOMAIN_NAME/health"
echo ""
echo "Service Management:"
echo "  - Start:   sudo systemctl start tutoring-platform"
echo "  - Stop:    sudo systemctl stop tutoring-platform"
echo "  - Restart: sudo systemctl restart tutoring-platform"
echo "  - Status:  sudo systemctl status tutoring-platform"
echo ""
echo "Logs:"
echo "  - Access:  /var/log/tutoring-platform/access.log"
echo "  - Error:   /var/log/tutoring-platform/error.log"
echo "  - Nginx:   /var/log/nginx/"
echo ""
echo "Backups:"
echo "  - Location: /var/backups/tutoring-platform"
echo "  - Schedule: Daily at 2:00 AM"
echo ""
echo "SSL Certificate:"
echo "  - Auto-renewal enabled via certbot"
echo "  - Renews automatically before expiration"
echo ""
echo "Database:"
echo "  - Name: tutoring_platform"
echo "  - User: tutoring_user"
echo "  - Host: localhost:5432"
echo ""
echo "Next steps:"
echo "  1. Create an admin user in the database"
echo "  2. Test the application at https://$DOMAIN_NAME"
echo "  3. Review logs to ensure everything is working"
echo "  4. Setup monitoring alerts (optional)"
echo ""
echo "================================================================================"
echo ""

log_success "Deployment complete!"
