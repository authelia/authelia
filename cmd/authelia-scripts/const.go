package main

// OutputDir the output directory where the built version of Authelia is located.
var OutputDir = "dist"

// DockerImageName the official name of Authelia docker image.
var DockerImageName = "authelia/authelia"

// IntermediateDockerImageName local name of the docker image.
var IntermediateDockerImageName = "authelia:dist"

var registries = []string{"docker.io", "ghcr.io"}

const dockerhub = "docker.io"
const ghcr = "ghcr.io"

const masterTag = "master"
const stringFalse = "false"
const stringTrue = "true"
const webDirectory = "web"
