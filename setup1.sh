#!/usr/bin/env bash
set -euo pipefail

# setup1.sh
# Installs system dependencies for gemfetch (requires sudo).
# Target: Debian/Ubuntu (apt).

if ! command -v sudo >/dev/null 2>&1; then
  echo "Error: sudo not found."
  exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
  echo "Error: apt-get not found. This script targets Debian/Ubuntu."
  exit 1
fi

echo "Updating apt index..."
sudo apt-get update

echo "Installing dependencies..."
sudo apt-get install -y 
  golang-go \
  coreutils \
  procps \
  util-linux \
  pciutils \
  x11-xserver-utils \
  iproute2 \
  network-manager \
  wireless-tools \
  iw \
  lm-sensors \
  wmctrl \
  upower

echo
echo "Done."
echo "- You can now run: ./setup2.sh"
echo "- Tip: run 'sudo sensors-detect' if CPU temp shows n/a."
