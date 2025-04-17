#!/bin/bash
set -euo pipefail

################################################################################
# Usage: build_sign_notarize.sh [--version VERSION] [--sign] [--notarize]
#
# This script:
#   1. Builds the .app (with Info.plist) [always required]
#   2. If --sign is requested:
#       - checks sign-specific env vars
#       - renders sign templates (entitlements + gon-sign.json)
#       - signs the .app with gon
#       - produces a signed .pkg (if MACOS_SIGN_INSTALLER_ID is set)
#   3. Else produces an unsigned .pkg
#   4. If --notarize is requested:
#       - checks notarization-specific env vars
#       - renders gon-notarize.json referencing the final .pkg
#       - notarizes using gon
#
# Core environment variables always required:
#   COMMON_PRODUCT_NAME, COMMON_BUILD_NAME, COMMON_PRODUCT_DESCRIPTION
#   MACOS_BUNDLE_ID, MACOS_BUILD_COMMAND, MACOS_APP_BUNDLE_PATH
#   MACOS_PKG_BUNDLE_PATH, MACOS_INFO_PLIST_PATH, MACOS_INFO_PLIST_TEMPLATE_PATH
#
# Additional env vars if --sign:
#   MACOS_ENTITLEMENTS_PATH, MACOS_ENTITLEMENTS_TEMPLATE_PATH (optional)
#   MACOS_GON_SIGN_JSON_PATH, MACOS_GON_SIGN_JSON_TEMPLATE_PATH
#   MACOS_SIGN_APPLE_USERNAME, MACOS_SIGN_APPLE_APP_PASSWORD, MACOS_SIGN_APPLE_TEAM_ID
#   MACOS_SIGN_APPLE_DEVELOPER_IDENTITY
#   MACOS_SIGN_INSTALLER_ID (for signing the .pkg)
#
# Additional env vars if --notarize:
#   MACOS_GON_NOTARIZE_JSON_PATH, MACOS_GON_NOTARIZE_JSON_TEMPLATE_PATH
#
# Command-line options:
#  --version <VERSION> : (Required) The version string for Info.plist, pkg name, etc.
#  --sign              : Perform code signing of the .app and produce a signed .pkg.
#  --notarize          : Notarize the .pkg with Apple.
################################################################################

################################################################################
# Parse CLI arguments
################################################################################

function usage() {
  cat <<EOF
Usage: $(basename "$0") [--version <version>] [--sign] [--notarize]

  --version <version>   Specify the app version (required, must be vX.Y.Z).
  --sign                Sign the .app (and sign the .pkg if MACOS_SIGN_INSTALLER_ID is set).
  --notarize            Notarize the resulting .pkg (requires --sign).
  -h, --help            Show this help.

Steps:
  1) Build the .app (always).
  2) If --sign, sign the .app and produce either a signed or unsigned .pkg.
  3) If --notarize, notarize that .pkg (requires --sign).

Ensure you have all required environment variables set for each requested step.
EOF
}

SIGN_APP=false
NOTARIZE_APP=false
VERSION_TAG=""

while [[ $# -gt 0 ]]; do
  case "$1" in
  --version)
    if [[ -z "${2:-}" ]]; then
      echo "Error: --version requires a value."
      exit 1
    fi
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

# Enforce version tag format: must start with v and be semantic version vX.Y.Z
if ! [[ "$VERSION_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: --version must be in format vX.Y.Z (e.g., v1.2.3)"
  exit 1
fi

# If --notarize is requested, --sign must also be requested
if [[ "$NOTARIZE_APP" == true && "$SIGN_APP" != true ]]; then
  echo "Error: --notarize requires --sign. Please specify both."
  usage
  exit 1
fi

################################################################################
# Check environment variables needed for ALWAYS-needed steps (building .app)
################################################################################

: "${COMMON_PRODUCT_NAME:?Must set COMMON_PRODUCT_NAME}"
: "${COMMON_BUILD_NAME:?Must set COMMON_BUILD_NAME}"
: "${COMMON_PRODUCT_DESCRIPTION:?Must set COMMON_PRODUCT_DESCRIPTION}"
: "${MACOS_BUNDLE_ID:?Must set MACOS_BUNDLE_ID}"
: "${MACOS_APP_BUNDLE_PATH:?Must set MACOS_APP_BUNDLE_PATH}"
: "${MACOS_PKG_BUNDLE_PATH:?Must set MACOS_PKG_BUNDLE_PATH}"
: "${MACOS_BUILD_COMMAND:?Must set MACOS_BUILD_COMMAND}"

# Templates and outputs always required for building the .app:
: "${MACOS_INFO_PLIST_TEMPLATE_PATH:?Must set MACOS_INFO_PLIST_TEMPLATE_PATH}"
: "${MACOS_INFO_PLIST_PATH:?Must set MACOS_INFO_PLIST_PATH}"

################################################################################
# Conditionally check environment if signing is requested
################################################################################

if [[ "$SIGN_APP" == true ]]; then
  : "${MACOS_GON_SIGN_JSON_TEMPLATE_PATH:?Must set MACOS_GON_SIGN_JSON_TEMPLATE_PATH if signing}"
  : "${MACOS_GON_SIGN_JSON_PATH:?Must set MACOS_GON_SIGN_JSON_PATH if signing}"

  : "${MACOS_SIGN_APPLE_USERNAME:?Must set MACOS_SIGN_APPLE_USERNAME if signing}"
  : "${MACOS_SIGN_APPLE_APP_PASSWORD:?Must set MACOS_SIGN_APPLE_APP_PASSWORD if signing}"
  : "${MACOS_SIGN_APPLE_TEAM_ID:?Must set MACOS_SIGN_APPLE_TEAM_ID if signing}"
  : "${MACOS_SIGN_APPLE_DEVELOPER_IDENTITY:?Must set MACOS_SIGN_APPLE_DEVELOPER_IDENTITY if signing}"

  # If using entitlements, check for them
  # (If your signing approach doesn't require separate entitlements, remove/modify these)
  : "${MACOS_ENTITLEMENTS_TEMPLATE_PATH:?Must set MACOS_ENTITLEMENTS_TEMPLATE_PATH if signing}"
  : "${MACOS_ENTITLEMENTS_PATH:?Must set MACOS_ENTITLEMENTS_PATH if signing}"
fi

################################################################################
# Conditionally check environment if notarization is requested
################################################################################

if [[ "$NOTARIZE_APP" == true ]]; then
  : "${MACOS_GON_NOTARIZE_JSON_TEMPLATE_PATH:?Must set MACOS_GON_NOTARIZE_JSON_TEMPLATE_PATH if notarizing}"
  : "${MACOS_GON_NOTARIZE_JSON_PATH:?Must set MACOS_GON_NOTARIZE_JSON_PATH if notarizing}"

  # Notarization needs signing too
fi

################################################################################
# Helper: Template rendering function using sed
################################################################################

render_template() {
  local template_file="$1"
  local dest_file="$2"

  sed -e "s|{{COMMON_PRODUCT_NAME}}|${COMMON_PRODUCT_NAME}|g" \
    -e "s|{{COMMON_BUILD_NAME}}|${COMMON_BUILD_NAME}|g" \
    -e "s|{{COMMON_PRODUCT_DESCRIPTION}}|${COMMON_PRODUCT_DESCRIPTION}|g" \
    -e "s|{{MACOS_BUNDLE_ID}}|${MACOS_BUNDLE_ID}|g" \
    -e "s|{{MACOS_APP_BUNDLE_PATH}}|${MACOS_APP_BUNDLE_PATH}|g" \
    -e "s|{{MACOS_PKG_BUNDLE_PATH}}|${MACOS_PKG_BUNDLE_PATH}|g" \
    -e "s|{{VERSION_TAG}}|${VERSION_TAG}|g" \
    -e "s|{{MACOS_SIGN_APPLE_USERNAME}}|${MACOS_SIGN_APPLE_USERNAME:-}|g" \
    -e "s|{{MACOS_SIGN_APPLE_APP_PASSWORD}}|${MACOS_SIGN_APPLE_APP_PASSWORD:-}|g" \
    -e "s|{{MACOS_SIGN_APPLE_TEAM_ID}}|${MACOS_SIGN_APPLE_TEAM_ID:-}|g" \
    -e "s|{{MACOS_SIGN_APPLE_DEVELOPER_IDENTITY}}|${MACOS_SIGN_APPLE_DEVELOPER_IDENTITY:-}|g" \
    -e "s|{{MACOS_SIGN_INSTALLER_ID}}|${MACOS_SIGN_INSTALLER_ID:-}|g" \
    -e "s|{{MACOS_ENTITLEMENTS_PATH}}|${MACOS_ENTITLEMENTS_PATH:-}|g" \
    "$template_file" >"$dest_file"
}

################################################################################
# Step 1: Render Info.plist (always needed), then build the .app
################################################################################

echo "Rendering Info.plist template..."
mkdir -p "$(dirname "$MACOS_INFO_PLIST_PATH")"
render_template "$MACOS_INFO_PLIST_TEMPLATE_PATH" "$MACOS_INFO_PLIST_PATH"

echo "Building .app with command: ${MACOS_BUILD_COMMAND}"
eval "${MACOS_BUILD_COMMAND}"

################################################################################
# Step 2: If signing is requested, render sign templates, sign the .app with gon
################################################################################

if [[ "$SIGN_APP" == true ]]; then
  echo "Signing requested. Rendering entitlements and gon-sign.json..."

  # Render entitlements
  mkdir -p "$(dirname "$MACOS_ENTITLEMENTS_PATH")"
  render_template "$MACOS_ENTITLEMENTS_TEMPLATE_PATH" "$MACOS_ENTITLEMENTS_PATH"

  # Render gon-sign.json
  mkdir -p "$(dirname "$MACOS_GON_SIGN_JSON_PATH")"
  render_template "$MACOS_GON_SIGN_JSON_TEMPLATE_PATH" "$MACOS_GON_SIGN_JSON_PATH"

  echo "Signing the app bundle with gon..."
  gon -log-level=info "$MACOS_GON_SIGN_JSON_PATH"
else
  echo "No signing requested. Skipping signing steps."
fi

################################################################################
# Step 3: Build the .pkg (signed if SIGN_APP and MACOS_SIGN_INSTALLER_ID are set)
################################################################################

mkdir -p "$(dirname "$MACOS_PKG_BUNDLE_PATH")"

if [[ "$SIGN_APP" == true && -n "${MACOS_SIGN_INSTALLER_ID:-}" ]]; then
  echo "Building signed installer pkg at: ${MACOS_PKG_BUNDLE_PATH}"
  productbuild \
    --sign "${MACOS_SIGN_INSTALLER_ID}" \
    --component "${MACOS_APP_BUNDLE_PATH}" "${MACOS_PKG_BUNDLE_PATH}"
else
  echo "Building unsigned installer pkg at: ${MACOS_PKG_BUNDLE_PATH}"
  productbuild \
    --component "${MACOS_APP_BUNDLE_PATH}" "${MACOS_PKG_BUNDLE_PATH}"
fi

################################################################################
# Step 4: If notarization requested, re-render gon-notarize.json (with final pkg)
#         and notarize the pkg.
################################################################################

if [[ "$SIGN_APP" == true && "$NOTARIZE_APP" == true ]]; then
  echo "Notarization requested. Rendering gon-notarize.json..."

  # If your gon-notarize template references MACOS_PKG_BUNDLE_PATH, re-render now
  mkdir -p "$(dirname "$MACOS_GON_NOTARIZE_JSON_PATH")"
  render_template "$MACOS_GON_NOTARIZE_JSON_TEMPLATE_PATH" "$MACOS_GON_NOTARIZE_JSON_PATH"

  echo "Notarizing .pkg with gon..."
  gon -log-level=info "$MACOS_GON_NOTARIZE_JSON_PATH"
else
  echo "No notarization requested. Skipping notarization steps."
fi

################################################################################
# Done
################################################################################

echo "Done! Final pkg is at: ${MACOS_PKG_BUNDLE_PATH}"
if [[ "$SIGN_APP" == true ]]; then
  echo "App was signed. Pkg may also be signed if MACOS_SIGN_INSTALLER_ID was set."
fi
if [[ "$NOTARIZE_APP" == true ]]; then
  echo "Pkg was notarized."
fi
