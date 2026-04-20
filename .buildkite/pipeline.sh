#!/usr/bin/env bash
DIVERGED=$(git merge-base --fork-point origin/master > /dev/null; echo $?)

BYPASS_REGEX='/^(CODE_OF_CONDUCT\.md|CONTRIBUTING\.md|README\.md|SECURITY\.md|crowdin\.yml|\.all-contributorsrc|\.editorconfig|\.github\/.*|docs\/.*|cmd\/authelia-gen\/templates\/.*|examples\/.*)/!{q1}'

changed() {
  git diff --name-only "${1}" | grep -q "^${2}"
}

bypass_check() {
  git diff --name-only "${1}" | sed -rn "${BYPASS_REGEX}" && echo true || echo false
}

BUILD_DUO="false"
BUILD_HAPROXY="false"
BUILD_SAMBA="false"
CI_BYPASS="false"
CI_MERGE_QUEUE="false"
CI_MERGE_QUEUE_BYPASS="false"
CI_PRIVATE="false"

if [[ ${DIVERGED} == 0 ]] && [[ ${BUILDKITE_TAG} == "" ]]; then
  if [[ ${BUILDKITE_BRANCH} == "master" ]]; then
    BASE_REF="HEAD~1"
  else
    BASE_REF=$(git merge-base --fork-point origin/master)
  fi

  changed "${BASE_REF}" "internal/suites/example/compose/duo-api/Dockerfile" && BUILD_DUO="true"
  changed "${BASE_REF}" "internal/suites/example/compose/haproxy/Dockerfile" && BUILD_HAPROXY="true"
  changed "${BASE_REF}" "internal/suites/example/compose/samba/Dockerfile" && BUILD_SAMBA="true"
  CI_BYPASS=$(bypass_check "${BASE_REF}")

  if [[ ${CI_BYPASS} == "true" ]]; then
    buildkite-agent annotate --style "info" --context "ctx-info" < .buildkite/annotations/bypass
  fi
fi

if [[ ${BUILDKITE_PULL_REQUEST_DRAFT} == "true" ]] && [[ ${BUILDKITE_BRANCH} =~ ^(dependabot|renovate) ]]; then
  CI_BYPASS="true"
  buildkite-agent annotate --style "info" --context "ctx-info" < .buildkite/annotations/draft
fi

if [[ ${BUILDKITE_BRANCH} =~ ^gh-readonly-queue/.* ]]; then
  CI_BYPASS="true"
  CI_MERGE_QUEUE="true"
  CI_MERGE_QUEUE_BYPASS=$(bypass_check "HEAD^..HEAD")
  buildkite-agent annotate --style "info" --context "ctx-info" < .buildkite/annotations/merge-queue
fi

if [[ ${BUILDKITE_PIPELINE_SLUG} == "authelia-cve" ]]; then
  CI_PRIVATE="true"
fi

cat << EOF
env:
  BUILD_DUO: ${BUILD_DUO}
  BUILD_HAPROXY: ${BUILD_HAPROXY}
  BUILD_SAMBA: ${BUILD_SAMBA}
  CI_BYPASS: ${CI_BYPASS}
  CI_MERGE_QUEUE: ${CI_MERGE_QUEUE}
  CI_MERGE_QUEUE_BYPASS: ${CI_MERGE_QUEUE_BYPASS}
  CI_PRIVATE: ${CI_PRIVATE}

steps:
  - label: ":service_dog: Linting"
    command: "lint.sh -reporter=github-check -filter-mode=nofilter -fail-level=error"
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.message !~ /\[(skip test|test skip)\]/

  - label: ":chrome: External Tests"
    command: "e2epackages.sh | buildkite-agent pipeline upload"
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.message !~ /\[(skip test|test skip)\]/ && build.env("CI_MERGE_QUEUE") != "true"

  - label: ":hammer_and_wrench: Unit Test"
    command: "authelia-scripts --log-level debug ci --buildkite"
    agents:
      build: "unit-test"
    artifact_paths:
      - "*.tar.gz"
      - "*.deb"
      - "*.sha256"
      - "*.sig"
      - "*.{c,sp}dx.json"
    key: "unit-test"
    env:
      NODE_OPTIONS: "--no-deprecation"
    if: build.env("CI_BYPASS") != "true"

  - label: ":grype: Vulnerability Scanning"
    command: "grypescans.sh"
EOF
if [[ ${CI_MERGE_QUEUE} != "true" ]]; then
cat << EOF
    depends_on:
      - "unit-test"
      - "build-docker-linux"
    if: build.env("CI_BYPASS") != "true" && build.message !~ /^docs/
EOF
else
cat << EOF
    if: build.env("CI_MERGE_QUEUE_BYPASS") != "true"
EOF
fi
if [[ ${BUILDKITE_TAG} != "" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [baseimage]"
    trigger: "baseimage"
    build:
      message: "${BUILDKITE_MESSAGE%%$'\n'*}"
      env:
        AUTHELIA_RELEASE: "${BUILDKITE_TAG//v}"
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"
    key: "baseimage"
    if: build.tag != null && build.env("CI_BYPASS") != "true"

EOF
fi
if [[ ${CI_BYPASS} != "true" ]]; then
if [[ ${BUILD_DUO} == "true" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [integration-duo]"
    trigger: "integration-duo"
    build:
      message: "${BUILDKITE_MESSAGE%%$'\n'*}"
      commit: "${BUILDKITE_COMMIT}"
      branch: "${BUILDKITE_BRANCH}"
      env:
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"

EOF
fi
if [[ ${BUILD_HAPROXY} == "true" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [integration-haproxy]"
    trigger: "integration-haproxy"
    build:
      message: "${BUILDKITE_MESSAGE%%$'\n'*}"
      commit: "${BUILDKITE_COMMIT}"
      branch: "${BUILDKITE_BRANCH}"
      env:
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"

EOF
fi
if [[ ${BUILD_SAMBA} == "true" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [integration-samba]"
    trigger: "integration-samba"
    build:
      message: "${BUILDKITE_MESSAGE%%$'\n'*}"
      commit: "${BUILDKITE_COMMIT}"
      branch: "${BUILDKITE_BRANCH}"
      env:
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"

EOF
fi
fi
cat << EOF
  - label: ":docker: Build Image [coverage]"
    command: "authelia-scripts docker build --container=coverage"
    retry:
      manual:
        permit_on_passed: true
    agents:
      build: "linux-coverage"
    artifact_paths:
      - "authelia-image-coverage.tar.zst"
    key: "build-docker-linux-coverage"
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true" && build.message !~ /\[(skip test|test skip)\]/

  - label: ":chrome: Integration Tests"
    command: "e2etests.sh | buildkite-agent pipeline upload"
    depends_on:
      - "build-docker-linux-coverage"
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true" && build.message !~ /\[(skip test|test skip)\]/

EOF
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
    key: "build-docker-linux"
EOF
if [[ ${BUILDKITE_BRANCH} == "master" ]]; then
cat << EOF
    concurrency: 1
    concurrency_group: "deployments"
EOF
fi
cat << EOF
    if: build.env("CI_BYPASS") != "true" && build.message !~ /^docs/

  - label: ":github: Deploy Artifacts"
    command: "ghartifacts.sh"
    depends_on:
      - "unit-test"
    retry:
      automatic: true
    agents:
      upload: "fast"
    key: "artifacts"
    if: build.tag != null && build.env("CI_BYPASS") != "true" && build.env("CI_PRIVATE") != "true"

  - label: ":linux: Deploy AUR"
    command: "aurpackages.sh | buildkite-agent pipeline upload"
    if: build.tag != null && build.env("CI_BYPASS") != "true" && build.env("CI_PRIVATE") != "true"

  - label: ":debian: :fedora: :ubuntu: Deploy APT"
    command: "aptdeploy.sh"
    depends_on:
      - "unit-test"
    agents:
      upload: "fast"
    if: build.tag != null && build.env("CI_BYPASS") != "true" && build.env("CI_PRIVATE") != "true"
EOF
