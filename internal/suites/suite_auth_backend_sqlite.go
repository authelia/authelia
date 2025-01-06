package suites

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

var authBackendSqliteSuiteName = "AuthBackendSqlite"

func init() {
	_ = os.MkdirAll("/tmp/authelia/AuthenticationBackendSQLiteSuite/", 0700)
	_ = os.WriteFile("/tmp/authelia/AuthenticationBackendSQLiteSuite/jwt", []byte("very_important_secret"), 0600)       //nolint:gosec
	_ = os.WriteFile("/tmp/authelia/AuthenticationBackendSQLiteSuite/session", []byte("unsecure_session_secret"), 0600) //nolint:gosec

	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/AuthenticationBackendSQLite/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/docker-compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/docker-compose.yml",
		"internal/suites/example/compose/nginx/portal/docker-compose.yml",
		"internal/suites/example/compose/smtp/docker-compose.yml",
	})

	createUser := func(provider *storage.SQLiteProvider, user model.User, groups ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer func() {
			cancel()
		}()

		if err := provider.CreateUser(ctx, user); err != nil {
			return err
		}

		if err := provider.AssignGroupsToUser(ctx, user.Username, groups...); err != nil {
			return err
		}

		return nil
	}

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, authBackendSqliteSuiteName); err != nil {
			return err
		}

		provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)

		var userJohn = model.User{
			Username:    "john",
			Email:       "john.doe@authelia.com",
			DisplayName: "John Doe",
			Password:    []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"),
		}

		if err = createUser(provider, userJohn, "admins", "dev"); err != nil {
			log.Warnf("could not create user %s (%s).\n", userJohn.Username, err)
		}

		var userHarry = model.User{
			Username:    "harry",
			Email:       "harry.potter@authelia.com",
			DisplayName: "Harry Potter",
			Password:    []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"),
		}

		if err = createUser(provider, userHarry); err != nil {
			log.Warnf("could not create user %s (%s).\n", userHarry.Username, err)
		}

		var userBob = model.User{
			Username:    "bob",
			Email:       "bob.dylan@authelia.com",
			DisplayName: "Bob Dylan",
			Password:    []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"),
		}

		if err = createUser(provider, userBob, "dev"); err != nil {
			log.Warnf("could not create user %s (%s).\n", userBob.Username, err)
		}

		var userJames = model.User{
			Username:    "james",
			Email:       "james.dean@authelia.com",
			DisplayName: "James Dean",
			Password:    []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"),
		}

		if err = createUser(provider, userJames); err != nil {
			log.Warnf("could not create user %s (%s).\n", userJames.Username, err)
		}

		return updateDevEnvFileForDomain(BaseDomain, true)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend")
	}

	teardown := func(suitePath string) error {
		if err := dockerEnvironment.Down(); err != nil {
			return fmt.Errorf("failed to tear down docker environment: %w", err)
		}

		if err := os.Remove("/tmp/db.sqlite3"); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove SQLite database: %w", err)
		}

		return nil
	}

	GlobalRegistry.Register(authBackendSqliteSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    5 * time.Minute,
		OnError:         displayAutheliaLogs,
		OnSetupTimeout:  displayAutheliaLogs,
		TearDown:        teardown,
		TestTimeout:     4 * time.Minute,
		TearDownTimeout: 2 * time.Minute,
		Description:     `This suite is used to test Authelia with SQLite as authentication backend`,
	})
}
