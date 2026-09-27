#!/bin/sh

# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

set -e

while true;
do
    AUTHELIA_SERVER_DISABLE_HEALTHCHECK=true CGO_ENABLED=1 dlv debug --headless --listen=:2345 --api-version=2 --output=./authelia --continue --accept-multiclient --build-flags="-tags=dev" cmd/authelia/*.go
    sleep 3
done
