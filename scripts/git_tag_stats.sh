#!/bin/bash

export GIT_PAGER=cat

# Usage: ./git_stats.sh <tag1: older tag> <tag2: newer tag>
tag1="$1"
tag2="$2"

if [[ -z "$tag1" || -z "$tag2" ]]; then
	echo "Usage: $0 <tag1> <tag2>"
	exit 1
fi

echo "=== Net Work (final diff between $tag1 and $tag2) ==="
git diff --shortstat "$tag1" "$tag2"

echo
echo "=== Actual Work (churn between $tag1 and $tag2) ==="

# Unique files touched (added, modified, or deleted)
unique_files=$(git log "$tag1".."$tag2" --name-only --pretty=format: | sort | uniq | wc -l)

# Files added
added_files=$(git log "$tag1".."$tag2" --diff-filter=A --name-only --pretty=format: | sort | uniq | wc -l)

# Files deleted
deleted_files=$(git log "$tag1".."$tag2" --diff-filter=D --name-only --pretty=format: | sort | uniq | wc -l)

# Files modified (touched but not added or deleted)
modified_files=$((unique_files - added_files - deleted_files))

# Lines added
lines_added=$(git log "$tag1".."$tag2" --pretty=tformat: --numstat | awk '{add+=$1} END {print add+0}')

# Lines removed
lines_removed=$(git log "$tag1".."$tag2" --pretty=tformat: --numstat | awk '{del+=$2} END {print del+0}')

echo "Files touched: $unique_files, added: $added_files, deleted: $deleted_files, modified: $modified_files"
echo "Lines added: $lines_added, removed: $lines_removed"
echo "==="
