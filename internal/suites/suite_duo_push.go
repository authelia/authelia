package suites

import (
	"os"
	"time"
)

var duoPushSuiteName = "DuoPush"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/DuoPush/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.yml",
		"internal/suites/example/compose/duo-api/docker-compose.yml",
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, duoPushSuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "duo-api")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		_ = os.Remove("/tmp/db.sqlite3")

		return err
	}

	GlobalRegistry.Register(duoPushSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     4 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,

		Description: `This suite has been created to test Authelia against
the Duo API for push notifications. It allows a user to validate second factor
with a mobile phone.`,
	})
}
