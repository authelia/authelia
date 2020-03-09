#! /bin/bash

GITTAG=$(git describe --long --tags | sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g')

echo "--- :linux: Deploy AUR package: ${PACKAGE}"
git clone ssh://aur@aur.archlinux.org/"${PACKAGE}".git
cd "${PACKAGE}" || exit

if [[ $PACKAGE != "authelia-git" ]]; then
  sed -i "/pkgver=/c\pkgver=${BUILDKITE_TAG//v/}" PKGBUILD && \
  docker run --rm -v $PWD:/build authelia/aurpackager bash -c "cd /build && updpkgsums"
else
  sed -i "/pkgver=/c\pkgver=${GITTAG}" PKGBUILD
fi

docker run --rm -v $PWD:/build authelia/aurpackager bash -c "cd /build && makepkg --printsrcinfo >| .SRCINFO" && \
git add . && \
if [[ $PACKAGE != "authelia-git" ]]; then
  git commit -m "Update to ${BUILDKITE_TAG}"
else
  git commit -m "Update to GIT version: ${GITTAG}"
fi
git push