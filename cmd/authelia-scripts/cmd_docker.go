package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var container string

var containers = []string{"dev", "coverage"}
var defaultContainer = "dev"
var ciBranch = os.Getenv("BUILDKITE_BRANCH")
var ciPullRequest = os.Getenv("BUILDKITE_PULL_REQUEST")
var ciTag = os.Getenv("BUILDKITE_TAG")
var dockerTags = regexp.MustCompile(`v(?P<Patch>(?P<Minor>(?P<Major>\d+)\.\d+)\.\d+.*)`)
var ignoredSuffixes = regexp.MustCompile("alpha|beta")
var publicRepo = regexp.MustCompile(`.*:.*`)
var tags = dockerTags.FindStringSubmatch(ciTag)

func init() {
	DockerBuildCmd.PersistentFlags().StringVar(&container, "container", defaultContainer, "target container among: "+strings.Join(containers, ", "))
}

func checkContainerIsSupported(container string) {
	for _, v := range containers {
		if container == v {
			return
		}
	}

	log.Fatal("Container is not supported. Please select one of " + strings.Join(containers, ", ") + ".")
}

func dockerBuildOfficialImage(arch string) error {
	docker := &Docker{}
	filename := "Dockerfile"
	dockerfile := fmt.Sprintf("%s.%s", filename, arch)

	flags, err := getXFlags(ciBranch, os.Getenv("BUILDKITE_BUILD_NUMBER"), "")
	if err != nil {
		log.Fatal(err)
	}

	return docker.Build(IntermediateDockerImageName, dockerfile, ".",
		strings.Join(flags, " "))
}

// DockerBuildCmd Command for building docker image of Authelia.
var DockerBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the docker image of Authelia",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Building Docker image %s...", DockerImageName)
		checkContainerIsSupported(container)
		err := dockerBuildOfficialImage(container)

		if err != nil {
			log.Fatal(err)
		}

		docker := &Docker{}
		err = docker.Tag(IntermediateDockerImageName, DockerImageName)

		if err != nil {
			log.Fatal(err)
		}
	},
}

// DockerManifestCmd Command for pushing Authelia docker manifest to DockerHub.
var DockerManifestCmd = &cobra.Command{
	Use:   "push-manifest",
	Short: "Publish Authelia docker manifest to Docker Hub",
	Run: func(cmd *cobra.Command, args []string) {
		publishDockerManifest()
	},
}

func login(docker *Docker, registry string) {
	username := ""
	password := ""

	switch registry {
	case dockerhub:
		username = os.Getenv("DOCKER_USERNAME")
		password = os.Getenv("DOCKER_PASSWORD")
	case ghcr:
		username = os.Getenv("GHCR_USERNAME")
		password = os.Getenv("GHCR_PASSWORD")
	}

	if username == "" {
		log.Fatal(errors.New("DOCKER_USERNAME/GHCR_USERNAME is empty"))
	}

	if password == "" {
		log.Fatal(errors.New("DOCKER_PASSWORD/GHCR_PASSWORD is empty"))
	}

	log.Infof("Login to %s as %s", registry, username)
	err := docker.Login(username, password, registry)

	if err != nil {
		log.Fatalf("Login to %s failed: %s", registry, err)
	}
}

func deployManifest(docker *Docker, tag string) {
	log.Infof("Docker manifest %s:%s will be deployed on %s and %s", DockerImageName, tag, dockerhub, ghcr)

	dockerhub := dockerhub + "/" + DockerImageName + ":" + tag
	ghcr := ghcr + "/" + DockerImageName + ":" + tag

	if err := docker.Manifest(dockerhub, ghcr); err != nil {
		log.Fatal(err)
	}
}

func publishDockerManifest() {
	docker := &Docker{}

	switch {
	case ciTag != "":
		if len(tags) == 4 {
			log.Infof("Detected tags: '%s' | '%s' | '%s'", tags[1], tags[2], tags[3])
			login(docker, dockerhub)
			login(docker, ghcr)
			deployManifest(docker, tags[1])
			publishDockerReadme(docker)

			if !ignoredSuffixes.MatchString(ciTag) {
				deployManifest(docker, tags[2])
				deployManifest(docker, tags[3])
				deployManifest(docker, "latest")
				publishDockerReadme(docker)
			}
		} else {
			log.Fatal("Docker manifest will not be published, the specified tag does not conform to the standard")
		}
	case ciBranch != masterTag && !publicRepo.MatchString(ciBranch):
		login(docker, dockerhub)
		login(docker, ghcr)
		deployManifest(docker, ciBranch)
	case ciBranch != masterTag && publicRepo.MatchString(ciBranch):
		login(docker, dockerhub)
		login(docker, ghcr)
		deployManifest(docker, "PR"+ciPullRequest)
	case ciBranch == masterTag && ciPullRequest == stringFalse:
		login(docker, dockerhub)
		login(docker, ghcr)
		deployManifest(docker, "master")
		publishDockerReadme(docker)
	default:
		log.Info("Docker manifest will not be published")
	}
}

func publishDockerReadme(docker *Docker) {
	log.Info("Docker pushing README.md to Docker Hub")

	if err := docker.PublishReadme(); err != nil {
		log.Fatal(err)
	}
}
