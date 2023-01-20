package suites

import (
	"fmt"
	"os"
	"time"
)

var traefik2SuiteName = "Traefik2"

var traefik2DockerEnvironment = NewDockerEnvironment([]string{
	"internal/suites/docker-compose.yml",
	"internal/suites/Traefik2/docker-compose.yml",
	"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
	"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
	"internal/suites/example/compose/redis/docker-compose.yml",
	"internal/suites/example/compose/nginx/backend/docker-compose.yml",
	"internal/suites/example/compose/traefik2/docker-compose.yml",
	"internal/suites/example/compose/smtp/docker-compose.yml",
	"internal/suites/example/compose/httpbin/docker-compose.yml",
})

func init() {
	if os.Getenv("CI") == t {
		traefik2DockerEnvironment = NewDockerEnvironment([]string{
			"internal/suites/docker-compose.yml",
			"internal/suites/Traefik2/docker-compose.yml",
			"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
			"internal/suites/example/compose/redis/docker-compose.yml",
			"internal/suites/example/compose/nginx/backend/docker-compose.yml",
			"internal/suites/example/compose/traefik2/docker-compose.yml",
			"internal/suites/example/compose/smtp/docker-compose.yml",
			"internal/suites/example/compose/httpbin/docker-compose.yml",
		})
	}

	setup := func(suitePath string) error {
		err := traefik2DockerEnvironment.Up()
		if err != nil {
			return err
		}

		err = waitUntilAutheliaIsReady(traefik2DockerEnvironment, traefik2SuiteName)
		if err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		backendLogs, err := traefik2DockerEnvironment.Logs("authelia-backend", nil)
		if err != nil {
			return err
		}

		fmt.Println(backendLogs)

		if os.Getenv("CI") != t {
			frontendLogs, err := traefik2DockerEnvironment.Logs("authelia-frontend", nil)
			if err != nil {
				return err
			}

			fmt.Println(frontendLogs)
		}

		redisLogs, err := traefik2DockerEnvironment.Logs("redis", nil)
		if err != nil {
			return err
		}

		fmt.Println(redisLogs)

		traefikLogs, err := traefik2DockerEnvironment.Logs("traefik", nil)
		if err != nil {
			return err
		}

		fmt.Println(traefikLogs)

		return nil
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
