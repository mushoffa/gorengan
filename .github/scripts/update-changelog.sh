#!/usr/bin/env bash
# Add the section for one new version to the top of CHANGELOG.md, leaving everything
# already in the file exactly as it is.
#
# Usage: update-changelog.sh vX.Y.Z [CHANGELOG.md]
#
# git-cliff only renders the unreleased commits:
#     git-cliff --unreleased --tag vX.Y.Z --strip all
# and this script splices that section in above the newest existing "## " entry. So
# the file's header and every older entry (including those written by
# standard-version) are kept byte for byte, and git-cliff never has to read, rewrite
# or de-duplicate the existing file. Without a CHANGELOG.md, one is created.
#
# When SECTION_FILE is set, that file is used as the rendered section instead of
# running git-cliff (the workflow renders it with git-cliff-action).
#
# The heading git-cliff renders is "## X.Y.Z (date)". When an earlier release tag
# exists it is turned into the linked form the older entries use:
#     ## [X.Y.Z](https://github.com/OWNER/REPO/compare/vPREV...vX.Y.Z) (date)
#
# Exits non-zero and leaves the file untouched when the version already has an
# entry, or when there are no changelog-worthy commits since the last tag.
set -euo pipefail

TAG="${1:?usage: update-changelog.sh vX.Y.Z [CHANGELOG.md]}"
FILE="${2:-CHANGELOG.md}"
CLIFF="${CLIFF:-git-cliff}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! [[ "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-+].+)?$ ]]; then
  echo "update-changelog: '${TAG}' is not a version like v1.2.3" >&2
  exit 1
fi
ver="${TAG#v}"

if [ -f "${FILE}" ] && bash "${HERE}/release-notes.sh" "${TAG}" "${FILE}" >/dev/null 2>&1; then
  echo "update-changelog: ${FILE} already has an entry for ${ver}" >&2
  exit 1
fi

if [ -n "${SECTION_FILE:-}" ]; then
  # The section was already rendered, for example by git-cliff-action in CI.
  if [ ! -f "${SECTION_FILE}" ]; then
    echo "update-changelog: SECTION_FILE '${SECTION_FILE}' does not exist" >&2
    exit 1
  fi
  section="$(cat "${SECTION_FILE}")"
else
  section="$("${CLIFF}" --unreleased --tag "${TAG}" --strip all)"
fi
section="$(printf '%s\n' "${section}" | sed '/./,$!d')" # drop leading blank lines

# Require a heading and at least one entry (a bullet line).
first="${section%%$'\n'*}"
if [[ "${first}" != "## "* ]] || ! printf '%s\n' "${section}" | grep -q -E '^[*-] '; then
  echo "update-changelog: no changelog-worthy commits (feat, fix, perf, refactor, docs) since the last release" >&2
  exit 1
fi

# Link the heading to the diff against the previous release, as older entries do.
prev="$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true)"
repo="${GITHUB_REPOSITORY:-}"
if [ -z "${repo}" ]; then
  url="$(git remote get-url origin 2>/dev/null || true)"
  if [[ "${url}" =~ github\.com[:/]([^/]+/[^/]+)$ ]]; then
    repo="${BASH_REMATCH[1]%.git}"
  fi
fi
# Without a previous tag or a GitHub remote the heading stays unlinked, never wrong.
if [ -n "${prev}" ] && [ -n "${repo}" ] && [[ "${first}" == "## ${ver} "* ]]; then
  rest="${first#"## ${ver} "}"
  linked="## [${ver}](https://github.com/${repo}/compare/${prev}...${TAG}) ${rest}"
  section="${linked}${section#"${first}"}"
fi

secfile="$(mktemp)"
tmp="$(mktemp)"
trap 'rm -f "${secfile}" "${tmp}"' EXIT
printf '%s\n' "${section}" >"${secfile}"

if [ ! -f "${FILE}" ]; then
  { printf '# Changelog\n\n'; cat "${secfile}"; } >"${tmp}"
elif grep -q '^## ' "${FILE}"; then
  # Insert above the first existing entry.
  awk -v secfile="${secfile}" '
    /^## / && !done {
      while ((getline line < secfile) > 0) print line
      print ""
      done = 1
    }
    { print }
  ' "${FILE}" >"${tmp}"
else
  # A file with a header but no entries yet: append.
  { cat "${FILE}"; printf '\n'; cat "${secfile}"; } >"${tmp}"
fi

cat "${tmp}" >"${FILE}"
echo "update-changelog: added ${ver} to ${FILE}"
