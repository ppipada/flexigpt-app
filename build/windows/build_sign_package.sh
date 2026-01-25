#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: build/windows/build_sign_package.sh --version vX.Y.Z [--sign]

Builds the Windows Wails app + NSIS installer (installer bundles build/licenses).
Optionally signs binaries if SIGN_WINDOWS_CERT/SIGN_WINDOWS_CERT_PASSWORD are set.
EOF
}

VERSION_TAG=""
DO_SIGN="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION_TAG="${2:-}"; shift 2 ;;
    --sign) DO_SIGN="true"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown arg: $1"; usage; exit 1 ;;
  esac
done

[[ -n "${VERSION_TAG}" ]] || { echo "ERROR: --version required"; usage; exit 1; }
[[ "${VERSION_TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "ERROR: version must be vX.Y.Z"; exit 1; }

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT_DIR}"

if [[ -f "build/buildvars.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "build/buildvars.env"
  set +a
fi

: "${WIN_BUILD_COMMAND:?Must set WIN_BUILD_COMMAND}"
: "${WIN_INSTALLER_PATH:?Must set WIN_INSTALLER_PATH}"

echo "==> Generating licenses..."
chmod +x build/licenses/gen_licenses.sh
build/licenses/gen_licenses.sh --version "${VERSION_TAG}"

echo "==> Building Windows app with: ${WIN_BUILD_COMMAND}"
export VERSION_TAG="${VERSION_TAG}"
eval "${WIN_BUILD_COMMAND}"


if [[ "${DO_SIGN}" == "true" ]]; then
  echo "==> Signing requested..."
  if [[ -z "${SIGN_WINDOWS_CERT:-}" || -z "${SIGN_WINDOWS_CERT_PASSWORD:-}" ]]; then
    echo "ERROR: --sign requested but SIGN_WINDOWS_CERT or SIGN_WINDOWS_CERT_PASSWORD not set"
    exit 1
  fi

  if command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File build/windows/sign_windows.ps1
  else
    echo "ERROR: powershell.exe not found; cannot sign"
    exit 1
  fi
fi

echo "==> Done."
echo "Installer: ${WIN_INSTALLER_PATH}"
