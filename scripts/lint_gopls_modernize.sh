#!/usr/bin/env bash
#
# gopls_modernize.sh  [PATH]
#
# Runs "gopls modernize" once for all packages under PATH
# (defaults to the current directory).  Exits 0 on success,
# non-zero otherwise.

set -euo pipefail

TARGET=${1:-'.'}

echo "==> gopls modernize"
tmp=$TARGET/frontend/dist/client/a.tmp
touch tmp
if out=$(go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest \
	"$TARGET/..." 2>&1); then
	echo "✓ gopls modernize passed"
	rm -f tmp
	exit 0
else
	printf '%s\n' "$out"
	echo "✗ gopls modernize failed"
	rm -f tmp
	exit 1
fi

