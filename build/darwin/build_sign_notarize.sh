#!/bin/bash
set -euo pipefail

################################################################################
# Usage: build_sign_notarize.sh [--version VERSION] [--sign] [--notarize]
################################################################################

function usage() {
  cat <<EOF
Usage: $(basename "$0") [--version <version>] [--sign] [--notarize]

  --version <version>   Specify the app version (required, must be vX.Y.Z).
  --sign                Sign the .app (and sign the .pkg if MACOS_SIGN_INSTALLER_ID is set).
  --notarize            Notarize the resulting .pkg (requires --sign).
  -h, --help            Show this help.
EOF
}

SIGN_APP=false
NOTARIZE_APP=false
VERSION_TAG=""

while [[ $# -gt 0 ]]; do
  case "$1" in
  --version)
    VERSION_TAG="$2"
    shift 2
    ;;
  --sign)
    SIGN_APP=true
    shift
    ;;
  --notarize)
    NOTARIZE_APP=true
    shift
    ;;
  -h | --help)
    usage
    exit 0
    ;;
  *)
    echo "Unknown argument: $1"
    usage
    exit 1
    ;;
  esac
done

if [[ -z "$VERSION_TAG" ]]; then
  echo "Error: --version is required."
  usage
  exit 1
fi

if ! [[ "$VERSION_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: --version must be in format vX.Y.Z (e.g., v1.2.3)"
  exit 1
fi

if [[ "$NOTARIZE_APP" == true && "$SIGN_APP" != true ]]; then
  echo "Error: --notarize requires --sign. Please specify both."
  usage
  exit 1
fi

################################################################################
# Step 1: Render all needed templates at the start
################################################################################

RENDER_FLAGS="--version $VERSION_TAG --info-plist"
if [[ "$SIGN_APP" == true ]]; then
  RENDER_FLAGS="$RENDER_FLAGS --entitlements --gon-sign"
fi
if [[ "$NOTARIZE_APP" == true ]]; then
  RENDER_FLAGS="$RENDER_FLAGS --gon-notarize"
fi

echo "Rendering templates..."
chmod +x build/darwin/render_templates.sh
build/darwin/render_templates.sh $RENDER_FLAGS

################################################################################
# Step 2: Build .app
################################################################################

: "${MACOS_BUILD_COMMAND:?Must set MACOS_BUILD_COMMAND}"
: "${MACOS_APP_BUNDLE_PATH:?Must set MACOS_APP_BUNDLE_PATH}"

echo "Building .app with command: ${MACOS_BUILD_COMMAND}"
eval "${MACOS_BUILD_COMMAND}"

################################################################################
# Step 3: Sign .app (if requested)
################################################################################

if [[ "$SIGN_APP" == true ]]; then
  : "${MACOS_GON_SIGN_JSON_PATH:?Must set MACOS_GON_SIGN_JSON_PATH for signing}"
  echo "Signing the app bundle with gon..."
  gon -log-level=info "$MACOS_GON_SIGN_JSON_PATH"
else
  echo "No signing requested. Skipping signing steps."
fi

################################################################################
# Step 4: Build .pkg (signed if requested)
################################################################################

: "${MACOS_PKG_BUNDLE_PATH:?Must set MACOS_PKG_BUNDLE_PATH}"
mkdir -p "$(dirname "$MACOS_PKG_BUNDLE_PATH")"

if [[ "$SIGN_APP" == true && -n "${MACOS_SIGN_INSTALLER_ID:-}" ]]; then
  : "${MACOS_SIGN_INSTALLER_ID:?Must set MACOS_SIGN_INSTALLER_ID to sign the installer pkg}"
  echo "Building signed installer pkg at: ${MACOS_PKG_BUNDLE_PATH}"
  productbuild \
    --sign "${MACOS_SIGN_INSTALLER_ID}" \
    --component "${MACOS_APP_BUNDLE_PATH}" /Applications \
    "${MACOS_PKG_BUNDLE_PATH}"
else
  echo "Building unsigned installer pkg at: ${MACOS_PKG_BUNDLE_PATH}"
  productbuild \
    --component "${MACOS_APP_BUNDLE_PATH}" /Applications \
    "${MACOS_PKG_BUNDLE_PATH}"
fi

################################################################################
# Step 5: Notarize (if requested)
################################################################################

if [[ "$SIGN_APP" == true && "$NOTARIZE_APP" == true ]]; then
  : "${MACOS_GON_NOTARIZE_JSON_PATH:?Must set MACOS_GON_NOTARIZE_JSON_PATH for notarization}"
  echo "Notarizing .pkg with gon..."
  gon -log-level=info "$MACOS_GON_NOTARIZE_JSON_PATH"
else
  echo "No notarization requested. Skipping notarization steps."
fi

echo "Done! Final pkg is at: ${MACOS_PKG_BUNDLE_PATH}"
if [[ "$SIGN_APP" == true ]]; then
  echo "App was signed. Pkg may also be signed if MACOS_SIGN_INSTALLER_ID was set."
fi
if [[ "$NOTARIZE_APP" == true ]]; then
  echo "Pkg was notarized."
fi
