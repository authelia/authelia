#!/usr/bin/env bash
DIVERGED=$(git merge-base --fork-point origin/master > /dev/null; echo $?)

if [[ ${DIVERGED} == 0 ]]; then
  if [[ ${BUILDKITE_TAG} == "" ]]; then
    if [[ ${BUILDKITE_BRANCH} == "master" ]]; then
      BUILD_DUO=$(git diff --name-only HEAD~1 | grep -q ^internal/suites/example/compose/duo-api/Dockerfile && echo true || echo false)
      BUILD_HAPROXY=$(git diff --name-only HEAD~1 | grep -q ^internal/suites/example/compose/haproxy/Dockerfile && echo true || echo false)
      BUILD_SAMBA=$(git diff --name-only HEAD~1 | grep -q ^internal/suites/example/compose/samba/Dockerfile && echo true || echo false)
      CI_BYPASS=$(git diff --name-only HEAD~1 | sed -rn '/^(CODE_OF_CONDUCT\.md|CONTRIBUTING\.md|README\.md|SECURITY\.md|crowdin\.yml|\.all-contributorsrc|\.editorconfig|\.github\/.*|docs\/.*|cmd\/authelia-gen\/templates\/.*|examples\/.*)/!{q1}' && echo true || echo false)
    else
      BUILD_DUO=$(git diff --name-only `git merge-base --fork-point origin/master` | grep -q ^internal/suites/example/compose/duo-api/Dockerfile && echo true || echo false)
      BUILD_HAPROXY=$(git diff --name-only `git merge-base --fork-point origin/master` | grep -q ^internal/suites/example/compose/haproxy/Dockerfile && echo true || echo false)
      BUILD_SAMBA=$(git diff --name-only `git merge-base --fork-point origin/master` | grep -q ^internal/suites/example/compose/samba/Dockerfile && echo true || echo false)
      CI_BYPASS=$(git diff --name-only `git merge-base --fork-point origin/master` | sed -rn '/^(CODE_OF_CONDUCT\.md|CONTRIBUTING\.md|README\.md|SECURITY\.md|crowdin\.yml|\.all-contributorsrc|\.editorconfig|\.github\/.*|docs\/.*|cmd\/authelia-gen\/templates\/.*|examples\/.*)/!{q1}' && echo true || echo false)
    fi

    if [[ ${CI_BYPASS} == "true" ]]; then
      buildkite-agent annotate --style "info" --context "ctx-info" < .buildkite/annotations/bypass
    fi
  else
    BUILD_DUO="false"
    BUILD_HAPROXY="false"
    BUILD_SAMBA="false"
    CI_BYPASS="false"
  fi
else
  BUILD_DUO="false"
  BUILD_HAPROXY="false"
  BUILD_SAMBA="false"
  CI_BYPASS="false"
fi

if [[ ${BUILDKITE_PULL_REQUEST_DRAFT} == "true" ]] && [[ ${BUILDKITE_BRANCH} =~ ^(dependabot|renovate) ]]; then
  CI_BYPASS="true"
  buildkite-agent annotate --style "info" --context "ctx-info" < .buildkite/annotations/draft
fi

cat << EOF
env:
  BUILD_DUO: ${BUILD_DUO}
  BUILD_HAPROXY: ${BUILD_HAPROXY}
  BUILD_SAMBA: ${BUILD_SAMBA}
  CI_BYPASS: ${CI_BYPASS}

steps:
  - label: ":service_dog: Linting"
    command: ".buildkite/lint.sh -reporter=github-check -filter-mode=nofilter -fail-level=error"
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.message !~ /\[(skip test|test skip)\]/

  - label: ":hammer_and_wrench: Unit Test"
    command: "authelia-scripts --log-level debug ci --buildkite"
    agents:
      build: "unit-test"
    artifact_paths:
      - "*.tar.gz"
      - "*.deb"
      - "*.sha256"
      - "*.sig"
    key: "unit-test"
    env:
      NODE_OPTIONS: "--no-deprecation"
    if: build.env("CI_BYPASS") != "true"

  - wait:
    if: build.env("CI_BYPASS") != "true"
EOF
if [[ ${BUILD_DUO} == "true" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [integration-duo]"
    trigger: "integration-duo"
    build:
      message: "${BUILDKITE_MESSAGE}"
      commit: "${BUILDKITE_COMMIT}"
      branch: "${BUILDKITE_BRANCH}"
      env:
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"
    depends_on: ~
EOF
fi
if [[ ${BUILD_HAPROXY} == "true" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [integration-haproxy]"
    trigger: "integration-haproxy"
    build:
      message: "${BUILDKITE_MESSAGE}"
      commit: "${BUILDKITE_COMMIT}"
      branch: "${BUILDKITE_BRANCH}"
      env:
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"
    depends_on: ~
EOF
fi
if [[ ${BUILD_SAMBA} == "true" ]]; then
cat << EOF
  - label: ":rocket: Trigger Pipeline [integration-samba]"
    trigger: "integration-samba"
    build:
      message: "${BUILDKITE_MESSAGE}"
      commit: "${BUILDKITE_COMMIT}"
      branch: "${BUILDKITE_BRANCH}"
      env:
        BUILDKITE_PULL_REQUEST: "${BUILDKITE_PULL_REQUEST}"
        BUILDKITE_PULL_REQUEST_BASE_BRANCH: "${BUILDKITE_PULL_REQUEST_BASE_BRANCH}"
        BUILDKITE_PULL_REQUEST_REPO: "${BUILDKITE_PULL_REQUEST_REPO}"
    depends_on: ~
EOF
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
    depends_on: ~
    key: "build-docker-linux-coverage"
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true" && build.message !~ /\[(skip test|test skip)\]/

  - wait:
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true" && build.message !~ /\[(skip test|test skip)\]/

  - label: ":chrome: Integration Tests"
    command: ".buildkite/steps/e2etests.sh | buildkite-agent pipeline upload"
    depends_on:
      - "build-docker-linux-coverage"
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/ && build.env("CI_BYPASS") != "true" && build.message !~ /\[(skip test|test skip)\]/
EOF
