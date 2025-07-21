#!/usr/bin/env bash
#
# lint_gopls_modernize.sh  [DIR]
#
# Runs gopls’ modernize analysis on all Go packages under DIR
# (defaults to ".").  In CI the front-end build artefacts do not exist;
# a dummy file is created so that //go:embed directives that point to
# frontend/dist/client do not break the build.
#

set -euo pipefail

TARGET=${1:-.}

echo "==> gopls modernize (${TARGET}/...)"

# --------------------------------------------------------------------------- #
# Dummy file so that //go:embed globs match at least one file
DUMMY="frontend/dist/client/.ci-placeholder"
mkdir -p "$(dirname "$DUMMY")"
touch "$DUMMY"
trap 'rm -f "$DUMMY"' EXIT      # make sure we delete it on success *or* failure
# --------------------------------------------------------------------------- #

# Prefer cached binary; fall back to 'go run' when run locally.
if command -v modernize >/dev/null 2>&1; then
  modernize "${TARGET}/..."
else
  go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest \
        "${TARGET}/..."
fi

echo "✓ gopls modernize passed"
