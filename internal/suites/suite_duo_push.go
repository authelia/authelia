package suites

import (
	"time"
)

var duoPushSuiteName = "DuoPush"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"docker-compose.yml",
		"example/compose/authelia/docker-compose.backend.yml",
		"example/compose/authelia/docker-compose.frontend.yml",
		"example/compose/nginx/backend/docker-compose.yml",
		"example/compose/nginx/portal/docker-compose.yml",
		"example/compose/duo-api/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		if err := dockerEnvironment.Up(suitePath); err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(dockerEnvironment)
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down(suitePath)
	}

	GlobalRegistry.Register(duoPushSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		TestTimeout:     1 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,

		Description: `This suite has been created to test Authelia against
the Duo API for push notifications. It allows a user to validate second factor
with a mobile phone.`,
	})
}
