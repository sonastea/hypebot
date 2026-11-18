#!/usr/bin/env bash
# Decide whether HEAD should be released, following Conventional Commits:
#   type!: ... or a "BREAKING CHANGE:" footer -> major
#   feat: ...                                 -> minor
#   fix: ... / perf: ...                      -> patch
#   anything else (chore, ci, docs, test, ...) -> no release
#
# Prints release=, version= and latest= lines for $GITHUB_OUTPUT.
set -euo pipefail

semver='^v[0-9]+\.[0-9]+\.[0-9]+$'

# Highest-versioned release tag from git tag output on stdin.
highest_tag() { awk -v re="$semver" '$0 ~ re && !found { print; found = 1 }'; }

highest=$(git tag --list 'v*' --sort=-v:refname | highest_tag)
current=$(git tag --points-at HEAD --list 'v*' --sort=-v:refname | highest_tag)

emit() { printf 'release=%s\nversion=%s\nlatest=%s\n' "$1" "${2:-}" "${3:-false}"; }

# Release a version; only the highest version becomes "latest" and is deployed.
release() {
  local latest=true
  if [[ -n $highest && $(printf '%s\n%s\n' "$1" "$highest" | sort -V | tail -n1) != "$1" ]]; then
    latest=false
  fi
  emit true "$1" "$latest"
}

if [[ -n $current ]]; then
  # A rerun of an already tagged commit: finish that release.
  echo "HEAD is already released as $current" >&2
  release "$current"
  exit 0
fi

if [[ -n $highest ]] && git merge-base --is-ancestor HEAD "$highest"; then
  echo "HEAD is already included in $highest" >&2
  emit false
  exit 0
fi

base=$(git tag --merged HEAD --list 'v*' --sort=-v:refname | highest_tag)
range=${base:+$base..}HEAD
subjects=$(git log --format=%s "$range")
bodies=$(git log --format=%b "$range")

scope='(\([^)]*\))?'
if grep -Eq "^[a-z]+$scope!:" <<<"$subjects" || grep -Eq '^BREAKING[ -]CHANGE:' <<<"$bodies"; then
  bump=major
elif grep -Eq "^feat$scope:" <<<"$subjects"; then
  bump=minor
elif grep -Eq "^(fix|perf)$scope:" <<<"$subjects"; then
  bump=patch
else
  echo "No feat, fix, perf or breaking commits since ${base:-the first commit}" >&2
  emit false
  exit 0
fi

IFS=. read -r major minor patch <<<"${base:-v0.0.0}"
major=${major#v}
case $bump in
  major) version="v$((major + 1)).0.0" ;;
  minor) version="v$major.$((minor + 1)).0" ;;
  patch) version="v$major.$minor.$((patch + 1))" ;;
esac

if git rev-parse --quiet --verify "refs/tags/$version" >/dev/null; then
  echo "$version already exists on another commit" >&2
  exit 1
fi

echo "Releasing $version ($bump bump from ${base:-nothing})" >&2
release "$version"
