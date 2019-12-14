package suites

import (
	"fmt"
	"time"
)

var traefikSuiteName = "Traefik"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"docker-compose.yml",
		"internal/suites/Traefik/docker-compose.yml",
		"example/compose/authelia/docker-compose.backend.yml",
		"example/compose/authelia/docker-compose.frontend.yml",
		"example/compose/nginx/backend/docker-compose.yml",
		"example/compose/traefik/docker-compose.yml",
		"example/compose/smtp/docker-compose.yml",
	})

	setup := func(suitePath string) error {
		err := dockerEnvironment.Up()

		if err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(dockerEnvironment)
	}

	onSetupTimeout := func() error {
		backendLogs, err := dockerEnvironment.Logs("authelia-backend", nil)
		if err != nil {
			return err
		}
		fmt.Println(backendLogs)

		frontendLogs, err := dockerEnvironment.Logs("authelia-frontend", nil)
		if err != nil {
			return err
		}
		fmt.Println(frontendLogs)
		return nil
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(traefikSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  onSetupTimeout,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
