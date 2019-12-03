package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/clems4ever/authelia/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var arch string
var supportedArch = []string{"amd64", "arm32v7", "arm64v8"}
var defaultArch = "amd64"
var travisBranch = os.Getenv("TRAVIS_BRANCH")
var travisPullRequest = os.Getenv("TRAVIS_PULL_REQUEST")
var travisTag = os.Getenv("TRAVIS_TAG")
var dockerTags = regexp.MustCompile(`(?P<Minor>(?P<Major>v\d+)\.\d+)\.\d+.*`)
var ignoredSuffixes = regexp.MustCompile("alpha|beta")
var tags = dockerTags.FindStringSubmatch(travisTag)

func init() {
	DockerBuildCmd.PersistentFlags().StringVar(&arch, "arch", defaultArch, "target architecture among: "+strings.Join(supportedArch, ", "))
	DockerPushCmd.PersistentFlags().StringVar(&arch, "arch", defaultArch, "target architecture among: "+strings.Join(supportedArch, ", "))

}

func checkArchIsSupported(arch string) {
	for _, a := range supportedArch {
		if arch == a {
			return
		}
	}
	log.Fatal("Architecture is not supported. Please select one of " + strings.Join(supportedArch, ", ") + ".")
}

func dockerBuildOfficialImage(arch string) error {
	docker := &Docker{}
	// Set default Architecture Dockerfile to amd64
	dockerfile := "Dockerfile"
	// Set version of QEMU
	qemuversion := "v4.1.1-1"

	// If not the default value
	if arch != defaultArch {
		dockerfile = fmt.Sprintf("%s.%s", dockerfile, arch)
	}

	if arch == "arm32v7" {
		err := utils.CommandWithStdout("docker", "run", "--rm", "--privileged", "multiarch/qemu-user-static", "--reset", "-p", "yes").Run()

		if err != nil {
			panic(err)
		}

		err = utils.CommandWithStdout("bash", "-c", "wget https://github.com/multiarch/qemu-user-static/releases/download/"+qemuversion+"/qemu-arm-static -O ./qemu-arm-static && chmod +x ./qemu-arm-static").Run()

		if err != nil {
			panic(err)
		}
	} else if arch == "arm64v8" {
		err := utils.CommandWithStdout("docker", "run", "--rm", "--privileged", "multiarch/qemu-user-static", "--reset", "-p", "yes").Run()

		if err != nil {
			panic(err)
		}

		err = utils.CommandWithStdout("bash", "-c", "wget https://github.com/multiarch/qemu-user-static/releases/download/"+qemuversion+"/qemu-aarch64-static -O ./qemu-aarch64-static && chmod +x ./qemu-aarch64-static").Run()

		if err != nil {
			panic(err)
		}
	}

	return docker.Build(IntermediateDockerImageName, dockerfile, ".")
}

// DockerBuildCmd Command for building docker image of Authelia.
var DockerBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the docker image of Authelia",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Building Docker image %s...", DockerImageName)
		checkArchIsSupported(arch)
		err := dockerBuildOfficialImage(arch)

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
	Use:   "push-image",
	Short: "Publish Authelia docker image to Dockerhub",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Pushing Docker image %s to dockerhub...", DockerImageName)
		checkArchIsSupported(arch)
		publishDockerImage(arch)
	},
}

// DockerManifestCmd Command for pushing Authelia docker manifest to Dockerhub
var DockerManifestCmd = &cobra.Command{
	Use:   "push-manifest",
	Short: "Publish Authelia docker manifest to Dockerhub",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Pushing Docker manifest of %s to dockerhub...", DockerImageName)
		publishDockerManifest()
	},
}

func login(docker *Docker) {
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")

	if username == "" {
		log.Fatal(errors.New("DOCKER_USERNAME is empty"))
	}

	if password == "" {
		log.Fatal(errors.New("DOCKER_PASSWORD is empty"))
	}

	log.Infof("Login to dockerhub as %s", username)
	err := docker.Login(username, password)

	if err != nil {
		log.Fatal("Login to dockerhub failed", err)
	}
}

func deploy(docker *Docker, tag string) {
	imageWithTag := DockerImageName + ":" + tag

	log.Infof("Docker image %s will be deployed on Dockerhub", imageWithTag)

	if err := docker.Tag(DockerImageName, imageWithTag); err != nil {
		log.Fatal(err)
	}

	if err := docker.Push(imageWithTag); err != nil {
		log.Fatal(err)
	}
}

func deployManifest(docker *Docker, tag string, amd64tag string, arm32v7tag string, arm64v8tag string) {
	dockerImagePrefix := DockerImageName + ":"

	log.Infof("Docker manifest %s%s will be deployed on Dockerhub", dockerImagePrefix, tag)

	err := docker.Manifest(dockerImagePrefix+tag, dockerImagePrefix+amd64tag, dockerImagePrefix+arm32v7tag, dockerImagePrefix+arm64v8tag)

	if err != nil {
		log.Fatal(err)
	}

	tags := []string{amd64tag, arm32v7tag, arm64v8tag}
	for _, t := range tags {
		log.Infof("Docker removing tag for %s%s on Dockerhub", dockerImagePrefix, t)

		if err := docker.CleanTag(t); err != nil {
			panic(err)
		}
	}

	log.Info("Docker pushing README.md to Dockerhub")

	if err := docker.PublishReadme(); err != nil {
		log.Fatal(err)
	}
}

func publishDockerImage(arch string) {
	docker := &Docker{}

	if travisBranch == "master" && travisPullRequest == "false" {
		login(docker)
		deploy(docker, "master-"+arch)
	} else if travisTag != "" {
		if len(tags) == 3 {
			login(docker)
			deploy(docker, tags[0]+"-"+arch)
		} else {
			log.Fatal("Docker image will not be published, the specified tag does not conform to the standard")
		}
		if !ignoredSuffixes.MatchString(travisTag) {
			deploy(docker, tags[1]+"-"+arch)
			deploy(docker, tags[2]+"-"+arch)
			deploy(docker, "latest-"+arch)
		}
	} else {
		log.Info("Docker image will not be published")
	}
}

func publishDockerManifest() {
	docker := &Docker{}

	if travisBranch == "master" && travisPullRequest == "false" {
		login(docker)
		deployManifest(docker, "master", "master-amd64", "master-arm32v7", "master-arm64v8")
	} else if travisTag != "" {
		if len(tags) == 3 {
			login(docker)
			deployManifest(docker, tags[0], tags[0]+"-amd64", tags[0]+"-arm32v7", tags[0]+"-arm64v8")
		} else {
			log.Fatal("Docker manifest will not be published, the specified tag does not conform to the standard")
		}
		if !ignoredSuffixes.MatchString(travisTag) {
			deployManifest(docker, tags[1], tags[1]+"-amd64", tags[1]+"-arm32v7", tags[1]+"-arm64v8")
			deployManifest(docker, tags[2], tags[2]+"-amd64", tags[2]+"-arm32v7", tags[2]+"-arm64v8")
			deployManifest(docker, "latest", "latest-amd64", "latest-arm32v7", "latest-arm64v8")
		}
	} else {
		fmt.Println("Docker manifest will not be published")
	}
}
