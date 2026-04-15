#!/usr/bin/env bash

export PATH=${PATH}:${PWD}/cmd/dev/:${PWD}/.buildkite/steps/:${GOPATH}/bin:${PWD}/web/node_modules/.bin:/tmp \
DOCKER_BUILDKIT=1

if [[ -z "${OLD_PS1}" ]]; then
  OLD_PS1="${PS1}"
  export PS1="(authelia) ${PS1}"
fi

if [[ $(id -u) = 0 ]]; then
  echo "Cannot run as root, defaulting to UID 1000"
  export USER_ID=1000
else
  USER_ID=$(id -u)
  export USER_ID
fi

if [[ $(id -g) = 0 ]]; then
  echo "Cannot run as root, defaulting to GID 1000"
  export GROUP_ID=1000
else
  GROUP_ID=$(id -g)
  export GROUP_ID
fi

if [[ "${CI}" != "true" ]]; then
  export CI=false
fi

echo "[BOOTSTRAP] Checking if Go is installed..."
if [[ ! -x "$(command -v go)" ]];
then
  echo "[ERROR] You must install Go on your machine." >&2
  return
fi

authelia-scripts bootstrap
