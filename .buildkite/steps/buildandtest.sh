#!/bin/bash
set -eo pipefail

echo "--- :gear: Installing pre-requisites"

apt update -y
apt install build-essential -y -q

if [ ! -d "/usr/local/go" ]; then
  cd /tmp
  curl -o go1.13.5.linux-amd64.tar.gz https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz
  tar -xf go1.13.5.linux-amd64.tar.gz
  mv go /usr/local
fi

if [ ! -d "$HOME/.nvm" ]; then
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.35.1/install.sh | bash
  [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
  nvm install v12 && nvm use v12
fi

cd $BUILDKITE_BUILD_CHECKOUT_PATH
nvm use v12
go mod download
source bootstrap.sh

echo "--- :hammer: Build and Test ${BUILDKITE_PIPELINE_SLUG}"
authelia-scripts --log-level debug ci
