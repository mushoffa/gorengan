#!/usr/bin/env bash
# Open the release pull request into main, or refresh its description when one is
# already open for the branch.
#
# Usage: open-release-pr.sh release/vX.Y.Z vX.Y.Z
#
# The PR description is the version's changelog section, read from CHANGELOG.md in
# the current checkout. Needs the GitHub CLI authenticated through GH_TOKEN.
set -euo pipefail

BRANCH="${1:?usage: open-release-pr.sh release/vX.Y.Z vX.Y.Z}"
TAG="${2:?missing tag}"
MAIN_BRANCH="${MAIN_BRANCH:-main}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

body="$(mktemp)"
trap 'rm -f "${body}"' EXIT

if ! bash "${HERE}/release-notes.sh" "${TAG}" CHANGELOG.md >"${body}"; then
  echo "open-release-pr: CHANGELOG.md has no section for ${TAG}" >&2
  exit 1
fi
printf '\n---\nMerging this PR tags %s on the merge commit and publishes the GitHub Release.\n' "${TAG}" >>"${body}"

number="$(gh pr list --head "${BRANCH}" --base "${MAIN_BRANCH}" --state open --json number --jq '.[0].number // empty')"
if [ -n "${number}" ]; then
  gh pr edit "${number}" --body-file "${body}"
  echo "open-release-pr: updated PR #${number}"
else
  gh pr create --base "${MAIN_BRANCH}" --head "${BRANCH}" --title "chore(release): ${TAG#v}" --body-file "${body}"
fi
