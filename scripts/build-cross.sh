#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

if [[ -e build && ! -w build ]]; then
  echo "build/ is not writable; if you previously ran wails with sudo, fix ownership:"
  echo "  sudo chown -R $(id -un):$(id -gn) build"
  echo "or remove it:"
  echo "  sudo rm -rf build"
  exit 1
fi

install_deps=1
build_darwin=1
osxcross_root="${OSXCROSS_ROOT:-$HOME/osxcross}"
osxcross_sdk_tarball="${OSXCROSS_SDK_TARBALL:-}"
osxcross_sdk_url="${OSXCROSS_SDK_URL:-}"

for arg in "$@"; do
  case "$arg" in
    --no-install-deps) install_deps=0 ;;
    --no-darwin) build_darwin=0 ;;
    --osxcross-root=*) osxcross_root="${arg#*=}" ;;
    --osxcross-sdk-tarball=*) osxcross_sdk_tarball="${arg#*=}" ;;
    --osxcross-sdk-url=*) osxcross_sdk_url="${arg#*=}" ;;
  esac
done

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
cd ..

echo "Building Linux target..."
wails build --platform linux/amd64

echo "Building Windows target (cross-compile)..."
wails build --platform windows/amd64 -nopackage -o poolcensus.exe

if [[ $build_darwin -eq 0 ]]; then
  exit 0
fi

if [[ -n "$osxcross_root" && -d "$osxcross_root/target/bin" ]]; then
  export OSXCROSS_ROOT="$osxcross_root"
  export PATH="$OSXCROSS_ROOT/target/bin:$PATH"
fi

find_cc() {
  local candidate
  for candidate in "$@"; do
    if command -v "$candidate" >/dev/null 2>&1; then
      echo "$candidate"
      return 0
    fi
  done
  return 1
}

cc_amd64="$(find_cc o64-clang x86_64-apple-darwin*-clang || true)"
cxx_amd64="$(find_cc o64-clang++ x86_64-apple-darwin*-clang++ || true)"
cc_arm64="$(find_cc oa64-clang aarch64-apple-darwin*-clang arm64-apple-darwin*-clang || true)"
cxx_arm64="$(find_cc oa64-clang++ aarch64-apple-darwin*-clang++ arm64-apple-darwin*-clang++ || true)"

if [[ -z "${cc_amd64}${cc_arm64}" ]]; then
  if [[ -x "./scripts/install-osxcross.sh" ]] && [[ -n "${osxcross_sdk_tarball}${osxcross_sdk_url}" ]]; then
    echo "osxcross toolchain not found; attempting install..."
    install_args=(--root "$osxcross_root")
    if [[ $install_deps -eq 0 ]]; then
      install_args+=(--no-deps)
    fi
    if [[ -n "$osxcross_sdk_tarball" ]]; then
      install_args+=(--sdk-tarball "$osxcross_sdk_tarball")
    fi
    if [[ -n "$osxcross_sdk_url" ]]; then
      install_args+=(--sdk-url "$osxcross_sdk_url")
    fi
    bash ./scripts/install-osxcross.sh "${install_args[@]}"
    export OSXCROSS_ROOT="$osxcross_root"
    export PATH="$OSXCROSS_ROOT/target/bin:$PATH"

    cc_amd64="$(find_cc o64-clang x86_64-apple-darwin*-clang || true)"
    cxx_amd64="$(find_cc o64-clang++ x86_64-apple-darwin*-clang++ || true)"
    cc_arm64="$(find_cc oa64-clang aarch64-apple-darwin*-clang arm64-apple-darwin*-clang || true)"
    cxx_arm64="$(find_cc oa64-clang++ aarch64-apple-darwin*-clang++ arm64-apple-darwin*-clang++ || true)"
  else
    echo "osxcross toolchain not found; skipping macOS targets."
    echo "To enable darwin builds:"
    echo "  - Install osxcross: bash scripts/install-osxcross.sh --sdk-tarball /path/to/MacOSX13.3.sdk.tar.xz"
    echo "  - Or set OSXCROSS_SDK_TARBALL and rerun this script"
    echo "  - Or pass --no-darwin to silence"
    exit 0
  fi
fi

echo "Building macOS targets via osxcross (-nopackage)..."
if [[ -n "$cc_amd64" && -n "$cxx_amd64" ]]; then
  echo " - darwin/amd64 (CC=$cc_amd64)"
  CC="$cc_amd64" CXX="$cxx_amd64" CGO_ENABLED=1 wails build --platform darwin/amd64 -nopackage
fi

if [[ -n "$cc_arm64" && -n "$cxx_arm64" ]]; then
  echo " - darwin/arm64 (CC=$cc_arm64)"
  CC="$cc_arm64" CXX="$cxx_arm64" CGO_ENABLED=1 wails build --platform darwin/arm64 -nopackage
fi

echo "Done. Note: darwin builds from Linux are -nopackage (no signed/notarized .app/.dmg)."
