package commands

import (
	"fmt"
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
			fmt.Printf("Error Loading Configuration: %s\n", err)
			os.Exit(2)
		}

		// TODO: Actually use the configuration to validate some providers like Notifier
		_, errs := configuration.Read(configPath)
		if len(errs) != 0 {
			str := "Errors"
			if len(errs) == 1 {
				str = "Error"
			}
			fmt.Printf("%s occurred parsing configuration:\n", str)
			for _, err := range errs {
				fmt.Printf("\t%s\n", err)
			}
			os.Exit(1)
		} else {
			fmt.Printf("Configuration parsed successfully without errors.\n")
			os.Exit(0)
		}
	},
	Args: cobra.MinimumNArgs(1),
}
