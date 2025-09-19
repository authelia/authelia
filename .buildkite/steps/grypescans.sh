#!/usr/bin/env bash
set -e

ciBranch="${BUILDKITE_BRANCH}"
ciPullRequest="${BUILDKITE_PULL_REQUEST}"
ciTag="${BUILDKITE_TAG}"
dockerImageName="authelia/authelia"
masterBranch="master"
publicRepoRegex='.*:.*'
grypeCmd="grype -f low"

if [[ -n "${ciTag}" ]]; then
  echo "--- :grype: Scanning ${dockerImageName}:${ciTag/v}"
  ${grypeCmd} ${dockerImageName}:${ciTag/v}
elif [[ "${ciBranch}" != "${masterBranch}" && ! "${ciBranch}" =~ ${publicRepoRegex} ]]; then
  echo "--- :grype: Scanning ${dockerImageName}:${ciBranch}"
  ${grypeCmd} ${dockerImageName}:${ciBranch}
elif [[ "${ciBranch}" != "${masterBranch}" && "${ciBranch}" =~ ${publicRepoRegex} ]]; then
  echo "--- :grype: Scanning ${dockerImageName}:PR${ciPullRequest}"
  ${grypeCmd} ${dockerImageName}:PR${ciPullRequest}
elif [[ "${ciBranch}" == "${masterBranch}" && "${ciPullRequest}" == "false" ]]; then
  echo "--- :grype: Scanning ${dockerImageName}:${masterBranch}"
  ${grypeCmd} ${dockerImageName}:${masterBranch}
fi

for file in *.spdx.json; do
  echo "--- :grype: Scanning ${file/.spdx.json}"
  ${grypeCmd} ${file}
done
