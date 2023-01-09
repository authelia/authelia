package suites

import (
	"fmt"
	"os"
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
	_ = os.MkdirAll("/tmp/authelia/MultiCookieDomainSuite/", 0700)
	_ = os.WriteFile("/tmp/authelia/MultiCookieDomainSuite/jwt", []byte("very_important_secret"), 0600)       //nolint:gosec
	_ = os.WriteFile("/tmp/authelia/MultiCookieDomainSuite/session", []byte("unsecure_session_secret"), 0600) //nolint:gosec

	if os.Getenv("CI") == t {
		multiCookieDomainDockerEnvironment = NewDockerEnvironment([]string{
			"internal/suites/docker-compose.yml",
			"internal/suites/MultiCookieDomain/docker-compose.yml",
			"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
			"internal/suites/example/compose/nginx/backend/docker-compose.yml",
			"internal/suites/example/compose/nginx/portal/docker-compose.yml",
			"internal/suites/example/compose/smtp/docker-compose.yml",
		})
	}

	setup := func(suitePath string) error {
		err := multiCookieDomainDockerEnvironment.Up()
		if err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(multiCookieDomainDockerEnvironment, multiCookieDomainSuiteName)
	}

	displayAutheliaLogs := func() error {
		backendLogs, err := multiCookieDomainDockerEnvironment.Logs("authelia-backend", nil)
		if err != nil {
			return err
		}

		fmt.Println(backendLogs)

		frontendLogs, err := multiCookieDomainDockerEnvironment.Logs("authelia-frontend", nil)
		if err != nil {
			return err
		}

		fmt.Println(frontendLogs)

		return nil
	}

	teardown := func(suitePath string) error {
		err := multiCookieDomainDockerEnvironment.Down()
		_ = os.Remove("/tmp/db.sqlite3")

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
