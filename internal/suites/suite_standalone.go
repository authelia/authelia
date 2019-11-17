package suites

import (
	"time"
)

var standaloneSuiteName = "Standalone"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"docker-compose.yml",
		"example/compose/authelia/docker-compose.backend.yml",
		"example/compose/authelia/docker-compose.frontend.yml",
		"example/compose/nginx/backend/docker-compose.yml",
		"example/compose/nginx/portal/docker-compose.yml",
		"example/compose/smtp/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		err := dockerEnvironment.Up(suitePath)

		if err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(dockerEnvironment)
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down(suitePath)
		return err
	}

	GlobalRegistry.Register(standaloneSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 5 * time.Minute,
		Description: `This suite is used to test Authelia in a standalone
configuration with in-memory sessions and a local sqlite db stored on disk`,
	})
}
