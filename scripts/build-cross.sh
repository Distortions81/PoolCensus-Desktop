#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

install_deps=1
if [[ "${1-}" == "--no-install-deps" ]]; then
  install_deps=0
fi

if ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then
  if [[ $install_deps -eq 1 ]] && command -v apt-get >/dev/null 2>&1; then
    echo "Installing mingw-w64 (Ubuntu 24.04)..."
    if [[ ${EUID:-$(id -u)} -eq 0 ]]; then
      apt-get update
      apt-get install -y mingw-w64
    else
      sudo apt-get update
      sudo apt-get install -y mingw-w64
    fi
  else
    echo "x86_64-w64-mingw32-gcc not found."
    echo "On Ubuntu 24.04, install it with: sudo apt-get install mingw-w64"
    echo "Or rerun without --no-install-deps"
    exit 1
  fi
fi

echo "Preparing frontend assets..."
cd frontend
npm install
npm run build

echo "Building Linux + Windows targets..."
cd ..
wails build --platform linux/amd64,windows/amd64
