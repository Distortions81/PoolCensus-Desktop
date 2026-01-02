#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

if [[ "$(uname -s)" == "Darwin" ]]; then
  echo "Building native macOS app via Wails..."
  cd frontend
  npm install
  npm run build
  cd ..
  wails build
else
  echo "Building macOS .app from Linux via osxcross..."
  bash ./scripts/build-cross.sh --macos-only "$@"
fi
