#!/usr/bin/env bash
set -euo pipefail

# setup2.sh
# Prepares a user-level environment without sudo:
# - builds gemfetch
# - installs it into ~/.local/bin (recommended)
# - prints a quick diagnostics of optional tools

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Project: ${ROOT_DIR}"

if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go is not installed (missing 'go')."
  echo "Run ./setup1.sh (requires sudo) or install Go yourself."
  exit 1
fi

BIN_DIR="${HOME}/.local/bin"
mkdir -p "${BIN_DIR}"

echo "Building gemfetch..."
cd "${ROOT_DIR}"
go build -o gemfetch

echo "Installing to ${BIN_DIR}/gemfetch ..."
cp -f gemfetch "${BIN_DIR}/gemfetch"

if ! echo "${PATH}" | tr ':' '\n' | grep -qx "${BIN_DIR}"; then
  echo
  echo "NOTE: ${BIN_DIR} is not in your PATH."
  echo "Add this to your shell config (e.g. ~/.bashrc):"
  echo "  export PATH=\"${BIN_DIR}:\$PATH\""
fi

echo
echo "Optional tools status (gemfetch will show 'n/a' if missing):"
check() {
  local cmd="$1"
  if command -v "${cmd}" >/dev/null 2>&1; then
    echo "  [OK]   ${cmd}"
  else
    echo "  [MISS] ${cmd}"
  fi
}

check uname
check uptime
check dpkg
check xrandr
check lspci
check lscpu
check free
check df
check ps
check ip
check nmcli
check iwgetid
check iw
check sensors
check wmctrl
check upower
check hostnamectl

echo
echo "Done."
echo "Try: gemfetch"
