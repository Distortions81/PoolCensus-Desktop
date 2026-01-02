#!/usr/bin/env bash
set -euo pipefail

# One entrypoint for "everything we can build from this machine".
# On Ubuntu: Linux + Windows always; macOS only if osxcross is available/configured.

cd "$(dirname "${BASH_SOURCE[0]}")"
bash ./build-cross.sh "$@"
