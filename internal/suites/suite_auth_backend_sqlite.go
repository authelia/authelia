package suites

import (
	"os"
	"time"
)

var authBackendSqliteSuiteName = "AuthBackendSqlite"

func init() {
	_ = os.MkdirAll("/tmp/authelia/AuthenticationBackendSQLiteSuite/", 0700)
	_ = os.WriteFile("/tmp/authelia/AuthenticationBackendSQLiteSuite/jwt", []byte("very_important_secret"), 0600)       //nolint:gosec
	_ = os.WriteFile("/tmp/authelia/AuthenticationBackendSQLiteSuite/session", []byte("unsecure_session_secret"), 0600) //nolint:gosec

	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/AuthenticationBackendSQLite/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, authBackendSqliteSuiteName); err != nil {
			return err
		}
		// dockerEnvironment.Exec()

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		_ = os.Remove("/tmp/db.sqlite3")

		return err
	}

	GlobalRegistry.Register(authBackendSqliteSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnError:         displayAutheliaLogs,
		OnSetupTimeout:  displayAutheliaLogs,
		TearDown:        teardown,
		TestTimeout:     4 * time.Minute,
		TearDownTimeout: 2 * time.Minute,
		Description:     `This suite is used to test Authelia with SQLite as authentication backend`,
	})
}
