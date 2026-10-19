#!/usr/bin/env bash
# Print the next release version (with a leading "v") derived from the latest git tag.
#
# Usage: scripts/next-version.sh major|minor|patch|X.Y.Z|vX.Y.Z
#
# Replaces the old `standard-version --dry-run --release-as <x> | sed ...` trick:
# the next version is computed from tags directly, so it does not depend on any
# tool's output format. With no existing tag, the base is v0.0.0.
set -euo pipefail

arg="${1:?usage: next-version.sh major|minor|patch|X.Y.Z|vX.Y.Z}"

prev="$(git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || true)"
prev="${prev:-v0.0.0}"

IFS=. read -r major minor patch <<<"${prev#v}"
patch="${patch%%[-+]*}" # drop any pre-release/build suffix such as -rc.1

case "$arg" in
  major) echo "v$((major + 1)).0.0" ;;
  minor) echo "v${major}.$((minor + 1)).0" ;;
  patch) echo "v${major}.${minor}.$((patch + 1))" ;;
  v[0-9]*.[0-9]*.[0-9]*) echo "$arg" ;;
  [0-9]*.[0-9]*.[0-9]*) echo "v$arg" ;;
  *)
    echo "next-version: '$arg' is not major, minor, patch or a version like 1.2.3" >&2
    exit 1
    ;;
esac
