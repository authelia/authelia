#!/bin/bash

if [ "$TRAVIS_BRANCH" == "master" ]; then
  TAG=latest
  if [ ! -z "$TRAVIS_TAG" ]; then
    TAG=$TRAVIS_TAG
  fi

  IMAGE_NAME=clems4ever/authelia

  docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD";
  docker tag $IMAGE_NAME $IMAGE_NAME:$TAG;
  docker push $IMAGE_NAME:$TAG;
fi

