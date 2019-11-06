package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func DockerBuildOfficialImage() error {
	docker := &Docker{}
	// Set default Architecture Dockerfile to amd64
	Dockerfile := "Dockerfile"
	if dockerfile := os.Getenv("DOCKERFILE"); dockerfile != ""{
		Dockerfile = dockerfile
	}
	return docker.Build(IntermediateDockerImageName, Dockerfile, ".")
}

// DockerBuildCmd Command for building docker image of Authelia.
var DockerBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the docker image of Authelia",
	Run: func(cmd *cobra.Command, args []string) {
		err := DockerBuildOfficialImage()

		if err != nil {
			log.Fatal(err)
		}

		docker := &Docker{}
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

// DockerManifestCmd Command for pushing Authelia docker manifest to Dockerhub
var DockerManifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Publish Authelia docker manifest to Dockerhub",
	Run: func(cmd *cobra.Command, args []string) {
		publishDockerManifest()
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

	err = docker.Push(imageWithTag)

	if err != nil {
		panic(err)
	}
}

func deployManifest(docker *Docker, tag string, amd64tag string, arm32v7tag string, arm64v8tag string) {
	imageWithTag := DockerImageName + ":" + tag
	fmt.Println("===================================================")
	fmt.Println("Docker manifest " + imageWithTag + " will be deployed on Dockerhub.")
	fmt.Println("===================================================")

	err := docker.Tag(DockerImageName, imageWithTag)

	if err != nil {
		panic(err)
	}

	err = docker.Manifest(imageWithTag, amd64tag, arm32v7tag, arm64v8tag)

	if err != nil {
		panic(err)
	}
}

func publishDockerImage() {
	docker := &Docker{}

	ARCH := os.Getenv("ARCH")
	travisBranch := os.Getenv("TRAVIS_BRANCH")
	travisPullRequest := os.Getenv("TRAVIS_PULL_REQUEST")
	travisTag := os.Getenv("TRAVIS_TAG")

	if travisBranch == "master" && travisPullRequest == "false" {
		login(docker)
		deploy(docker, "master-" + ARCH)
	} else if travisTag != "" {
		login(docker)
		deploy(docker, travisTag + "-" + ARCH)
		deploy(docker, "latest-" + ARCH)
	} else {
		fmt.Println("Docker image will not be published")
	}
}

func publishDockerManifest() {
	docker := &Docker{}

	travisBranch := os.Getenv("TRAVIS_BRANCH")
	travisPullRequest := os.Getenv("TRAVIS_PULL_REQUEST")
	travisTag := os.Getenv("TRAVIS_TAG")

	if travisBranch == "master" && travisPullRequest == "false" {
		login(docker)
		deployManifest(docker, "master", "master-amd64", "master-arm32v7", "master-arm64v8")
	} else if travisTag != "" {
		login(docker)
		deployManifest(docker, travisTag, travisTag + "-amd64", travisTag + "-arm32v7", travisTag + "-arm64v8")
		deployManifest(docker, "latest", "latest-amd64", "latest-arm32v7", "latest-arm64v8")
	} else {
		fmt.Println("Docker manifest will not be published")
	}
}
