package suites

import (
	"os"
	"time"
)

var traefik2SuiteName = "Traefik2"

var traefik2DockerEnvironment = NewDockerEnvironment([]string{
	"internal/suites/compose.yml",
	"internal/suites/Traefik/compose.yml",
	"internal/suites/example/compose/authelia/compose.backend.{}.yml",
	"internal/suites/example/compose/authelia/compose.frontend.{}.yml",
	"internal/suites/example/compose/redis/compose.yml",
	"internal/suites/example/compose/nginx/backend/compose.yml",
	"internal/suites/example/compose/traefik/compose.yml",
	"internal/suites/example/compose/traefik/compose.v2.yml",
	"internal/suites/example/compose/smtp/compose.yml",
	"internal/suites/example/compose/httpbin/compose.yml",
})

func init() {
	if os.Getenv("CI") == t {
		traefik2DockerEnvironment = NewDockerEnvironment([]string{
			"internal/suites/compose.yml",
			"internal/suites/Traefik/compose.yml",
			"internal/suites/example/compose/authelia/compose.backend.{}.yml",
			"internal/suites/example/compose/redis/compose.yml",
			"internal/suites/example/compose/nginx/backend/compose.yml",
			"internal/suites/example/compose/traefik/compose.yml",
			"internal/suites/example/compose/traefik/compose.v2.yml",
			"internal/suites/example/compose/smtp/compose.yml",
			"internal/suites/example/compose/httpbin/compose.yml",
		})
	}

	setup := func(suitePath string) (err error) {
		if err = traefik2DockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(traefik2DockerEnvironment, traefik2SuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return traefik2DockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "redis", "traefik")
	}

	teardown := func(suitePath string) error {
		err := traefik2DockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(traefik2SuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
