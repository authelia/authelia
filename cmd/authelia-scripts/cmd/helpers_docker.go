package cmd

import (
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
func (d *Docker) Manifest(tag1, tag2 string) error {
	args := []string{"build", "-t", tag1, "-t", tag2}

	buildMetaData, err := getBuild(ciBranch, os.Getenv("BUILDKITE_BUILD_NUMBER"), "")
	if err != nil {
		return err
	}

	for label, value := range buildMetaData.ContainerLabels() {
		if value == "" {
			continue
		}

		args = append(args, "--label", fmt.Sprintf("%s=%s", label, value))
	}

	baseImageTag := "3.16.1"

	resp, err := http.Get("https://hub.docker.com/v2/repositories/library/alpine/tags/" + baseImageTag + "/images")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	images := DockerImages{}

	if err = json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return err
	}

	for _, platform := range []string{"linux/amd64", "linux/arm/v7", "linux/arm64"} {
		var digest string

		for _, image := range images {
			arch := image.Architecture
			if image.Variant != nil {
				arch += "/" + image.Variant.(string)
			}

			if arch == platform {
				digest = image.Digest

				break
			}
		}

		if digest == "" {
			fmt.Printf("Skipping %s\n", platform)
			continue
		}

		var finalArgs []string

		copy(finalArgs, args)

		finalArgs = append(finalArgs, "--label", "org.opencontainers.image.base.name=library/alpine:"+baseImageTag, "--label", "org.opencontainers.image.base.digest="+digest, "--platform", platform, "--builder", "buildx", "--push", ".")

		if err = utils.CommandWithStdout("docker", finalArgs...).Run(); err != nil {
			return err
		}
	}

	return nil
}

// PublishReadme push README.md to dockerhub.
func (d *Docker) PublishReadme() error {
	return utils.CommandWithStdout("bash", "-c", `token=$(curl -fs --retry 3 -H "Content-Type: application/json" -X "POST" -d '{"username": "'$DOCKER_USERNAME'", "password": "'$DOCKER_PASSWORD'"}' https://hub.docker.com/v2/users/login/ | jq -r .token) && jq -n --arg msg "$(cat README.md | sed -r 's/(\<img\ src\=\")(\.\/)/\1https:\/\/github.com\/authelia\/authelia\/raw\/master\//' | sed 's/\.\//https:\/\/github.com\/authelia\/authelia\/blob\/master\//g' | sed '/start \[contributing\]/ a <a href="https://github.com/authelia/authelia/graphs/contributors"><img src="https://opencollective.com/authelia-sponsors/contributors.svg?width=890" /></a>' | sed '/Thanks goes to/,/### Backers/{/### Backers/!d}')" '{"registry":"registry-1.docker.io","full_description": $msg }' | curl -fs --retry 3 -o /dev/null -L -X "PATCH" -H "Content-Type: application/json" -H "Authorization: JWT $token" -d @- https://hub.docker.com/v2/repositories/authelia/authelia/`).Run()
}
