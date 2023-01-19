package suites

import (
	"fmt"
	"time"
)

var postgresSuiteName = "Postgres"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/Postgres/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
		"internal/suites/example/compose/postgres/docker-compose.yml",
		"internal/suites/example/compose/ldap/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		err := dockerEnvironment.Up()
		if err != nil {
			return err
		}

		err = waitUntilAutheliaIsReady(dockerEnvironment, postgresSuiteName)
		if err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		backendLogs, err := dockerEnvironment.Logs("authelia-backend", nil)
		if err != nil {
			return err
		}

		fmt.Println(backendLogs)

		frontendLogs, err := dockerEnvironment.Logs("authelia-frontend", nil)
		if err != nil {
			return err
		}

		fmt.Println(frontendLogs)

		return nil
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
