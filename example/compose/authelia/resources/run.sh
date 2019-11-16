#!/bin/bash

set -e

# Build the binary
go build -o /tmp/authelia/authelia-tmp cmd/authelia/main.go

# Run the temporary binary
cd $SUITE_PATH
/tmp/authelia/authelia-tmp -config ${SUITE_PATH}/configuration.yml