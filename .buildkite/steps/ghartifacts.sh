#!/usr/bin/env bash
set -eu

readonly REPO="authelia/authelia"
PREV_TAG="$(git describe --tags --abbrev=0 "${BUILDKITE_TAG}^")"
readonly PREV_TAG

assets=()

for FILE in \
  checksums.sha256{,.sig} \
  authelia-${BUILDKITE_TAG}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.{tar.gz,tar.gz.{c,sp}dx.json} \
  authelia_${BUILDKITE_TAG/v/}-1_{amd64,armhf,arm64}.deb
do
  assets+=("${FILE}")
done

COMPARE_FILE="$(mktemp)"
MAP_FILE="$(mktemp)"
NOTES_FILE="$(mktemp)"
trap 'rm -f "${COMPARE_FILE}" "${MAP_FILE}" "${NOTES_FILE}"' EXIT

gh api "repos/${REPO}/compare/${PREV_TAG}...${BUILDKITE_TAG}" --paginate > "${COMPARE_FILE}"
jq -r '.commits[] | "\(.sha)\t\(.author.login // "")"' "${COMPARE_FILE}" > "${MAP_FILE}"

echo "--- :github: Build release notes for: ${BUILDKITE_TAG}"

conventional-changelog -p angular -o /dev/stdout -r 2 \
  | sed -e '1,3d' \
  | awk -v map="${MAP_FILE}" '
      BEGIN {
        while ((getline line < map) > 0) {
          split(line, a, "\t")
          if (a[2] != "") logins[a[1]] = a[2]
        }
      }
      /^\* / {
        if (match($0, /\/commit\/[0-9a-f]{40}/)) {
          sha = substr($0, RSTART + 8, 40)
          if (sha in logins) {
            printf "%s by @%s\n", $0, logins[sha]
            next
          }
        }
      }
      { print }
    ' > "${NOTES_FILE}"

NEW_CONTRIBUTORS="$(
  gh api -X POST "repos/${REPO}/releases/generate-notes" \
    -f tag_name="${BUILDKITE_TAG}" \
    -f previous_tag_name="${PREV_TAG}" \
    --jq '.body' \
  | awk '/^## New Contributors$/{f=1; sub(/^## /, "### "); print; next}
         /^## /{f=0}
         /^\*\*Full Changelog\*\*/{f=0}
         f'
)"

{
  echo
  if [[ -n "${NEW_CONTRIBUTORS}" ]]; then
    printf '%s\n\n' "${NEW_CONTRIBUTORS}"
  fi
  echo "### Docker Container"
  echo "* \`docker pull authelia/authelia:${BUILDKITE_TAG/v/}\`"
  echo "* \`docker pull ghcr.io/authelia/authelia:${BUILDKITE_TAG/v/}\`"
} >> "${NOTES_FILE}"

echo "--- :github: Deploy artifacts for release: ${BUILDKITE_TAG}"
gh release create "${BUILDKITE_TAG}" "${assets[@]}" \
  --title "${BUILDKITE_TAG}" \
  --notes-file "${NOTES_FILE}"; EXIT=$?

if [[ "${EXIT}" == 0 ]]; then
  echo "--- :github: Sync master and tags to authelia/authelia-cve"
  # shellcheck disable=SC2016 # ${GITHUB_TOKEN} is expanded by git's helper subshell, not here.
  cveHelper='!f() { echo "username=x-access-token"; echo "password=${GITHUB_TOKEN}"; }; f'
  git remote remove cve 2>/dev/null || true
  git remote add cve https://github.com/authelia/authelia-cve.git
  git fetch origin master --tags
  git -c "credential.helper=${cveHelper}" push cve refs/remotes/origin/master:refs/heads/master --force || echo ":warning: Failed to sync master to authelia/authelia-cve"
  git -c "credential.helper=${cveHelper}" push cve --tags --force || echo ":warning: Failed to sync tags to authelia/authelia-cve"
  git remote remove cve
  exit
else
  gh release delete "${BUILDKITE_TAG}" --yes && false
fi
