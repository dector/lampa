#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

PACKAGE="${PACKAGE:-./exported}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/build/android}"
AAR_NAME="${AAR_NAME:-lampa-server.aar}"
TARGET="${TARGET:-android}"
ANDROID_API="${ANDROID_API:-21}"

if ! go tool gomobile version >/dev/null 2>&1; then
  echo "gomobile Go tool is not available in this module."
  echo "Add it with: go get -tool golang.org/x/mobile/cmd/gomobile"
  echo "Then run: go mod tidy"
  exit 1
fi

if [[ -z "${ANDROID_SDK_ROOT:-}" ]]; then
  if command -v mise >/dev/null 2>&1 && mise which sdkmanager >/dev/null 2>&1; then
    SDKMANAGER_BIN="$(mise which sdkmanager)"
    ANDROID_SDK_ROOT="$(echo "$SDKMANAGER_BIN" | sed -E 's#/cmdline-tools/[^/]+/bin/sdkmanager$##')"
  fi
fi

if [[ -n "${ANDROID_SDK_ROOT:-}" ]]; then
  export ANDROID_SDK_ROOT
  export ANDROID_HOME="${ANDROID_HOME:-$ANDROID_SDK_ROOT}"
fi

if [[ -z "${ANDROID_SDK_ROOT:-}" || ! -d "$ANDROID_SDK_ROOT" ]]; then
  echo "ANDROID_SDK_ROOT is not set or invalid."
  echo "Set it to your Android SDK path (example):"
  echo "  export ANDROID_SDK_ROOT=\"$HOME/Android/Sdk\""
  echo "  export ANDROID_HOME=\"$ANDROID_SDK_ROOT\""
  exit 1
fi

if [[ ! -d "$ANDROID_SDK_ROOT/platform-tools" ]]; then
  echo "Missing Android SDK component: platform-tools"
  echo "Install with: sdkmanager --sdk_root=\"$ANDROID_SDK_ROOT\" \"platform-tools\""
  exit 1
fi

if [[ ! -d "$ANDROID_SDK_ROOT/build-tools" ]] || ! compgen -G "$ANDROID_SDK_ROOT/build-tools/*" >/dev/null; then
  echo "Missing Android SDK component: build-tools"
  echo "Install with: sdkmanager --sdk_root=\"$ANDROID_SDK_ROOT\" \"build-tools;34.0.0\""
  exit 1
fi

if [[ ! -d "$ANDROID_SDK_ROOT/platforms" ]] || ! compgen -G "$ANDROID_SDK_ROOT/platforms/android-*" >/dev/null; then
  echo "Missing Android SDK component: platforms"
  echo "Install with: sdkmanager --sdk_root=\"$ANDROID_SDK_ROOT\" \"platforms;android-34\""
  exit 1
fi

HAS_COMPATIBLE_PLATFORM=0
for platform_dir in "$ANDROID_SDK_ROOT"/platforms/android-*; do
  platform_api="${platform_dir##*-}"
  if [[ "$platform_api" =~ ^[0-9]+$ ]] && (( platform_api >= ANDROID_API )); then
    HAS_COMPATIBLE_PLATFORM=1
    break
  fi
done
if (( HAS_COMPATIBLE_PLATFORM == 0 )); then
  echo "No installed Android platform is compatible with ANDROID_API=$ANDROID_API"
  echo "Install a newer platform, for example:"
  echo "  sdkmanager --sdk_root=\"$ANDROID_SDK_ROOT\" \"platforms;android-34\""
  exit 1
fi

if [[ ! -d "$ANDROID_SDK_ROOT/licenses" ]]; then
  echo "Android SDK licenses are missing."
  echo "Accept licenses with: sdkmanager --sdk_root=\"$ANDROID_SDK_ROOT\" --licenses"
  exit 1
fi

NDK_DIR="${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-}}"
if [[ -z "$NDK_DIR" ]]; then
  NDK_META_PATHS=("$ANDROID_SDK_ROOT"/ndk/*/meta/platforms.json)
  if [[ -e "${NDK_META_PATHS[0]}" ]]; then
    latest_meta="$(printf '%s\n' "${NDK_META_PATHS[@]}" | sort | tail -n1)"
    NDK_DIR="$(dirname "$(dirname "$latest_meta")")"
  fi
fi

if [[ -z "$NDK_DIR" || ! -f "$NDK_DIR/meta/platforms.json" ]]; then
  echo "Missing Android SDK component: NDK"
  echo "Install with: sdkmanager --sdk_root=\"$ANDROID_SDK_ROOT\" \"ndk;26.3.11579264\""
  exit 1
fi

export ANDROID_NDK_HOME="$NDK_DIR"
export ANDROID_NDK_ROOT="$NDK_DIR"

PKG_NAME="$(go list -f '{{.Name}}' "$PACKAGE")"
if [[ "$PKG_NAME" == "main" ]]; then
  echo "Cannot build AAR from '$PACKAGE' because it is package main."
  echo "gomobile bind requires an importable (non-main) package."
  echo "Use the exported package and run:"
  echo "  PACKAGE=./exported $0"
  exit 1
fi

mkdir -p "$OUT_DIR"

echo "Using ANDROID_SDK_ROOT=$ANDROID_SDK_ROOT"
echo "Initializing gomobile toolchain..."
go tool gomobile init

echo "Building AAR..."
go tool gomobile bind \
  -v \
  -target "$TARGET" \
  -androidapi "$ANDROID_API" \
  -o "$OUT_DIR/$AAR_NAME" \
  "$PACKAGE"

echo "Done: $OUT_DIR/$AAR_NAME"
