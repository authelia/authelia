package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
)

// ValidateConfigCmd uses the internal configuration reader to validate the configuration.
var ValidateConfigCmd = &cobra.Command{
	Use:   "validate-config [yaml]",
	Short: "Check a configuration against the internal configuration validation mechanisms.",
	Run: func(cobraCmd *cobra.Command, args []string) {
		configPath := args[0]
		if _, err := os.Stat(configPath); err != nil {
			log.Fatalf("Error Loading Configuration: %s\n", err)
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
			log.Fatalf("%s occurred parsing configuration:\n%s", str, errors)
		} else {
			log.Println("Configuration parsed successfully without errors.")
		}
	},
	Args: cobra.MinimumNArgs(1),
}
