package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/server"
)

var configPathFlag string

func startServer() {
	logger := logging.Logger()
	config, errs := configuration.Read(configPathFlag)

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(1)
	}

	providers, nonFatalErrs, errs := configureProviders(config)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(2)
	}

	if len(nonFatalErrs) > 0 {
		for _, err := range nonFatalErrs {
			logger.Warn(err)
		}
	}

	if err := logging.InitializeLogger(config.Logging.Format, config.Logging.FilePath, config.Logging.KeepStdout); err != nil {
		logger.Fatalf("Cannot initialize logger: %v", err)
	}

	switch config.Logging.Level {
	case "error":
		logger.Info("Logging severity set to error")
		logging.SetLevel(logrus.ErrorLevel)
	case "warn":
		logger.Info("Logging severity set to warn")
		logging.SetLevel(logrus.WarnLevel)
	case "info":
		logger.Info("Logging severity set to info")
		logging.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.Info("Logging severity set to debug")
		logging.SetLevel(logrus.DebugLevel)
	case "trace":
		logger.Info("Logging severity set to trace")
		logging.SetLevel(logrus.TraceLevel)
	}

	if os.Getenv("ENVIRONMENT") == "dev" {
		logger.Info("===> Authelia is running in development mode. <===")
	}

	server.StartServer(*config, providers)
}

func main() {
	logger := logging.Logger()
	rootCmd := &cobra.Command{
		Use: "authelia",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}

	rootCmd.Flags().StringVar(&configPathFlag, "config", "", "Configuration file")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of Authelia",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Authelia version %s, build %s\n", BuildTag, BuildCommit)
		},
	}

	rootCmd.AddCommand(versionCmd, commands.HashPasswordCmd,
		commands.ValidateConfigCmd, commands.CertificatesCmd,
		commands.RSACmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}
