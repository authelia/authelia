#!/usr/bin/env bash
set -e

ciBranch="${BUILDKITE_BRANCH}"
ciPullRequest="${BUILDKITE_PULL_REQUEST}"
ciTag="${BUILDKITE_TAG}"
dockerImageName="authelia/authelia"
masterBranch="master"
publicRepoRegex='.*:.*'
grypeCmd=(grype -f low)

IMAGE=""
if [[ "${CI_MERGE_QUEUE}" != "true" ]]; then
  if [[ -n "${ciTag}" ]]; then
    IMAGE="${dockerImageName}:${ciTag/v}"
  elif [[ "${ciBranch}" != "${masterBranch}" && ! "${ciBranch}" =~ ${publicRepoRegex} ]]; then
    IMAGE="${dockerImageName}:${ciBranch}"
  elif [[ "${ciBranch}" != "${masterBranch}" && "${ciBranch}" =~ ${publicRepoRegex} ]]; then
    IMAGE="${dockerImageName}:PR${ciPullRequest}"
  elif [[ "${ciBranch}" == "${masterBranch}" && "${ciPullRequest}" == "false" ]]; then
    IMAGE="${dockerImageName}:${masterBranch}"
  fi
fi

if [[ -n "${IMAGE}" ]]; then
  echo "--- :grype: Scanning ${IMAGE}"
  "${grypeCmd[@]}" "${IMAGE}"
fi

for file in *.spdx.json; do
  echo "--- :grype: Scanning ${file/.spdx.json}"
  "${grypeCmd[@]}" "${file}"
done
