package main

import (
	"github.com/clems4ever/authelia/internal/utils"
)

// Docker a docker object
type Docker struct{}

// Build build a docker image
func (d *Docker) Build(tag, dockerfile, target string) error {
	return utils.CommandWithStdout("docker", "build", "-t", tag, "-f", dockerfile, target).Run()
}

// Tag tag a docker image.
func (d *Docker) Tag(image, tag string) error {
	return utils.CommandWithStdout("docker", "tag", image, tag).Run()
}

// Login login to the dockerhub registry.
func (d *Docker) Login(username, password string) error {
	return utils.CommandWithStdout("docker", "login", "-u", username, "-p", password).Run()
}

// Push push a docker image to dockerhub.
func (d *Docker) Push(tag string) error {
	return utils.CommandWithStdout("docker", "push", tag).Run()
}

// Manifest push a docker manifest to dockerhub.
func (d *Docker) Manifest(tag, amd64tag, arm32v7tag, arm64v8tag string) error {
	err := utils.CommandWithStdout("docker", "manifest", "create", tag, amd64tag, arm32v7tag, arm64v8tag).Run()

	if err != nil {
		panic(err)
	}

	err = utils.CommandWithStdout("docker", "manifest", "annotate", tag, arm32v7tag, "--os", "linux", "--arch", "arm").Run()

	if err != nil {
		panic(err)
	}

	err = utils.CommandWithStdout("docker", "manifest", "annotate", tag, arm64v8tag, "--os", "linux", "--arch", "arm64", "--variant", "v8").Run()

	if err != nil {
		panic(err)
	}

	return utils.CommandWithStdout("docker", "manifest", "push", "--purge", tag).Run()
}

// CleanTag remove a tag from dockerhub.
func (d *Docker) CleanTag(tag string) error {
	return utils.CommandWithStdout("bash", "-c", `token=$(curl -fs --retry 3 -H "Content-Type: application/json" -X "POST" -d '{"username": "'$DOCKER_USERNAME'", "password": "'$DOCKER_PASSWORD'"}' https://hub.docker.com/v2/users/login/ | jq -r .token) && curl -fs --retry 3 -o /dev/null -L -X "DELETE" -H "Authorization: JWT $token" https://hub.docker.com/v2/repositories/"+DockerImageName+"/tags/"+tag+"/`).Run()
}

// PublishReadme push README.md to dockerhub.
func (d *Docker) PublishReadme() error {
	return utils.CommandWithStdout("bash", "-c", `token=$(curl -fs --retry 3 -H "Content-Type: application/json" -X "POST" -d '{"username": "'$DOCKER_USERNAME'", "password": "'$DOCKER_PASSWORD'"}' https://hub.docker.com/v2/users/login/ | jq -r .token) && jq -n --arg msg "$(cat README.md | sed -r 's/(\<img\ src\=\")(\.\/)/\1https:\/\/github.com\/clems4ever\/authelia\/raw\/master\//' | sed 's/\.\//https:\/\/github.com\/clems4ever\/authelia\/blob\/master\//g')" '{"registry":"registry-1.docker.io","full_description": $msg }' | curl -fs --retry 3 -o /dev/null -L -X "PATCH" -H "Content-Type: application/json" -H "Authorization: JWT $token" -d @- https://hub.docker.com/v2/repositories/clems4ever/authelia/`).Run()
}
