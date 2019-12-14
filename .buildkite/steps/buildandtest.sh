#!/bin/bash
set -eo pipefail

echo "--- :hammer: Build and Test ${BUILDKITE_PIPELINE_SLUG}"
authelia-scripts --log-level debug ci