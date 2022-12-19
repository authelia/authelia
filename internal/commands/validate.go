package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newValidateConfigCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "validate-config",
		Short:   cmdAutheliaValidateConfigShort,
		Long:    cmdAutheliaValidateConfigLong,
		Example: cmdAutheliaValidateConfigExample,
		Args:    cobra.NoArgs,
		PreRunE: ctx.ChainRunE(
			ctx.ConfigLoadRunE,
			ctx.ConfigValidateKeysRunE,
			ctx.ConfigValidateRunE,
		),
		RunE: ctx.ValidateConfigRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

// ValidateConfigRunE is the RunE for the authelia validate-config command.
func (ctx *CmdCtx) ValidateConfigRunE(_ *cobra.Command, _ []string) (err error) {
	switch {
	case ctx.cconfig.validator.HasErrors():
		fmt.Println("Configuration parsed and loaded with errors:")
		fmt.Println("")

		for _, err = range ctx.cconfig.validator.Errors() {
			fmt.Printf("\t - %v\n", err)
		}

		fmt.Println("")

		if !ctx.cconfig.validator.HasWarnings() {
			break
		}

		fallthrough
	case ctx.cconfig.validator.HasWarnings():
		fmt.Println("Configuration parsed and loaded with warnings:")
		fmt.Println("")

		for _, err = range ctx.cconfig.validator.Warnings() {
			fmt.Printf("\t - %v\n", err)
		}

		fmt.Println("")
	default:
		fmt.Println("Configuration parsed and loaded successfully without errors.")
		fmt.Println("")
	}

	return nil
}
