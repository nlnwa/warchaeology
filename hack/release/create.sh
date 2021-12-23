#!/usr/bin/env bash
#
# Copyright 2021 National Library of Norway.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#

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
go run hack/release/md_docs.go ${1}
make_commit "v${1}"
add_tag "v${1}"

# print follow-up instructions
echo ""
echo "Created commit for v${1}, you should now:"
echo " - git push ${UPSTREAM} v${1}"
echo " - Create a GitHub release from the pushed tag v${1}"
