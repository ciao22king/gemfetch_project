#!/usr/bin/env bash
set -euo pipefail

# setup.sh - One script to rule them all

echo "🔧 Installing dependencies..."
sudo apt-get update
sudo apt-get install -y \
  git golang-go pciutils x11-xserver-utils \
  network-manager wireless-tools iw lm-sensors \
  wmctrl upower

echo "📦 Building fetchy..."
go build -ldflags="-s -w" -o fetchy

echo "✅ Done! Run ./fetchy to see your system info."
