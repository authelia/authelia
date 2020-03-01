#!/bin/bash
set -eu

artifacts=()

for FILES in \
  authelia-linux-amd64.tar.gz authelia-linux-amd64.tar.gz.sha256 \
  authelia-linux-arm32v7.tar.gz authelia-linux-arm32v7.tar.gz.sha256 \
  authelia-linux-arm64v8.tar.gz authelia-linux-arm64v8.tar.gz.sha256;
do
  artifacts+=(-a "${FILES}")
done

echo "--- :github: Deploy artifacts for release: ${BUILDKITE_TAG}"
git fetch
hub release create "${BUILDKITE_TAG}" "${artifacts[@]}" -F <(echo -e "${BUILDKITE_TAG}\n\n## Changelog\n$(git log --oneline --pretty='* %h %s' $(git describe --abbrev=0 --tags $(git rev-list --tags --skip=1 --max-count=1))...$(git describe --abbrev=0 --tags))\n\n## Docker Container\n* \`docker pull authelia/authelia:${BUILDKITE_TAG//v}\`"); EXIT=$?

if [[ $EXIT -eq 0 ]];
  then
    exit
  else
    hub release delete "${BUILDKITE_TAG}" && false
fi
