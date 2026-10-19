#!/usr/bin/env bash
# Commit CHANGELOG.md and push it to its release branch.
#
# Usage: commit-changelog.sh vX.Y.Z release/vX.Y.Z
#
# Does nothing when CHANGELOG.md has no changes. Uses the github-actions[bot]
# identity unless git already has one configured. A push made with the workflow's
# default GITHUB_TOKEN does not start another workflow run.
set -euo pipefail

TAG="${1:?usage: commit-changelog.sh vX.Y.Z release/vX.Y.Z}"
BRANCH="${2:?missing branch}"

git config user.name >/dev/null 2>&1 || git config user.name "github-actions[bot]"
git config user.email >/dev/null 2>&1 || git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

git add CHANGELOG.md
if git diff --cached --quiet; then
  echo "commit-changelog: CHANGELOG.md has no changes; nothing to commit"
  exit 0
fi

git commit -q -m "chore(release): ${TAG#v}"
git push -q origin "HEAD:refs/heads/${BRANCH}"
echo "commit-changelog: pushed the changelog for ${TAG} to ${BRANCH}"
