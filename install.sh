#!/bin/bash
# GhostGate install script for Debian/Ubuntu and Arch Linux
set -e

GHOSTGATE_SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
GHOSTGATE_BIN="ghostgate"
GHOSTGATE_SERVICE="ghostgate.service"
GHOSTGATE_CONF="gate.conf"
GHOSTGATE_CONFD="conf.d"
GHOSTGATE_STATIC="static"

# Detect OS
if [ -f /etc/debian_version ]; then
    OS="debian"
    BIN_DIR="/usr/local/bin"
    CONF_DIR="/etc/ghostgate"
    SYSTEMD_DIR="/etc/systemd/system"
elif [ -f /etc/arch-release ]; then
    OS="arch"
    BIN_DIR="/usr/bin"
    CONF_DIR="/etc/ghostgate"
    SYSTEMD_DIR="/etc/systemd/system"
else
    echo "Unsupported OS. Only Debian/Ubuntu and Arch Linux are supported."
    exit 1
fi

echo "[+] Building GhostGate binary..."
cd "$GHOSTGATE_SRC_DIR"
go build -o "$GHOSTGATE_BIN"

# Install binary
sudo install -Dm755 "$GHOSTGATE_BIN" "$BIN_DIR/$GHOSTGATE_BIN"

# Install config and directories
sudo install -d "$CONF_DIR"
sudo install -Dm644 "$GHOSTGATE_CONF" "$CONF_DIR/gate.conf"
sudo cp -r "$GHOSTGATE_CONFD" "$CONF_DIR/"
if [ -d "$GHOSTGATE_STATIC" ]; then
    sudo cp -r "$GHOSTGATE_STATIC" "$CONF_DIR/"
fi

# Install systemd service
sudo install -Dm644 "$GHOSTGATE_SERVICE" "$SYSTEMD_DIR/$GHOSTGATE_SERVICE"

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable ghostgate.service
sudo systemctl restart ghostgate.service

echo "[+] GhostGate installed and started!"
echo "    Binary: $BIN_DIR/$GHOSTGATE_BIN"
echo "    Config: $CONF_DIR/gate.conf"
echo "    Service: $SYSTEMD_DIR/$GHOSTGATE_SERVICE"
echo "    To check status: sudo systemctl status ghostgate.service"
