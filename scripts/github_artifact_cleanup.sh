#!/usr/bin/env bash
# Change the date under CUTOFF_DATE to change how far back you want to delete
# Install the GitHub CLI tool by following the instructions in the official documentation: https://cli.github.com/manual/installation
# Make sure you auth first to github with 'gh auth login'

REPO_OWNER="ppipada"
REPO_NAME="flexigpt-app"
# CUTOFF_DATE=$(date --date='10 days ago' +'%Y-%m-%dT%H:%M:%SZ')
CUTOFF_DATE=$(date -u -v-10d +'%Y-%m-%dT%H:%M:%SZ')
PAGE=1

while true; do
	# Retrieve a page of artifacts
	ART_EXIST=$(gh api repos/$REPO_OWNER/$REPO_NAME/actions/artifacts?per_page=100\&page=$PAGE | jq -r '.artifacts[]')
	ARTIFACTS=$(gh api repos/$REPO_OWNER/$REPO_NAME/actions/artifacts?per_page=100\&page=$PAGE | jq -r '.artifacts[] | select(.created_at < "'"$CUTOFF_DATE"'") | .id')
	echo $PAGE
	# If there are no more artifacts, exit the loop
	if [[ -z "$ART_EXIST" ]]; then
		break
	fi

	# Loop through the artifacts on this page and delete the old ones
	for ARTIFACT_ID in $ARTIFACTS; do
		ARTIFACT_NAME=$(gh api repos/$REPO_OWNER/$REPO_NAME/actions/artifacts/$ARTIFACT_ID | jq -r '.name')
		echo "Deleting artifact $ARTIFACT_NAME (ID: $ARTIFACT_ID)..."
		gh api repos/$REPO_OWNER/$REPO_NAME/actions/artifacts/$ARTIFACT_ID -X DELETE
	done

	# Increment the page counter
	PAGE=$((PAGE + 1))
done
