package suites

import (
	"fmt"
	"time"
)

var standaloneSuiteName = "Standalone"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/Standalone/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.https.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		err := dockerEnvironment.Up()

		if err != nil {
			return err
		}

		return waitUntilAutheliaBackendIsReady(dockerEnvironment)
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

	GlobalRegistry.Register(standaloneSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnError:         displayAutheliaLogs,
		OnSetupTimeout:  displayAutheliaLogs,
		TearDown:        teardown,
		TestTimeout:     3 * time.Minute,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite is used to test Authelia in a standalone
configuration with in-memory sessions and a local sqlite db stored on disk`,
	})
}
