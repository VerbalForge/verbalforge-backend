#!/bin/bash

# HTTPS Setup Script for VerbalForge Backend
# This script sets up Nginx as a reverse proxy with Let's Encrypt SSL

set -e

echo "🔧 Setting up HTTPS for VerbalForge Backend..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Variables
DOMAIN="${1:-api.verbalforge.com}"  # Replace with your actual domain
EMAIL="${2:-admin@verbalforge.com}"  # Replace with your email
BACKEND_PORT="8080"

echo "📝 Configuration:"
echo "   Domain: $DOMAIN"
echo "   Email: $EMAIL"
echo "   Backend Port: $BACKEND_PORT"
echo ""

# Update system
echo "📦 Updating system packages..."
apt-get update
apt-get upgrade -y

# Install Nginx
echo "🌐 Installing Nginx..."
apt-get install -y nginx

# Install Certbot for Let's Encrypt
echo "🔐 Installing Certbot..."
apt-get install -y certbot python3-certbot-nginx

# Stop Nginx temporarily
systemctl stop nginx

# Create Nginx configuration
echo "⚙️  Creating Nginx configuration..."
cat > /etc/nginx/sites-available/verbalforge-backend <<EOF
server {
    listen 80;
    server_name $DOMAIN;

    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name $DOMAIN;

    # SSL certificates (will be added by certbot)
    ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Proxy settings
    location / {
        proxy_pass http://localhost:$BACKEND_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # CORS headers (if needed)
        add_header Access-Control-Allow-Origin "https://black-dune-04f31470f.3.azurestaticapps.net" always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS, PATCH" always;
        add_header Access-Control-Allow-Headers "Origin, Content-Type, Authorization" always;
        add_header Access-Control-Allow-Credentials "true" always;
        
        # Handle preflight requests
        if (\$request_method = 'OPTIONS') {
            return 204;
        }
    }

    # File upload size limit
    client_max_body_size 10M;

    # Logging
    access_log /var/log/nginx/verbalforge-backend-access.log;
    error_log /var/log/nginx/verbalforge-backend-error.log;
}
EOF

# Enable the site
ln -sf /etc/nginx/sites-available/verbalforge-backend /etc/nginx/sites-enabled/

# Remove default site
rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
echo "✅ Testing Nginx configuration..."
nginx -t

# Obtain SSL certificate
echo "🔐 Obtaining SSL certificate from Let's Encrypt..."
echo "⚠️  Make sure DNS is pointing to this server!"
read -p "Press Enter to continue..."

certbot --nginx -d $DOMAIN --non-interactive --agree-tos -m $EMAIL

# Start Nginx
echo "🚀 Starting Nginx..."
systemctl start nginx
systemctl enable nginx

# Setup automatic certificate renewal
echo "⏰ Setting up automatic certificate renewal..."
systemctl enable certbot.timer
systemctl start certbot.timer

echo ""
echo "✅ HTTPS setup complete!"
echo ""
echo "📋 Next steps:"
echo "   1. Update DNS to point $DOMAIN to this server IP: $(curl -s ifconfig.me)"
echo "   2. Update GitHub secret NEXT_PUBLIC_API_URL to: https://$DOMAIN"
echo "   3. Update backend FRONTEND_URL env var to: https://black-dune-04f31470f.3.azurestaticapps.net"
echo ""
echo "🔍 Check status:"
echo "   systemctl status nginx"
echo "   certbot certificates"
echo ""
