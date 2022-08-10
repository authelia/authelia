package cmd

import (
	"bufio"
	"fmt"
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
func (d *Docker) Manifest(tag string, registries []string) error {
	args := []string{"buildx", "bake", "-f", "docker-bake.hcl", "--builder", "buildx", "--push"}

	buildMetaData, err := getBuild(ciBranch, os.Getenv("BUILDKITE_BUILD_NUMBER"), "")
	if err != nil {
		return err
	}

	var baseImageTag string

	from, err := getDockerfileDirective("Dockerfile", "FROM")
	if err == nil {
		baseImageTag = from[strings.IndexRune(from, ':')+1:]
	}

	flags := buildMetaData.BakeSetFlags("docker.io/library/alpine:"+baseImageTag, "", "", "")

	for key, value := range flags {
		args = append(args, "--set", fmt.Sprintf("%s=%s", key, value))
	}

	fmt.Printf("Building with docker %s\n", strings.Join(args, " "))

	cmd := utils.CommandWithStdout("docker", args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "IMAGE_TAG="+tag)

	if err = cmd.Run(); err != nil {
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
