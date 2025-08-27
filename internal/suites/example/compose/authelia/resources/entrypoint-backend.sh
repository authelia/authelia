#!/bin/sh

set -x

# We move out of the workspace to not include the modules as dependencies of the project.
cd /

echo "Use hot reloaded version of Authelia backend"
go install github.com/cespare/reflex@13e5691dcde5f7c29c144d1cb8c34f453d78505d
go install github.com/go-delve/delve/cmd/dlv@f498dc8c5a8ad01334a9d782893c10bd0addb510

cd /app

reflex -c /resources/reflex.conf
