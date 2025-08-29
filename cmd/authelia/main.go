package main

import (
	"errors"
	"os"

	"github.com/authelia/authelia/v4/internal/commands"
)

func main() {
	if err := commands.NewRootCmd().Execute(); err != nil {
		switch {
		case errors.Is(err, commands.ErrConfigCreated):
			os.Exit(0)
		default:
			os.Exit(1)
		}
	}
}
