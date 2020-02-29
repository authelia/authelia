#!/bin/bash

set -e

rm -rf authelia
git clone git@github.com:authelia/authelia.git --branch gh-pages

pushd docs
bundle install
bundle exec jekyll build -d ../authelia
popd

COMMIT=$(git show -s --format=%h)

pushd authelia
git add -A
git commit -m "synchronize docs of commit ${COMMIT}"
git push
popd

rm -rf authelia