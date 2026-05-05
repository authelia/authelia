package commands

import (
	"github.com/spf13/cobra"
)

func newUsersCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "users",
		Short:   "Manage users",
		Long:    "Manage users in Authelia. This subcommand has several methods to interact with the Authelia SQL Database. This allows doing several advanced operations that would be difficult normally.",
		Example: "authelia users --help",
		PersistentPreRunE: ctx.ChainRunE(
			ctx.ConfigRequireExplicitRunE,
			ctx.ConfigStorageCommandLineConfigRunE,
			ctx.HelperConfigLoadRunE,
			ctx.ConfigValidateStorageRunE,
			ctx.ConfigValidateAdministrationRunE,
			ctx.ConfigValidateUserBackendRunE,
			ctx.LoadProvidersStorageRunE,
			ctx.LoadProvidersUserBackendRunE,
		),
		Args: cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newUsersGetCmd(ctx),
		newUsersListCmd(ctx),
		newUsersAddCmd(ctx),
		newUsersUpdateCmd(ctx),
		newUsersDeleteCmd(ctx),
		newUsersGroupsCmd(ctx),
		newUsersSchemaCmd(ctx),
	)

	return cmd
}

func newUsersGetCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "get <username>",
		Short:             "Get details for a user",
		Long:              "Get details for a user",
		Example:           "authelia users get <username>",
		Args:              cobra.ExactArgs(1),
		RunE:              ctx.UsersGetRunE(),
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "output format, options are 'table' or 'json'")
	cmd.Flags().StringSlice(cmdFlagNameFields, nil, "fields to display as columns, defaults to the backend's required attributes (excluding password)")

	return cmd
}

func newUsersListCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "list",
		Short:             "Get details for all users",
		Long:              "Get details for all users",
		Example:           "authelia users list --help",
		Args:              cobra.NoArgs,
		RunE:              ctx.UsersListRunE(),
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "output format, options are 'table' or 'json'")
	cmd.Flags().StringSlice(cmdFlagNameFields, nil, "fields to display as columns, defaults to the backend's required attributes (excluding password)")

	return cmd
}

func newUsersAddCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "add",
		Short:             "Add a user",
		Long:              "Add a user",
		Example:           "authelia users add\nauthelia users add --file user.json\nauthelia users add --file - < user.json",
		Args:              cobra.NoArgs,
		RunE:              ctx.UsersAddRunE(),
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(cmdFlagNameFile, "f", "", "path to a JSON file describing the user, use '-' for stdin")

	return cmd
}

func newUsersUpdateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "update",
		Short:             "Modify a user",
		Long:              "Modify a user",
		Example:           "authelia users update --help",
		Args:              cobra.NoArgs,
		RunE:              ctx.UsersUpdateRunE(),
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUsersDeleteCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "delete <username>",
		Short:             "Delete a user",
		Long:              "Delete a user",
		Example:           "authelia users delete <username>",
		Args:              cobra.ExactArgs(1),
		RunE:              ctx.UsersDeleteRunE(),
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUsersGroupsCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "groups",
		Short:   "Manage user groups",
		Long:    "Manage user groups in Authelia. This subcommand has several methods to interact with Authelia Users. This allows doing several advanced operations that would be difficult normally.",
		Example: "authelia users groups --help",
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newUsersGroupsListCmd(ctx),
		newUsersGroupsAddCmd(ctx),
		newUsersGroupsDeleteCmd(ctx),
	)

	return cmd
}

func newUsersGroupsListCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "list",
		Short:             "List all user groups",
		Long:              "List all user groups",
		Example:           "authelia users groups list --help",
		Args:              cobra.NoArgs,
		RunE:              ctx.UsersGroupsListRunE(),
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUsersGroupsAddCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "add",
		Short:             "Add a group",
		Long:              "Add a group",
		Example:           "authelia users groups add --help",
		Args:              cobra.ExactArgs(1),
		RunE:              ctx.UsersGroupsAddRunE(),
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUsersGroupsDeleteCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "delete",
		Short:             "Delete a group",
		Long:              "Delete a group",
		Example:           "authelia users groups delete --help",
		Args:              cobra.ExactArgs(1),
		RunE:              ctx.UsersGroupsDeleteRunE(),
		DisableAutoGenTag: true,
	}

	return cmd
}

func newUsersSchemaCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "schema",
		Short:             "View user schema",
		Long:              "View user schema",
		Example:           "authelia users schema",
		Args:              cobra.NoArgs,
		RunE:              ctx.UsersSchemaPrintRunE(),
		DisableAutoGenTag: true,
	}

	return cmd
}
