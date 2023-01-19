package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// Docker a docker object.
type Docker struct{}

// Build build a docker image.
func (d *Docker) Build(tag, dockerfile, target string, buildMetaData *Build) error {
	args := []string{"build", "-t", tag, "-f", dockerfile, "--progress=plain"}

	for label, value := range buildMetaData.ContainerLabels() {
		if value == "" {
			continue
		}

		args = append(args, "--label", fmt.Sprintf("%s=%s", label, value))
	}

	args = append(args, "--build-arg", "LDFLAGS_EXTRA="+strings.Join(buildMetaData.XFlags(), " "), target)

	return utils.CommandWithStdout("docker", args...).Run()
}

// Tag tag a docker image.
func (d *Docker) Tag(image, tag string) error {
	return utils.CommandWithStdout("docker", "tag", image, tag).Run()
}

// Login login to the dockerhub registry.
func (d *Docker) Login(username, password, registry string) error {
	return utils.CommandWithStdout("bash", "-c", `echo `+password+` | docker login `+registry+` --password-stdin -u `+username).Run()
}

// Manifest push a docker manifest to dockerhub.
func (d *Docker) Manifest(tags []string) error {
	args := []string{"build"}

	for _, tag := range tags {
		args = append(args, "-t", tag)
	}

	annotations := ""

	buildMetaData, err := getBuild(ciBranch, os.Getenv("BUILDKITE_BUILD_NUMBER"), "")
	if err != nil {
		return err
	}

	for label, value := range buildMetaData.ContainerLabels() {
		if value == "" {
			continue
		}

		annotations += fmt.Sprintf("annotation.%s=%s,", label, value)
		args = append(args, "--label", fmt.Sprintf("%s=%s", label, value))
	}

	var baseImageTag string

	from, err := getDockerfileDirective("Dockerfile", "FROM")
	if err == nil {
		baseImageTag = from[strings.IndexRune(from, ':')+1:]
		args = append(args, "--label", "org.opencontainers.image.base.name=docker.io/library/alpine:"+baseImageTag)
	}

	resp, err := http.Get("https://hub.docker.com/v2/repositories/library/alpine/tags/" + baseImageTag + "/images")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	images := DockerImages{}

	if err = json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return err
	}

	var (
		digestAMD64, digestARM, digestARM64 string
	)

	for _, platform := range []string{"linux/amd64", "linux/arm/v7", "linux/arm64"} {
		for _, image := range images {
			if !image.Match(platform) {
				continue
			}

			switch platform {
			case "linux/amd64":
				digestAMD64 = image.Digest
			case "linux/arm/v7":
				digestARM = image.Digest
			case "linux/arm64":
				digestARM64 = image.Digest
			}
		}
	}

	finalArgs := make([]string, len(args))

	copy(finalArgs, args)

	finalArgs = append(finalArgs, "--output", "type=image,\"name="+dockerhub+"/"+DockerImageName+","+ghcr+"/"+DockerImageName+"\","+annotations+"annotation.org.opencontainers.image.base.name=docker.io/library/alpine:"+baseImageTag+",annotation[linux/amd64].org.opencontainers.image.base.digest="+digestAMD64+",annotation[linux/arm/v7].org.opencontainers.image.base.digest="+digestARM+",annotation[linux/arm64].org.opencontainers.image.base.digest="+digestARM64, "--platform", "linux/amd64,linux/arm/v7,linux/arm64", "--builder", "buildx", "--push", ".")

	if err = utils.CommandWithStdout("docker", finalArgs...).Run(); err != nil {
		return err
	}

	return nil
}

// PublishReadme push README.md to dockerhub.
func (d *Docker) PublishReadme() error {
	return utils.CommandWithStdout("bash", "-c", `token=$(curl -fs --retry 3 -H "Content-Type: application/json" -X "POST" -d '{"username": "'$DOCKER_USERNAME'", "password": "'$DOCKER_PASSWORD'"}' https://hub.docker.com/v2/users/login/ | jq -r .token) && jq -n --arg msg "$(cat README.md | sed -r 's/(\<img\ src\=\")(\.\/)/\1https:\/\/github.com\/authelia\/authelia\/raw\/master\//' | sed 's/\.\//https:\/\/github.com\/authelia\/authelia\/blob\/master\//g' | sed '/start \[contributing\]/ a <a href="https://github.com/authelia/authelia/graphs/contributors"><img src="https://opencollective.com/authelia-sponsors/contributors.svg?width=890" /></a>' | sed '/Thanks goes to/,/### Backers/{/### Backers/!d}')" '{"registry":"registry-1.docker.io","full_description": $msg }' | curl -fs --retry 3 -o /dev/null -L -X "PATCH" -H "Content-Type: application/json" -H "Authorization: JWT $token" -d @- https://hub.docker.com/v2/repositories/authelia/authelia/`).Run()
}

func getDockerfileDirective(filePath, directive string) (from string, err error) {
	var f *os.File

	if f, err = os.Open(filePath); err != nil {
		return "", err
	}

	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		data := s.Text()

		if strings.HasPrefix(data, directive+" ") {
			return data[5:], nil
		}
	}

	return "", nil
}
