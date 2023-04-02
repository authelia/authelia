#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

if [[ $(uname) == "Darwin" ]]; then
  echo "Authelia's development workflow currently isn't supported on macOS"
  exit
fi

export PATH=$PATH:./cmd/authelia-scripts/:./.buildkite/steps/:$GOPATH/bin:./web/node_modules/.bin:/tmp \
DOCKER_BUILDKIT=1

if [[ -z "$OLD_PS1" ]]; then
  OLD_PS1="$PS1"
  export PS1="(authelia) $PS1"
fi

if [[ $(id -u) = 0 ]]; then
  echo "Cannot run as root, defaulting to UID 1000"
  export USER_ID=1000
else
  export USER_ID=$(id -u)
fi

if [[ $(id -g) = 0 ]]; then
  echo "Cannot run as root, defaulting to GID 1000"
  export GROUP_ID=1000
else
  export GROUP_ID=$(id -g)
fi

if [[ "$CI" != "true" ]]; then
  export CI=false
fi

echo "[BOOTSTRAP] Checking if Go is installed..."
if [[ ! -x "$(command -v go)" ]];
then
  echo "[ERROR] You must install Go on your machine.";
  return
fi

authelia-scripts bootstrap