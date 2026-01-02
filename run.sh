#!/usr/bin/env bash
set -euo pipefail

mode="${1:-dev}"
shift || true

case "$mode" in
  dev)
    command -v wails >/dev/null 2>&1 || {
      echo "error: 'wails' not found in PATH"
      echo "hint: install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
      exit 1
    }
    exec wails dev "$@"
    ;;
  prod|production)
    if [[ ! -d "frontend/dist" ]]; then
      echo "error: frontend/dist not found"
      echo "hint: build the frontend first: (cd frontend && npm install && npm run build)"
      exit 1
    fi
    exec go run -tags production . "$@"
    ;;
  *)
    echo "usage: ./run.sh [dev|prod] [-- <extra args>]"
    echo "  dev  : runs 'wails dev' (default)"
    echo "  prod : runs 'go run -tags production .' using embedded frontend/dist"
    exit 2
    ;;
esac
