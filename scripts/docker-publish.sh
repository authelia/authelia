#!/bin/bash

if [ "$TRAVIS_BRANCH" == "master" ]; then
  echo "======================================="
  echo "Authelia will be deployed on Dockerhub."
  echo "======================================="
  echo "TRAVIS_TAG='$TRAVIS_TAG'"

  TAG=latest
  if [ ! -z "$TRAVIS_TAG" ]; then
    TAG=$TRAVIS_TAG
  fi

  IMAGE_NAME=clems4ever/authelia
  IMAGE_WITH_TAG=$IMAGE_NAME:$TAG

  echo "Login to Dockerhub."
  docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD";

  echo "Docker image $IMAGE_WITH_TAG will be deployed on Dockerhub."
  docker tag $IMAGE_NAME $IMAGE_WITH_TAG;
  docker push $IMAGE_WITH_TAG;
  echo "Docker image deployed successfully."

else
  echo "Docker image will not be deployed on Dockerhub."
fi

