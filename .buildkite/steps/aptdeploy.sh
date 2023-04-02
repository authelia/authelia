#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

for FILE in authelia_amd64.deb authelia_arm64.deb authelia_armhf.deb; do
  mv ${FILE} ${FILE/authelia_/authelia_${BUILDKITE_TAG//v}-1_}
done

for ARCH in amd64 arm64 armhf; do
  echo "--- :debian: :fedora: :ubuntu: Deploy APT repository package for architecture: ${ARCH}"
  curl -s -H "Authorization: Bearer ${BALTO_TOKEN}" \
  -F "distribution=all" \
  -F "package=@authelia_${BUILDKITE_TAG//v}-1_${ARCH}.deb" \
  --form-string "readme=$(cat README.md | sed -r 's/(\<img\ src\=\")(\.\/)/\1https:\/\/github.com\/authelia\/authelia\/raw\/master\//' | sed 's/\.\//https:\/\/github.com\/authelia\/authelia\/blob\/master\//g')" \
  https://apt.authelia.com/stable/debian/upload/
  echo -e "\n"
done
