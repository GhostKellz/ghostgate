#!/bin/bash

# This script starts the GhostGate server temporarily to renew Let's Encrypt certificates.
# Ensure this script is executable and scheduled in a cron job if needed.

GHOSTGATE_DIR="/home/chris/ghostgate"
CONFIG_FILE="$GHOSTGATE_DIR/gate.conf"
CONF_DIR="$GHOSTGATE_DIR/conf.d"

# Start GhostGate server temporarily
/usr/bin/env go run "$GHOSTGATE_DIR/main.go" -config "$CONFIG_FILE" -conf-dir "$CONF_DIR" &
GHOSTGATE_PID=$!

# Allow some time for certificate renewal
sleep 300

# Stop the temporary GhostGate server
kill $GHOSTGATE_PID

# Ensure the process is terminated
wait $GHOSTGATE_PID 2>/dev/null
