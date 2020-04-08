#!/bin/sh

set -e

# Build the binary
go build -o /tmp/authelia/authelia-tmp cmd/authelia/*.go
while true;
do
    /tmp/authelia/authelia-tmp --config /etc/authelia/configuration.yml
    sleep 10
done