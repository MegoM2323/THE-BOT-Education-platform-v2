#!/bin/bash

################################################################################
# SSL Certificate Setup Script
#
# This script obtains and configures SSL certificates using Let's Encrypt
# for the Tutoring Platform
#
# Usage: sudo ./certbot-setup.sh <domain> <email>
################################################################################

set -e

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
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (use sudo)"
    exit 1
fi

# Parse arguments
DOMAIN=$1
EMAIL=$2

if [ -z "$DOMAIN" ] || [ -z "$EMAIL" ]; then
    log_error "Usage: $0 <domain> <email>"
    echo "Example: $0 tutoring.example.com admin@example.com"
    exit 1
fi

log_info "Setting up SSL certificate for $DOMAIN"

################################################################################
# Install Certbot if not present
################################################################################

if ! command -v certbot &> /dev/null; then
    log_info "Installing Certbot..."
    apt update -qq
    apt install -y -qq certbot python3-certbot-nginx
    log_success "Certbot installed"
else
    log_info "Certbot already installed"
fi

################################################################################
# Verify DNS configuration
################################################################################

log_info "Verifying DNS configuration for $DOMAIN..."

# Get server's public IP
SERVER_IP=$(curl -s ifconfig.me)
log_info "Server IP: $SERVER_IP"

# Resolve domain
DOMAIN_IP=$(dig +short $DOMAIN | tail -n1)

if [ -z "$DOMAIN_IP" ]; then
    log_warning "Could not resolve $DOMAIN"
    log_warning "Make sure DNS is configured correctly"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
elif [ "$DOMAIN_IP" != "$SERVER_IP" ]; then
    log_warning "Domain $DOMAIN points to $DOMAIN_IP, but server IP is $SERVER_IP"
    log_warning "Make sure DNS is configured correctly"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    log_success "DNS configured correctly"
fi

################################################################################
# Check Nginx configuration
################################################################################

log_info "Checking Nginx configuration..."

if ! nginx -t &> /dev/null; then
    log_error "Nginx configuration is invalid"
    nginx -t
    exit 1
fi

log_success "Nginx configuration is valid"

################################################################################
# Obtain SSL certificate
################################################################################

log_info "Obtaining SSL certificate from Let's Encrypt..."

# Check if certificate already exists
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    log_warning "Certificate for $DOMAIN already exists"
    read -p "Renew certificate? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        certbot renew --cert-name $DOMAIN --nginx --non-interactive
        log_success "Certificate renewed"
    else
        log_info "Skipping certificate renewal"
    fi
else
    # Obtain new certificate
    certbot --nginx \
        -d $DOMAIN \
        --non-interactive \
        --agree-tos \
        -m $EMAIL \
        --redirect \
        --hsts \
        --staple-ocsp

    if [ $? -eq 0 ]; then
        log_success "SSL certificate obtained and configured"
    else
        log_error "Failed to obtain SSL certificate"
        exit 1
    fi
fi

################################################################################
# Verify certificate
################################################################################

log_info "Verifying SSL certificate..."

certbot certificates | grep -A 3 $DOMAIN

################################################################################
# Configure automatic renewal
################################################################################

log_info "Configuring automatic renewal..."

# Enable certbot timer
systemctl enable certbot.timer
systemctl start certbot.timer

# Check timer status
if systemctl is-active --quiet certbot.timer; then
    log_success "Certbot renewal timer enabled"
else
    log_warning "Failed to enable certbot timer"
fi

# Add renewal hook to reload nginx
cat > /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh <<'EOF'
#!/bin/bash
systemctl reload nginx
EOF

chmod +x /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh

log_success "Nginx reload hook configured"

################################################################################
# Test automatic renewal
################################################################################

log_info "Testing automatic renewal (dry run)..."

certbot renew --dry-run

if [ $? -eq 0 ]; then
    log_success "Automatic renewal test passed"
else
    log_warning "Automatic renewal test failed"
fi

################################################################################
# Display certificate information
################################################################################

echo ""
echo "================================================================================"
echo -e "${GREEN}SSL Certificate Setup Complete${NC}"
echo "================================================================================"
echo ""
echo "Domain: $DOMAIN"
echo "Certificate location: /etc/letsencrypt/live/$DOMAIN"
echo ""
echo "Certificate details:"
certbot certificates --cert-name $DOMAIN
echo ""
echo "Auto-renewal:"
echo "  - Certbot timer: $(systemctl is-active certbot.timer)"
echo "  - Next run: $(systemctl list-timers certbot.timer --no-pager | grep certbot)"
echo ""
echo "To manually renew:"
echo "  sudo certbot renew --cert-name $DOMAIN"
echo ""
echo "To test renewal:"
echo "  sudo certbot renew --dry-run"
echo ""
echo "================================================================================"
echo ""

log_success "SSL setup complete!"
