#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
  echo "This script must run via sudo because it installs packages."
  exit 1
fi

apt update
apt install -y --no-install-recommends \
  curl \
  build-essential \
  pkg-config \
  libgtk-3-dev \
  libwebkit2gtk-4.1-dev \
  nodejs \
  npm \
  git \
  ca-certificates

PKGCONFIG_DIR="/usr/lib/x86_64-linux-gnu/pkgconfig"
if [[ -f "$PKGCONFIG_DIR/webkit2gtk-4.1.pc" ]]; then
  ln -sf "webkit2gtk-4.1.pc" "$PKGCONFIG_DIR/webkit2gtk-4.0.pc"
fi

export PATH="$HOME/go/bin:$PATH"

if ! command -v go >/dev/null 2>&1; then
  apt install -y --no-install-recommends golang-go
fi

go install github.com/wailsapp/wails/v2/cmd/wails@latest

cd "$(dirname "${BASH_SOURCE[0]}")/.."
node -v
npm -v

cd frontend
npm install
npm run build

cd ..
wails build
