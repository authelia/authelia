package suites

import (
	"os"
	"time"
)

var oidcTraefikSuiteName = "OIDCTraefik"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/OIDCTraefik/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/traefik/docker-compose.yml",
		"internal/suites/example/compose/traefik/docker-compose.v3.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
		"internal/suites/example/compose/oidc-client/docker-compose.yml",
		"internal/suites/example/compose/redis/docker-compose.yml",
	})

	if os.Getenv("CI") == t {
		dockerEnvironment = NewDockerEnvironment([]string{
			"internal/suites/docker-compose.yml",
			"internal/suites/OIDCTraefik/docker-compose.yml",
			"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
			"internal/suites/example/compose/nginx/backend/docker-compose.yml",
			"internal/suites/example/compose/traefik/docker-compose.yml",
			"internal/suites/example/compose/traefik/docker-compose.v3.yml",
			"internal/suites/example/compose/smtp/docker-compose.yml",
			"internal/suites/example/compose/oidc-client/docker-compose.yml",
			"internal/suites/example/compose/redis/docker-compose.yml",
		})
	}

	setup := func(suitePath string) (err error) {
		// TODO(c.michaud): use version in tags for oidc-client but in the meantime we pull the image to make sure it's
		// up to date.
		if err = dockerEnvironment.Pull("oidc-client"); err != nil {
			return err
		}

		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, oidcTraefikSuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "oidc-client")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(oidcTraefikSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
