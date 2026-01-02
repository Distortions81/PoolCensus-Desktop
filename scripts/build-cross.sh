#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

assert_writable_dir() {
  local path="$1"
  if [[ -e "$path" && ! -w "$path" ]]; then
    echo "${path} is not writable; if you previously ran builds with sudo, fix ownership:"
    echo "  sudo chown -R $(id -un):$(id -gn) ${path}"
    echo "or remove it:"
    echo "  sudo rm -rf ${path}"
    exit 1
  fi
}

assert_writable_dir build
assert_writable_dir frontend
assert_writable_dir frontend/dist

install_deps=1
build_linux=1
build_windows=1
build_darwin=1
osxcross_root="${OSXCROSS_ROOT:-$HOME/osxcross}"
osxcross_sdk_tarball="${OSXCROSS_SDK_TARBALL:-}"
osxcross_sdk_url="${OSXCROSS_SDK_URL:-}"
macos_min_amd64="${MACOSX_DEPLOYMENT_TARGET_AMD64:-10.14}"
macos_min_arm64="${MACOSX_DEPLOYMENT_TARGET_ARM64:-11.0}"
macos_plist_min="${MACOSX_PLIST_MIN_VERSION:-}"
build_macos_arm64=0
clean_outputs=1

for arg in "$@"; do
  case "$arg" in
    -h|--help)
      cat <<'EOF'
Usage: scripts/build-cross.sh [options]

Builds frontend once, then produces platform artifacts under build/dist/.

Options:
  --no-install-deps   Don't apt-get missing deps (mingw-w64/zip/etc)
  --no-clean          Keep previous build outputs
  --no-linux          Skip Linux build/zip
  --no-windows        Skip Windows build/zip
  --no-darwin         Skip macOS build/zip
  --linux-only        Build only Linux
  --windows-only      Build only Windows
  --macos-only        Build only macOS (.app zip via osxcross)

macOS (osxcross):
  --osxcross-root=DIR         (default: $OSXCROSS_ROOT or $HOME/osxcross)
  --osxcross-sdk-tarball=PATH optional auto-install helper input
  --osxcross-sdk-url=URL      optional auto-install helper input
  --macos-min=VERSION         Intel minimum (default: 10.14)
  --macos-arm64               Also build Apple Silicon (default off)
EOF
      exit 0
      ;;
    --no-install-deps) install_deps=0 ;;
    --no-linux) build_linux=0 ;;
    --no-windows) build_windows=0 ;;
    --no-darwin) build_darwin=0 ;;
    --linux-only) build_linux=1; build_windows=0; build_darwin=0 ;;
    --windows-only) build_linux=0; build_windows=1; build_darwin=0 ;;
    --macos-only) build_linux=0; build_windows=0; build_darwin=1 ;;
    --no-clean) clean_outputs=0 ;;
    --osxcross-root=*) osxcross_root="${arg#*=}" ;;
    --osxcross-sdk-tarball=*) osxcross_sdk_tarball="${arg#*=}" ;;
    --osxcross-sdk-url=*) osxcross_sdk_url="${arg#*=}" ;;
    --macos-min=*) macos_min_amd64="${arg#*=}" ;;
    --macos-min-amd64=*) macos_min_amd64="${arg#*=}" ;;
    --macos-min-arm64=*) macos_min_arm64="${arg#*=}" ;;
    --macos-plist-min=*) macos_plist_min="${arg#*=}" ;;
    --no-macos-arm64) build_macos_arm64=0 ;;
    --macos-arm64) build_macos_arm64=1 ;;
  esac
done

if [[ $clean_outputs -eq 1 ]]; then
  rm -rf build/dist
  rm -rf build/bin/PoolCensus.app
  rm -f build/bin/poolcensus build/bin/poolcensus.exe
  rm -f build/bin/poolcensus-darwin-amd64 build/bin/poolcensus-darwin-arm64
  rm -f build/dist/PoolCensus-Linux-x86_64.zip build/dist/PoolCensus-Windows-x86_64.zip build/dist/PoolCensus-macOS.app.zip
  rm -rf frontend/dist
fi

if [[ $build_windows -eq 1 ]] && ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then
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

if [[ $build_linux -eq 1 ]] && command -v pkg-config >/dev/null 2>&1; then
  if ! pkg-config --exists webkit2gtk-4.0 && pkg-config --exists webkit2gtk-4.1; then
    PKGCONFIG_DIR="/usr/lib/x86_64-linux-gnu/pkgconfig"
    if [[ -f "${PKGCONFIG_DIR}/webkit2gtk-4.1.pc" && ! -f "${PKGCONFIG_DIR}/webkit2gtk-4.0.pc" ]]; then
      echo "Ubuntu 24.04 WebKitGTK detected; creating webkit2gtk-4.0 pkg-config alias..."
      if [[ ${EUID:-$(id -u)} -eq 0 ]]; then
        ln -sf "webkit2gtk-4.1.pc" "${PKGCONFIG_DIR}/webkit2gtk-4.0.pc"
      else
        sudo ln -sf "webkit2gtk-4.1.pc" "${PKGCONFIG_DIR}/webkit2gtk-4.0.pc"
      fi
    fi
  fi
fi

if [[ $build_linux -eq 1 ]]; then
  echo "Building Linux target..."
  wails build --platform linux/amd64
fi

if [[ $build_windows -eq 1 ]]; then
  echo "Building Windows target (cross-compile)..."
  wails build --platform windows/amd64 -nopackage -o poolcensus.exe
fi

have() { command -v "$1" >/dev/null 2>&1; }

ensure_cmd() {
  local cmd="$1"
  local pkg="${2:-$1}"
  if have "$cmd"; then
    return 0
  fi
  if [[ $install_deps -eq 1 ]] && have apt-get; then
    echo "Installing $pkg..."
    if [[ ${EUID:-$(id -u)} -eq 0 ]]; then
      apt-get update -qq
      apt-get install -y "$pkg"
    else
      sudo apt-get update -qq
      sudo apt-get install -y "$pkg"
    fi
    return 0
  fi
  echo "$cmd not found; please install $pkg (or re-run without --no-install-deps)" >&2
  return 1
}

dist_dir="build/dist"
mkdir -p "$dist_dir"
ensure_cmd zip zip

if [[ $build_linux -eq 1 ]]; then
  echo "Packaging Linux .zip artifact..."
  zip -q -j -r "${dist_dir}/PoolCensus-Linux-x86_64.zip" "build/bin/poolcensus"
fi

if [[ $build_windows -eq 1 ]]; then
  echo "Packaging Windows .zip artifact..."
  zip -q -j -r "${dist_dir}/PoolCensus-Windows-x86_64.zip" "build/bin/poolcensus.exe"
fi

if [[ $build_darwin -eq 0 ]]; then
  echo "Dist zips:"
  if [[ $build_linux -eq 1 ]]; then
    echo " - ${dist_dir}/PoolCensus-Linux-x86_64.zip"
  fi
  if [[ $build_windows -eq 1 ]]; then
    echo " - ${dist_dir}/PoolCensus-Windows-x86_64.zip"
  fi
  exit 0
fi

if [[ -n "$osxcross_root" && -d "$osxcross_root/target/bin" ]]; then
  export OSXCROSS_ROOT="$osxcross_root"
  export PATH="$OSXCROSS_ROOT/target/bin:$PATH"
fi

if ! have o64-clang && ! have oa64-clang; then
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
  else
    echo "osxcross toolchain not found; skipping macOS targets."
    echo "To enable .app builds from Ubuntu:"
    echo "  - Install osxcross: bash scripts/install-osxcross.sh --sdk-tarball /path/to/MacOSX13.3.sdk.tar.xz"
    echo "  - Or set OSXCROSS_SDK_TARBALL and rerun this script"
    echo "  - Or pass --no-darwin to silence"
    exit 0
  fi
fi

echo "Building macOS binaries via osxcross..."

ldflags="-s -w"
darwin_amd64_bin="build/bin/poolcensus-darwin-amd64"
darwin_arm64_bin="build/bin/poolcensus-darwin-arm64"

if [[ $build_macos_arm64 -eq 0 ]]; then
  rm -f "$darwin_arm64_bin"
fi

if have o64-clang; then
  echo " - darwin/amd64 (min macOS ${macos_min_amd64})"
  darwin_amd64_ldflags="${ldflags} -linkmode external -extldflags=-mmacosx-version-min=${macos_min_amd64}"
  env \
    GOOS=darwin GOARCH=amd64 \
    CGO_ENABLED=1 \
    MACOSX_DEPLOYMENT_TARGET="${macos_min_amd64}" \
    CGO_CFLAGS="-mmacosx-version-min=${macos_min_amd64}" \
    CGO_LDFLAGS="-mmacosx-version-min=${macos_min_amd64} -Wl,-weak_framework,UniformTypeIdentifiers" \
    CC=o64-clang CXX=o64-clang++ \
    go build -trimpath -tags production -ldflags "$darwin_amd64_ldflags" -o "$darwin_amd64_bin" .
fi

if [[ $build_macos_arm64 -eq 1 ]] && have oa64-clang; then
  echo " - darwin/arm64 (min macOS ${macos_min_arm64})"
  darwin_arm64_ldflags="${ldflags} -linkmode external -extldflags=-mmacosx-version-min=${macos_min_arm64}"
  env \
    GOOS=darwin GOARCH=arm64 \
    CGO_ENABLED=1 \
    MACOSX_DEPLOYMENT_TARGET="${macos_min_arm64}" \
    CGO_CFLAGS="-mmacosx-version-min=${macos_min_arm64}" \
    CGO_LDFLAGS="-mmacosx-version-min=${macos_min_arm64}" \
    CC=oa64-clang CXX=oa64-clang++ \
    go build -trimpath -tags production -ldflags "$darwin_arm64_ldflags" -o "$darwin_arm64_bin" .
fi

app_name="PoolCensus"
bundle_dir="build/bin/${app_name}.app"
macos_dir="${bundle_dir}/Contents/MacOS"
resources_dir="${bundle_dir}/Contents/Resources"
plist_path="${bundle_dir}/Contents/Info.plist"
icon_src="build/appicon.png"
icon_dest="${resources_dir}/PoolCensus.icns"

echo "Creating ${app_name}.app bundle..."
rm -rf "$bundle_dir"
mkdir -p "$macos_dir" "$resources_dir"

bundle_bin="${macos_dir}/${app_name}"
if [[ -f "$darwin_amd64_bin" ]]; then
  cp "$darwin_amd64_bin" "$bundle_bin"
elif [[ -f "$darwin_arm64_bin" ]]; then
  cp "$darwin_arm64_bin" "$bundle_bin"
else
  echo "No darwin binaries built; skipping .app creation." >&2
  exit 1
fi

chmod +x "$bundle_bin"

app_min_version="${macos_plist_min:-$macos_min_amd64}"

if [[ -f "$icon_src" ]]; then
  if have convert || ensure_cmd convert imagemagick; then
    convert "$icon_src" -define icon:auto-resize=16,32,64,128,256,512 "$icon_dest" || true
  else
    echo "convert/imagemagick not available; skipping .icns generation."
  fi
fi

cat <<EOF >"$plist_path"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>${app_name}</string>
  <key>CFBundleIdentifier</key>
  <string>com.poolcensus.desktop</string>
  <key>CFBundleName</key>
  <string>${app_name}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>LSMinimumSystemVersion</key>
  <string>${app_min_version}</string>
  <key>CFBundleVersion</key>
  <string>1.0.0</string>
  <key>CFBundleShortVersionString</key>
  <string>1.0.0</string>
  <key>CFBundleIconFile</key>
  <string>PoolCensus.icns</string>
</dict>
</plist>
EOF

if have rcodesign; then
  echo "Ad-hoc signing ${app_name}.app with rcodesign..."
  rcodesign sign "$bundle_dir" || echo "rcodesign sign failed, continuing" >&2
  rcodesign verify --verbose "$bundle_bin" || echo "rcodesign verify failed, continuing" >&2
else
  echo "rcodesign not found; skipping macOS signing. (Some Macs may quarantine/deny unsigned apps.)" >&2
fi

ensure_cmd zip zip

(
  cd build/bin
  zip -q -r "../dist/${app_name}-macOS.app.zip" "${app_name}.app"
)

echo "Done. macOS output:"
 echo " - ${bundle_dir}"
 echo " - ${dist_dir}/${app_name}-macOS.app.zip"

echo "Dist zips:"
if [[ $build_linux -eq 1 ]]; then
  echo " - ${dist_dir}/PoolCensus-Linux-x86_64.zip"
fi
if [[ $build_windows -eq 1 ]]; then
  echo " - ${dist_dir}/PoolCensus-Windows-x86_64.zip"
fi
echo " - ${dist_dir}/${app_name}-macOS.app.zip"
