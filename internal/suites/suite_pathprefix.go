package suites

import (
	"fmt"
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
		"internal/suites/example/compose/traefik2/docker-compose.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
		"internal/suites/example/compose/httpbin/docker-compose.yml",
	})

	if os.Getenv("CI") == t {
		dockerEnvironment = NewDockerEnvironment([]string{
			"internal/suites/docker-compose.yml",
			"internal/suites/PathPrefix/docker-compose.yml",
			"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
			"internal/suites/example/compose/nginx/backend/docker-compose.yml",
			"internal/suites/example/compose/traefik2/docker-compose.yml",
			"internal/suites/example/compose/smtp/docker-compose.yml",
			"internal/suites/example/compose/httpbin/docker-compose.yml",
		})
	}

	setup := func(suitePath string) error {
		if err := dockerEnvironment.Up(); err != nil {
			return err
		}

		err := waitUntilAutheliaIsReady(dockerEnvironment, pathPrefixSuiteName)
		if err != nil {
			return err
		}

		err = updateDevEnvFileForDomain(BaseDomain)
		if err != nil {
			return err
		}

		return nil
	}

	displayAutheliaLogs := func() error {
		backendLogs, err := dockerEnvironment.Logs("authelia-backend", nil)
		if err != nil {
			return err
		}

		fmt.Println(backendLogs)

		if os.Getenv("CI") != t {
			frontendLogs, err := dockerEnvironment.Logs("authelia-frontend", nil)
			if err != nil {
				return err
			}

			fmt.Println(frontendLogs)
		}

		return nil
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
