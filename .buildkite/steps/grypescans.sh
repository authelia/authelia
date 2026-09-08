#!/usr/bin/env bash
set -e

# shellcheck source=/dev/null
source "$(dirname "${BASH_SOURCE[0]}")/../libs/common.sh"

ciTag="${BUILDKITE_TAG}"
dockerImageName="authelia/authelia"
masterBranch="master"
grypeCmd=(grype -f low --only-fixed)

if [[ "${CI_PRIVATE}" == "true" ]]; then
  dockerImageName="${dockerImageName}-cve"
fi

IMAGE=""
if [[ "${CI_MERGE_QUEUE}" != "true" ]]; then
  if [[ -n "${ciTag}" ]]; then
    IMAGE="${dockerImageName}:${ciTag/v}"
  else
    resolve_tag_suffix "${masterBranch}" "false"
    [[ -n "${TAG_SUFFIX}" ]] && IMAGE="${dockerImageName}:${TAG_SUFFIX}"
  fi
fi

STATUS=0

if [[ -n "${IMAGE}" ]]; then
  echo "--- :grype: Scanning ${IMAGE}"
  "${grypeCmd[@]}" "registry:${IMAGE}" || STATUS=1
fi

for file in *.cdx.json; do
  echo "--- :grype: Scanning ${file/.cdx.json}"
  "${grypeCmd[@]}" "${file}" || STATUS=1
done

exit "${STATUS}"
