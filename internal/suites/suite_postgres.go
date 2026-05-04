package suites

import (
	"time"
)

var postgresSuiteName = "Postgres"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		"internal/suites/Postgres/compose.yml",
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathNginxPortal,
		composePathSMTP,
		"internal/suites/example/compose/postgres/compose.yml",
		composePathLDAP,
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, postgresSuiteName); err != nil {
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

	GlobalRegistry.Register(postgresSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
