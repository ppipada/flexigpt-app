#!/usr/bin/env bash
set -euo pipefail
################################################################################
#
# Renders Flatpak templates (.desktop & .appdata.xml)
#
# Usage:
#   ./render_flatpak_templates.sh --version vX.Y.Z
#
# See build/buildvars.env for env variables required
################################################################################

###############################################################################
# CLI parsing
###############################################################################
VERSION_TAG=""

usage() {
    grep '^#' "\$0" | sed -e 's/^#//'
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "\$1" in
    --version)
        VERSION_TAG="${2:-}"
        shift 2
        ;;
    -h | --help) usage ;;
    *)
        echo "Unknown arg: \$1"
        usage
        ;;
    esac
done

[[ -z "$VERSION_TAG" ]] && {
    echo "ERROR: --version is required"
    usage
}

[[ "$VERSION_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] ||
    {
        echo "ERROR: version must match vX.Y.Z"
        exit 1
    }

APPDATA_DATE="$(date +%Y-%m-%d)"
###############################################################################
# Env-var sanity checks
###############################################################################
: "${FLATPAK_ID:?}"
: "${COMMON_BUILD_NAME:?}"
: "${COMMON_PRODUCT_NAME:?}"
: "${COMMON_PRODUCT_GENERIC_NAME:?}"
: "${COMMON_PRODUCT_DESCRIPTION:?}"
: "${COMMON_DEVELOPER_NAME:?}"
: "${COMMON_PROJECT_LICENSE:?}"

: "${FLATPAK_APPDATA_TEMPLATE_PATH:?}"
: "${FLATPAK_APPDATA_PATH:?}"
: "${FLATPAK_DESKTOP_TEMPLATE_PATH:?}"
: "${FLATPAK_DESKTOP_PATH:?}"

###############################################################################
# Helper: simple template renderer with sed substitution
###############################################################################
render_tpl() {
    local tpl="\$1" out="\$2"
    sed \
        -e "s|{{COMMON_BUILD_NAME}}|${COMMON_BUILD_NAME}|g" \
        -e "s|{{COMMON_PRODUCT_NAME}}|${COMMON_PRODUCT_NAME}|g" \
        -e "s|{{COMMON_PRODUCT_GENERIC_NAME}}|${COMMON_PRODUCT_GENERIC_NAME}|g" \
        -e "s|{{COMMON_PRODUCT_DESCRIPTION}}|${COMMON_PRODUCT_DESCRIPTION}|g" \
        -e "s|{{COMMON_DEVELOPER_NAME}}|${COMMON_DEVELOPER_NAME}|g" \
        -e "s|{{COMMON_PROJECT_LICENSE}}|${COMMON_PROJECT_LICENSE}|g" \
        -e "s|{{FLATPAK_ID}}|${FLATPAK_ID}|g" \
        -e "s|{{VERSION_TAG}}|${VERSION_TAG}|g" \
        -e "s|{{APPDATA_DATE}}|${APPDATA_DATE}|g" \
        "$tpl" >"$out"
}

echo "🔧  Rendering Flatpak templates..."
mkdir -p "$(dirname "$FLATPAK_APPDATA_PATH")" "$(dirname "$FLATPAK_DESKTOP_PATH")"
render_tpl "$FLATPAK_APPDATA_TEMPLATE_PATH" "$FLATPAK_APPDATA_PATH"
render_tpl "$FLATPAK_DESKTOP_TEMPLATE_PATH" "$FLATPAK_DESKTOP_PATH"
echo "   • $FLATPAK_APPDATA_PATH"
echo "   • $FLATPAK_DESKTOP_PATH"
echo "✅  Done."
