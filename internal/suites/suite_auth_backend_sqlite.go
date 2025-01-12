package suites

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
)

var authBackendSqliteSuiteName = "AuthBackendSqlite"

func createDemoUsers(userProvider *authentication.DBUserProvider) (err error) {
	if err = userProvider.StartupCheck(); err != nil {
		return err
	}

	userList := []struct {
		Username    string
		Email       string
		DisplayName string
		Groups      []string
		Password    string
	}{
		{
			Username:    "john",
			Email:       "john.doe@authelia.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Password:    "password",
		},
		{
			Username:    "harry",
			Email:       "harry.potter@authelia.com",
			DisplayName: "Harry Potter",
			Password:    "password",
		},
		{
			Username:    "bob",
			Email:       "bob.dylan@authelia.com",
			DisplayName: "Bob Dylan",
			Groups:      []string{"dev"},
			Password:    "password",
		},
		{
			Username:    "james",
			Email:       "james.dean@authelia.com",
			DisplayName: "James Dean",
			Password:    "password",
		},
	}

	var errors []error

	for _, user := range userList {
		if err := userProvider.AddUser(
			user.Username,
			user.DisplayName,
			user.Password,
			authentication.WithEmail(user.Email),
			authentication.WithGroups(user.Groups)); err != nil {
			errors = append(errors, fmt.Errorf("\n   user %s: %w", user.Username, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to create demo users:%v", errors)
	}

	return nil
}

func init() {
	_ = os.MkdirAll("/tmp/authelia/AuthenticationBackendSQLiteSuite/", 0700)
	_ = os.WriteFile("/tmp/authelia/AuthenticationBackendSQLiteSuite/jwt", []byte("very_important_secret"), 0600)       //nolint:gosec
	_ = os.WriteFile("/tmp/authelia/AuthenticationBackendSQLiteSuite/session", []byte("unsecure_session_secret"), 0600) //nolint:gosec

	composeFiles := defaultComposeFiles
	composeFiles = append(composeFiles,
		"internal/suites/AuthenticationBackendSQLite/docker-compose.yml",
	)

	dockerEnvironment := NewDockerEnvironment(composeFiles)

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, authBackendSqliteSuiteName); err != nil {
			return err
		}

		storageProvider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
		userProvider := authentication.NewDBUserProvider(&schema.DefaultDBAuthenticationBackendConfig, storageProvider)

		if err = createDemoUsers(userProvider); err != nil {
			log.Warn(err)
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
