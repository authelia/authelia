package suites

import (
	"time"
)

var ldapSuiteName = "LDAP"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/LDAP/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
		"internal/suites/example/compose/ldap/docker-compose.yml",
		"internal/suites/example/compose/ldap/docker-compose.admin.yml",
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, ldapSuiteName); err != nil {
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

	GlobalRegistry.Register(ldapSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		TestTimeout:     120 * time.Second,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		OnError:         displayAutheliaLogs,
	})
}
