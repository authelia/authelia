#!/bin/bash

if [ "$TRAVIS_BRANCH" == "master" ]; then
  TAG=latest
  if [ ! -z "$TRAVIS_TAG" ]; then
    TAG=$TRAVIS_TAG
  fi

  docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD";
  docker push clems4ever/authelia:$TAG;
fi

