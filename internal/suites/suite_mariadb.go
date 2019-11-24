package suites

import (
	"time"
)

var mariadbSuiteName = "Mariadb"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"docker-compose.yml",
		"internal/suites/Mariadb/docker-compose.yml",
		"example/compose/authelia/docker-compose.backend.yml",
		"example/compose/authelia/docker-compose.frontend.yml",
		"example/compose/nginx/backend/docker-compose.yml",
		"example/compose/nginx/portal/docker-compose.yml",
		"example/compose/smtp/docker-compose.yml",
		"example/compose/mariadb/docker-compose.yml",
		"example/compose/ldap/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		if err := dockerEnvironment.Up(); err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(dockerEnvironment)
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(mariadbSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    3 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
