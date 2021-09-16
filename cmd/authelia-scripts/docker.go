package main

import (
	"github.com/authelia/authelia/v4/internal/utils"
)

// Docker a docker object.
type Docker struct{}

// Build build a docker image.
func (d *Docker) Build(tag, dockerfile, target, ldflags string) error {
	return utils.CommandWithStdout(
		"docker", "build", "-t", tag, "-f", dockerfile,
		"--build-arg", "LDFLAGS_EXTRA="+ldflags,
		target).Run()
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
	return utils.CommandWithStdout("docker", "build", "-t", tag1, "-t", tag2, "--platform", "linux/amd64,linux/arm/v7,linux/arm64", "--builder", "buildx", "--push", ".").Run()
}

// PublishReadme push README.md to dockerhub.
func (d *Docker) PublishReadme() error {
	return utils.CommandWithStdout("bash", "-c", `token=$(curl -fs --retry 3 -H "Content-Type: application/json" -X "POST" -d '{"username": "'$DOCKER_USERNAME'", "password": "'$DOCKER_PASSWORD'"}' https://hub.docker.com/v2/users/login/ | jq -r .token) && jq -n --arg msg "$(cat README.md | sed -r 's/(\<img\ src\=\")(\.\/)/\1https:\/\/github.com\/authelia\/authelia\/raw\/master\//' | sed 's/\.\//https:\/\/github.com\/authelia\/authelia\/blob\/master\//g' | sed '/start \[contributing\]/ a <a href="https://github.com/authelia/authelia/graphs/contributors"><img src="https://opencollective.com/authelia-sponsors/contributors.svg?width=890" /></a>' | sed '/Thanks goes to/,/### Backers/{/### Backers/!d}')" '{"registry":"registry-1.docker.io","full_description": $msg }' | curl -fs --retry 3 -o /dev/null -L -X "PATCH" -H "Content-Type: application/json" -H "Authorization: JWT $token" -d @- https://hub.docker.com/v2/repositories/authelia/authelia/`).Run()
}
