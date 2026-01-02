#!/usr/bin/env bash
set -euo pipefail

# install-osxcross.sh — bootstrap osxcross on Ubuntu (or any apt-based distro).
#
# IMPORTANT:
# - osxcross requires a macOS SDK that you must supply and have rights to use.
# - This script does NOT fetch Xcode/Apple SDKs automatically by default.
#
# Usage:
#   bash scripts/install-osxcross.sh --sdk-tarball /path/to/MacOSX13.3.sdk.tar.xz
#
# After success:
#   export OSXCROSS_ROOT="$HOME/osxcross"
#   export PATH="$OSXCROSS_ROOT/target/bin:$PATH"
#
# Then:
#   bash scripts/build-cross.sh

have() { command -v "$1" >/dev/null 2>&1; }
msg() { printf "[osxcross] %s\n" "$*"; }
err() { printf "[osxcross][ERROR] %s\n" "$*" >&2; }

print_help() {
	cat <<'EOF'
install-osxcross.sh — bootstrap osxcross and a macOS SDK

Flags:
  --root DIR           Install location (default: $HOME/osxcross)
  --sdk-tarball PATH   Copy an existing MacOSX*.sdk.tar.* into osxcross/tarballs (recommended)
  --sdk-url URL        Download SDK tarball into osxcross/tarballs (optional; you are responsible for licensing)
  --branch NAME        osxcross git branch (default: master)
  --no-deps            Skip apt-get dependency installation
  -h, --help           Show this help

Notes:
  - You must supply a macOS SDK tarball via --sdk-tarball, OR pre-place one in $OSXCROSS_ROOT/tarballs.
  - After success, add "$OSXCROSS_ROOT/target/bin" to your PATH.
EOF
}

OSXCROSS_ROOT="${OSXCROSS_ROOT:-$HOME/osxcross}"
SDK_TARBALL=""
SDK_URL=""
OSXCROSS_BRANCH="${OSXCROSS_BRANCH:-master}"
INSTALL_DEPS=1

while [ $# -gt 0 ]; do
	case "$1" in
	--root)
		OSXCROSS_ROOT="$2"
		shift 2
		;;
	--sdk-tarball)
		SDK_TARBALL="$2"
		shift 2
		;;
	--sdk-url)
		SDK_URL="$2"
		shift 2
		;;
	--branch)
		OSXCROSS_BRANCH="$2"
		shift 2
		;;
	--no-deps)
		INSTALL_DEPS=0
		shift
		;;
	-h | --help)
		print_help
		exit 0
		;;
	*)
		err "Unknown arg: $1"
		print_help
		exit 2
		;;
	esac
done

msg "Target root: $OSXCROSS_ROOT (branch: $OSXCROSS_BRANCH)"

if [ "$INSTALL_DEPS" = 1 ] && have apt-get; then
	msg "Installing dependencies via apt-get..."
	sudo apt-get update -qq
	sudo apt-get install -y git cmake ninja-build clang llvm lldb \
		build-essential g++ pkg-config \
		libxml2-dev uuid-dev libssl-dev libbz2-dev zlib1g-dev \
		cpio unzip zip xz-utils curl
else
	msg "Skipping dependency installation or non-apt system. Ensure required packages are present."
fi

mkdir -p "$OSXCROSS_ROOT"
if [ ! -d "$OSXCROSS_ROOT/.git" ]; then
	msg "Cloning osxcross into $OSXCROSS_ROOT ..."
	git clone --depth 1 --branch "$OSXCROSS_BRANCH" https://github.com/tpoechtrager/osxcross.git "$OSXCROSS_ROOT"
else
	msg "osxcross already present at $OSXCROSS_ROOT"
fi

mkdir -p "$OSXCROSS_ROOT/tarballs"

if [ -n "$SDK_URL" ]; then
	fname="$(basename "$SDK_URL")"
	if [ ! -f "$OSXCROSS_ROOT/tarballs/$fname" ]; then
		msg "Downloading SDK: $SDK_URL"
		curl -L "$SDK_URL" -o "$OSXCROSS_ROOT/tarballs/$fname"
	else
		msg "SDK already downloaded: $fname"
	fi
fi

if [ -n "$SDK_TARBALL" ]; then
	if [ ! -f "$SDK_TARBALL" ]; then
		err "SDK tarball not found: $SDK_TARBALL"
		exit 1
	fi
	msg "Copying SDK tarball into tarballs/ ..."
	cp -n "$SDK_TARBALL" "$OSXCROSS_ROOT/tarballs/"
fi

sdk_file="$(ls -1 "$OSXCROSS_ROOT"/tarballs/MacOSX*.sdk.tar.* 2>/dev/null | head -n1 || true)"
if [ -z "$sdk_file" ]; then
	err "No SDK tarball found in $OSXCROSS_ROOT/tarballs."
	err "Provide --sdk-tarball (recommended) or pre-place MacOSX*.sdk.tar.* there."
	exit 1
fi

sdk_base="$(basename "$sdk_file")"
msg "Using SDK: $sdk_base"

(
	cd "$OSXCROSS_ROOT"
	msg "Building osxcross (this can take a while)..."
	UNATTENDED=1 ./build.sh
)

TARGET_BIN="$OSXCROSS_ROOT/target/bin"
if [ ! -x "$TARGET_BIN/o64-clang" ] && [ ! -x "$TARGET_BIN/oa64-clang" ]; then
	err "osxcross build did not produce o64-clang/oa64-clang in $TARGET_BIN"
	exit 1
fi

msg "Done. Add to your shell profile or current env:"
echo "  export OSXCROSS_ROOT=\"$OSXCROSS_ROOT\""
echo "  export PATH=\"$TARGET_BIN:\$PATH\""
