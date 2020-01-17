package suites

import (
	"fmt"
	"time"
)

var highAvailabilitySuiteName = "HighAvailability"

var haDockerEnvironment = NewDockerEnvironment([]string{
	"docker-compose.yml",
	"internal/suites/HighAvailability/docker-compose.yml",
	"example/compose/authelia/docker-compose.backend.{}.yml",
	"example/compose/authelia/docker-compose.frontend.{}.yml",
	"example/compose/mariadb/docker-compose.yml",
	"example/compose/redis/docker-compose.yml",
	"example/compose/nginx/backend/docker-compose.yml",
	"example/compose/nginx/portal/docker-compose.yml",
	"example/compose/smtp/docker-compose.yml",
	"example/compose/httpbin/docker-compose.yml",
	"example/compose/ldap/docker-compose.admin.yml", // This is just used for administration, not for testing.
	"example/compose/ldap/docker-compose.yml",
})

func init() {
	setup := func(suitePath string) error {
		if err := haDockerEnvironment.Up(); err != nil {
			return err
		}

		return waitUntilAutheliaBackendIsReady(haDockerEnvironment)
	}

	onSetupTimeout := func() error {
		backendLogs, err := haDockerEnvironment.Logs("authelia-backend", nil)
		if err != nil {
			return err
		}
		fmt.Println(backendLogs)

		frontendLogs, err := haDockerEnvironment.Logs("authelia-frontend", nil)
		if err != nil {
			return err
		}
		fmt.Println(frontendLogs)
		return nil
	}

	teardown := func(suitePath string) error {
		return haDockerEnvironment.Down()
	}

	GlobalRegistry.Register(highAvailabilitySuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  onSetupTimeout,
		TestTimeout:     5 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite is made to test Authelia in a *complete*
environment, that is, with all components making Authelia highly available.`,
	})
}
