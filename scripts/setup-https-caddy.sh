#!/bin/bash

# Quick HTTPS setup using Caddy (easier alternative to Nginx)
# Caddy automatically handles SSL certificates

set -e

echo "🔧 Setting up HTTPS with Caddy (Automatic SSL)..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Variables
DOMAIN="${1:-api.verbalforge.com}"
EMAIL="${2:-admin@verbalforge.com}"
BACKEND_PORT="8080"

echo "📝 Configuration:"
echo "   Domain: $DOMAIN"
echo "   Email: $EMAIL"
echo "   Backend Port: $BACKEND_PORT"
echo ""

# Install Caddy
echo "📦 Installing Caddy..."
apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt install -y caddy

# Create Caddyfile
echo "⚙️  Creating Caddy configuration..."
cat > /etc/caddy/Caddyfile <<EOF
$DOMAIN {
    # Automatic HTTPS with Let's Encrypt
    email $EMAIL

    # Reverse proxy to Go backend
    reverse_proxy localhost:$BACKEND_PORT {
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }

    # CORS headers
    header {
        Access-Control-Allow-Origin "https://black-dune-04f31470f.3.azurestaticapps.net"
        Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS, PATCH"
        Access-Control-Allow-Headers "Origin, Content-Type, Authorization"
        Access-Control-Allow-Credentials "true"
    }

    # Handle CORS preflight
    @options {
        method OPTIONS
    }
    respond @options 204

    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Frame-Options "SAMEORIGIN"
        X-Content-Type-Options "nosniff"
    }

    # File upload limit
    request_body {
        max_size 10MB
    }

    # Logging
    log {
        output file /var/log/caddy/verbalforge-backend.log
        format json
    }
}
EOF

# Test Caddy configuration
echo "✅ Testing Caddy configuration..."
caddy validate --config /etc/caddy/Caddyfile

# Reload Caddy
echo "🚀 Starting Caddy..."
systemctl enable caddy
systemctl restart caddy

echo ""
echo "✅ HTTPS setup complete with Caddy!"
echo ""
echo "📋 Next steps:"
echo "   1. Make sure DNS points $DOMAIN to: $(curl -s ifconfig.me)"
echo "   2. Wait 1-2 minutes for SSL certificate"
echo "   3. Update GitHub secret: NEXT_PUBLIC_API_URL=https://$DOMAIN"
echo "   4. Update backend env: FRONTEND_URL=https://black-dune-04f31470f.3.azurestaticapps.net"
echo ""
echo "🔍 Check status:"
echo "   systemctl status caddy"
echo "   curl https://$DOMAIN/health"
echo ""
