package cmd

// OutputDir the output directory where the built version of Authelia is located.
var OutputDir = "dist"

// DockerImageName the official name of Authelia docker image.
var DockerImageName = "authelia/authelia"

// IntermediateDockerImageName local name of the docker image.
var IntermediateDockerImageName = "authelia:dist"

const (
	dockerhub = "docker.io"
	ghcr      = "ghcr.io"
)

const (
	masterTag    = "master"
	stringFalse  = "false"
	webDirectory = "web"
)

const (
	pathPNPMStore  = "/.local/share/pnpm/store"
	pathAuthelia   = "/authelia"
	extTarballGzip = ".tar.gz"
)

const (
	txtDirectoryTidle = "` directory"
	txtRunningSuite   = "Running suite ("
)

const fmtLDFLAGSX = "-X 'github.com/authelia/authelia/v4/internal/utils.%s=%s'"

const (
	cmdRootShort = "A utility used in the Authelia development process."

	cmdRootLong = `The authelia-scripts utility is utilized by developers and the CI/CD pipeline for configuring
testing suites and various other aspects of the environment.

It can be used to automate or manually run unit testing, integration testing, etc.`

	cmdRootExample = `authelia-scripts help`

	cmdBootstrapShort = "Prepare environment for development and testing"

	cmdBootstrapLong = `Prepare environment for development and testing.`

	cmdBootstrapExample = `authelia-scripts bootstrap`

	cmdBuildShort = "Build Authelia binary and static assets"

	cmdBuildLong = `Build Authelia binary and static assets.`

	cmdBuildExample = `authelia-scripts build`

	cmdCleanShort = "Clean build artifacts"

	cmdCleanLong = `Clean build artifacts.`

	cmdCleanExample = `authelia-scripts clean`

	cmdCIShort = "Run the continuous integration script"

	cmdCILong = `Run the continuous integration script.`

	cmdCIExample = `authelia-scripts ci`

	cmdDockerShort = "Commands related to building and publishing docker image"

	cmdDockerLong = `Commands related to building and publishing docker image.`

	cmdDockerExample = `authelia-scripts docker`

	cmdDockerBuildShort = "Build the docker image of Authelia"

	cmdDockerBuildLong = `Build the docker image of Authelia.`

	cmdDockerBuildExample = `authelia-scripts docker build`

	cmdDockerPushManifestShort = "Push Authelia docker manifest to the Docker registries"

	cmdDockerPushManifestLong = `Push Authelia docker manifest to the Docker registries.`

	cmdDockerPushManifestExample = `authelia-scripts docker push-manifest`

	cmdServeShort = "Serve compiled version of Authelia"

	cmdServeLong = `Serve compiled version of Authelia.`

	cmdServeExample = `authelia-scripts serve test.yml`

	cmdSuitesShort = "Commands related to suites management"

	cmdSuitesLong = `Commands related to suites management.`

	cmdSuitesExample = `authelia-scripts suites`

	cmdSuitesListShort = "List available suites"

	cmdSuitesListLong = `List available suites.

Suites can be ran with the authelia-scripts suites test [suite] command.`

	cmdSuitesListExample = `authelia-scripts suites list`

	cmdSuitesTestShort = "Run a test suite"

	cmdSuitesTestLong = `Run a test suite.

Suites can be listed with the authelia-scripts suites list command.`

	cmdSuitesTestExample = `authelia-scripts suites test Standalone`

	cmdSuitesSetupShort = "Setup a test suite environment"

	cmdSuitesSetupLong = `Setup a test suite environment.

Suites can be listed with the authelia-scripts suites list command.`

	cmdSuitesSetupExample = `authelia-scripts suites setup Standalone`

	cmdSuitesTeardownShort = "Teardown a test suite environment"

	cmdSuitesTeardownLong = `Teardown a test suite environment.

Suites can be listed with the authelia-scripts suites list command.`

	cmdSuitesTeardownExample = `authelia-scripts suites setup Standalone`

	cmdUnitTestShort = "Run unit tests"

	cmdUnitTestLong = `Run unit tests.`

	cmdUnitTestExample = `authelia-scripts unittest`

	cmdXFlagsShort = "Generate X LDFlags for building Authelia"

	cmdXFlagsLong = `Generate X LDFlags for building Authelia.`

	cmdXFlagsExample = `authelia-scripts xflags`
)
