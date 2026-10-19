#!/usr/bin/env bash
# Check that a release pull request is consistent before it is merged or published.
#
# Usage: verify-release.sh vX.Y.Z <head-sha> <main-sha>
#   head-sha  tip of the release branch (the commit that regenerates CHANGELOG.md)
#   main-sha  tip of main that the release branch has to be built on
#
# Set REQUIRE_NEW_TAG=1 to also fail when the tag already exists (used while the
# PR is open; after the merge the tag may legitimately exist from an earlier run).
#
# Checks:
#   1. CHANGELOG.md on the release branch has a section for this version.
#   2. main has not moved since the branch was created. If it had, commits on main
#      would be released without appearing in the changelog.
set -euo pipefail

TAG="${1:?usage: verify-release.sh vX.Y.Z <head-sha> <main-sha>}"
HEAD_SHA="${2:?missing head-sha}"
MAIN_SHA="${3:?missing main-sha}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! [[ "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-+].+)?$ ]]; then
  echo "verify-release: '${TAG}' is not a version like v1.2.3 (release branches must be named release/vX.Y.Z)" >&2
  exit 1
fi

if ! git show "${HEAD_SHA}:CHANGELOG.md" 2>/dev/null | bash "${HERE}/release-notes.sh" "${TAG}" >/dev/null 2>&1; then
  echo "verify-release: CHANGELOG.md on the release branch has no section for ${TAG}; create the branch with 'make release-*'" >&2
  exit 1
fi

if ! git merge-base --is-ancestor "${MAIN_SHA}" "${HEAD_SHA}"; then
  echo "verify-release: main has new commits since the release branch was created, so the changelog would miss them." >&2
  echo "verify-release: close this PR, delete the branch and run 'make release-*' again." >&2
  exit 1
fi

if [ "${REQUIRE_NEW_TAG:-0}" = "1" ] && git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
  echo "verify-release: tag ${TAG} already exists" >&2
  exit 1
fi

echo "verify-release: ${TAG} looks consistent"
