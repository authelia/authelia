#!/bin/sh

set -x

cd /resources

echo "Use hot reloaded version of Authelia backend"
go install github.com/cespare/reflex
go install github.com/go-delve/delve/cmd/dlv

cd /app

reflex -c /resources/reflex.conf
