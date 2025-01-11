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
		newUserAddCmd(ctx),
		newUserDeleteCmd(ctx),
		newUserNameCmd(ctx),
		newUserEmailCmd(ctx),
		newUserGroupsCmd(ctx),
		newListUsersCmd(ctx),
		newUserEnableCmd(ctx),
		newUserDisableCmd(ctx),
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

func newUserAddCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "add",
		Short:             cmdAutheliaUserAddShort,
		Long:              cmdAutheliaUserAddLong,
		Example:           cmdAutheliaUserAddExample,
		Args:              cobra.MinimumNArgs(1),
		ArgAliases:        []string{"username"},
		RunE:              ctx.UserAddRunE,
		DisableAutoGenTag: true,
	}

	cmd.Flags().String("password", "", "the new user's password")
	cmd.Flags().String("display-name", "", "the new user's display name")
	cmd.Flags().String("email", "", "the new user's email")
	cmd.Flags().StringSlice("group", []string{}, "assign the group to the user. ")

	return cmd
}

func newUserDeleteCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "del",
		Short:             cmdAutheliaUserDeleteShort,
		Long:              cmdAutheliaUserDeleteLong,
		Example:           cmdAutheliaUserDeleteExample,
		Args:              cobra.MinimumNArgs(1),
		ArgAliases:        []string{"username"},
		RunE:              ctx.UserDeleteRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUserNameCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "display-name",
		Short:             cmdAutheliaUserDeleteShort,
		Long:              cmdAutheliaUserDeleteLong,
		Example:           cmdAutheliaUserDeleteExample,
		Args:              cobra.MinimumNArgs(1),
		ArgAliases:        []string{"username", "display-name"},
		RunE:              ctx.UserChangeNameRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUserEmailCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "email",
		Short:             cmdAutheliaUserDeleteShort,
		Long:              cmdAutheliaUserDeleteLong,
		Example:           cmdAutheliaUserDeleteExample,
		Args:              cobra.MinimumNArgs(1),
		ArgAliases:        []string{"username", "email"},
		RunE:              ctx.UserChangeEmailRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUserGroupsCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "groups",
		Short:             cmdAutheliaUserDeleteShort,
		Long:              cmdAutheliaUserDeleteLong,
		Example:           cmdAutheliaUserDeleteExample,
		Args:              cobra.MinimumNArgs(2),
		ArgAliases:        []string{"username", "group"},
		RunE:              ctx.UserChangeGroupsRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newListUsersCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "list",
		Short:             cmdAutheliaUserDeleteShort,
		Long:              cmdAutheliaUserDeleteLong,
		Example:           cmdAutheliaUserDeleteExample,
		RunE:              ctx.UserListRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUserDisableCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "disable",
		Short:             cmdAutheliaUserDeleteShort,
		Long:              cmdAutheliaUserDeleteLong,
		Example:           cmdAutheliaUserDeleteExample,
		Args:              cobra.MinimumNArgs(1),
		ArgAliases:        []string{"username"},
		RunE:              ctx.UserDisableRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUserEnableCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "enable",
		Short:             cmdAutheliaUserDeleteShort,
		Long:              cmdAutheliaUserDeleteLong,
		Example:           cmdAutheliaUserDeleteExample,
		Args:              cobra.MinimumNArgs(1),
		ArgAliases:        []string{"username"},
		RunE:              ctx.UserEnableRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}
