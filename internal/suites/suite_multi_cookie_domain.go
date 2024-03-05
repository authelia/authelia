package suites

import (
	"time"
)

var multiCookieDomainSuiteName = "MultiCookieDomain"

var multiCookieDomainDockerEnvironment = NewDockerEnvironment([]string{
	"internal/suites/docker-compose.yml",
	"internal/suites/MultiCookieDomain/docker-compose.yml",
	"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
	"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
	"internal/suites/example/compose/nginx/backend/docker-compose.yml",
	"internal/suites/example/compose/nginx/portal/docker-compose.yml",
	"internal/suites/example/compose/smtp/docker-compose.yml",
})

func init() {
	setup := func(suitePath string) (err error) {
		if err = multiCookieDomainDockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(multiCookieDomainDockerEnvironment, multiCookieDomainSuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return multiCookieDomainDockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		err := multiCookieDomainDockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(multiCookieDomainSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnError:         displayAutheliaLogs,
		OnSetupTimeout:  displayAutheliaLogs,
		TearDown:        teardown,
		TestTimeout:     4 * time.Minute,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite is used to test Authelia in a multi cookie domain
configuration with in-memory sessions and a local sqlite db stored on disk`,
	})
}
