package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newValidateConfigCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "validate-config",
		Short:   cmdAutheliaValidateConfigShort,
		Long:    cmdAutheliaValidateConfigLong,
		Example: cmdAutheliaValidateConfigExample,
		Args:    cobra.NoArgs,
		RunE:    cmdValidateConfigRunE,

		DisableAutoGenTag: true,
	}

	cmdWithConfigFlags(cmd, false, []string{"configuration.yml"})

	return cmd
}

func cmdValidateConfigRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		configs []string
		val     *schema.StructValidator
	)

	if configs, err = cmd.Flags().GetStringSlice("config"); err != nil {
		return err
	}

	config, val, err = loadConfig(configs, true, true)
	if err != nil {
		return fmt.Errorf("error occurred loading configuration: %v", err)
	}

	switch {
	case val.HasErrors():
		fmt.Println("Configuration parsed and loaded with errors:")
		fmt.Println("")

		for _, err = range val.Errors() {
			fmt.Printf("\t - %v\n", err)
		}

		fmt.Println("")

		if !val.HasWarnings() {
			break
		}

		fallthrough
	case val.HasWarnings():
		fmt.Println("Configuration parsed and loaded with warnings:")
		fmt.Println("")

		for _, err = range val.Warnings() {
			fmt.Printf("\t - %v\n", err)
		}

		fmt.Println("")
	default:
		fmt.Println("Configuration parsed and loaded successfully without errors.")
		fmt.Println("")
	}

	return nil
}
