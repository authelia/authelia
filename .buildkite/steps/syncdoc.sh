#!/bin/bash

set -e

rm -rf authelia
git clone git@github.com:authelia/authelia.git --single-branch --branch gh-pages

pushd docs
bundle install
bundle exec jekyll build -d ../authelia
popd

COMMIT=$(git show -s --format=%h)

pushd authelia
git config user.name "Authelia[bot]"
git config user.email "autheliabot@gmail.com"

git status | grep "nothing to commit" && exit
git add -A
git commit -m "Synchronize docs of commit: ${COMMIT}"
git push
popd

rm -rf authelia
