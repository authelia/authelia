#!/usr/bin/env bash
set -u

# shellcheck source=/dev/null
source "$(dirname "${BASH_SOURCE[0]}")/libs/common.sh"

DIRECTORY="unset"
GROUP="unset"
PREFIX="authelia/"
CREATED=$(date -u +%Y-%m-%dT%H:%M:%SZ)
SOURCE="https://github.com/authelia/authelia"

resolve_tag_suffix
TAG="${TAG_SUFFIX:-unset}"

if [[ "${BUILDKITE_PIPELINE_NAME}" == "integration-duo" ]]; then
  DIRECTORY="internal/suites/example/compose/duo-api"
  GROUP="duo-deployments"
elif [[ "${BUILDKITE_PIPELINE_NAME}" == "integration-haproxy" ]]; then
  DIRECTORY="internal/suites/example/compose/haproxy"
  GROUP="haproxy-deployments"
elif [[ "${BUILDKITE_PIPELINE_NAME}" == "integration-samba" ]]; then
  DIRECTORY="internal/suites/example/compose/samba"
  GROUP="samba-deployments"
fi

REVISION=$(git log -1 --format=%H -- "${DIRECTORY}")

cat << EOF
steps:
  - label: ":docker: Build and Deploy"
    commands:
      - "cd ${DIRECTORY}"
      - "docker build \
        --tag ${PREFIX}${BUILDKITE_PIPELINE_NAME}:${TAG} \
        --label org.opencontainers.image.created=${CREATED} \
        --label org.opencontainers.image.revision=${REVISION} \
        --label org.opencontainers.image.source=${SOURCE} \
        --label com.authelia.ci.build-url=${BUILDKITE_BUILD_URL} \
        --platform linux/amd64,linux/arm64 \
        --provenance mode=max,reproducible=true \
        --sbom true \
        --builder buildx \
        --pull --push ."
    concurrency: 1
    concurrency_group: "${GROUP}"
    agents:
      upload: "fast"
EOF
