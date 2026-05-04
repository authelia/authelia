package suites

import (
	"time"
)

var mysqlSuiteName = "MySQL"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		"internal/suites/MySQL/compose.yml",
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathNginxPortal,
		composePathSMTP,
		"internal/suites/example/compose/mysql/compose.yml",
		composePathLDAP,
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, mysqlSuiteName); err != nil {
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

	GlobalRegistry.Register(mysqlSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
