#!/bin/bash
set -u

if [[ $BUILDKITE_BRANCH == "master" ]]; then
  CI_DOCS_BYPASS=$(git diff --name-only HEAD~1 | sed -rn '/^docs\/.*/!{q1}' && echo true || echo false)
else
  CI_DOCS_BYPASS=$(git diff --name-only `git merge-base --fork-point origin/master` | sed -rn '/^docs\/.*/!{q1}' && echo true || echo false)
fi

cat << EOF
env:
  CI_DOCS_BYPASS: ${CI_DOCS_BYPASS}

steps:
  - label: ":hammer_and_wrench: Unit Test"
    command: "authelia-scripts --log-level debug ci"
    if: build.env("CI_DOCS_BYPASS") != "true"

  - wait:
    if: build.env("CI_DOCS_BYPASS") != "true"

  - label: ":docker: Image Builds"
    command: ".buildkite/steps/buildimages.sh | buildkite-agent pipeline upload"
    depends_on: ~
    if: build.env("CI_DOCS_BYPASS") != "true"

  - wait:
    if: build.env("CI_DOCS_BYPASS") != "true"

  - label: ":chrome: Integration Tests"
    command: ".buildkite/steps/e2etests.sh | buildkite-agent pipeline upload"
    depends_on:
      - "build-docker-amd64"
    if: build.env("CI_DOCS_BYPASS") != "true"
EOF