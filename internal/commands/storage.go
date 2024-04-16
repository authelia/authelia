package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newStorageCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "storage",
		Short:   cmdAutheliaStorageShort,
		Long:    cmdAutheliaStorageLong,
		Example: cmdAutheliaStorageExample,
		PersistentPreRunE: ctx.ChainRunE(
			ctx.ConfigStorageCommandLineConfigRunE,
			ctx.HelperConfigLoadRunE,
			ctx.ConfigValidateStorageRunE,
			ctx.LoadProvidersStorageRunE,
		),
		Args: cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.PersistentFlags().String(cmdFlagNameEncryptionKey, "", "the storage encryption key to use")

	cmd.PersistentFlags().String(cmdFlagNameSQLite3Path, "", "the SQLite database path")

	cmd.PersistentFlags().String(cmdFlagNameMySQLHost, "", "the MySQL hostname")
	cmd.PersistentFlags().Int(cmdFlagNameMySQLPort, 3306, "the MySQL port")
	cmd.PersistentFlags().String(cmdFlagNameMySQLDatabase, "authelia", "the MySQL database name")
	cmd.PersistentFlags().String(cmdFlagNameMySQLUsername, "authelia", "the MySQL username")
	cmd.PersistentFlags().String(cmdFlagNameMySQLPassword, "", "the MySQL password")

	cmd.PersistentFlags().String(cmdFlagNamePostgreSQLHost, "", "the PostgreSQL hostname")
	cmd.PersistentFlags().Int(cmdFlagNamePostgreSQLPort, 5432, "the PostgreSQL port")
	cmd.PersistentFlags().String(cmdFlagNamePostgreSQLDatabase, "authelia", "the PostgreSQL database name")
	cmd.PersistentFlags().String(cmdFlagNamePostgreSQLSchema, "public", "the PostgreSQL schema name")
	cmd.PersistentFlags().String(cmdFlagNamePostgreSQLUsername, "authelia", "the PostgreSQL username")
	cmd.PersistentFlags().String(cmdFlagNamePostgreSQLPassword, "", "the PostgreSQL password")
	cmd.PersistentFlags().String("postgres.ssl.mode", "disable", "the PostgreSQL ssl mode")
	cmd.PersistentFlags().String("postgres.ssl.root_certificate", "", "the PostgreSQL ssl root certificate file location")
	cmd.PersistentFlags().String("postgres.ssl.certificate", "", "the PostgreSQL ssl certificate file location")
	cmd.PersistentFlags().String("postgres.ssl.key", "", "the PostgreSQL ssl key file location")

	cmd.AddCommand(
		newStorageMigrateCmd(ctx),
		newStorageSchemaInfoCmd(ctx),
		newStorageEncryptionCmd(ctx),
		newStorageUserCmd(ctx),
	)

	return cmd
}

func newStorageEncryptionCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "encryption",
		Short:   cmdAutheliaStorageEncryptionShort,
		Long:    cmdAutheliaStorageEncryptionLong,
		Example: cmdAutheliaStorageEncryptionExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newStorageEncryptionChangeKeyCmd(ctx),
		newStorageEncryptionCheckCmd(ctx),
	)

	return cmd
}

func newStorageEncryptionCheckCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "check",
		Short:   cmdAutheliaStorageEncryptionCheckShort,
		Long:    cmdAutheliaStorageEncryptionCheckLong,
		Example: cmdAutheliaStorageEncryptionCheckExample,
		RunE:    ctx.StorageSchemaEncryptionCheckRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().Bool(cmdFlagNameVerbose, false, "enables verbose checking of every row of encrypted data")

	return cmd
}

func newStorageEncryptionChangeKeyCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "change-key",
		Short:   cmdAutheliaStorageEncryptionChangeKeyShort,
		Long:    cmdAutheliaStorageEncryptionChangeKeyLong,
		Example: cmdAutheliaStorageEncryptionChangeKeyExample,
		RunE:    ctx.StorageSchemaEncryptionChangeKeyRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().String(cmdFlagNameNewEncryptionKey, "", "the new key to encrypt the data with")

	return cmd
}

func newStorageUserCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "user",
		Short:   cmdAutheliaStorageUserShort,
		Long:    cmdAutheliaStorageUserLong,
		Example: cmdAutheliaStorageUserExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newStorageUserIdentifiersCmd(ctx),
		newStorageUserTOTPCmd(ctx),
		newStorageUserWebAuthnCmd(ctx),
	)

	return cmd
}

func newStorageUserIdentifiersCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "identifiers",
		Short:   cmdAutheliaStorageUserIdentifiersShort,
		Long:    cmdAutheliaStorageUserIdentifiersLong,
		Example: cmdAutheliaStorageUserIdentifiersExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newStorageUserIdentifiersExportCmd(ctx),
		newStorageUserIdentifiersImportCmd(ctx),
		newStorageUserIdentifiersGenerateCmd(ctx),
		newStorageUserIdentifiersAddCmd(ctx),
	)

	return cmd
}

func newStorageUserIdentifiersExportCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseExport,
		Short:   cmdAutheliaStorageUserIdentifiersExportShort,
		Long:    cmdAutheliaStorageUserIdentifiersExportLong,
		Example: cmdAutheliaStorageUserIdentifiersExportExample,
		RunE:    ctx.StorageUserIdentifiersExportRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(cmdFlagNameFile, "f", "authelia.export.opaque-identifiers.yml", "The file name for the YAML export")

	return cmd
}

func newStorageUserIdentifiersImportCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseImportFileName,
		Short:   cmdAutheliaStorageUserIdentifiersImportShort,
		Long:    cmdAutheliaStorageUserIdentifiersImportLong,
		Example: cmdAutheliaStorageUserIdentifiersImportExample,
		RunE:    ctx.StorageUserIdentifiersImportRunE,
		Args:    cobra.ExactArgs(1),

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageUserIdentifiersGenerateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "generate",
		Short:   cmdAutheliaStorageUserIdentifiersGenerateShort,
		Long:    cmdAutheliaStorageUserIdentifiersGenerateLong,
		Example: cmdAutheliaStorageUserIdentifiersGenerateExample,
		RunE:    ctx.StorageUserIdentifiersGenerateRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringSlice(cmdFlagNameUsers, nil, "The list of users to generate the opaque identifiers for")
	cmd.Flags().StringSlice(cmdFlagNameServices, []string{identifierServiceOpenIDConnect}, fmt.Sprintf("The list of services to generate the opaque identifiers for, valid values are: %s", strings.Join(validIdentifierServices, ", ")))
	cmd.Flags().StringSlice(cmdFlagNameSectors, []string{""}, "The list of sectors to generate identifiers for")

	return cmd
}

func newStorageUserIdentifiersAddCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "add <username>",
		Short:   cmdAutheliaStorageUserIdentifiersAddShort,
		Long:    cmdAutheliaStorageUserIdentifiersAddLong,
		Example: cmdAutheliaStorageUserIdentifiersAddExample,
		RunE:    ctx.StorageUserIdentifiersAddRunE,
		Args:    cobra.ExactArgs(1),

		DisableAutoGenTag: true,
	}

	cmd.Flags().String(cmdFlagNameIdentifier, "", "The optional version 4 UUID to use, if not set a random one will be used")
	cmd.Flags().String(cmdFlagNameService, identifierServiceOpenIDConnect, fmt.Sprintf("The service to add the identifier for, valid values are: %s", strings.Join(validIdentifierServices, ", ")))
	cmd.Flags().String(cmdFlagNameSector, "", "The sector identifier to use (should usually be blank)")

	return cmd
}

func newStorageUserWebAuthnCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "webauthn",
		Short:   cmdAutheliaStorageUserWebAuthnShort,
		Long:    cmdAutheliaStorageUserWebAuthnLong,
		Example: cmdAutheliaStorageUserWebAuthnExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newStorageUserWebAuthnListCmd(ctx),
		newStorageUserWebAuthnDeleteCmd(ctx),
		newStorageUserWebAuthnExportCmd(ctx),
		newStorageUserWebAuthnImportCmd(ctx),
	)

	return cmd
}

func newStorageUserWebAuthnImportCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseImportFileName,
		Short:   cmdAutheliaStorageUserWebAuthnImportShort,
		Long:    cmdAutheliaStorageUserWebAuthnImportLong,
		Example: cmdAutheliaStorageUserWebAuthnImportExample,
		RunE:    ctx.StorageUserWebAuthnImportRunE,
		Args:    cobra.ExactArgs(1),

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageUserWebAuthnExportCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseExport,
		Short:   cmdAutheliaStorageUserWebAuthnExportShort,
		Long:    cmdAutheliaStorageUserWebAuthnExportLong,
		Example: cmdAutheliaStorageUserWebAuthnExportExample,
		RunE:    ctx.StorageUserWebAuthnExportRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(cmdFlagNameFile, "f", "authelia.export.webauthn.yaml", "The file name for the YAML export")

	return cmd
}

func newStorageUserWebAuthnListCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "list [username]",
		Short:   cmdAutheliaStorageUserWebAuthnListShort,
		Long:    cmdAutheliaStorageUserWebAuthnListLong,
		Example: cmdAutheliaStorageUserWebAuthnListExample,
		RunE:    ctx.StorageUserWebAuthnListRunE,
		Args:    cobra.MaximumNArgs(1),

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageUserWebAuthnDeleteCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "delete [username]",
		Short:   cmdAutheliaStorageUserWebAuthnDeleteShort,
		Long:    cmdAutheliaStorageUserWebAuthnDeleteLong,
		Example: cmdAutheliaStorageUserWebAuthnDeleteExample,
		RunE:    ctx.StorageUserWebAuthnDeleteRunE,
		Args:    cobra.MaximumNArgs(1),

		DisableAutoGenTag: true,
	}

	cmd.Flags().Bool(cmdFlagNameAll, false, "delete all of the users WebAuthn credentials")
	cmd.Flags().String(cmdFlagNameDescription, "", "delete a users WebAuthn credential by description")
	cmd.Flags().String(cmdFlagNameKeyID, "", "delete a users WebAuthn credential by key id")

	return cmd
}

func newStorageUserTOTPCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "totp",
		Short:   cmdAutheliaStorageUserTOTPShort,
		Long:    cmdAutheliaStorageUserTOTPLong,
		Example: cmdAutheliaStorageUserTOTPExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newStorageUserTOTPGenerateCmd(ctx),
		newStorageUserTOTPDeleteCmd(ctx),
		newStorageUserTOTPExportCmd(ctx),
		newStorageUserTOTPImportCmd(ctx),
	)

	return cmd
}

func newStorageUserTOTPGenerateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "generate <username>",
		Short:   cmdAutheliaStorageUserTOTPGenerateShort,
		Long:    cmdAutheliaStorageUserTOTPGenerateLong,
		Example: cmdAutheliaStorageUserTOTPGenerateExample,
		RunE:    ctx.StorageUserTOTPGenerateRunE,
		Args:    cobra.ExactArgs(1),

		DisableAutoGenTag: true,
	}

	cmd.Flags().String(cmdFlagNameSecret, "", "set the shared secret as base32 encoded bytes (no padding), it's recommended that you do not use this option unless you're restoring a configuration")
	cmd.Flags().Uint(cmdFlagNameSecretSize, schema.TOTPSecretSizeDefault, "set the secret size")
	cmd.Flags().Uint(cmdFlagNamePeriod, 30, "set the period between rotations")
	cmd.Flags().Uint(cmdFlagNameDigits, 6, "set the number of digits")
	cmd.Flags().String(cmdFlagNameAlgorithm, "SHA1", "set the algorithm to either SHA1 (supported by most applications), SHA256, or SHA512")
	cmd.Flags().String(cmdFlagNameIssuer, "Authelia", "set the issuer description")
	cmd.Flags().BoolP(cmdFlagNameForce, "f", false, "forces the configuration to be generated regardless if it exists or not")
	cmd.Flags().StringP(cmdFlagNamePath, "p", "", "path to a file to create a PNG file with the QR code (optional)")

	return cmd
}

func newStorageUserTOTPDeleteCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "delete <username>",
		Short:   cmdAutheliaStorageUserTOTPDeleteShort,
		Long:    cmdAutheliaStorageUserTOTPDeleteLong,
		Example: cmdAutheliaStorageUserTOTPDeleteExample,
		RunE:    ctx.StorageUserTOTPDeleteRunE,
		Args:    cobra.ExactArgs(1),

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageUserTOTPImportCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseImportFileName,
		Short:   cmdAutheliaStorageUserTOTPImportShort,
		Long:    cmdAutheliaStorageUserTOTPImportLong,
		Example: cmdAutheliaStorageUserTOTPImportExample,
		RunE:    ctx.StorageUserTOTPImportRunE,
		Args:    cobra.ExactArgs(1),

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageUserTOTPExportCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseExport,
		Short:   cmdAutheliaStorageUserTOTPExportShort,
		Long:    cmdAutheliaStorageUserTOTPExportLong,
		Example: cmdAutheliaStorageUserTOTPExportExample,
		RunE:    ctx.StorageUserTOTPExportRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newStorageUserTOTPExportCSVCmd(ctx),
		newStorageUserTOTPExportPNGCmd(ctx),
		newStorageUserTOTPExportURICmd(ctx),
	)

	cmd.Flags().StringP(cmdFlagNameFile, "f", "authelia.export.totp.yaml", "The file name for the YAML export")

	return cmd
}

func newStorageUserTOTPExportURICmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "uri",
		Short:   cmdAutheliaStorageUserTOTPExportURIShort,
		Long:    cmdAutheliaStorageUserTOTPExportURILong,
		Example: cmdAutheliaStorageUserTOTPExportURIExample,
		RunE:    ctx.StorageUserTOTPExportURIRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageUserTOTPExportCSVCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "csv",
		Short:   cmdAutheliaStorageUserTOTPExportCSVShort,
		Long:    cmdAutheliaStorageUserTOTPExportCSVLong,
		Example: cmdAutheliaStorageUserTOTPExportCSVExample,
		RunE:    ctx.StorageUserTOTPExportCSVRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(cmdFlagNameFile, "f", "authelia.export.totp.csv", "The file name for the CSV export")

	return cmd
}

func newStorageUserTOTPExportPNGCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "png",
		Short:   cmdAutheliaStorageUserTOTPExportPNGShort,
		Long:    cmdAutheliaStorageUserTOTPExportPNGLong,
		Example: cmdAutheliaStorageUserTOTPExportPNGExample,
		RunE:    ctx.StorageUserTOTPExportPNGRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().String(cmdFlagNameDirectory, "", "The directory where all exported png files will be saved to")

	return cmd
}

func newStorageSchemaInfoCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "schema-info",
		Short:   cmdAutheliaStorageSchemaInfoShort,
		Long:    cmdAutheliaStorageSchemaInfoLong,
		Example: cmdAutheliaStorageSchemaInfoExample,
		RunE:    ctx.StorageSchemaInfoRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	return cmd
}

// newStorageMigrateCmd returns a new Migration Cmd.
func newStorageMigrateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "migrate",
		Short:   cmdAutheliaStorageMigrateShort,
		Long:    cmdAutheliaStorageMigrateLong,
		Example: cmdAutheliaStorageMigrateExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newStorageMigrateUpCmd(ctx), newStorageMigrateDownCmd(ctx),
		newStorageMigrateListUpCmd(ctx), newStorageMigrateListDownCmd(ctx),
		newStorageMigrateHistoryCmd(ctx),
	)

	return cmd
}

func newStorageMigrateHistoryCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "history",
		Short:   cmdAutheliaStorageMigrateHistoryShort,
		Long:    cmdAutheliaStorageMigrateHistoryLong,
		Example: cmdAutheliaStorageMigrateHistoryExample,
		RunE:    ctx.StorageMigrateHistoryRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageMigrateListUpCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "list-up",
		Short:   cmdAutheliaStorageMigrateListUpShort,
		Long:    cmdAutheliaStorageMigrateListUpLong,
		Example: cmdAutheliaStorageMigrateListUpExample,
		RunE:    ctx.NewStorageMigrateListRunE(true),
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageMigrateListDownCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "list-down",
		Short:   cmdAutheliaStorageMigrateListDownShort,
		Long:    cmdAutheliaStorageMigrateListDownLong,
		Example: cmdAutheliaStorageMigrateListDownExample,
		RunE:    ctx.NewStorageMigrateListRunE(false),
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newStorageMigrateUpCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     storageMigrateDirectionUp,
		Short:   cmdAutheliaStorageMigrateUpShort,
		Long:    cmdAutheliaStorageMigrateUpLong,
		Example: cmdAutheliaStorageMigrateUpExample,
		RunE:    ctx.NewStorageMigrationRunE(true),
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().IntP(cmdFlagNameTarget, "t", 0, "sets the version to migrate to, by default this is the latest version")

	return cmd
}

func newStorageMigrateDownCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     storageMigrateDirectionDown,
		Short:   cmdAutheliaStorageMigrateDownShort,
		Long:    cmdAutheliaStorageMigrateDownLong,
		Example: cmdAutheliaStorageMigrateDownExample,
		RunE:    ctx.NewStorageMigrationRunE(false),
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().IntP(cmdFlagNameTarget, "t", 0, "sets the version to migrate to")
	cmd.Flags().Bool(cmdFlagNameDestroyData, false, "confirms you want to destroy data with this migration")

	return cmd
}
