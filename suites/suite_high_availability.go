package suites

import (
	"time"
)

var highAvailabilitySuiteName = "HighAvailability"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"docker-compose.yml",
		"example/compose/authelia/docker-compose.backend.yml",
		"example/compose/authelia/docker-compose.frontend.yml",
		"example/compose/mariadb/docker-compose.yml",
		"example/compose/redis/docker-compose.yml",
		"example/compose/nginx/backend/docker-compose.yml",
		"example/compose/nginx/portal/docker-compose.yml",
		"example/compose/smtp/docker-compose.yml",
		"example/compose/httpbin/docker-compose.yml",
		"example/compose/ldap/docker-compose.admin.yml", // This is just used for administration, not for testing.
		"example/compose/ldap/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		if err := dockerEnvironment.Up(suitePath); err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(dockerEnvironment)
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down(suitePath)
	}

	GlobalRegistry.Register(highAvailabilitySuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		TestTimeout:     1 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite is made to test Authelia in a *complete*
environment, that is, with all components making Authelia highly available.`,
	})
}
