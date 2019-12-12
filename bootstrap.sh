#!/bin/bash

export PATH=$PATH:./cmd/authelia-scripts/:./node_modules/.bin:/tmp

if [ -z "$OLD_PS1" ]; then
  OLD_PS1="$PS1"
  export PS1="(authelia) $PS1"
fi

if [ $(id -u) = 0 ]; then
  export USER_ID=1000
else
  export USER_ID=$(id -u)
fi
if [ $(id -g) = 0 ]; then
  export GROUP_ID=1000
else
  export GROUP_ID=$(id -g)
fi
if [ $CI == "true" ]; then
  echo "Running in CI don't overwrite variable"
else
  export CI=false
fi

echo "[BOOTSTRAP] Checking if Go is installed..."
if [ ! -x "$(command -v go)" ];
then
  echo "[ERROR] You must install Go on your machine.";
  return
fi

authelia-scripts bootstrap
