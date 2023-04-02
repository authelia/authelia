#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

GITTAG=$(git describe --long --tags | sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g')

echo "--- :linux: Deploy AUR package: ${PACKAGE}"
git clone ssh://aur@aur.archlinux.org/"${PACKAGE}".git
cd "${PACKAGE}" || exit

if [[ "${PACKAGE}" != "authelia-git" ]]; then
  sed -i -e "/^pkgver=/c pkgver=${BUILDKITE_TAG//v/}" \
  -e '/^pkgrel=/c pkgrel=1' PKGBUILD && \
  docker run --rm -v $PWD:/build authelia/aurpackager bash -c "cd /build && updpkgsums"
else
  sed -i -e "/^pkgver=/c pkgver=${GITTAG}" \
  -e '/^pkgrel=/c pkgrel=1' PKGBUILD
fi

docker run --rm -v $PWD:/build authelia/aurpackager bash -c "cd /build && makepkg --printsrcinfo >| .SRCINFO" && \
git add . && \
if [[ "${PACKAGE}" != "authelia-git" ]]; then
  git commit -m "Update to ${BUILDKITE_TAG}"
else
  git commit -m "Update to GIT version: ${GITTAG}"
fi
git push
