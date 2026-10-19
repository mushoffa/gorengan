#!/usr/bin/env bash
# Tag the merge commit of a merged release PR and publish its GitHub Release.
#
# Usage: publish-release.sh vX.Y.Z <merge-sha> <head-sha>
#   merge-sha  the commit the PR produced on main (merge, squash or rebase commit)
#   head-sha   the tip of the release branch
#
# Needs: git with full history and tags (including the PR head commit), the GitHub
# CLI authenticated via GH_TOKEN, run from a checkout of the repository.
#
# The tag is created by GitHub together with the release, on the merge commit, so
# it contains the new CHANGELOG.md and is never moved. The release notes are that
# version's changelog section followed by a link to the full diff against the
# previous tag. Re-running is safe: an existing release is updated, not recreated.
set -euo pipefail

TAG="${1:?usage: publish-release.sh vX.Y.Z <merge-sha> <head-sha>}"
MERGE_SHA="${2:?missing merge-sha}"
HEAD_SHA="${3:?missing head-sha}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

MERGE_SHA="$(git rev-parse --verify "${MERGE_SHA}^{commit}")"
MAIN_BEFORE="$(git rev-parse --verify "${MERGE_SHA}^1")"

# Same checks as on the open PR: the changelog section exists and main did not
# move since the branch was created.
bash "${HERE}/verify-release.sh" "${TAG}" "${HEAD_SHA}" "${MAIN_BEFORE}"

# Repository as owner/name: from Actions, otherwise from the origin URL.
REPO="${GITHUB_REPOSITORY:-}"
if [ -z "${REPO}" ]; then
  url="$(git remote get-url origin 2>/dev/null || true)"
  if [[ "${url}" =~ github\.com[:/]([^/]+/[^/]+)$ ]]; then
    REPO="${BASH_REMATCH[1]%.git}"
  fi
fi

body="$(mktemp)"
trap 'rm -f "${body}"' EXIT

# Notes come from the changelog in the merged commit itself.
if ! git show "${MERGE_SHA}:CHANGELOG.md" | bash "${HERE}/release-notes.sh" "${TAG}" >"${body}"; then
  echo "publish-release: the merged commit has no changelog section for ${TAG}" >&2
  exit 1
fi

prev="$(git describe --tags --abbrev=0 --match 'v[0-9]*' "${MERGE_SHA}^" 2>/dev/null || true)"
if [ -n "${prev}" ] && [ -n "${REPO}" ]; then
  printf '\n**Full Changelog**: https://github.com/%s/compare/%s...%s\n' "${REPO}" "${prev}" "${TAG}" >>"${body}"
fi

# An existing tag must already point at the merge commit; it is never moved.
tag_exists=0
if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
  tag_exists=1
  if [ "$(git rev-parse "${TAG}^{commit}")" != "${MERGE_SHA}" ]; then
    echo "publish-release: tag ${TAG} exists but does not point at the merge commit ${MERGE_SHA}; refusing to continue" >&2
    exit 1
  fi
fi

flags=()
case "${TAG}" in
  *-*) flags+=(--prerelease) ;;
esac

if gh release view "${TAG}" >/dev/null 2>&1; then
  gh release edit "${TAG}" --notes-file "${body}"
  echo "Updated release ${TAG}"
else
  target=()
  if [ "${tag_exists}" = "0" ]; then
    target+=(--target "${MERGE_SHA}")
  fi
  gh release create "${TAG}" --title "${TAG}" --notes-file "${body}" \
    ${target[@]+"${target[@]}"} ${flags[@]+"${flags[@]}"}
  echo "Created release ${TAG} at ${MERGE_SHA}"
fi
