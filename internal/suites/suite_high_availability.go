package suites

import (
	"time"
)

var highAvailabilitySuiteName = "HighAvailability"

var haDockerEnvironment = NewDockerEnvironment([]string{
	"internal/suites/docker-compose.yml",
	"internal/suites/HighAvailability/docker-compose.yml",
	"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
	"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
	"internal/suites/example/compose/mariadb/docker-compose.yml",
	"internal/suites/example/compose/redis-sentinel/docker-compose.yml",
	"internal/suites/example/compose/nginx/backend/docker-compose.yml",
	"internal/suites/example/compose/nginx/portal/docker-compose.yml",
	"internal/suites/example/compose/smtp/docker-compose.yml",
	"internal/suites/example/compose/httpbin/docker-compose.yml",
	"internal/suites/example/compose/ldap/docker-compose.admin.yml", // This is just used for administration, not for testing.
	"internal/suites/example/compose/ldap/docker-compose.yml",
})

func init() {
	setup := func(suitePath string) (err error) {
		if err = haDockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(haDockerEnvironment, highAvailabilitySuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return haDockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		return haDockerEnvironment.Down()
	}

	GlobalRegistry.Register(highAvailabilitySuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		TestTimeout:     6 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		OnError:         displayAutheliaLogs,
		Description: `This suite is made to test Authelia in a *complete*
environment, that is, with all components making Authelia highly available.`,
	})
}
