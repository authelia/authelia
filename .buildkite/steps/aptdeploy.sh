#!/usr/bin/env bash

APT_CONF_ROOT="./apt/conf"
APT_DB_ROOT="./apt/db"
APT_REPO_ROOT="./apt/static"
APT_DISTS=("stable")
APT_ARCHS=("binary-amd64" "binary-arm64" "binary-armhf")

echo "--- :debian: :fedora: :ubuntu: Updating APT repository for release: ${BUILDKITE_TAG}"
git clone git@github.com:authelia/apt.git --single-branch --depth 1

mkdir -p ${APT_DB_ROOT} ${APT_REPO_ROOT}/pool/main/dists

for DIST in ${APT_DISTS[@]}; do
  mkdir -p ${APT_REPO_ROOT}/dists/${DIST}/main
  for ARCH in ${APT_ARCHS[@]}; do
    mkdir -p ${APT_REPO_ROOT}/dists/${DIST}/main/${ARCH}
  done
done

for FILE in *.deb; do
  echo "/pool/main/dists/stable/${FILE} https://github.com/authelia/authelia/releases/download/${BUILDKITE_TAG}/${FILE} 301" >> ${APT_REPO_ROOT}/_redirects
  mv ${FILE} /buildkite/.cache/apt/
done

cp -a /buildkite/.cache/apt ${APT_REPO_ROOT}/pool/main/dists/stable

apt-ftparchive generate ${APT_CONF_ROOT}/apt-ftparchive.conf

for DIST in ${APT_DISTS[@]}; do
  apt-ftparchive -c ${APT_CONF_ROOT}/${DIST}.conf release ${APT_REPO_ROOT}/dists/${DIST} > ${APT_REPO_ROOT}/dists/${DIST}/Release
  rm -f ${APT_REPO_ROOT}/dists/${DIST}/Release.gpg ${APT_REPO_ROOT}/dists/${DIST}/InRelease
  gpg --batch --pinentry-mode loopback -u security@authelia.com --passphrase ${GPG_PASSWORD} -b -o ${APT_REPO_ROOT}/dists/${DIST}/Release.gpg ${APT_REPO_ROOT}/dists/${DIST}/Release
  gpg --batch --pinentry-mode loopback -u security@authelia.com --passphrase ${GPG_PASSWORD} --clearsign -o ${APT_REPO_ROOT}/dists/${DIST}/InRelease ${APT_REPO_ROOT}/dists/${DIST}/Release
done

cd apt
git add -A
GIT_AUTHOR_NAME="autheliabot" GIT_AUTHOR_EMAIL="autheliabot@users.noreply.github.com" GIT_COMMITTER_NAME="autheliabot" GIT_COMMITTER_EMAIL="autheliabot@users.noreply.github.com" \
git commit -m "release: ${BUILDKITE_TAG}"
git push
