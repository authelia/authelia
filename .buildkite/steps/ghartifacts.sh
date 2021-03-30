#!/usr/bin/env bash
set -eu

artifacts=()

for FILES in \
  authelia-linux-amd64.tar.gz authelia-linux-amd64.tar.gz.sha256 \
  authelia-linux-arm32v7.tar.gz authelia-linux-arm32v7.tar.gz.sha256 \
  authelia-linux-arm64v8.tar.gz authelia-linux-arm64v8.tar.gz.sha256 \
  authelia-public_html.tar.gz authelia-public_html.tar.gz.sha256;
do
  artifacts+=(-a "${FILES}")
done

echo "--- :github: Deploy artifacts for release: ${BUILDKITE_TAG}"
hub release create "${BUILDKITE_TAG}" "${artifacts[@]}" -F <(echo -e "${BUILDKITE_TAG}\n$(conventional-changelog -p angular -o /dev/stdout -r 2 | sed -e '1,3d')\n\n### Docker Container\n* \`docker pull authelia/authelia:${BUILDKITE_TAG//v}\`\n* \`docker pull ghcr.io/authelia/authelia:${BUILDKITE_TAG//v}\`"); EXIT=$?

if [[ $EXIT == 0 ]];
  then
    exit
  else
    hub release delete "${BUILDKITE_TAG}" && false
fi
