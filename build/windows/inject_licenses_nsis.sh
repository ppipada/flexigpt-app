#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT_DIR}"

if [[ -f "build/buildvars.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "build/buildvars.env"
  set +a
fi

: "${WIN_INSTALLER_PATH:?Must set WIN_INSTALLER_PATH}"

LICENSE_DIR="${ROOT_DIR}/build/licenses"
[[ -d "${LICENSE_DIR}" ]] || { echo "ERROR: ${LICENSE_DIR} missing"; exit 1; }

# Prefer Wails default template locations, then fallback to searching.
NSI_CANDIDATES=(
  "${ROOT_DIR}/build/windows/installer/project.nsi"
  "${ROOT_DIR}/build/windows/installer/installer.nsi"
)

NSI=""
for f in "${NSI_CANDIDATES[@]}"; do
  if [[ -f "$f" ]]; then NSI="$f"; break; fi
done
if [[ -z "${NSI}" ]]; then
  NSI="$(find "${ROOT_DIR}/build" -type f -name "*.nsi" | head -n 1 || true)"
fi
[[ -n "${NSI}" && -f "${NSI}" ]] || { echo "ERROR: Could not locate a .nsi file under build/. Cannot inject licenses."; exit 1; }

echo "Using NSIS script: ${NSI}"

# Convert license dir to a Windows-friendly path for NSIS (C:/... style)
if command -v cygpath >/dev/null 2>&1; then
  LICENSE_DIR_WIN="$(cygpath -m "${LICENSE_DIR}")"
elif pwd -W >/dev/null 2>&1; then
  # MSYS sometimes supports this
  LICENSE_DIR_WIN="$(cd "${LICENSE_DIR}" && pwd -W | sed 's|\\|/|g')"
else
  LICENSE_DIR_WIN="${LICENSE_DIR}"
fi

PATCHED="${NSI%.nsi}.with-licenses.nsi"
cp -f "${NSI}" "${PATCHED}"

SENTINEL="FlexiGPT: bundled licenses"
if grep -q "${SENTINEL}" "${PATCHED}"; then
  echo "NSIS script already patched; rebuilding..."
else
  cat >> "${PATCHED}" <<EOF

; --- ${SENTINEL} ---
Section "-Licenses"
  SetOutPath "\$INSTDIR\\licenses"
  File /r "${LICENSE_DIR_WIN}/*"
SectionEnd
EOF
fi

MAKENSIS="${MAKENSIS:-}"
if [[ -z "${MAKENSIS}" ]]; then
  if command -v makensis.exe >/dev/null 2>&1; then
    MAKENSIS="$(command -v makensis.exe)"
  elif [[ -x "/c/Program Files (x86)/NSIS/makensis.exe" ]]; then
    MAKENSIS="/c/Program Files (x86)/NSIS/makensis.exe"
  else
    MAKENSIS="C:/Program Files (x86)/NSIS/makensis.exe"
  fi
fi

echo "Using makensis: ${MAKENSIS}"

MSYS2_ARG_CONV_EXCL="${MSYS2_ARG_CONV_EXCL:-};/V*" "${MAKENSIS}" /V2 "${PATCHED}"

[[ -f "${WIN_INSTALLER_PATH}" ]] || {
  echo "ERROR: Installer not found at expected path after rebuild: ${WIN_INSTALLER_PATH}"
  exit 1
}

echo "Installer rebuilt with licenses: ${WIN_INSTALLER_PATH}"
