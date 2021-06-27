package commands

import (
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
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

		if ensureConfigExists {
			created, err := configuration.EnsureConfigurationExists(configs)
			if err != nil {
				logger.Fatal(err)
			}

			if created {
				logger.Warnf("configuration did not exist so a default one has been generated at %s, you will need to configure this", configs[0])
			}
		}

		provider := configuration.GetProvider()

		err = provider.LoadSources(configuration.NewDefaultSources(configs, provider)...)
		if err != nil {
			logger.Errorf("Error loading configuration sources: %v", err)

			for _, err := range provider.Errors() {
				logger.Errorf("%+v", err)
			}
		}

		err = provider.UnmarshalToConfiguration()
		if err != nil {
			logger.Fatalf("Error unmarshalling configuration: %v", err)
		}

		s := schema.NewStructValidator()
		validator.ValidateKeys(s, provider.Keys())
		validator.ValidateConfiguration(provider.Configuration(), s)

		warns := s.Warnings()
		if len(warns) != 0 {
			logger.Warnf("Warnings occurred while validating configuration:")

			for _, warn := range warns {
				logger.Warnf("  %v", warn)
			}
		}

		errs := s.Errors()
		if len(errs) != 0 {
			logger.Errorf("Errors occurred while validating configuration:")

			for _, err := range errs {
				logger.Errorf("  %v", err)
			}

			logger.Fatalf("Exiting due to configuration validation errors above.")
		}

		provider.Clear()
	}
}
