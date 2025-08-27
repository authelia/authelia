#!/usr/bin/env bash
set -eu

artifacts=()

for FILE in \
  authelia-linux-amd64.tar.gz authelia-linux-amd64.tar.gz.sha256 authelia-linux-amd64.tar.gz.sha256.sig \
  authelia-linux-arm.tar.gz authelia-linux-arm.tar.gz.sha256 authelia-linux-arm.tar.gz.sha256.sig \
  authelia-linux-arm64.tar.gz authelia-linux-arm64.tar.gz.sha256 authelia-linux-arm64.tar.gz.sha256.sig \
  authelia-linux-amd64-musl.tar.gz authelia-linux-amd64-musl.tar.gz.sha256 authelia-linux-amd64-musl.tar.gz.sha256.sig \
  authelia-linux-arm-musl.tar.gz authelia-linux-arm-musl.tar.gz.sha256 authelia-linux-arm-musl.tar.gz.sha256.sig \
  authelia-linux-arm64-musl.tar.gz authelia-linux-arm64-musl.tar.gz.sha256 authelia-linux-arm64-musl.tar.gz.sha256.sig \
  authelia-freebsd-amd64.tar.gz authelia-freebsd-amd64.tar.gz.sha256 authelia-freebsd-amd64.tar.gz.sha256.sig \
  authelia-public_html.tar.gz authelia-public_html.tar.gz.sha256 authelia-public_html.tar.gz.sha256.sig;
do
  # Add the version to the artifact name
  mv ${FILE} ${FILE/authelia-/authelia-${BUILDKITE_TAG}-}
  artifacts+=(-a "${FILE/authelia-/authelia-${BUILDKITE_TAG}-}")
done

for FILE in \
  authelia_amd64.deb authelia_amd64.deb.sha256 authelia_amd64.deb.sha256.sig \
  authelia_arm64.deb authelia_arm64.deb.sha256 authelia_arm64.deb.sha256.sig \
  authelia_armhf.deb authelia_armhf.deb.sha256 authelia_armhf.deb.sha256.sig;
do
  # Add the version to the artifact name
  mv ${FILE} ${FILE/authelia_/authelia_${BUILDKITE_TAG}_}
  artifacts+=(-a "${FILE/authelia_/authelia_${BUILDKITE_TAG}_}")
done

echo "--- :github: Deploy artifacts for release: ${BUILDKITE_TAG}"
hub release create "${BUILDKITE_TAG}" "${artifacts[@]}" -F <(echo -e "${BUILDKITE_TAG}\n$(conventional-changelog -p angular -o /dev/stdout -r 2 | sed -e '1,3d')\n\n### Docker Container\n* \`docker pull authelia/authelia:${BUILDKITE_TAG//v}\`\n* \`docker pull ghcr.io/authelia/authelia:${BUILDKITE_TAG//v}\`"); EXIT=$?

if [[ "${EXIT}" == 0 ]];
  then
    exit
  else
    hub release delete "${BUILDKITE_TAG}" && false
fi
