package main

// Docker a docker object
type Docker struct{}

// Build build a docker image
func (d *Docker) Build(tag string, target string) error {
	return CommandWithStdout("docker", "build", "-t", tag, target).Run()
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
