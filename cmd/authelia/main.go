// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Command authelia is the Authelia daemon which provides authentication and authorization for your applications.
package main

import (
	"errors"
	"os"

	"github.com/authelia/authelia/v4/internal/commands"
	"github.com/authelia/authelia/v4/internal/service"
)

func main() {
	for {
		// A new root command is built for every iteration as the reload must not reuse any state from the previous
		// run, most notably the loaded configuration and the provisioned providers.
		if err := commands.NewRootCmd().Execute(); err != nil {
			switch {
			case errors.Is(err, service.ErrApplicationReload):
				continue
			case errors.Is(err, commands.ErrConfigCreated):
				os.Exit(0)
			default:
				os.Exit(1)
			}
		}

		return
	}
}
