#!/bin/sh

set -x

cd /resources

echo "Installing pinned CLI tools from go.mod"
go run .

cd /app

echo "Use hot reloaded version of Authelia backend"
reflex -c /resources/reflex.conf
