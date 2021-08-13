package main

import (
	"github.com/authelia/authelia/v4/internal/commands"
	"github.com/authelia/authelia/v4/internal/logging"
)

func main() {
	logger := logging.Logger()

	if err := commands.NewRootCmd().Execute(); err != nil {
		logger.Fatal(err)
	}
}
