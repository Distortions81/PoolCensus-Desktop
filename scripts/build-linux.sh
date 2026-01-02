#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

echo "Installing frontend dependencies..."
cd frontend
npm install
npm run build

echo "Running Wails build for Linux..."
cd ..
wails build
