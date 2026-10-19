#!/usr/bin/env bash
# Inspect a release branch before the workflow changes anything.
#
# Usage: release-inspect.sh release/vX.Y.Z
#
# Prints, in GITHUB_OUTPUT format:
#   tag=vX.Y.Z
#   needs_changelog=true|false   (false when CHANGELOG.md already has this version)
#
# Refuses, with a message on stderr and a non-zero exit, when:
#   - the branch is not named release/vX.Y.Z
#   - the tag already exists
#   - the version is not newer than the latest release tag
set -euo pipefail

BRANCH="${1:?usage: release-inspect.sh release/vX.Y.Z}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! [[ "${BRANCH}" =~ ^release/(v[0-9]+\.[0-9]+\.[0-9]+([-+].+)?)$ ]]; then
  echo "release-inspect: '${BRANCH}' is not a release branch named release/vX.Y.Z" >&2
  exit 1
fi
TAG="${BASH_REMATCH[1]}"

if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
  echo "release-inspect: tag ${TAG} already exists; use a new version" >&2
  exit 1
fi

# A new release must be newer than the latest one (plain versions only; pre-release
# suffixes are not ordered here).
prev="$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true)"
if [ -n "${prev}" ] && [[ "${prev}" != *-* ]] && [[ "${TAG}" != *-* ]]; then
  newest="$(printf '%s\n%s\n' "${prev}" "${TAG}" | sort -V | tail -n 1)"
  if [ "${newest}" != "${TAG}" ]; then
    echo "release-inspect: ${TAG} is not newer than the latest release ${prev}" >&2
    exit 1
  fi
fi

if [ -f CHANGELOG.md ] && bash "${HERE}/release-notes.sh" "${TAG}" CHANGELOG.md >/dev/null 2>&1; then
  needs=false
else
  needs=true
fi

echo "tag=${TAG}"
echo "needs_changelog=${needs}"
