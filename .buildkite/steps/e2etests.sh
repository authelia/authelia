#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

set -eu

declare -A SUITE_AGENTS=(
  [ActiveDirectory]="activedirectory"
  [HighAvailability]="highavailability"
  [Kubernetes]="kubernetes"
  [Standalone]="standalone"
)

declare -A SUITE_TIMEOUTS=(
  [Kubernetes]="30"
  [OIDCConformance]="120"
)

declare -A SUITE_NO_FAILFAST=(
  [OIDCConformance]="true"
)

DEBUG_REGEX='\[(debug test|test debug)\]'
SUITE_DEBUG="false"

if [[ "${BUILDKITE_MESSAGE:-}" =~ ${DEBUG_REGEX} ]]; then
  SUITE_DEBUG="true"
fi

for SUITE_NAME in $(authelia-scripts suites list); do
  AGENT="${SUITE_AGENTS[${SUITE_NAME}]:-all}"
  TIMEOUT="${SUITE_TIMEOUTS[${SUITE_NAME}]:-20}"
  FAILFAST="--failfast"

  if [[ "${SUITE_NO_FAILFAST[${SUITE_NAME}]:-false}" == "true" ]]; then
    FAILFAST=""
  fi
cat << EOF
  - label: ":selenium: ${SUITE_NAME} Suite"
    command: "authelia-scripts --log-level debug suites test ${SUITE_NAME} ${FAILFAST} --headless"
    artifact_paths:
      - "oidc-conformance-plans/*.zip"
      - "screenshots/**/*.access.log"
      - "screenshots/**/*.console.json"
      - "screenshots/**/*.containers.log"
      - "screenshots/**/*.html"
      - "screenshots/**/*.png"
      - "screenshots/**/*.resources.json"
      - "test-results-*.json"
      - "test-results-*.xml"
    timeout_in_minutes: ${TIMEOUT}
    retry:
      automatic:
        - exit_status: "*"
          limit: 2
      manual:
        permit_on_passed: true
    agents:
      suite: "${AGENT}"
    env:
      SUITE: "${SUITE_NAME}"
      SUITE_DEBUG: "${SUITE_DEBUG}"
EOF
done
