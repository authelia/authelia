package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
)

func newValidateConfigCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "validate-config [yaml]",
		Short: "Check a configuration against the internal configuration validation mechanisms",
		Args:  cobra.MinimumNArgs(1),
		RunE:  cmdValidateConfigRunE,
	}

	return cmd
}

func cmdValidateConfigRunE(_ *cobra.Command, args []string) (err error) {
	configPath := args[0]
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("Error Loading Configuration: %w\n", err)
	}

	// TODO: Actually use the configuration to validate some providers like Notifier
	_, errs := configuration.Read(configPath)
	if len(errs) != 0 {
		str := "Errors"
		if len(errs) == 1 {
			str = "Error"
		}

		errors := ""
		for _, err := range errs {
			errors += fmt.Sprintf("\t%s\n", err.Error())
		}

		return fmt.Errorf("%s occurred parsing configuration:\n%s", str, errors)
	}

	log.Println("Configuration parsed successfully without errors.")

	return nil
}
