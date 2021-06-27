package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/logging"
)

func newValidateConfigCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "validate-config [yaml]",
		Short: "Check a configuration against the internal configuration validation mechanisms",
		Args:  cobra.MinimumNArgs(1),
		Run:   cmdValidateConfigRun,
	}

	return cmd
}

func cmdValidateConfigRun(_ *cobra.Command, args []string) {
	logger := logging.Logger()

	configPath := args[0]
	if _, err := os.Stat(configPath); err != nil {
		logger.Fatalf("Error Loading Configuration: %v\n", err)
	}

	provider := configuration.NewProvider()

	err := provider.LoadSources(configuration.NewYAMLFileSource(configPath))
	if err != nil {
		logger.Fatalf("Error loading file configuration: %v", err)
	}

	err = provider.UnmarshalToConfiguration()
	if err != nil {
		logger.Fatalf("Error unmarshalling configuration: %v", err)
	}

	provider.Validate()

	// TODO: Actually use the configuration to validate some providers like Notifier
	errs := provider.Errors()
	if len(errs) != 0 {
		str := "Errors"
		if len(errs) == 1 {
			str = "Error"
		}

		errors := ""
		for _, err := range errs {
			errors += fmt.Sprintf("\t%s\n", err.Error())
		}

		logger.Fatalf("%s occurred parsing configuration:\n%s", str, errors)
	}

	log.Println("Configuration parsed successfully without errors.")
}
