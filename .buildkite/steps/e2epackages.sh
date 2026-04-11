#!/usr/bin/env bash
set -eu

for SUITE_NAME in $(authelia-scripts suites external list); do
cat << EOF
  - label: ":nodejs: ${SUITE_NAME} External Suite"
    command: "authelia-scripts --log-level debug suites external test ${SUITE_NAME} --failfast --headless"
    retry:
      automatic: true
      manual:
        permit_on_passed: true
    env:
      SUITE: "${SUITE_NAME}"
EOF
done
