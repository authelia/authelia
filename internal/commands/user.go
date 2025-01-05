package commands

import (
	"github.com/spf13/cobra"
)

func newUserCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "user",
		Short:   cmdAutheliaUserShort,
		Long:    cmdAutheliaUserLong,
		Example: cmdAutheliaUserExample,
		PersistentPreRunE: ctx.ChainRunE(
			ctx.ConfigStorageCommandLineConfigRunE,
			ctx.HelperConfigLoadRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
			ctx.ConfigValidateStorageRunE,
			ctx.LoadProvidersStorageRunE,
			ctx.LoadProvidersAuthenticationRunE,
		),
		Args: cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newUserPasswordCmd(ctx),
		newUserShowCmd(ctx),
	)

	return cmd
}

func newUserPasswordCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "password",
		Short:             cmdAutheliaUserPasswordShort,
		Long:              cmdAutheliaUserPasswordLong,
		Example:           cmdAutheliaUserPasswordExample,
		Args:              cobra.MinimumNArgs(2),
		ArgAliases:        []string{"username", "password"},
		RunE:              ctx.UserChangePasswordRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUserShowCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "show",
		Short:             cmdAutheliaUserShowShort,
		Long:              cmdAutheliaUserShowLong,
		Example:           cmdAutheliaUserShowExample,
		Args:              cobra.MinimumNArgs(1),
		ArgAliases:        []string{"username"},
		RunE:              ctx.UserShowInfoRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}
