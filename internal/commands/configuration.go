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

func newCmdWithConfigPreRun(ensureConfigExists bool) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		logger := logging.Logger()

		configs, err := cmd.Root().Flags().GetStringSlice("config")
		if err != nil {
			logger.Fatalf("Error reading flags: %v", err)
		}

		if ensureConfigExists && len(configs) == 1 {
			created, err := configuration.EnsureConfigurationExists(configs[0])
			if err != nil {
				logger.Fatal(err)
			}

			if created {
				logger.Warnf("configuration did not exist so a default one has been generated at %s, you will need to configure this", configs[0])
			}
		}

		provider := configuration.GetProvider()

		errs := provider.LoadSources(configuration.NewDefaultSources(configs)...)

		if len(errs) != 0 {
			logger.Error("Error loading configuration sources:")

			for _, err := range errs {
				logger.Errorf("  %+v", err)
			}

			logger.Fatalf("Can't continue due to the errors loading the configuration sources")
		}

		warns, errs := provider.Unmarshal()

		if len(warns) != 0 {
			logger.Warnf("Warnings occurred while validating configuration:")

			for _, warn := range warns {
				logger.Warnf("  %v", warn)
			}
		}

		if len(errs) != 0 {
			logger.Errorf("Errors occurred while validating configuration:")

			for _, err := range errs {
				logger.Errorf("  %v", err)
			}

			logger.Fatalf("Exiting due to configuration validation errors above.")
		}

		provider.Validator.Clear()
	}
}
