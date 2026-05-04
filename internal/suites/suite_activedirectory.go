package suites

import (
	"time"
)

var activedirectorySuiteName = "ActiveDirectory"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		"internal/suites/ActiveDirectory/compose.yml",
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathNginxPortal,
		composePathSMTP,
		"internal/suites/example/compose/samba/compose.yml",
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, activedirectorySuiteName); err != nil {
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

	GlobalRegistry.Register(activedirectorySuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		TestTimeout:     120 * time.Second,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		OnError:         displayAutheliaLogs,
	})
}
