#!/usr/bin/env bash
set -euo pipefail # use bash for pipefail, otherwise identical

FINDPATH=${1:-'.'}
JOBS=${2:-4}

FILES=$(find "$FINDPATH" -type f -name '*.go')

echo "==> gopls check - up to $JOBS parallel jobs"

tmpfile=$(mktemp)
trap 'rm -f "$tmpfile"' EXIT
export TMPFILE=$tmpfile # let the subshells see it

# Find all *.go files and run gopls in parallel

printf '%s\n' $FILES |
	xargs -P"$JOBS" -I{} bash -c '
		file=$1
		# Capture gopls output
		out=$(gopls check -- "$file" 2>&1)
		if [ -n "$out" ]; then
				# show diagnostics
				printf "%s\n" "$out"
				echo "$file:1" >> "$TMPFILE"   # mark as failed
		else
				echo "$file:0" >> "$TMPFILE"
		fi
		' _ {}

# If any file was marked :1, fail the script
# cat $tmpfile
if grep -q ":1$" "$tmpfile"; then
	echo "✗ gopls check failed on at least one file"
	exit 1
fi

echo "gopls check passed"
