#!/bin/bash

export PATH=$PATH:./cmd/authelia-scripts/:./.buildkite/steps/:./node_modules/.bin:/tmp

if [[ -z "$OLD_PS1" ]]; then
  OLD_PS1="$PS1"
  export PS1="(authelia) $PS1"
fi

if [[ $(id -u) = 0 ]]; then
  echo "Cannot run as root, defaulting to UID 1000"
  export USER_ID=1000
elif [[ $(uname) == "Darwin" ]]; then
  echo "Normalise for OSX, defaulting to UID 1000"
  export USER_ID=1000
else
  export USER_ID=$(id -u)
fi

if [[ $(id -g) = 0 ]]; then
  echo "Cannot run as root, defaulting to GID 1000"
  export GROUP_ID=1000
elif [[ $(uname) == "Darwin" ]]; then
  echo "Normalise for OSX, defaulting to GID 1000"
  export GROUP_ID=1000
else
  export GROUP_ID=$(id -g)
fi

if [[ "$CI" == "true" ]]; then
  true
else
  export CI=false
fi

echo "[BOOTSTRAP] Checking if Go is installed..."
if [[ ! -x "$(command -v go)" ]];
then
  echo "[ERROR] You must install Go on your machine.";
  return
fi

authelia-scripts bootstrap