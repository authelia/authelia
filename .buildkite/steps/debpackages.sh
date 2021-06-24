#!/usr/bin/env bash

GITTAG=$(git describe --long --tags | sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g')

if [[ ${BUILDKITE_TAG} == "" ]]; then
  VERSION="pkgver=${GITTAG}"
else
  VERSION="pkgver=${BUILDKITE_TAG//v/}"
fi

wget https://aur.archlinux.org/cgit/aur.git/plain/PKGBUILD?h=authelia-bin -O PKGBUILD && \
sed -i -e '/^pkgname=/c\pkgname=authelia' -e "/pkgver=/c\pkgver=$VERSION" -e '10,14d' \
-e 's/source_x86_64.*/source_x86_64=("authelia-linux-amd64.tar.gz")/' \
-e 's/source_aarch64.*/source_aarch64=("authelia-linux-arm64v8.tar.gz")/' \
-e 's/source_armv7h.*/source_armv7h=("authelia-linux-arm32v7.tar.gz")/' \
-e 's/sha256sums_x86_64.*/sha256sums_x86_64=("SKIP")/' \
-e 's/sha256sums_aarch64.*/sha256sums_aarch64=("SKIP")/' \
-e 's/sha256sums_armv7h.*/sha256sums_armv7h=("SKIP")/' PKGBUILD && \
docker run --rm -v $PWD:/build authelia/aurpackager bash -c "cd /build && makedeb" && \
docker run --rm -v $PWD:/build nightah/debpackager:armhf bash -c "cd /build && makedeb" && \
docker run --rm -v $PWD:/build nightah/debpackager:arm64 bash -c "cd /build && makedeb"
