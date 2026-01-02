#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

echo "Ensuring frontend bundle is current..."
cd frontend
npm install
npm run build

echo "Packaging macOS app with Wails..."
cd ..
wails build
