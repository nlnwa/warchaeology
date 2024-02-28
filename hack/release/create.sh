#!/usr/bin/env bash

# creates a release
# Use like: create.sh <release-version>
# EG: create.sh 0.3.0
set -o errexit -o nounset -o pipefail

UPSTREAM='https://github.com/nlnwa/warchaeology.git'

# cd to the repo root
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
cd "${REPO_ROOT}"

# check for arguments
if [ "$#" -ne 1 ]; then
    echo "Usage: create.sh release-version"
    exit 1
fi

# make a commit denoting the version ($1)
make_commit() {
  git add docs
  git commit -m "version ${1}"
  echo "Created commit for ${1}"
}

# add a git tag with $1
add_tag() {
  git tag "${1}"
  echo "Tagged ${1}"
}

# create the release docs and tag it
go run hack/release/md_docs.go
make_commit "v${1}"
add_tag "v${1}"

# print follow-up instructions
echo ""
echo "Created commit for v${1}, you should now:"
echo "Push the generated docs on the main branch"
echo " - git push"
echo "Push the generated tag. Github actions will then create a release"
echo " - git push ${UPSTREAM} v${1}"
