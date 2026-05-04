package suites

import (
	"os"
	"time"
)

var pathPrefixSuiteName = "PathPrefix"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		composePathSuitePathPrefix,
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathTraefik,
		composePathTraefikV3,
		composePathSMTP,
		composePathHTTPBin,
	})

	if os.Getenv("CI") == t {
		dockerEnvironment = NewDockerEnvironment([]string{
			composePathBase,
			composePathSuitePathPrefix,
			composePathAutheliaBackend,
			composePathNginxBackend,
			composePathTraefik,
			composePathTraefikV3,
			composePathSMTP,
			composePathHTTPBin,
		})
	}

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, pathPrefixSuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(pathPrefixSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
