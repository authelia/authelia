#!/usr/bin/env bash
set -u

DIVERGED=$(git merge-base --fork-point origin/master > /dev/null; echo $?)

if [[ "${DIVERGED}" == 0 ]]; then
  if [[ "${BUILDKITE_TAG}" == "" ]]; then
    if [[ "${BUILDKITE_BRANCH}" == "master" ]]; then
      CI_BYPASS=$(git diff --name-only HEAD~1 | sed -rn '/^(CODE_OF_CONDUCT\.md|CONTRIBUTING\.md|README\.md|SECURITY\.md|crowdin\.yml|\.all-contributorsrc|\.editorconfig|\.github\/.*|docs\/.*|cmd\/authelia-gen\/templates\/.*|examples\/.*)/!{q1}' && echo true || echo false)
    else
      CI_BYPASS=$(git diff --name-only `git merge-base --fork-point origin/master` | sed -rn '/^(CODE_OF_CONDUCT\.md|CONTRIBUTING\.md|README\.md|SECURITY\.md|crowdin\.yml|\.all-contributorsrc|\.editorconfig|\.github\/.*|docs\/.*|cmd\/authelia-gen\/templates\/.*|examples\/.*)/!{q1}' && echo true || echo false)
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
EOF
if [[ "${BUILDKITE_TAG}" != "" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [baseimage]"
    trigger: "baseimage"
    build:
      message: "${BUILDKITE_MESSAGE}"
      env:
        AUTHELIA_RELEASE: "${BUILDKITE_TAG//v}"
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"
    depends_on: ~
    key: "baseimage"
    if: build.tag != null && build.env("CI_BYPASS") != "true"

EOF
fi
cat << EOF
  - label: ":docker: Deploy Manifest"
    command: "authelia-scripts docker push-manifest"
    depends_on:
      - "unit-test"
EOF
if [[ "${BUILDKITE_TAG}" != "" ]]; then
cat << EOF
      - "baseimage"
EOF
fi
cat << EOF
    retry:
      manual:
        permit_on_passed: true
    agents:
      upload: "fast"
    if: build.env("CI_BYPASS") != "true"

  - label: ":github: Deploy Artifacts"
    command: "ghartifacts.sh"
    depends_on:
      - "unit-test"
    retry:
      automatic: true
    agents:
      upload: "fast"
    key: "artifacts"
    if: build.tag != null && build.env("CI_BYPASS") != "true"

  - label: ":linux: Deploy AUR"
    command: ".buildkite/steps/aurpackages.sh | buildkite-agent pipeline upload"
    depends_on: ~
    if: build.tag != null && build.env("CI_BYPASS") != "true"

  - label: ":debian: :fedora: :ubuntu: Deploy APT"
    command: "aptdeploy.sh"
    depends_on:
      - "unit-test"
    agents:
      upload: "fast"
    if: build.tag != null && build.env("CI_BYPASS") != "true"
EOF
