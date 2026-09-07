#!/usr/bin/env bash
git fetch --quiet origin master
DIVERGED=$(git merge-base origin/master HEAD > /dev/null; echo $?)

BYPASS_REGEX='/^(CODE_OF_CONDUCT\.md|CONTRIBUTING\.md|README\.md|SECURITY\.md|crowdin\.yml|\.all-contributorsrc|\.editorconfig|\.github\/.*|docs\/.*|cmd\/authelia-gen\/templates\/.*|examples\/.*)/!{q1}'

changed() {
  git diff --name-only "${1}" | grep -q "^${2}"
}

bypass_check() {
  git diff --name-only "${1}" | sed -rn "${BYPASS_REGEX}" && echo true || echo false
}

# shellcheck source=.buildkite/lib/resolve_tag.sh disable=SC1091 # external-sources is disabled repo-wide
source "$(dirname "${BASH_SOURCE[0]}")/lib/resolve_tag.sh"

extract_tag_override() {
  local name="${1}"

  [[ "${BUILDKITE_MESSAGE}" =~ \[${name}\ tag:([^]]+)\] ]] && echo "${BASH_REMATCH[1]}"
}

force_build() {
  local name="${1}"

  [[ "${BUILDKITE_MESSAGE}" =~ \[${name}\ build\] ]] && echo "true" || echo "false"
}

image_needs_build() {
  local dir="${1}" image="${2}" tag last_commit published_revision

  tag=$(resolve_tag)
  [[ -z "${tag}" ]] && return 0

  last_commit=$(git log -1 --format=%H -- "${dir}")
  [[ -z "${last_commit}" ]] && return 0

  published_revision=$(docker buildx imagetools inspect "${image}:${tag}" --format '{{ index (index .Image "linux/amd64").Config.Labels "org.opencontainers.image.revision" }}' 2> /dev/null)

  [[ -n "${published_revision}" ]] && [[ "${published_revision}" == "${last_commit}" ]] && return 1

  return 0
}

image_exists() {
  local image="${1}" tag="${2}"

  [[ -z "${tag}" ]] && return 1

  docker buildx imagetools inspect "${image}:${tag}" > /dev/null 2>&1
}

BUILD_DUO="false"
BUILD_HAPROXY="false"
BUILD_SAMBA="false"
USE_DUO="false"
USE_HAPROXY="false"
USE_SAMBA="false"
DUO_TAG_OVERRIDE=""
HAPROXY_TAG_OVERRIDE=""
SAMBA_TAG_OVERRIDE=""
CI_BYPASS="false"
CI_MERGE_QUEUE="false"
CI_MERGE_QUEUE_BYPASS="false"
CI_PRIVATE="false"
LINT_REPORTER="github-check"

if [[ ${DIVERGED} == 0 ]] && [[ ${BUILDKITE_TAG} == "" ]]; then
  if [[ ${BUILDKITE_BRANCH} == "master" ]]; then
    BASE_REF="HEAD~1"
  else
    BASE_REF=$(git merge-base origin/master HEAD)
  fi

  changed "${BASE_REF}" "internal/suites/example/compose/duo-api/" && USE_DUO="true"
  changed "${BASE_REF}" "internal/suites/example/compose/haproxy/" && USE_HAPROXY="true"
  changed "${BASE_REF}" "internal/suites/example/compose/samba/" && USE_SAMBA="true"

  DUO_TAG_OVERRIDE=$(extract_tag_override "duo")
  HAPROXY_TAG_OVERRIDE=$(extract_tag_override "haproxy")
  SAMBA_TAG_OVERRIDE=$(extract_tag_override "samba")

  FORCE_BUILD_DUO=$(force_build "duo")
  FORCE_BUILD_HAPROXY=$(force_build "haproxy")
  FORCE_BUILD_SAMBA=$(force_build "samba")

  if [[ -n "${DUO_TAG_OVERRIDE}" ]]; then
    USE_DUO="true"
    { [[ ${FORCE_BUILD_DUO} == "true" ]] || ! image_exists "authelia/integration-duo" "${DUO_TAG_OVERRIDE}"; } && BUILD_DUO="true"
  elif [[ ${FORCE_BUILD_DUO} == "true" ]]; then
    USE_DUO="true"
    BUILD_DUO="true"
  elif [[ ${USE_DUO} == "true" ]] && image_needs_build "internal/suites/example/compose/duo-api/" "authelia/integration-duo"; then
    BUILD_DUO="true"
  fi

  if [[ -n "${HAPROXY_TAG_OVERRIDE}" ]]; then
    USE_HAPROXY="true"
    { [[ ${FORCE_BUILD_HAPROXY} == "true" ]] || ! image_exists "authelia/integration-haproxy" "${HAPROXY_TAG_OVERRIDE}"; } && BUILD_HAPROXY="true"
  elif [[ ${FORCE_BUILD_HAPROXY} == "true" ]]; then
    USE_HAPROXY="true"
    BUILD_HAPROXY="true"
  elif [[ ${USE_HAPROXY} == "true" ]] && image_needs_build "internal/suites/example/compose/haproxy/" "authelia/integration-haproxy"; then
    BUILD_HAPROXY="true"
  fi

  if [[ -n "${SAMBA_TAG_OVERRIDE}" ]]; then
    USE_SAMBA="true"
    { [[ ${FORCE_BUILD_SAMBA} == "true" ]] || ! image_exists "authelia/integration-samba" "${SAMBA_TAG_OVERRIDE}"; } && BUILD_SAMBA="true"
  elif [[ ${FORCE_BUILD_SAMBA} == "true" ]]; then
    USE_SAMBA="true"
    BUILD_SAMBA="true"
  elif [[ ${USE_SAMBA} == "true" ]] && image_needs_build "internal/suites/example/compose/samba/" "authelia/integration-samba"; then
    BUILD_SAMBA="true"
  fi

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
  LINT_REPORTER="local"
fi

cat << EOF
env:
  BUILD_DUO: ${BUILD_DUO}
  BUILD_HAPROXY: ${BUILD_HAPROXY}
  BUILD_SAMBA: ${BUILD_SAMBA}
  USE_DUO: ${USE_DUO}
  USE_HAPROXY: ${USE_HAPROXY}
  USE_SAMBA: ${USE_SAMBA}
  DUO_TAG_OVERRIDE: ${DUO_TAG_OVERRIDE}
  HAPROXY_TAG_OVERRIDE: ${HAPROXY_TAG_OVERRIDE}
  SAMBA_TAG_OVERRIDE: ${SAMBA_TAG_OVERRIDE}
  CI_BYPASS: ${CI_BYPASS}
  CI_MERGE_QUEUE: ${CI_MERGE_QUEUE}
  CI_MERGE_QUEUE_BYPASS: ${CI_MERGE_QUEUE_BYPASS}
  CI_PRIVATE: ${CI_PRIVATE}

steps:
  - label: ":service_dog: Linting"
    command: "lint.sh -reporter=${LINT_REPORTER} -filter-mode=nofilter -fail-level=error"
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
        TAG_OVERRIDE: "${DUO_TAG_OVERRIDE}"

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
        TAG_OVERRIDE: "${HAPROXY_TAG_OVERRIDE}"

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
        TAG_OVERRIDE: "${SAMBA_TAG_OVERRIDE}"

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
