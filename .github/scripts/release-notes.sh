#!/usr/bin/env bash
# Print the CHANGELOG.md section for one version, without its "## ..." heading.
#
# Usage: release-notes.sh vX.Y.Z [CHANGELOG.md]   (reads stdin when no file is given)
#
# Understands both heading shapes git-cliff writes with cliff.toml:
#   ## [0.5.0](https://github.com/.../compare/v0.4.0...v0.5.0) (2025-03-07)
#   ## 0.1.0 (2025-03-03)
# Exits non-zero, printing nothing, when the version has no section.
set -euo pipefail

ver="${1:?usage: release-notes.sh vX.Y.Z [CHANGELOG.md]}"
ver="${ver#v}"

out="$(
  awk -v v="${ver}" '
    /^## / {
      if (found) exit
      if (index($0, "## [" v "]") == 1 || index($0, "## " v " ") == 1 || $0 == "## " v) {
        found = 1
        next
      }
    }
    found { print }
  ' "${2:--}" | sed '/./,$!d'
)"

if [ -z "${out}" ]; then
  echo "release-notes: no CHANGELOG section found for ${ver}" >&2
  exit 1
fi

printf '%s\n' "${out}"
