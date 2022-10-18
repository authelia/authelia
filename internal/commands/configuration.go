package commands

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/logging"
)

// cmdWithConfigFlags is used for commands which require access to the configuration to add the flag to the command.
func cmdWithConfigFlags(cmd *cobra.Command, persistent bool, configs []string) {
	if persistent {
		cmd.PersistentFlags().StringSliceP(cmdFlagNameConfig, "c", configs, "configuration files to load")
	} else {
		cmd.Flags().StringSliceP(cmdFlagNameConfig, "c", configs, "configuration files to load")
	}
}

var config *schema.Configuration

func newCmdWithConfigPreRun(ensureConfigExists, validateKeys, validateConfiguration bool) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		var (
			logger  *logrus.Logger
			configs []string
			err     error
		)

		logger = logging.Logger()

		if configs, err = cmd.Flags().GetStringSlice(cmdFlagNameConfig); err != nil {
			logger.Fatalf("Error reading flags: %v", err)
		}

		if ensureConfigExists && len(configs) == 1 {
			created, err := configuration.EnsureConfigurationExists(configs[0])
			if err != nil {
				logger.Fatal(err)
			}

			if created {
				logger.Warnf("Configuration did not exist so a default one has been generated at %s, you will need to configure this", configs[0])
				os.Exit(0)
			}
		}

		var (
			val *schema.StructValidator
		)

		config, val, err = loadConfig(configs, validateKeys, validateConfiguration)
		if err != nil {
			logger.Fatalf("Error occurred loading configuration: %v", err)
		}

		warnings := val.Warnings()
		if len(warnings) != 0 {
			for _, warning := range warnings {
				logger.Warnf("Configuration: %+v", warning)
			}
		}

		errs := val.Errors()
		if len(errs) != 0 {
			for _, err := range errs {
				logger.Errorf("Configuration: %+v", err)
			}

			logger.Fatalf("Can't continue due to the errors loading the configuration")
		}
	}
}

func loadConfig(configs []string, validateKeys, validateConfiguration bool) (c *schema.Configuration, val *schema.StructValidator, err error) {
	var keys []string

	val = schema.NewStructValidator()

	if keys, c, err = configuration.Load(val,
		configuration.NewDefaultSources(
			configs,
			configuration.DefaultEnvPrefix,
			configuration.DefaultEnvDelimiter)...); err != nil {
		return nil, nil, err
	}

	if validateKeys {
		validator.ValidateKeys(keys, configuration.DefaultEnvPrefix, val)
	}

	if validateConfiguration {
		validator.ValidateConfiguration(c, val)
	}

	return c, val, nil
}
