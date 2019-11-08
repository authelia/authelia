package main

// Docker a docker object
type Docker struct{}

// Build build a docker image
func (d *Docker) Build(tag, dockerfile, target string) error {
	return CommandWithStdout("docker", "build", "-t", tag, "-f", dockerfile, target).Run()
}

// Tag tag a docker image.
func (d *Docker) Tag(image, tag string) error {
	return CommandWithStdout("docker", "tag", image, tag).Run()
}

// Login login to the dockerhub registry.
func (d *Docker) Login(username, password string) error {
	return CommandWithStdout("docker", "login", "-u", username, "-p", password).Run()
}

// Push push a docker image to dockerhub.
func (d *Docker) Push(tag string) error {
	return CommandWithStdout("docker", "push", tag).Run()
}

// Manifest push a docker manifest to dockerhub.
func (d *Docker) Manifest(tag, amd64tag, arm32v7tag, arm64v8tag string) error {
	err := CommandWithStdout("docker", "manifest", "create", tag, amd64tag, arm32v7tag, arm64v8tag).Run()

	if err != nil {
		panic(err)
	}

	err = CommandWithStdout("docker", "manifest", "annotate", tag, arm32v7tag, "--os", "linux", "--arch", "arm").Run()

	if err != nil {
		panic(err)
	}

	err = CommandWithStdout("docker", "manifest", "annotate", tag, arm64v8tag, "--os", "linux", "--arch", "arm64", "--variant", "v8").Run()

	if err != nil {
		panic(err)
	}

	return CommandWithStdout("docker", "manifest", "push", "--purge", tag).Run()
}

// CleanTag remove a tag from dockerhub.
func (d *Docker) CleanTag(username, password, tag string) error {
	return CommandWithStdout("curl", "-u", username+":"+password, "-X", "DELETE", "https://cloud.docker.com/v2/repositories/"+DockerImageName+"/tags/"+tag+"/").Run()
}
