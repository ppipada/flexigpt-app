#!/bin/bash

export GIT_PAGER=cat

# Usage: ./scripts/git_tag_stats.sh [<tag1: older tag>] [<tag2: newer tag>]
# Start of time to now.
# ./scripts/git_tag_stats.sh  "$(git rev-list --max-parents=0 HEAD)" HEAD
# Last two tags example.
# ./scripts/git_tag_stats.sh v0.0.30 v0.0.31
# No args:
#   - last two semantic-version tags (older -> newer)
#   - if only one semver tag: initial commit -> that tag
#   - if no semver tags: initial commit -> HEAD

tag1="$1"
tag2="$2"

# Determine revisions to use
mode=""
if [[ -n "$tag1" && -n "$tag2" ]]; then
  mode="provided-2"
elif [[ -n "$tag1" && -z "$tag2" ]]; then
  tag2="HEAD"
  mode="provided-1"
else
  # Auto-pick revisions based on available semantic version tags (v-prefixed)
  mapfile -t _tags < <(git tag -l 'v[0-9]*' --sort=-v:refname | head -n 2)
  init_commit="$(git rev-list --max-parents=0 HEAD | head -n 1)"
  if [[ ${#_tags[@]} -ge 2 ]]; then
    tag1="${_tags[1]}"
    tag2="${_tags[0]}"
    mode="auto-2semver"
  elif [[ ${#_tags[@]} -eq 1 ]]; then
    tag1="$init_commit"
    tag2="${_tags[0]}"
    mode="auto-1semver"
  else
    tag1="$init_commit"
    tag2="HEAD"
    mode="auto-none"
  fi
fi

# Validate chosen revisions resolve to commits
for rev in "$tag1" "$tag2"; do
  if ! git rev-parse --verify --quiet "$rev^{commit}" >/dev/null; then
    echo "Error: revision '$rev' is not a valid commit-ish" >&2
    exit 1
  fi
done

# Announce what will be used
case "$mode" in
  provided-2)  echo "Using provided revisions: older='$tag1', newer='$tag2'";;
  provided-1)  echo "Using provided revision and HEAD: older='$tag1', newer='$tag2'";;
  auto-2semver) echo "Auto-selected last two semver tags: older='$tag1', newer='$tag2'";;
  auto-1semver) echo "Only one semver tag found; using initial commit -> '$tag2'";;
  auto-none)   echo "No semver tags found; using initial commit -> HEAD";;
esac

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
