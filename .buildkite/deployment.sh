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
  - label: ":docker: Deploy Manifest"
    command: "authelia-scripts docker push-manifest"
    depends_on:
      - "unit-test"
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
      - "build-deb-package-amd64"
      - "build-deb-package-armhf"
      - "build-deb-package-arm64"
    retry:
      automatic: true
    agents:
      upload: "fast"
    key: "artifacts"
    if: build.tag != null

  - label: ":linux: Deploy AUR"
    command: ".buildkite/steps/aurpackages.sh | buildkite-agent pipeline upload"
    depends_on: ~
    if: build.tag != null && build.env("CI_BYPASS") != "true"

  - label: ":debian: :fedora: :ubuntu: Deploy APT"
    command: "aptdeploy.sh"
    depends_on:
      - "build-deb-package-amd64"
      - "build-deb-package-arm64"
      - "build-deb-package-armhf"
    agents:
      upload: "fast"
    if: build.tag != null
EOF
