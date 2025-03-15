#!/usr/bin/env bash

GITTAG=$(git describe --long --tags | sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g')

if [[ "${BUILDKITE_TAG}" == "" ]]; then
  VERSION="pkgver=${GITTAG}"
else
  VERSION="pkgver=${BUILDKITE_TAG//v/}"
fi

wget https://aur.archlinux.org/cgit/aur.git/plain/PKGBUILD?h=authelia-bin -qO PKGBUILD && \
sed -i -e '/^pkgname=/c pkgname=authelia' -e "/pkgver=/c $VERSION" -e '10,14d' -e "s/'etc/'\/etc/g" \
-e 's/source_x86_64.*/source_amd64=("authelia-linux-amd64.tar.gz")/' \
-e 's/source_aarch64.*/source_arm64=("authelia-linux-arm64.tar.gz")/' \
-e 's/source_armv7h.*/source_armhf=("authelia-linux-arm.tar.gz")/' \
-e 's/sha256sums_x86_64.*/sha256sums_amd64=("SKIP")/' \
-e 's/sha256sums_aarch64.*/sha256sums_arm64=("SKIP")/' \
-e 's/sha256sums_armv7h.*/sha256sums_armhf=("SKIP")/' \
-e 's/x86_64/amd64/g' -e 's/aarch64/arm64/g' -e 's/armv7h/armhf/g' \
-e 's/CARCH/MAKEDEB_DPKG_ARCHITECTURE/g' \
-e "s/provides=('authelia')/postinst='authelia.postinst'\nprovides=('authelia')/g" PKGBUILD

tee -a authelia.postinst > /dev/null << AUTHELIAPLUSULTRA
#!/bin/sh

# Trigger a reload of sysusers.
systemd-sysusers

ROOT="/etc/authelia"

f=`stat -c "%g" ${ROOT}`
c=`stat -c "%g" ${ROOT}/configuration.yml`

# Check permissions of /etc/authelia and /etc/authelia/configuration.yml.
#
# The intent behind this is if either the /etc/authelia or /etc/authelia/configuration.yml file is currently owned
# by root that we update it to be the authelia user which should have just been created.
#
# This effectively lets anyone update the permissions/mode as they see fit as long as they don't modify the grp.

if [ "$f" = "0" ] || [ "$c" = "0" ]; then
        chgrp -R authelia ${ROOT}
        chmod -R 750 ${ROOT}
        chmod 740 ${ROOT}/configuration.yml
fi


AUTHELIAPLUSULTRA

chmod +x authelia.postinst

if [[ "${PACKAGE}" == "amd64" ]]; then
  docker run --rm -v $PWD:/build authelia/debpackager bash -c "cd /build && makedeb"
elif [[ "${PACKAGE}" == "armhf" ]]; then
  docker run --rm --platform linux/arm/v7 -v $PWD:/build authelia/debpackager bash -c "cd /build && makedeb"
else
  docker run --rm --platform linux/arm64 -v $PWD:/build authelia/debpackager bash -c "cd /build && makedeb"
fi
