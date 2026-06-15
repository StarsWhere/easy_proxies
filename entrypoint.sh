#!/bin/sh
# Auto-generate config and fix permissions, then start easy_proxies

CONFIG_DIR="/etc/easy_proxies"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
NODES_FILE="$CONFIG_DIR/nodes.txt"
EXAMPLE_CONFIG="/app/config.example.yaml"

# Get current user uid/gid
CURRENT_UID=$(id -u 2>/dev/null || echo "10001")
CURRENT_GID=$(id -g 2>/dev/null || echo "10001")

# Auto-generate config.yaml if not exists
if [ ! -f "$CONFIG_FILE" ]; then
    if [ -w "$CONFIG_DIR" ]; then
        cp "$EXAMPLE_CONFIG" "$CONFIG_FILE"
        echo "[easy_proxies] Generated default config from $EXAMPLE_CONFIG"
    else
        echo "[easy_proxies] WARNING: Cannot write to $CONFIG_DIR (permission denied)"
        echo "[easy_proxies] Please ensure the mounted directory has correct permissions:"
        echo "  mkdir -p data && chown -R $(id -u):$(id -g) data"
        exit 1
    fi
fi

# Auto-create nodes.txt if not exists
if [ ! -f "$NODES_FILE" ]; then
    if [ -w "$CONFIG_DIR" ]; then
        touch "$NODES_FILE"
        echo "[easy_proxies] Created empty nodes.txt"
    else
        echo "[easy_proxies] WARNING: Cannot write to $CONFIG_DIR (permission denied)"
        echo "[easy_proxies] Please ensure the mounted directory has correct permissions:"
        echo "  mkdir -p data && chown -R $(id -u):$(id -g) data"
        exit 1
    fi
fi

# Verify read access to config files
if [ ! -r "$CONFIG_FILE" ]; then
    echo "[easy_proxies] ERROR: Cannot read $CONFIG_FILE (permission denied)"
    echo "[easy_proxies] Please fix permissions: chown -R $(id -u):$(id -g) data"
    exit 1
fi

exec /usr/local/bin/easy_proxies "$@"
