#!/bin/bash

set -e

# 1. Download binary
BIN_URL="https://github.com/paulozagaloneves/kvm-compose/releases/download/0.2.0/kvm-compose-linux-amd64"
BIN_DEST="/usr/local/bin/kvm-compose"
echo "Downloading kvm-compose binary..."
curl -L "$BIN_URL" -o /tmp/kvm-compose
sudo mv /tmp/kvm-compose "$BIN_DEST"
sudo chmod +x "$BIN_DEST"
echo "kvm-compose installed to $BIN_DEST"

# 2. Create config directories
CONFIG_DIR="$HOME/.config/kvm-compose"
echo "Creating configuration directories..."
mkdir -p "$CONFIG_DIR/images/vm"
mkdir -p "$CONFIG_DIR/images/upstream"
mkdir -p "$CONFIG_DIR/templates"

# 3. Download config.ini.example
CONFIG_URL="https://raw.githubusercontent.com/paulozagaloneves/kvm-compose/main/config.ini.example"
echo "Downloading config.ini.example..."
curl -L "$CONFIG_URL" -o "$CONFIG_DIR/config.ini.example"

# 4. Download all template files
TEMPLATES=(
  "almalinux10.ini"
  "debian13.ini"
  "fedora43.ini"
  "meta-data.tmpl"
  "network-config-almalinux.tmpl"
  "network-config.tmpl"
  "ubuntu24.04.ini"
  "user-data.tmpl"
)
TEMPLATE_BASE="https://raw.githubusercontent.com/paulozagaloneves/kvm-compose/main/templates"
echo "Downloading template files..."
for tmpl in "${TEMPLATES[@]}"; do
  curl -L "$TEMPLATE_BASE/$tmpl" -o "$CONFIG_DIR/templates/$tmpl"
done

echo "kvm-compose installation and configuration complete."
