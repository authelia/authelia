package main

// OutputDir the output directory where the built version of Authelia is located.
var OutputDir = "dist"

// DockerImageName the official name of Authelia docker image.
var DockerImageName = "authelia/authelia"

// IntermediateDockerImageName local name of the docker image.
var IntermediateDockerImageName = "authelia:dist"

const dockerhub = "docker.io"
const ghcr = "ghcr.io"

const masterTag = "master"
const stringFalse = "false"
const webDirectory = "web"

const fmtLDFLAGSX = "-X 'github.com/authelia/authelia/v4/internal/utils.%s=%s'"
