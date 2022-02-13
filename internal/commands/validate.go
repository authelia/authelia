package commands

import (
	"log"

	"github.com/spf13/cobra"
)

func newValidateConfigCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:    "validate-config",
		Short:  "Check a configuration against the internal configuration validation mechanisms",
		Args:   cobra.NoArgs,
		PreRun: newCmdWithConfigPreRun(false, true, true),
		Run:    cmdValidateConfigRun,
	}

	cmdWithConfigFlags(cmd, false, []string{"config.yml"})

	return cmd
}

func cmdValidateConfigRun(_ *cobra.Command, _ []string) {
	log.Println("Configuration parsed successfully without errors.")
}
