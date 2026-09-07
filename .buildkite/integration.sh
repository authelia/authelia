#!/usr/bin/env bash
set -u

DIRECTORY="unset"
GROUP="unset"
PREFIX="authelia/"
TAG="unset"
CREATED=$(date -u +%Y-%m-%dT%H:%M:%SZ)
SOURCE="https://github.com/authelia/authelia"

if [[ "${BUILDKITE_BRANCH}" =~ ^renovate- ]]; then
  TAG="renovate"
elif [[ "${BUILDKITE_BRANCH}" != "master" ]] && [[ ! "${BUILDKITE_BRANCH}" =~ .*:.* ]]; then
  TAG="${BUILDKITE_BRANCH}"
elif [[ "${BUILDKITE_BRANCH}" != "master" ]] && [[ "${BUILDKITE_BRANCH}" =~ .*:.* ]]; then
  TAG="PR${BUILDKITE_PULL_REQUEST}"
elif [[ "${BUILDKITE_BRANCH}" == "master" ]] && [[ "${BUILDKITE_PULL_REQUEST}" == "false" ]]; then
  TAG="latest"
fi

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

cat << EOF
steps:
  - label: ":docker: Build and Deploy"
    commands:
      - "cd ${DIRECTORY}"
      - "docker build \
          --tag ${PREFIX}${BUILDKITE_PIPELINE_NAME}:${TAG} \
          --label org.opencontainers.image.revision=${BUILDKITE_COMMIT} \
          --label org.opencontainers.image.created=${CREATED} \
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
