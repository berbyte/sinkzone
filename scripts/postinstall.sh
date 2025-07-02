#!/bin/bash

# Post-installation script for sinkzone

set -e

# Create sinkzone user if it doesn't exist
if ! id "sinkzone" &>/dev/null; then
    useradd -r -s /bin/false -d /var/lib/sinkzone sinkzone
fi

# Create data directory
mkdir -p /var/lib/sinkzone
chown sinkzone:sinkzone /var/lib/sinkzone

# Create systemd service
cat > /etc/systemd/system/sinkzone.service << 'EOF'
[Unit]
Description=Sinkzone DNS Filter
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/sinkzone dns start
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "Sinkzone post-installation completed successfully!"
echo "To start the service: sudo systemctl start sinkzone"
echo "To enable on boot: sudo systemctl enable sinkzone" 