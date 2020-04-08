#!/bin/bash
set -u

DIVERGED=$(git merge-base --fork-point origin/master > /dev/null; echo $?)

if [[ $DIVERGED -eq 0 ]]; then
  if [[ $BUILDKITE_TAG == "" ]]; then
    if [[ $BUILDKITE_BRANCH == "master" ]]; then
      CI_BYPASS=$(git diff --name-only HEAD~1 | sed -rn '/^(BREAKING.md|CONTRIBUTING.md|README.md|docs\/.*)/!{q1}' && echo true || echo false)
    else
      CI_BYPASS=$(git diff --name-only `git merge-base --fork-point origin/master` | sed -rn '/^(BREAKING.md|CONTRIBUTING.md|README.md|docs\/.*)/!{q1}' && echo true || echo false)
    fi

    if [[ $CI_BYPASS == "true" ]]; then
      cat .buildkite/annotations/bypass | buildkite-agent annotate --style "info" --context "ctx-info"
    fi
  else
    CI_BYPASS="false"
  fi
else
  CI_BYPASS="false"
fi

cat << EOF
env:
  CI_BYPASS: ${CI_BYPASS}

steps:
  - label: ":hammer_and_wrench: Unit Test"
    command: "authelia-scripts --log-level debug ci"
    if: build.branch !~ /^(master)|(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true"

  - wait:
    if: build.env("CI_BYPASS") != "true"

  - label: ":docker: Image Builds"
    command: ".buildkite/steps/buildimages.sh | buildkite-agent pipeline upload"
    depends_on: ~
    if: build.env("CI_BYPASS") != "true"

  - wait:
    if: build.branch !~ /^(master)|(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true"

  - label: ":chrome: Integration Tests"
    command: ".buildkite/steps/e2etests.sh | buildkite-agent pipeline upload"
    depends_on:
      - "build-docker-linux-amd64"
    if: build.branch !~ /^(master)|(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true"
EOF