#!/bin/sh

set -x

echo "Use hot reloaded version of Authelia backend"
go get github.com/cespare/reflex

# Fake index.html because Authelia reads it as a template at startup to inject nonces.
# This prevents a crash of Authelia in dev mode.
mkdir -p /tmp/authelia-web
touch /tmp/authelia-web/index.html

# Sleep 10 seconds to wait the end of npm install updating web directory
# and making reflex reload multiple times.
sleep 10

reflex -c /resources/reflex.conf