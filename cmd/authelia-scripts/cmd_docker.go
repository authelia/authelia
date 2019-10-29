package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// DockerBuildCmd Command for building docker image of Authelia.
var DockerBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the docker image of Authelia",
	Run: func(cmd *cobra.Command, args []string) {
		docker := &Docker{}
		err := docker.Build(IntermediateDockerImageName, ".")
		if err != nil {
			panic(err)
		}

		err = docker.Tag(IntermediateDockerImageName, DockerImageName)

		if err != nil {
			panic(err)
		}
	},
}

// DockerPushCmd Command for pushing Authelia docker image to Dockerhub
var DockerPushCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish Authelia docker image to Dockerhub",
	Run: func(cmd *cobra.Command, args []string) {
		publishDockerImage()
	},
}

func login(docker *Docker) {
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")

	if username == "" {
		panic(errors.New("DOCKER_USERNAME is empty"))
	}

	if password == "" {
		panic(errors.New("DOCKER_PASSWORD is empty"))
	}

	fmt.Println("Login to dockerhub as " + username)
	err := docker.Login(username, password)

	if err != nil {
		fmt.Println("Login to dockerhub failed")
		panic(err)
	}
}

func deploy(docker *Docker, tag string) {
	imageWithTag := DockerImageName + ":" + tag
	fmt.Println("===================================================")
	fmt.Println("Docker image " + imageWithTag + " will be deployed on Dockerhub.")
	fmt.Println("===================================================")

	err := docker.Tag(DockerImageName, imageWithTag)

	if err != nil {
		panic(err)
	}

	docker.Push(imageWithTag)

	if err != nil {
		panic(err)
	}
}

func publishDockerImage() {
	docker := &Docker{}

	travisBranch := os.Getenv("TRAVIS_BRANCH")
	travisPullRequest := os.Getenv("TRAVIS_PULL_REQUEST")
	travisTag := os.Getenv("TRAVIS_TAG")

	if travisBranch == "master" && travisPullRequest == "false" {
		login(docker)
		deploy(docker, "master")
	} else if travisTag != "" {
		login(docker)
		deploy(docker, travisTag)
		deploy(docker, "latest")
	} else {
		fmt.Println("Docker image will not be built")
	}
}
