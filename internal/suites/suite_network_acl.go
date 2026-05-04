package suites

import (
	"time"
)

var networkACLSuiteName = "NetworkACL"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		composePathBase,
		"internal/suites/NetworkACL/compose.yml",
		composePathAutheliaBackend,
		composePathAutheliaFrontend,
		composePathNginxBackend,
		composePathNginxPortal,
		"internal/suites/example/compose/squid/compose.yml",
		composePathSMTP,
		// To debug headers.
		composePathHTTPBin,
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, networkACLSuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down()
	}

	GlobalRegistry.Register(networkACLSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     1 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite has been created to test Authelia with basic feature in a non highly-available setup.
Authelia basically use an in-memory cache to store user sessions and persist data on disk instead
of using a remote database. Also, the user accounts are stored in file-based database.`,
	})
}
