#!/usr/bin/env bash

resolve_tag() {
  if [[ "${BUILDKITE_BRANCH}" =~ ^renovate- ]]; then
    echo "renovate"
  elif [[ "${BUILDKITE_BRANCH}" != "master" ]] && [[ ! "${BUILDKITE_BRANCH}" =~ .*:.* ]]; then
    echo "${BUILDKITE_BRANCH}"
  elif [[ "${BUILDKITE_BRANCH}" != "master" ]] && [[ "${BUILDKITE_BRANCH}" =~ .*:.* ]]; then
    echo "PR${BUILDKITE_PULL_REQUEST}"
  elif [[ "${BUILDKITE_BRANCH}" == "master" ]] && [[ "${BUILDKITE_PULL_REQUEST}" == "false" ]]; then
    echo "latest"
  fi
}
