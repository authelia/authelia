#!/usr/bin/env bash

GITTAG=$(git describe --long --tags | sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g')

if [[ "${BUILDKITE_TAG}" == "" ]]; then
  VERSION="pkgver=${GITTAG}"
else
  VERSION="pkgver=${BUILDKITE_TAG//v/}"
fi

wget https://aur.archlinux.org/cgit/aur.git/plain/PKGBUILD?h=authelia-bin -qO PKGBUILD && \
sed -i -e '/^pkgname=/c pkgname=authelia' -e "/pkgver=/c $VERSION" -e '10,14d' \
-e 's/source_x86_64.*/source_x86_64=("authelia-linux-amd64.tar.gz")/' \
-e 's/source_aarch64.*/source_aarch64=("authelia-linux-arm64.tar.gz")/' \
-e 's/source_armv7h.*/source_armv7l=("authelia-linux-arm.tar.gz")/' \
-e 's/sha256sums_x86_64.*/sha256sums_x86_64=("SKIP")/' \
-e 's/sha256sums_aarch64.*/sha256sums_aarch64=("SKIP")/' \
-e 's/sha256sums_armv7h.*/sha256sums_armv7l=("SKIP")/' PKGBUILD

if [[ "${PACKAGE}" == "amd64" ]]; then
  docker run --rm -v $PWD:/build authelia/aurpackager bash -c "cd /build && makedeb"
elif [[ "${PACKAGE}" == "armhf" ]]; then
  docker run --rm --platform linux/arm/v7 -v $PWD:/build authelia/debpackager bash -c "cd /build && makedeb -A"
else
  docker run --rm --platform linux/arm64 -v $PWD:/build authelia/debpackager bash -c "cd /build && makedeb"
fi
