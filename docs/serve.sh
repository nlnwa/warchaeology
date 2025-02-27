#!/usr/bin/env bash

# Serve documentation on http://localhost:1313
set -o errexit -o nounset -o pipefail

# cd to the repo root
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "${REPO_ROOT}"

GITHUB_REF_NAME=${GITHUB_REF_NAME:-next} HUGO_BASEURL=http://localhost/ hugo serve --buildDrafts --source docs
