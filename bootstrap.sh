#!/bin/bash

export PATH=$PATH:./cmd/authelia-scripts/:./node_modules/.bin:/tmp

if [ -z "$OLD_PS1" ]; then
  OLD_PS1="$PS1"
  export PS1="(authelia) $PS1"
fi

export USER_ID=$(id -u)
export GROUP_ID=$(id -g)


echo "[BOOTSTRAP] Checking if Go is installed..."
if [ ! -x "$(command -v go)" ];
then
  echo "[ERROR] You must install Go on your machine.";
  return
fi

authelia-scripts bootstrap
