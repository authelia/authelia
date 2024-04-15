#!/bin/sh

set -x

# We move out of the workspace to not include the modules as dependencies of the project.
cd /

echo "Use hot reloaded version of Authelia backend"
go install github.com/cespare/reflex@latest
go install github.com/go-delve/delve/cmd/dlv@latest

cd /app

reflex -c /resources/reflex.conf
