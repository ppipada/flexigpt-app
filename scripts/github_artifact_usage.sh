#!/bin/bash

OWNER="ppipada"
REPO="flexigpt-app"

gh api -X GET "repos/$OWNER/$REPO/actions/artifacts" --paginate \
	--jq '.artifacts[] | "\(.name)\t\(.size_in_bytes / 1024 / 1024 | tostring) MB"' |
	column -t -s $'\t'
