#!/usr/bin/env bash
set -euo pipefail

if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew is required. Please install it from https://brew.sh"
  exit 1
fi

brew update
brew install go node pkg-config

export PATH="$HOME/go/bin:$PATH"
go install github.com/wailsapp/wails/v2/cmd/wails@latest

cd "$(dirname "${BASH_SOURCE[0]}")/.."
cd frontend
npm install
npm run build

cd ..
wails build
