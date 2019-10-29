package main

// OutputDir the output directory where the built version of Authelia is located
var OutputDir = "dist"

// DockerImageName the official name of authelia docker image
var DockerImageName = "clems4ever/authelia"

// IntermediateDockerImageName local name of the docker image
var IntermediateDockerImageName = "authelia:dist"

// RunningSuiteFile name of the file containing the currently running suite
var RunningSuiteFile = ".suite"
