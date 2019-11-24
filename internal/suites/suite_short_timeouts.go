package suites

import (
	"time"
)

var shortTimeoutsSuiteName = "ShortTimeouts"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"docker-compose.yml",
		"internal/suites/ShortTimeouts/docker-compose.yml",
		"example/compose/authelia/docker-compose.backend.yml",
		"example/compose/authelia/docker-compose.frontend.yml",
		"example/compose/nginx/backend/docker-compose.yml",
		"example/compose/nginx/portal/docker-compose.yml",
		"example/compose/smtp/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		if err := dockerEnvironment.Up(); err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(dockerEnvironment)
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down()
	}

	GlobalRegistry.Register(shortTimeoutsSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		TestTimeout:     1 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
		Description: `This suite has been created to configure Authelia with short timeouts for sessions expiration
in order to test the inactivity feature and the remember me feature.`,
	})
}
