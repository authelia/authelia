package suites

import (
	"os"
	"time"
)

var standaloneSuiteName = "Standalone"

func init() {
	_ = os.MkdirAll("/tmp/authelia/StandaloneSuite/", 0700)
	_ = os.WriteFile("/tmp/authelia/StandaloneSuite/jwt", []byte("very_important_secret"), 0600)       //nolint:gosec
	_ = os.WriteFile("/tmp/authelia/StandaloneSuite/session", []byte("unsecure_session_secret"), 0600) //nolint:gosec

	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		"internal/suites/Standalone/compose.yml",
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathNginxPortal,
		composePathSMTP,
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, standaloneSuiteName); err != nil {
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

	GlobalRegistry.Register(standaloneSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnError:         displayAutheliaLogs,
		OnSetupTimeout:  displayAutheliaLogs,
		TearDown:        teardown,
		TestTimeout:     5 * time.Minute,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite is used to test Authelia in a standalone
configuration with in-memory sessions and a local sqlite db stored on disk`,
	})
}
