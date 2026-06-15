#!/bin/bash
# Ensure config directory exists with correct permissions and start easy_proxies

set -e

echo "🚀 Starting easy_proxies..."

# Create directories if they don't exist
echo "📁 Creating data and logs directories..."
mkdir -p data logs

# Fix permissions for the data directory (ensure current user can write)
echo "🔒 Setting permissions..."
chmod -R u+w data logs 2>/dev/null || true

# Export current user UID/GID for docker-compose
export UID=$(id -u)
export GID=$(id -g)
echo "👤 Running as UID:GID = $UID:$GID"

# Pull latest image, stop old container, and start new one
echo "🐳 Starting Docker containers..."
docker compose pull && docker compose down && docker compose up -d

echo ""
echo "✅ easy_proxies started successfully!"
echo ""
echo "📊 Access WebUI: http://localhost:9091"
echo "🔍 View logs: docker compose logs -f"
echo "🛑 Stop: docker compose down"
echo ""
