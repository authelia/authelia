package suites

import (
	"os"
	"time"
)

var securitySettingsSuiteName = "SecuritySettings"

func init() {
	_ = os.MkdirAll("/tmp/authelia/SecuritySettings/Suite/", 0700)
	_ = os.WriteFile("/tmp/authelia/SecuritySettings/Suite/jwt", []byte("very_important_secret"), 0600)       //nolint:gosec
	_ = os.WriteFile("/tmp/authelia/SecuritySettings/Suite/session", []byte("unsecure_session_secret"), 0600) //nolint:gosec

	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/SecuritySettings/docker-compose.yml",
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

		if err = waitUntilAutheliaIsReady(dockerEnvironment, securitySettingsSuiteName); err != nil {
			return err
		}

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

	GlobalRegistry.Register(securitySettingsSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnError:         displayAutheliaLogs,
		OnSetupTimeout:  displayAutheliaLogs,
		TearDown:        teardown,
		TestTimeout:     4 * time.Minute,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite is used to test Authelia in a password change
		configuration with in-memory sessions and a local sqlite db stored on disk`,
	})
}
