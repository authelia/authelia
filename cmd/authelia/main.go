package main

import (
	"errors"
	"os"

	"github.com/authelia/authelia/v4/internal/commands"
	"github.com/authelia/authelia/v4/internal/service"
)

func main() {
	reload := service.IsConfigReloadEnabled()

	cmd := commands.NewRootCmd()

	for {
		if err := cmd.Execute(); err != nil {
			if reload && errors.Is(err, service.ErrConfigReload) {
				continue
			}

			os.Exit(1)
		}

		break
	}
}
