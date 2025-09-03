#!/usr/bin/env bash

for FILE in *.deb; do
  echo "--- :debian: :fedora: :ubuntu: Deploy APT repository package for architecture: $(basename ${FILE##*_} .deb)"
    curl -s -H "Authorization: Bearer ${BALTO_TOKEN}" \
    -F "distribution=all" \
    -F "package=@${FILE}" \
    --form-string "readme=$(cat README.md | sed -r 's/(\<img\ src\=\")(\.\/)/\1https:\/\/github.com\/authelia\/authelia\/raw\/master\//' | sed 's/\.\//https:\/\/github.com\/authelia\/authelia\/blob\/master\//g')" \
    https://apt.authelia.com/stable/debian/upload/
    echo -e "\n"
done
