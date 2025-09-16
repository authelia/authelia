#!/usr/bin/env bash
set -eu

artifacts=()

for FILE in \
  checksums.sha256{,.sig} \
  authelia-${BUILDKITE_TAG}-{linux-{amd64,arm,arm64,amd64-musl,arm-musl,arm64-musl},freebsd-amd64,public_html}.{tar.gz,tar.gz.{c,sp}dx.json} \
  authelia_${BUILDKITE_TAG/v/}-1_{amd64,armhf,arm64}.deb
do
  artifacts+=(-a "${FILE}")
done

echo "--- :github: Deploy artifacts for release: ${BUILDKITE_TAG}"
hub release create "${BUILDKITE_TAG}" "${artifacts[@]}" -F <(echo -e "${BUILDKITE_TAG}\n$(conventional-changelog -p angular -o /dev/stdout -r 2 | sed -e '1,3d')\n\n### Docker Container\n* \`docker pull authelia/authelia:${BUILDKITE_TAG//v}\`\n* \`docker pull ghcr.io/authelia/authelia:${BUILDKITE_TAG//v}\`"); EXIT=$?

if [[ "${EXIT}" == 0 ]];
  then
    exit
  else
    hub release delete "${BUILDKITE_TAG}" && false
fi
