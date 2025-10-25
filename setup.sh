#!/bin/bash

set -e

# Configuration
APP_NAME="verbalforge-backend"
APP_DIR="$HOME/$APP_NAME"
SERVICE_USER="verbalforge"
DOMAIN="${1:-api.verbalforge.xyz}"
EMAIL="${2:-newrex2002@gmail.com}"
BACKEND_PORT="8080"

# Detect if this is first-time setup or deployment
FIRST_TIME_SETUP=false
if [ ! -f "/etc/systemd/system/$APP_NAME.service" ]; then
    FIRST_TIME_SETUP=true
fi

echo "🚀 VerbalForge Backend Setup"
echo "   Mode: $([ "$FIRST_TIME_SETUP" = true ] && echo "First-time installation" || echo "Update deployment")"
echo "   Domain: $DOMAIN"
echo "   Email: $EMAIL"
echo ""

# ============================================================================
# FIRST-TIME SETUP ONLY
# ============================================================================
if [ "$FIRST_TIME_SETUP" = true ]; then
    echo "📦 Installing dependencies..."
    
    # Update system
    sudo apt-get update
    sudo apt-get install -y gnupg curl wget git
    
    # Install Go
    echo "📦 Installing Go 1.23.2..."
    wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz
    rm go1.23.2.linux-amd64.tar.gz
    
    # Add Go to PATH for current session
    export PATH=$PATH:/usr/local/go/bin
    
    # Add Go to PATH permanently
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    
    # Install MongoDB
    echo "📦 Installing MongoDB..."
    curl -fsSL https://pgp.mongodb.com/server-7.0.asc | sudo gpg --dearmor -o /usr/share/keyrings/mongodb-server-7.0.gpg
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list
    sudo apt-get update
    sudo apt-get install -y mongodb-org
    
    sudo systemctl enable mongod
    sudo systemctl start mongod
    
    # Create service user
    echo "👤 Creating service user..."
    sudo useradd -r -s /bin/false $SERVICE_USER || true
    
    # Install Caddy
    echo "📦 Installing Caddy..."
    sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
    sudo apt-get update
    sudo apt-get install -y caddy
    
    echo "✅ Dependencies installed successfully!"
fi

# ============================================================================
# BUILD AND DEPLOY (runs every time)
# ============================================================================
echo "🔨 Building application..."

# Ensure we're in the app directory
cd $APP_DIR

# Optional cache cleanup (skip by exporting SKIP_CACHE_CLEAN=true)
if [ "$SKIP_CACHE_CLEAN" != "true" ]; then
    echo "🧹 Cleaning Go build & module caches..."
    /usr/local/go/bin/go clean -cache -modcache -testcache || echo "⚠️ Cache clean failed (non-critical)"
    echo "🧹 Removing old binary (if present)..."
    rm -f "$APP_DIR/$APP_NAME" || true
fi

# Build the application (fresh after cache cleanup)
export PATH=$PATH:/usr/local/go/bin
/usr/local/go/bin/go mod download
/usr/local/go/bin/go build -o $APP_NAME ./cmd/server

# Set ownership
sudo chown -R $SERVICE_USER:$SERVICE_USER $APP_DIR

# ============================================================================
# SYSTEMD SERVICE SETUP
# ============================================================================
echo "⚙️  Configuring systemd service..."

sudo tee /etc/systemd/system/$APP_NAME.service > /dev/null <<EOF
[Unit]
Description=VerbalForge Backend API
After=network.target mongod.service
Requires=mongod.service

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$APP_NAME
Restart=always
RestartSec=5
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable $APP_NAME
sudo systemctl restart $APP_NAME

echo "✅ Backend service restarted!"

# ============================================================================
# CADDY HTTPS SETUP (runs every time to capture config changes)
# ============================================================================
echo "🔒 Configuring Caddy reverse proxy..."

# Stop Nginx if running (to free port 80) - first time only
if systemctl is-active --quiet nginx; then
    echo "⚠️  Stopping Nginx to free port 80..."
    sudo systemctl stop nginx
    sudo systemctl disable nginx
fi

# Always update Caddyfile to capture any configuration changes
sudo tee /etc/caddy/Caddyfile > /dev/null <<EOF
{
    email $EMAIL
    auto_https disable_redirects
}

$DOMAIN {
    # Reverse proxy to Go backend
    reverse_proxy localhost:$BACKEND_PORT {
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }

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
}
EOF

# Validate Caddyfile syntax
echo "🔍 Validating Caddyfile..."
sudo caddy validate --config /etc/caddy/Caddyfile

# Enable and reload Caddy (reload preserves existing connections and certificates)
sudo systemctl enable caddy

# Use reload if Caddy is already running, otherwise start it
if systemctl is-active --quiet caddy; then
    echo "🔄 Reloading Caddy configuration..."
    sudo systemctl reload caddy
else
    echo "🚀 Starting Caddy..."
    sudo systemctl start caddy
fi

# Wait a moment and check if Caddy is running successfully
sleep 2
if systemctl is-active --quiet caddy; then
    echo "✅ Caddy configured and running successfully!"
else
    echo "❌ Caddy failed to start. Check logs with: sudo journalctl -u caddy -n 50"
    exit 1
fi

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "✅ Setup complete!"
echo ""
echo "📊 Service Status:"
sudo systemctl status $APP_NAME --no-pager -l | head -10
echo ""

if [ "$FIRST_TIME_SETUP" = true ]; then
    echo "🎉 First-time setup completed!"
    echo ""
    echo "📋 Next steps:"
    echo "   1. Ensure DNS points $DOMAIN to this server"
    echo "   2. Wait 1-2 minutes for SSL certificate"
    echo "   3. Test: curl https://$DOMAIN/health"
    echo "   4. Update GitHub secrets with required environment variables:"
    echo "      - MONGODB_URI"
    echo "      - JWT_SECRET"
    echo "      - FRONTEND_URLS (comma-separated list of allowed origins)"
    echo "      - LETS_ENCRYPT_EMAIL"
else
    echo "🔄 Deployment complete!"
    echo "   Backend is running on https://$DOMAIN"
fi
echo ""
