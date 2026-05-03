#!/bin/sh

set -x

cd /resources || exit

echo "Installing pinned CLI tools from go.mod"
go run .

cd /app || exit

echo "Use hot reloaded version of Authelia backend"
reflex -c /resources/reflex.conf
