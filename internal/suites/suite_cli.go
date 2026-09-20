// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"os"
	"syscall"
	"time"
)

var cliSuiteName = "CLI"

const cliSuiteFIFOPath = "/tmp/authelia/CLISuite/notification.fifo"

func init() {
	_ = os.MkdirAll("/tmp/authelia/CLISuite/", 0o700)
	_ = os.Remove(cliSuiteFIFOPath)
	_ = syscall.Mkfifo(cliSuiteFIFOPath, 0o600)

	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/compose.yml",
		"internal/suites/CLI/compose.yml",
		"internal/suites/example/compose/authelia/compose.backend.{}.yml",
		"internal/suites/example/compose/nginx/cli/compose.yml",
	})

	setup := func(suitePath string) (err error) {
		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		return waitUntilAutheliaIsReady(dockerEnvironment, cliSuiteName)
	}

	displayAutheliaLogs := func() error {
		return dockerEnvironment.PrintLogs("authelia-backend")
	}

	teardown := func(suitePath string) error {
		err := dockerEnvironment.Down()
		_ = os.Remove(SuiteTmpPath("db.sqlite3"))
		_ = os.Remove(SuiteTmpPath("db.sqlite"))
		_ = os.RemoveAll(SuiteTmpPath("qr"))
		_ = os.RemoveAll(SuiteTmpPath("out"))
		_ = os.Remove(SuiteTmpPath("qr.png"))
		_ = os.Remove(SuiteTmpPath(cliSuiteFIFOPath))

		return err
	}

	GlobalRegistry.Register(cliSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    2 * time.Minute,
		OnSetupTimeout:  displayAutheliaLogs,
		OnError:         displayAutheliaLogs,
		TestTimeout:     3 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 1 * time.Minute,
		Description: `This suite has been created to test the Authelia command line interface, covering the
configuration validation, password hashing, certificate generation, storage and access control subcommands.`,
	})
}
