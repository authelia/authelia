package suites

import (
	"os"
	"time"
)

var haproxySuiteName = "HAProxy"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		composePathSuiteHAProxy,
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathHAProxy,
		composePathSMTP,
		composePathHTTPBin,
	})

	if os.Getenv("CI") == t {
		dockerEnvironment = NewDockerEnvironment([]string{
			composePathBase,
			composePathSuiteHAProxy,
			composePathAutheliaBackend,
			composePathNginxBackend,
			composePathHAProxy,
			composePathSMTP,
			composePathHTTPBin,
		})
	}

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, haproxySuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "haproxy")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(haproxySuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
