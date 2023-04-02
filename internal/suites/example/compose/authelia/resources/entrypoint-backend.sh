#!/bin/sh

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

set -x

# We move out of the workspace to not include the modules as dependencies of the project.
cd /

echo "Use hot reloaded version of Authelia backend"
go install github.com/cespare/reflex@latest
go install github.com/go-delve/delve/cmd/dlv@latest

cd /app

# Sleep 10 seconds to wait the end of npm install updating web directory
# and making reflex reload multiple times.
sleep 10

reflex -c /resources/reflex.conf