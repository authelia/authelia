#!/bin/sh

set -e

while true;
do
    AUTHELIA_SERVER_DISABLE_HEALTHCHECK=true CGO_ENABLED=1 dlv --listen 0.0.0.0:2345 --headless=true --output=./authelia --continue --accept-multiclient debug cmd/authelia/*.go
    sleep 3
done
