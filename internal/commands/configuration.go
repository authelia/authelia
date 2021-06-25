package commands

import (
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/logging"
)

// cmdWithConfigFlags is used for commands which require access to the configuration to add the flag to the command.
func cmdWithConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("config", "c", []string{}, "Configuration files")
}

// cmdWithConfigPreRun is used for commands which require access to the configuration to load the configuration in the PreRun.
func cmdWithConfigPreRun(cmd *cobra.Command, _ []string) {
	logger := logging.Logger()

	configs, err := cmd.Root().Flags().GetStringSlice("config")
	if err != nil {
		logger.Fatalf("Error reading flags: %v", err)
	}

	provider := configuration.GetProvider()

	err = provider.LoadPaths(configs)
	if err != nil {
		logger.Fatalf("Error loading file configuration: %v", err)
	}

	err = provider.LoadEnvironment()
	if err != nil {
		logger.Fatalf("Error loading environment configuration: %v", err)
	}

	err = provider.LoadSecrets()
	if err != nil {
		for _, err := range provider.Errors() {
			logger.Errorf("%+v", err)
		}

		logger.Fatalf("Errors loading secrets configuration: %v", err)
	}

	err = provider.UnmarshalToConfiguration()
	if err != nil {
		logger.Fatalf("Error unmarshalling configuration: %v", err)
	}

	provider.Validate()

	warns := provider.Warnings()
	if len(warns) != 0 {
		logger.Warnf("Warnings occurred while validating configuration:")

		for _, warn := range warns {
			logger.Warnf("  %v", warn)
		}
	}

	errs := provider.Errors()
	if len(errs) != 0 {
		logger.Errorf("Errors occurred while validating configuration:")

		for _, err := range errs {
			logger.Errorf("  %v", err)
		}

		logger.Fatalf("Exiting due to configuration validation errors above.")
	}

	provider.Clear()
}
