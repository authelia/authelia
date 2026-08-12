#!/usr/bin/env bash
set -eu

for SUITE_NAME in $(authelia-scripts suites external list); do
cat << EOF
  - label: ":nodejs: ${SUITE_NAME} Suite"
    command: "authelia-scripts --log-level debug suites external test ${SUITE_NAME} --failfast --headless"
    artifact_paths:
      - "test-results-*.json"
      - "test-results-*.xml"
      - "screenshots/**/*.png"
      - "internal/suites/testdata/*.actual.png"
    timeout_in_minutes: 20
    retry:
      automatic:
        - exit_status: "*"
          limit: 2
      manual:
        permit_on_passed: true
    env:
      SUITE: "${SUITE_NAME}"
EOF
done
