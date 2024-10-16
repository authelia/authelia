package suites

import (
	"os"
	"time"
)

var pathPrefixSuiteName = "PathPrefix"

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/PathPrefix/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/traefik/docker-compose.yml",
		"internal/suites/example/compose/traefik/docker-compose.v3.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
		"internal/suites/example/compose/httpbin/docker-compose.yml",
	})

	if os.Getenv("CI") == t {
		dockerEnvironment = NewDockerEnvironment([]string{
			"internal/suites/docker-compose.yml",
			"internal/suites/PathPrefix/docker-compose.yml",
			"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
			"internal/suites/example/compose/nginx/backend/docker-compose.yml",
			"internal/suites/example/compose/traefik/docker-compose.yml",
			"internal/suites/example/compose/traefik/docker-compose.v3.yml",
			"internal/suites/example/compose/smtp/docker-compose.yml",
			"internal/suites/example/compose/httpbin/docker-compose.yml",
		})
	}

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, pathPrefixSuiteName); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		return err
	}

	GlobalRegistry.Register(pathPrefixSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     2 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 2 * time.Minute,
	})
}
