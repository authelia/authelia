package main

import (
	"errors"
	"os"

	"github.com/authelia/authelia/v4/internal/commands"
	"github.com/authelia/authelia/v4/internal/service"
)

func main() {
	cmd := commands.NewRootCmd()

	for {
		reload := service.IsConfigFileWatcherEnabled()

		if err := cmd.Execute(); err != nil {
			if reload && errors.Is(err, service.ErrApplicationReload) {
				continue
			}

			os.Exit(1)
		}

		break
	}
}
