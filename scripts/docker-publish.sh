#!/bin/bash

# Parameters:
# TAG - The name of the tag to use for publishing in Dockerhub

function login_to_dockerhub {
  echo "Logging in to Dockerhub as $DOCKER_USERNAME."
  docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD
  if [ "$?" -ne "0" ];
  then
    echo "Logging in to Dockerhub failed.";
    exit 1
  fi
}

function deploy_on_dockerhub {
  TAG=$1
  IMAGE_NAME=clems4ever/authelia
  IMAGE_WITH_TAG=$IMAGE_NAME:$TAG

  echo "==========================================================="
  echo "Docker image $IMAGE_WITH_TAG will be deployed on Dockerhub."
  echo "==========================================================="
  docker build -t $IMAGE_NAME .
  docker tag $IMAGE_NAME $IMAGE_WITH_TAG;
  docker push $IMAGE_WITH_TAG;
  echo "Docker image deployed successfully."
}


if [ "$TRAVIS_BRANCH" == "master" ]; then
  login_to_dockerhub
  deploy_on_dockerhub master
elif [ "$TRAVIS_BRANCH" == "develop" ]; then
  login_to_dockerhub
  deploy_on_dockerhub develop
elif [ ! -z "$TRAVIS_TAG" ]; then
  login_to_dockerhub
  deploy_on_dockerhub $TRAVIS_TAG
  deploy_on_dockerhub latest
else
  echo "Docker image will not be deployed on Dockerhub."
fi

