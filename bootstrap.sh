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
else
  export SUITE_PULL_POLICY=always
fi

echo "[BOOTSTRAP] Checking if Go is installed..."
if [[ ! -x "$(command -v go)" ]];
then
  echo "[ERROR] You must install Go on your machine." >&2
  return
fi

# Working trees that share a Docker daemon are each given a slot so their suites cannot collide on the compose project,
# the network subnet, the debug ports or the temporary directory. CI supplies the slot per agent; locally every working
# tree is handed one of its own so several of them can run suites at the same time. Set SUITE_SLOT_AUTO=false to opt
# out. Everything falls back to the historical values when the slot is unset.
if [[ -z "${SUITE_SLOT}" && "${CI}" != "true" && "${SUITE_SLOT_AUTO}" != "false" ]]; then
  if SUITE_SLOT="$(authelia-scripts suites slot)"; then
    export SUITE_SLOT
    echo "[BOOTSTRAP] Using suite slot ${SUITE_SLOT} for ${PWD}"
  else
    unset SUITE_SLOT
    echo "[BOOTSTRAP] Could not allocate a suite slot, falling back to the shared defaults" >&2
  fi
fi

if [[ -n "${SUITE_SLOT}" ]]; then
  export COMPOSE_PROJECT_NAME="authelia-${SUITE_SLOT}"
  export SUITE_SUBNET="10.240.${SUITE_SLOT}"
  export LDAP_ADMIN_PORT="$((9090 + SUITE_SLOT))"
  export ENVOY_ADMIN_PORT="$((9901 + SUITE_SLOT))"

  # The agent container remaps SUITE_TMP onto its own /tmp, so only a local shell has to move the path it reads and
  # writes through as well.
  if [[ "${CI}" != "true" ]]; then
    export SUITE_TMP="${SUITE_TMP:-/tmp/authelia-suite-${SUITE_SLOT}}"
    export SUITE_TMP_PATH="${SUITE_TMP_PATH:-${SUITE_TMP}}"
  fi
fi

if [[ -n "${SUITE_TMP}" ]]; then
  mkdir -p "${SUITE_TMP}"
fi

authelia-scripts bootstrap
