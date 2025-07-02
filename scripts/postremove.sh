#!/bin/bash

# Post-removal script for sinkzone

set -e

# Stop and disable service if running
if systemctl is-active --quiet sinkzone; then
    systemctl stop sinkzone
fi

if systemctl is-enabled --quiet sinkzone; then
    systemctl disable sinkzone
fi

# Remove systemd service file
rm -f /etc/systemd/system/sinkzone.service

# Reload systemd
systemctl daemon-reload

# Remove sinkzone user if it exists and has no home directory
if id "sinkzone" &>/dev/null; then
    if [ "$(eval echo ~sinkzone)" = "/var/lib/sinkzone" ]; then
        userdel sinkzone
        rmdir /var/lib/sinkzone 2>/dev/null || true
    fi
fi

echo "Sinkzone post-removal completed successfully!" 