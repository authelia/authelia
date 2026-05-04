package suites

import (
	"os"
	"time"
)

var envoySuiteName = "Envoy"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		composePathSuiteEnvoy,
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathEnvoy,
		composePathSMTP,
		composePathHTTPBin,
	})

	if os.Getenv("CI") == t {
		dockerEnvironment = NewDockerEnvironment([]string{
			composePathBase,
			composePathSuiteEnvoy,
			composePathAutheliaBackend,
			composePathNginxBackend,
			composePathEnvoy,
			composePathSMTP,
			composePathHTTPBin,
		})
	}

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, envoySuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "envoy")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(envoySuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
