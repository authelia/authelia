#!/bin/sh

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

set -e

while true;
do
    AUTHELIA_SERVER_DISABLE_HEALTHCHECK=true CGO_ENABLED=1 dlv --listen 0.0.0.0:2345 --headless=true --output=./authelia --continue --accept-multiclient debug cmd/authelia/*.go
    sleep 10
done
