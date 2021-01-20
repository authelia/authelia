#!/bin/sh

set -e

while true;
do
    dlv --listen 0.0.0.0:2345 --headless=true --output=./authelia --continue --accept-multiclient debug cmd/authelia/*.go -- --config /config/configuration.yml
    sleep 10
done