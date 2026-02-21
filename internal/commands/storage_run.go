package commands

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/go-webauthn/webauthn/metadata"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/totp"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// LoadProvidersStorageRunE is a special PreRunE that loads the storage provider into the CmdCtx.
func (ctx *CmdCtx) LoadProvidersStorageRunE(cmd *cobra.Command, args []string) (err error) {
	switch warns, errs := ctx.LoadTrustedCertificates(); {
	case len(errs) != 0:
		err = fmt.Errorf("had the following errors loading the trusted certificates")

		for _, e := range errs {
			err = fmt.Errorf("%+v: %w", err, e)
		}

		return err
	case len(warns) != 0:
		err = fmt.Errorf("had the following warnings loading the trusted certificates")

		for _, e := range errs {
			err = fmt.Errorf("%+v: %w", err, e)
		}

		return err
	default:
		ctx.providers.StorageProvider = getStorageProvider(ctx)

		return nil
	}
}

// ConfigStorageCommandLineConfigRunE configures the storage command mapping.
func (ctx *CmdCtx) ConfigStorageCommandLineConfigRunE(cmd *cobra.Command, _ []string) (err error) {
	//nolint:gosec // This maps flasg to field names.
	flagsMap := map[string]string{
		cmdFlagNameEncryptionKey: "storage.encryption_key",

		cmdFlagNameSQLite3Path: "storage.local.path",

		cmdFlagNameMySQLAddress:  "storage.mysql.address",
		cmdFlagNameMySQLDatabase: "storage.mysql.database",
		cmdFlagNameMySQLUsername: "storage.mysql.username",
		cmdFlagNameMySQLPassword: "storage.mysql.password",

		cmdFlagNamePostgreSQLAddress:  "storage.postgres.address",
		cmdFlagNamePostgreSQLDatabase: "storage.postgres.database",
		cmdFlagNamePostgreSQLSchema:   "storage.postgres.schema",
		cmdFlagNamePostgreSQLUsername: "storage.postgres.username",
		cmdFlagNamePostgreSQLPassword: "storage.postgres.password",

		cmdFlagNamePeriod:     "totp.period",
		cmdFlagNameDigits:     "totp.digits",
		cmdFlagNameAlgorithm:  "totp.algorithm",
		cmdFlagNameIssuer:     "totp.issuer",
		cmdFlagNameSecretSize: "totp.secret_size",
	}

	return ctx.HelperConfigSetFlagsMapRunE(cmd.Flags(), flagsMap, true, false)
}

// ConfigValidateStorageRunE validates the storage config before running commands using it.
func (ctx *CmdCtx) ConfigValidateStorageRunE(_ *cobra.Command, _ []string) (err error) {
	if errs := ctx.cconfig.validator.Errors(); len(errs) != 0 {
		var (
			i int
			e error
		)

		for i, e = range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%w, %v", err, e)
		}

		return err
	}

	validator.ValidateStorage(ctx.config.Storage, ctx.cconfig.validator)

	validator.ValidateTOTP(ctx.config, ctx.cconfig.validator)

	if errs := ctx.cconfig.validator.Errors(); len(errs) != 0 {
		var (
			i int
			e error
		)

		for i, e = range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%w, %v", err, e)
		}

		return err
	}

	return nil
}

func (ctx *CmdCtx) StorageCacheDeleteRunE(name, description string) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		if err = ctx.CheckSchema(); err != nil {
			return storageWrapCheckSchemaErr(err)
		}

		if err = ctx.providers.StorageProvider.DeleteCachedData(ctx, name); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted cached %s data.\n", description)

		return nil
	}
}

func (ctx *CmdCtx) StorageCacheMDS3StatusRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	return runStorageCacheMDS3Status(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, ctx.config)
}

func runStorageCacheMDS3Status(ctx context.Context, w io.Writer, store storage.Provider, config *schema.Configuration) (err error) {
	if !config.WebAuthn.Metadata.Enabled {
		return fmt.Errorf("webauthn metadata is disabled")
	}

	provider, err := webauthn.NewMetaDataProvider(config, store)
	if err != nil {
		return err
	}

	var (
		mds *metadata.Metadata

		valid, initialized, outdated bool
	)

	if mds, _, err = provider.LoadCache(ctx); err == nil {
		valid = true

		if mds != nil {
			initialized = true
			outdated = provider.Outdated()
		}
	}

	_, _ = fmt.Fprintf(w, "WebAuthn MDS3 Cache Status:\n\n\tValid: %t\n\tInitialized: %t\n\tOutdated: %t\n", valid, initialized, outdated)

	if initialized {
		_, _ = fmt.Fprintf(w, "\tVersion: %d\n", mds.Parsed.Number)

		if !outdated {
			_, _ = fmt.Fprintf(w, "\tNext Update: %s\n", mds.Parsed.NextUpdate.Format("January 2, 2006"))
		}
	}

	return nil
}

func (ctx *CmdCtx) StorageCacheMDS3DumpRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if !ctx.config.WebAuthn.Metadata.Enabled {
		return fmt.Errorf("webauthn metadata is disabled")
	}

	var path string
	if path, err = cmd.Flags().GetString("path"); err != nil {
		return err
	}

	return runStorageCacheMDS3Dump(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, ctx.config, path)
}

func runStorageCacheMDS3Dump(ctx context.Context, w io.Writer, store storage.Provider, config *schema.Configuration, path string) (err error) {
	var (
		provider webauthn.MetaDataProvider
		mds      *metadata.Metadata
		data     []byte
	)

	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("error dumping metadata: path must not be blank")
	}

	if provider, err = webauthn.NewMetaDataProvider(config, store); err != nil {
		return err
	}

	if mds, data, err = provider.LoadCache(ctx); err != nil {
		return err
	} else if mds == nil {
		return fmt.Errorf("error dumping metadata: no metadata is in the cache")
	}

	var f *os.File
	if f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("error writing data to file: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Successfully dumped WebAuthn MDS3 data with version %d from cache to file '%s'.\n", mds.Parsed.Number, path)

	return nil
}

func (ctx *CmdCtx) StorageCacheMDS3UpdateRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		path  string
		force bool
	)

	if force, err = cmd.Flags().GetBool(cmdFlagNameForce); err != nil {
		return err
	}

	if path, err = cmd.Flags().GetString(cmdFlagNamePath); err != nil {
		return err
	}

	return runStorageCacheMDS3Update(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, ctx.config, path, force)
}

func runStorageCacheMDS3Update(ctx context.Context, w io.Writer, store storage.Provider, config *schema.Configuration, path string, force bool) (err error) {
	if !config.WebAuthn.Metadata.Enabled {
		return fmt.Errorf("webauthn metadata is disabled")
	}

	var (
		provider webauthn.MetaDataProvider
		mds      *metadata.Metadata
		data     []byte
	)

	if provider, err = webauthn.NewMetaDataProvider(config, store); err != nil {
		return err
	}

	if mds, _, err = provider.LoadCache(ctx); err != nil {
		return err
	} else if mds != nil && !force && !provider.Outdated() {
		_, _ = fmt.Fprintf(w, "WebAuthn MDS3 cache data with version %d due for update on %s does not require an update.\n", mds.Parsed.Number, mds.Parsed.NextUpdate.Format("January 2, 2006"))

		return nil
	}

	switch {
	case path != "":
		mds, data, err = provider.LoadFile(ctx, path)
	case force:
		mds, data, err = provider.LoadForce(ctx)
	default:
		mds, data, err = provider.Load(ctx)
	}

	if err != nil {
		return err
	}

	if data == nil {
		return fmt.Errorf("error updating metadata: no data was returned")
	}

	if provider.Outdated() && !force {
		_, _ = fmt.Fprintf(w, "Provided WebAuthn MDS3 data with version %d was due for update on %s and can't be used.\n", mds.Parsed.Number, mds.Parsed.NextUpdate.Format("January 2, 2006"))

		return fmt.Errorf("error updating metadata: the metadata is outdated")
	}

	if err = provider.SaveCache(ctx, data); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "WebAuthn MDS3 cache data updated to version %d and is due for update on %s.\n", mds.Parsed.Number, mds.Parsed.NextUpdate.Format("January 2, 2006"))

	return nil
}

func (ctx *CmdCtx) StorageSchemaEncryptionCheckRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	var verbose bool

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if verbose, err = cmd.Flags().GetBool(cmdFlagNameVerbose); err != nil {
		return err
	}

	return runStorageSchemaEncryptionCheckKey(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, verbose)
}

//nolint:unparam
func runStorageSchemaEncryptionCheckKey(ctx context.Context, w io.Writer, store storage.Provider, verbose bool) (err error) {
	var result storage.EncryptionValidationResult
	if result, err = store.SchemaEncryptionCheckKey(ctx, verbose); err != nil {
		switch {
		case errors.Is(err, storage.ErrSchemaEncryptionVersionUnsupported):
			_, _ = fmt.Fprintf(w, "Storage Encryption Key Validation: FAILURE\n\n\tCause: The schema version doesn't support encryption.\n")
		default:
			_, _ = fmt.Fprintf(w, "Storage Encryption Key Validation: UNKNOWN\n\n\tCause: %v.\n", err)
		}

		return nil
	}

	if result.Success() {
		_, _ = fmt.Fprintln(w, "Storage Encryption Key Validation: SUCCESS")
	} else {
		_, _ = fmt.Fprintf(w, "Storage Encryption Key Validation: FAILURE\n\n\tCause: %v.\n", storage.ErrSchemaEncryptionInvalidKey)
	}

	if verbose {
		_, _ = fmt.Fprintf(w, "\nTables:")

		tables := make([]string, 0, len(result.Tables))

		for name := range result.Tables {
			tables = append(tables, name)
		}

		sort.Strings(tables)

		for _, name := range tables {
			table := result.Tables[name]

			_, _ = fmt.Fprintf(w, "\n\n\tTable (%s): %s\n\t\tInvalid Rows: %d\n\t\tTotal Rows: %d", name, table.ResultDescriptor(), table.Invalid, table.Total)
		}

		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// StorageSchemaEncryptionChangeKeyRunE is the RunE for the authelia storage encryption change-key command.
func (ctx *CmdCtx) StorageSchemaEncryptionChangeKeyRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var key string
	if key, err = cmd.Flags().GetString(cmdFlagNameNewEncryptionKey); err != nil {
		return err
	}

	return runStorageSchemaEncryptionChangeKey(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, key, !cmd.Flags().Changed(cmdFlagNameNewEncryptionKey))
}

func runStorageSchemaEncryptionChangeKey(ctx context.Context, w io.Writer, store storage.Provider, key string, read bool) (err error) {
	var version int
	if version, err = store.SchemaVersion(ctx); err != nil {
		return err
	}

	if version <= 0 {
		return errors.New("schema version must be at least version 1 to change the encryption key")
	}

	if read || key == "" {
		if key, err = termReadPasswordWithPrompt("Enter New Storage Encryption Key: ", cmdFlagNameNewEncryptionKey); err != nil {
			return err
		}
	}

	switch {
	case key == "":
		return errors.New("the new encryption key must not be blank")
	case len(key) < 20:
		return errors.New("the new encryption key must be at least 20 characters")
	}

	if err = store.SchemaEncryptionChangeKey(ctx, key); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w, "Completed the encryption key change. Please adjust your configuration to use the new key.")

	return nil
}

// StorageMigrateHistoryRunE is the RunE for the authelia storage migrate history command.
func (ctx *CmdCtx) StorageMigrateHistoryRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	return runStorageMigrateHistory(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider)
}

func runStorageMigrateHistory(ctx context.Context, w io.Writer, store storage.Provider) (err error) {
	var (
		version    int
		migrations []model.Migration
	)

	if version, err = store.SchemaVersion(ctx); err != nil {
		return err
	}

	if version <= 0 {
		_, _ = fmt.Fprintln(w, "No migration history is available for schemas that are not version 1 or above.")
		return
	}

	if migrations, err = store.SchemaMigrationHistory(ctx); err != nil {
		return err
	}

	if len(migrations) == 0 {
		return errors.New("no migration history found which may indicate a broken schema")
	}

	_, _ = fmt.Fprintf(w, "Migration History:\n\n")

	tw := tabwriter.NewWriter(w, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(tw, "ID\tDate\tBefore\tAfter\tAuthelia Version")

	for _, m := range migrations {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%d\t%d\t%s\n", m.ID, m.Applied.Format("2006-01-02 15:04:05 -0700"), m.Before, m.After, m.Version)
	}

	return tw.Flush()
}

// NewStorageMigrateListRunE creates the RunE for the authelia storage migrate list command.
func (ctx *CmdCtx) NewStorageMigrateListRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		return runStorageMigrateList(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, up)
	}
}

func runStorageMigrateList(ctx context.Context, w io.Writer, store storage.Provider, up bool) (err error) {
	var (
		migrations   []model.SchemaMigration
		directionStr string
	)

	if up {
		migrations, err = store.SchemaMigrationsUp(ctx, 0)
		directionStr = "Up"
	} else {
		migrations, err = store.SchemaMigrationsDown(ctx, 0)
		directionStr = "Down"
	}

	if err != nil && !errors.Is(err, storage.ErrNoAvailableMigrations) && !errors.Is(err, storage.ErrMigrateCurrentVersionSameAsTarget) {
		return err
	}

	if len(migrations) == 0 {
		_, _ = fmt.Fprintf(w, "Storage Schema Migration List (%s)\n\nNo Migrations Available\n", directionStr)

		return nil
	}

	_, _ = fmt.Fprintf(w, "Storage Schema Migration List (%s)\n\n", directionStr)

	tw := tabwriter.NewWriter(w, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(tw, "Version\tDescription")

	for _, migration := range migrations {
		_, _ = fmt.Fprintf(tw, "%d\t%s\n", migration.Version, migration.Name)
	}

	return tw.Flush()
}

// NewStorageMigrationRunE creates the RunE for the authelia storage migrate command.
func (ctx *CmdCtx) NewStorageMigrationRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		var (
			target  int
			destroy bool
		)

		if target, err = cmd.Flags().GetInt(cmdFlagNameTarget); err != nil {
			return err
		}

		if !up {
			if destroy, err = cmd.Flags().GetBool(cmdFlagNameDestroyData); err != nil {
				return err
			}
		}

		return runStorageMigration(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, up, target, destroy, cmd.Flags().Changed(cmdFlagNameTarget))
	}
}

func runStorageMigration(ctx context.Context, _ io.Writer, store storage.Provider, up bool, version int, destroy, useTarget bool) (err error) {
	if !useTarget {
		if up {
			version = storage.SchemaLatest
		} else {
			return errors.New("you must set a target version")
		}
	}

	if !up {
		var confirmed bool

		if destroy {
			confirmed = true
		} else {
			if confirmed, err = termReadConfirmation("Schema Down Migrations may DESTROY data, type 'DESTROY' and press return to continue: ", "DESTROY"); err != nil {
				return err
			}
		}

		if !confirmed {
			return errors.New("cancelling down migration due to user not accepting data destruction")
		}
	}

	return store.SchemaMigrate(ctx, up, version)
}

// StorageSchemaInfoRunE is the RunE for the authelia storage schema info command.
func (ctx *CmdCtx) StorageSchemaInfoRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	return runStorageSchemaInfo(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider)
}

func runStorageSchemaInfo(ctx context.Context, w io.Writer, store storage.Provider) (err error) {
	var (
		upgradeStr, tablesStr string

		tables []string

		version, latest int
	)

	if version, err = store.SchemaVersion(ctx); err != nil && err.Error() != "unknown schema state" {
		return err
	}

	if tables, err = store.SchemaTables(ctx); err != nil {
		return err
	}

	if len(tables) == 0 {
		tablesStr = "N/A"
	} else {
		tablesStr = strings.Join(tables, ", ")
	}

	if latest, err = store.SchemaLatestVersion(); err != nil {
		return err
	}

	if latest > version {
		upgradeStr = fmt.Sprintf("yes - version %d", latest)
	} else {
		upgradeStr = "no"
	}

	var (
		encryption string
		result     storage.EncryptionValidationResult
	)

	switch result, err = store.SchemaEncryptionCheckKey(ctx, false); {
	case err != nil:
		if errors.Is(err, storage.ErrSchemaEncryptionVersionUnsupported) {
			encryption = "unsupported (schema version)"
		} else {
			encryption = invalid
		}
	case !result.Success():
		encryption = invalid
	default:
		encryption = "valid"
	}

	_, _ = fmt.Fprintf(w, "Schema Version: %s\nSchema Upgrade Available: %s\nSchema Tables: %s\nSchema Encryption Key: %s\n", storage.SchemaVersionToString(version), upgradeStr, tablesStr, encryption)

	return nil
}

func (ctx *CmdCtx) StorageBansListRunE(use string) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		if err = ctx.CheckSchema(); err != nil {
			return storageWrapCheckSchemaErr(err)
		}

		switch use {
		case cmdUseIP:
			return runStorageBansListIP(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider)
		case cmdUseUser:
			return runStorageBansListUser(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider)
		default:
			return fmt.Errorf("unknown command %q", use)
		}
	}
}

func runStorageBansListIP(ctx context.Context, w io.Writer, store storage.Provider) (err error) {
	var results []model.BannedIP

	limit := 10
	count := 0

	for page := 0; true; page++ {
		var bans []model.BannedIP

		if bans, err = store.LoadBannedIPs(ctx, limit, page); err != nil {
			return err
		}

		l := len(bans)

		count += l

		results = append(results, bans...)

		if l < limit {
			break
		}
	}

	if count == 0 {
		_, _ = fmt.Fprintf(w, "No results.\n")

		return nil
	}

	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)

	_, _ = fmt.Fprintln(tw, "ID\tIP\tExpires\tSource\tReason")

	for _, ban := range results {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", ban.ID, ban.IP, regulation.FormatExpiresShort(ban.Expires), ban.Source, ban.Reason.String)
	}

	return tw.Flush()
}

func runStorageBansListUser(ctx context.Context, w io.Writer, store storage.Provider) (err error) {
	var results []model.BannedUser

	limit := 10
	count := 0

	for page := 0; true; page++ {
		var bans []model.BannedUser

		if bans, err = store.LoadBannedUsers(ctx, limit, page); err != nil {
			return err
		}

		l := len(bans)

		count += l

		results = append(results, bans...)

		if l < limit {
			break
		}
	}

	if count == 0 {
		_, _ = fmt.Fprintf(w, "No results.\n")

		return nil
	}

	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)

	_, _ = fmt.Fprintln(tw, "ID\tUsername\tExpires\tSource\tReason")

	for _, ban := range results {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", ban.ID, ban.Username, regulation.FormatExpiresShort(ban.Expires), ban.Source, ban.Reason.String)
	}

	return tw.Flush()
}

func (ctx *CmdCtx) StorageBansRevokeRunE(use string) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		if err = ctx.CheckSchema(); err != nil {
			return storageWrapCheckSchemaErr(err)
		}

		var (
			id     int
			target string
		)

		if id, err = cmd.Flags().GetInt("id"); err != nil {
			return err
		}

		if len(args) != 0 {
			target = args[0]
		}

		switch use {
		case cmdUseIP:
			return runStorageBansRevokeIP(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, id, target)
		case cmdUseUser:
			return runStorageBansRevokeUser(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, id, target)
		default:
			return fmt.Errorf("unknown command %q", use)
		}
	}
}

func runStorageBansRevokeIP(ctx context.Context, w io.Writer, store storage.Provider, id int, target string) (err error) {
	ip := net.ParseIP(target)

	var bans []model.BannedIP

	if id == 0 {
		if target == "" {
			return fmt.Errorf("either the ip or id is required")
		}

		if bans, err = store.LoadBannedIP(ctx, model.NewIP(ip)); err != nil {
			return err
		}
	} else {
		var ban model.BannedIP

		if ban, err = store.LoadBannedIPByID(ctx, id); err != nil {
			return err
		}

		bans = []model.BannedIP{ban}
	}

	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)

	_, _ = fmt.Fprintf(tw, "ID\tIP\tResult\tInformation\n")

	for _, ban := range bans {
		if ban.Revoked {
			_, _ = fmt.Fprintf(tw, "%d\t%s\tSKIPPED\tBan has already been revoked\n", ban.ID, ban.IP)
		} else {
			if err = store.RevokeBannedIP(ctx, ban.ID, time.Now()); err != nil {
				_, _ = fmt.Fprintf(tw, "%d\t%s\tFAILURE\tError: %+v\n", ban.ID, ban.IP, err)
			} else {
				_, _ = fmt.Fprintf(tw, "%d\t%s\tSUCCESS\tN/A\n", ban.ID, ban.IP)
			}
		}
	}

	return tw.Flush()
}

func runStorageBansRevokeUser(ctx context.Context, w io.Writer, store storage.Provider, id int, target string) (err error) {
	var bans []model.BannedUser

	if id == 0 {
		if target == "" {
			return fmt.Errorf("either the username or id is required")
		}

		if bans, err = store.LoadBannedUser(ctx, target); err != nil {
			return err
		}
	} else {
		var ban model.BannedUser

		if ban, err = store.LoadBannedUserByID(ctx, id); err != nil {
			return err
		}

		bans = []model.BannedUser{ban}
	}

	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)

	_, _ = fmt.Fprintf(tw, "ID\tUsername\tResult\tInformation\n")

	for _, ban := range bans {
		if ban.Revoked {
			_, _ = fmt.Fprintf(tw, "%d\t%s\tSKIPPED\tBan has already been revoked\n", ban.ID, ban.Username)
		} else {
			if err = store.RevokeBannedUser(ctx, ban.ID, time.Now()); err != nil {
				_, _ = fmt.Fprintf(tw, "%d\t%s\tFAILURE\tError: %+v\n", ban.ID, ban.Username, err)
			} else {
				_, _ = fmt.Fprintf(tw, "%d\t%s\tSUCCESS\tN/A\n", ban.ID, ban.Username)
			}
		}
	}

	return tw.Flush()
}

func (ctx *CmdCtx) StorageBansAddRunE(use string) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err := ctx.providers.StorageProvider.Close(); err != nil {
				panic(err)
			}
		}()

		if err = ctx.CheckSchema(); err != nil {
			return storageWrapCheckSchemaErr(err)
		}

		var (
			permanent           bool
			reason, durationStr string
		)

		if permanent, err = cmd.Flags().GetBool("permanent"); err != nil {
			return err
		}

		if reason, err = cmd.Flags().GetString("reason"); err != nil {
			return err
		}

		if durationStr, err = cmd.Flags().GetString("duration"); err != nil {
			return err
		}

		duration, err := utils.ParseDurationString(durationStr)
		if err != nil {
			return fmt.Errorf("failed to parse duration string: %w", err)
		}

		if duration <= 0 {
			return fmt.Errorf("duration must be a positive value")
		}

		target := args[0]

		switch use {
		case cmdUseIP:
			return runStorageBansAddIP(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, target, reason, duration, permanent)
		case cmdUseUser:
			return runStorageBansAddUser(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, target, reason, duration, permanent)
		default:
			return fmt.Errorf("unknown command %q", use)
		}
	}
}

func runStorageBansAddIP(ctx context.Context, w io.Writer, store storage.Provider, target, reason string, duration time.Duration, permanent bool) (err error) {
	// TODO: Check for existing ban and revoke it?
	ip := net.ParseIP(target)

	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", target)
	}

	ban := &model.BannedIP{
		IP:     model.NewIP(ip),
		Source: "cli",
	}

	if reason != "" {
		ban.Reason = sql.NullString{Valid: true, String: reason}
	}

	if !permanent {
		ban.Expires = sql.NullTime{Valid: true, Time: time.Now().Add(duration)}
	}

	if err = store.SaveBannedIP(ctx, ban); err != nil {
		return err
	}

	if permanent {
		_, _ = fmt.Fprintf(w, "Successfully banned IP '%s' permanently.\n", ban.IP)
	} else {
		_, _ = fmt.Fprintf(w, "Successfully banned IP '%s' until '%s'.\n", ban.IP, time.Now().Add(duration).Format(time.RFC3339))
	}

	return nil
}

func runStorageBansAddUser(ctx context.Context, w io.Writer, store storage.Provider, target, reason string, duration time.Duration, permanent bool) (err error) {
	// TODO: Check for existing ban and revoke it?
	ban := &model.BannedUser{
		Username: target,
		Source:   "cli",
	}

	if reason != "" {
		ban.Reason = sql.NullString{Valid: true, String: reason}
	}

	if !permanent {
		ban.Expires = sql.NullTime{Valid: true, Time: time.Now().Add(duration)}
	}

	if err = store.SaveBannedUser(ctx, ban); err != nil {
		return err
	}

	if permanent {
		_, _ = fmt.Fprintf(w, "Successfully banned user '%s' permanently.\n", ban.Username)
	} else {
		_, _ = fmt.Fprintf(w, "Successfully banned user '%s' until '%s'.\n", ban.Username, time.Now().Add(duration).Format(time.RFC3339))
	}

	return nil
}

func (ctx *CmdCtx) StorageUserWebAuthnExportRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		filename string
	)

	if filename, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	switch _, err = os.Stat(filename); {
	case err == nil:
		return fmt.Errorf("must specify a file that doesn't exist but '%s' exists", filename)
	case !os.IsNotExist(err):
		return fmt.Errorf("error occurred opening '%s': %w", filename, err)
	}

	return runStorageUserWebAuthnExport(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, filename)
}

func runStorageUserWebAuthnExport(ctx context.Context, w io.Writer, store storage.Provider, filename string) (err error) {
	limit := 10
	count := 0

	var (
		credentials []model.WebAuthnCredential
	)

	export := &model.WebAuthnCredentialExport{
		WebAuthnCredentials: []model.WebAuthnCredential{},
	}

	for page := 0; true; page++ {
		if credentials, err = store.LoadWebAuthnCredentials(ctx, limit, page); err != nil {
			return err
		}

		export.WebAuthnCredentials = append(export.WebAuthnCredentials, credentials...)

		l := len(credentials)

		count += l

		if l < limit {
			break
		}
	}

	if len(export.WebAuthnCredentials) == 0 {
		return fmt.Errorf("no data to export")
	}

	var f *os.File

	if f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if err = exportYAMLWithJSONSchema(f, "export.webauthn", export); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	_, _ = fmt.Fprintf(w, cliOutputFmtSuccessfulUserExportFile, count, "WebAuthn credentials", "YAML", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserWebAuthnImportRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	return runStorageUserWebAuthnImport(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, args[0])
}

func runStorageUserWebAuthnImport(ctx context.Context, w io.Writer, store storage.Provider, filename string) (err error) {
	var (
		stat os.FileInfo
		data []byte
	)

	if stat, err = os.Stat(filename); err != nil {
		return fmt.Errorf("must specify a filename that exists but '%s' had an error opening it: %w", filename, err)
	}

	if stat.IsDir() {
		return fmt.Errorf("must specify a filename that exists but '%s' is a directory", filename)
	}

	if data, err = os.ReadFile(filename); err != nil {
		return err
	}

	export := &model.WebAuthnCredentialExport{}

	if err = yaml.Unmarshal(data, export); err != nil {
		return err
	}

	if len(export.WebAuthnCredentials) == 0 {
		return fmt.Errorf("can't import a YAML file without WebAuthn credentials data")
	}

	for _, credential := range export.WebAuthnCredentials {
		if err = store.SaveWebAuthnCredential(ctx, credential); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(w, cliOutputFmtSuccessfulUserImportFile, len(export.WebAuthnCredentials), "WebAuthn credentials", "YAML", filename)

	return nil
}

// StorageUserWebAuthnListRunE is the RunE for the authelia storage user webauthn list command.
func (ctx *CmdCtx) StorageUserWebAuthnListRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if len(args) == 0 || args[0] == "" {
		return runStorageUserWebAuthnListAll(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider)
	}

	return runStorageUserWebAuthnList(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, args[0])
}

func runStorageUserWebAuthnList(ctx context.Context, w io.Writer, store storage.Provider, user string) (err error) {
	var credentials []model.WebAuthnCredential

	credentials, err = store.LoadWebAuthnCredentialsByUsername(ctx, "", user)

	switch {
	case len(credentials) == 0 || (err != nil && errors.Is(err, storage.ErrNoWebAuthnCredential)):
		return fmt.Errorf("user '%s' has no WebAuthn credentials", user)
	case err != nil:
		return fmt.Errorf("can't list credentials for user '%s': %w", user, err)
	default:
		_, _ = fmt.Fprintf(w, "WebAuthn Credentials for user '%s':\n\n", user)

		tw := tabwriter.NewWriter(w, 1, 1, 4, ' ', 0)

		_, _ = fmt.Fprintln(tw, "ID\tKID\tDescription")

		for _, credential := range credentials {
			_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\n", credential.ID, credential.KID, credential.Description)
		}

		return tw.Flush()
	}
}

// StorageUserWebAuthnListAllRunE is the RunE for the authelia storage user webauthn list command when no args are specified.
func (ctx *CmdCtx) StorageUserWebAuthnListAllRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	return runStorageUserWebAuthnListAll(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider)
}

func runStorageUserWebAuthnListAll(ctx context.Context, w io.Writer, store storage.Provider) (err error) {
	var credentials []model.WebAuthnCredential

	limit := 10

	_, _ = fmt.Fprintf(w, "WebAuthn Credentials:\n\n")

	tw := tabwriter.NewWriter(w, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(tw, "ID\tRPID\tKID\tDescription\tUsername")

	for page := 0; true; page++ {
		if credentials, err = store.LoadWebAuthnCredentials(ctx, limit, page); err != nil {
			return fmt.Errorf("failed to list credentials: %w", err)
		}

		if page == 0 && len(credentials) == 0 {
			return errors.New("no WebAuthn credentials in database")
		}

		for _, credential := range credentials {
			_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", credential.ID, credential.RPID, credential.KID, credential.Description, credential.Username)
		}

		if len(credentials) < limit {
			break
		}
	}

	return tw.Flush()
}

// StorageUserWebAuthnVerifyRunE is the RunE for the authelia storage user webauthn verify command when no args are specified.
func (ctx *CmdCtx) StorageUserWebAuthnVerifyRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	return runStorageUserWebAuthnVerify(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, ctx.config)
}

func runStorageUserWebAuthnVerify(ctx context.Context, w io.Writer, store storage.Provider, config *schema.Configuration) (err error) {
	var (
		provider    webauthn.MetaDataProvider
		credentials []model.WebAuthnCredential
	)

	if provider, err = webauthn.NewMetaDataProvider(config, store); err != nil {
		return err
	}

	limit := 10

	_, _ = fmt.Fprintf(w, "WebAuthn Credential Verifications:\n\n")

	tw := tabwriter.NewWriter(w, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(tw, "ID\tRPID\tKID\tUsername\tAAGUID\tStatement\tBackup\tMDS")

	for page := 0; true; page++ {
		if credentials, err = store.LoadWebAuthnCredentials(ctx, limit, page); err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}

		if page == 0 && len(credentials) == 0 {
			return errors.New("no WebAuthn credentials in database")
		}

		for _, credential := range credentials {
			result := webauthn.VerifyCredential(&config.WebAuthn, &credential, provider)

			strAAGUID, strStatement, strBackup, strMDS := wordYes, wordYes, wordYes, wordYes

			if result.IsProhibitedAAGUID {
				strAAGUID = wordNo
			}

			if result.MissingStatement {
				strStatement = wordNo
			}

			if result.IsProhibitedBackupEligibility {
				strBackup = wordNo
			}

			if result.Malformed {
				strMDS = "Malformed"
			} else if result.MetaDataValidationError {
				strMDS = wordNo
			}

			_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", credential.ID, credential.RPID, credential.KID, credential.Username, strAAGUID, strStatement, strBackup, strMDS)
		}

		if len(credentials) < limit {
			break
		}
	}

	return tw.Flush()
}

// StorageUserWebAuthnDeleteRunE is the RunE for the authelia storage user webauthn delete command.
func (ctx *CmdCtx) StorageUserWebAuthnDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		all, byKID             bool
		description, kid, user string
	)

	if all, byKID, description, kid, user, err = storageWebAuthnDeleteRunEOptsFromFlags(cmd.Flags(), args); err != nil {
		return err
	}

	return runStorageUserWebAuthnDelete(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, all, byKID, description, kid, user)
}

func runStorageUserWebAuthnDelete(ctx context.Context, w io.Writer, store storage.Provider, all, byKID bool, description, kid, user string) (err error) {
	if byKID {
		if err = store.DeleteWebAuthnCredential(ctx, kid); err != nil {
			return fmt.Errorf("failed to delete WebAuthn credential with kid '%s': %w", kid, err)
		}

		_, _ = fmt.Fprintf(w, "Successfully deleted WebAuthn credential with key id '%s'\n", kid)
	} else {
		err = store.DeleteWebAuthnCredentialByUsername(ctx, user, description)

		if all {
			if err != nil {
				return fmt.Errorf("failed to delete all WebAuthn credentials with username '%s': %w", user, err)
			}

			_, _ = fmt.Fprintf(w, "Successfully deleted all WebAuthn credentials for user '%s'\n", user)
		} else {
			if err != nil {
				return fmt.Errorf("failed to delete WebAuthn credential with username '%s' and description '%s': %w", user, description, err)
			}

			_, _ = fmt.Fprintf(w, "Successfully deleted WebAuthn credential with description '%s' for user '%s'\n", description, user)
		}
	}

	return nil
}

// StorageUserTOTPGenerateRunE is the RunE for the authelia storage user totp generate command.
func (ctx *CmdCtx) StorageUserTOTPGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	var (
		force            bool
		filename, secret string
	)

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if force, filename, secret, err = storageTOTPGenerateRunEOptsFromFlags(cmd.Flags()); err != nil {
		return err
	}

	return runStorageUserTOTPGenerate(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, ctx.config, filename, args[0], secret, force)
}

func runStorageUserTOTPGenerate(ctx context.Context, w io.Writer, store storage.Provider, config *schema.Configuration, filename, username, secret string, force bool) (err error) {
	var (
		c    *model.TOTPConfiguration
		file *os.File
		img  image.Image
	)

	if _, err = store.LoadTOTPConfiguration(ctx, username); err == nil && !force {
		return fmt.Errorf("%s already has a TOTP configuration, use --force to overwrite", username)
	} else if err != nil && !errors.Is(err, storage.ErrNoTOTPConfiguration) {
		return err
	}

	totpProvider := totp.NewTimeBasedProvider(config.TOTP)

	if c, err = totpProvider.GenerateCustom(totp.NewContext(ctx, &clock.Real{}, &random.Cryptographical{}), username, config.TOTP.DefaultAlgorithm, secret, uint32(config.TOTP.DefaultDigits), uint(config.TOTP.DefaultPeriod), uint(config.TOTP.SecretSize)); err != nil { //nolint:gosec // Validated at runtime.
		return err
	}

	extraInfo := ""

	if filename != "" {
		if _, err = os.Stat(filename); !os.IsNotExist(err) {
			return errors.New("image output filepath already exists")
		}

		if file, err = os.Create(filename); err != nil {
			return err
		}

		defer func() {
			if err := file.Close(); err != nil {
				panic(err)
			}
		}()

		if img, err = c.Image(256, 256); err != nil {
			return err
		}

		if err = png.Encode(file, img); err != nil {
			return err
		}

		extraInfo = fmt.Sprintf(" and saved it as a PNG image at the path '%s'", filename)
	}

	if err = store.SaveTOTPConfiguration(ctx, *c); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Successfully generated TOTP configuration for user '%s' with URI '%s'%s\n", username, c.URI(), extraInfo)

	return nil
}

// StorageUserTOTPDeleteRunE is the RunE for the authelia storage user totp delete command.
func (ctx *CmdCtx) StorageUserTOTPDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	user := args[0]

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if _, err = ctx.providers.StorageProvider.LoadTOTPConfiguration(ctx, user); err != nil {
		return fmt.Errorf("failed to delete TOTP configuration for user '%s': %+v", user, err)
	}

	if err = ctx.providers.StorageProvider.DeleteTOTPConfiguration(ctx, user); err != nil {
		return fmt.Errorf("failed to delete TOTP configuration for user '%s': %+v", user, err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted TOTP configuration for user '%s'\n", user)

	return nil
}

// StorageUserTOTPExportRunE is the RunE for the authelia storage user totp export command.
func (ctx *CmdCtx) StorageUserTOTPExportRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		filename string
	)

	if filename, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	switch _, err = os.Stat(filename); {
	case err == nil:
		return fmt.Errorf("must specify a file that doesn't exist but '%s' exists", filename)
	case !os.IsNotExist(err):
		return fmt.Errorf("error occurred opening '%s': %w", filename, err)
	}

	return runStorageUserTOTPExport(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, filename)
}

func runStorageUserTOTPExport(ctx context.Context, w io.Writer, store storage.Provider, filename string) (err error) {
	limit := 10
	count := 0

	var (
		configs []model.TOTPConfiguration
	)

	export := &model.TOTPConfigurationExport{}

	for page := 0; true; page++ {
		if configs, err = store.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		export.TOTPConfigurations = append(export.TOTPConfigurations, configs...)

		l := len(configs)

		count += l

		if l < limit {
			break
		}
	}

	if len(export.TOTPConfigurations) == 0 {
		return fmt.Errorf("no data to export")
	}

	var f *os.File

	if f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if err = exportYAMLWithJSONSchema(f, "export.totp", export); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	_, _ = fmt.Fprintf(w, cliOutputFmtSuccessfulUserExportFile, count, "TOTP configurations", "YAML", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPImportRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	return runStorageUserTOTPImport(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, args[0])
}

func runStorageUserTOTPImport(ctx context.Context, w io.Writer, store storage.Provider, filename string) (err error) {
	var (
		stat os.FileInfo
		data []byte
	)

	if stat, err = os.Stat(filename); err != nil {
		return fmt.Errorf("must specify a filename that exists but '%s' had an error opening it: %w", filename, err)
	}

	if stat.IsDir() {
		return fmt.Errorf("must specify a filename that exists but '%s' is a directory", filename)
	}

	if data, err = os.ReadFile(filename); err != nil {
		return err
	}

	export := &model.TOTPConfigurationExport{}

	if err = yaml.Unmarshal(data, export); err != nil {
		return err
	}

	if len(export.TOTPConfigurations) == 0 {
		return fmt.Errorf("can't import a YAML file without TOTP configuration data")
	}

	for _, config := range export.TOTPConfigurations {
		if err = store.SaveTOTPConfiguration(ctx, config); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(w, cliOutputFmtSuccessfulUserImportFile, len(export.TOTPConfigurations), "TOTP configurations", "YAML", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPExportURIRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	return runStorageUserTOTPExportURI(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider)
}

func runStorageUserTOTPExportURI(ctx context.Context, w io.Writer, store storage.Provider) (err error) {
	limit := 10
	count := 0

	buf := new(bytes.Buffer)

	var (
		configs []model.TOTPConfiguration
	)

	for page := 0; true; page++ {
		if configs, err = store.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		for _, c := range configs {
			_, _ = fmt.Fprintf(buf, "%s\n", c.URI())
		}

		l := len(configs)

		count += l

		if l < limit {
			break
		}
	}

	_, _ = buf.WriteTo(w)

	buf.Reset()

	_, _ = fmt.Fprintf(w, "\n\nSuccessfully exported %d TOTP configurations as TOTP URI's and printed them to the console\n", count)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPExportCSVRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var filename string
	if filename, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	return runStorageUserTOTPExportCSV(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, filename, 10)
}

func runStorageUserTOTPExportCSV(ctx context.Context, w io.Writer, store storage.Provider, filename string, limit int) (err error) {
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("must specify a filename to export to")
	}

	count := 0

	buf := new(bytes.Buffer)

	buf.WriteString("issuer,username,algorithm,digits,period,secret\n")

	var configs []model.TOTPConfiguration

	for page := 0; true; page++ {
		if configs, err = store.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		for _, c := range configs {
			_, _ = fmt.Fprintf(buf, "%s,%s,%s,%d,%d,%s\n", c.Issuer, c.Username, c.Algorithm, c.Digits, c.Period, string(c.Secret))
		}

		l := len(configs)

		count += l

		if l < limit {
			break
		}
	}

	if err = os.WriteFile(filename, buf.Bytes(), 0600); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, cliOutputFmtSuccessfulUserExportFile, count, "TOTP configurations", "CSV", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPExportPNGRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var dir string
	if dir, err = cmd.Flags().GetString(cmdFlagNameDirectory); err != nil {
		return err
	}

	return runStorageUserTOTPExportPNG(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, dir)
}

func runStorageUserTOTPExportPNG(ctx context.Context, w io.Writer, store storage.Provider, dir string) (err error) {
	if dir == "" {
		rand := &random.Cryptographical{}
		dir = rand.StringCustom(8, random.CharSetAlphaNumeric)
	}

	if _, err = os.Stat(dir); !os.IsNotExist(err) {
		return errors.New("output directory must not exist")
	}

	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	limit := 10
	count := 0

	var (
		file    *os.File
		configs []model.TOTPConfiguration
		img     image.Image
	)

	for page := 0; true; page++ {
		if configs, err = store.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		for _, c := range configs {
			if file, err = os.Create(filepath.Join(dir, fmt.Sprintf("%s.png", c.Username))); err != nil {
				return err
			}

			if img, err = c.Image(256, 256); err != nil {
				_ = file.Close()

				return err
			}

			if err = png.Encode(file, img); err != nil {
				_ = file.Close()

				return err
			}

			_ = file.Close()
		}

		l := len(configs)

		count += l

		if l < limit {
			break
		}
	}

	_, _ = fmt.Fprintf(w, "Successfully exported %d TOTP configuration as QR codes in PNG format to the '%s' directory\n", count, dir)

	return nil
}

// StorageUserIdentifiersExportRunE is the RunE for the authelia storage user identifiers export command.
func (ctx *CmdCtx) StorageUserIdentifiersExportRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		filename string
	)

	if filename, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	return runStorageUserIdentifiersExport(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, filename)
}

func runStorageUserIdentifiersExport(ctx context.Context, w io.Writer, store storage.Provider, filename string) (err error) {
	switch _, err = os.Stat(filename); {
	case err == nil:
		return fmt.Errorf("must specify a file that doesn't exist but '%s' exists", filename)
	case !os.IsNotExist(err):
		return fmt.Errorf("error occurred opening '%s': %w", filename, err)
	}

	export := &model.UserOpaqueIdentifiersExport{
		Identifiers: nil,
	}

	if export.Identifiers, err = store.LoadUserOpaqueIdentifiers(ctx); err != nil {
		return err
	}

	if len(export.Identifiers) == 0 {
		return fmt.Errorf("no data to export")
	}

	var f *os.File

	if f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if err = exportYAMLWithJSONSchema(f, "export.identifiers", export); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	_, _ = fmt.Fprintf(w, cliOutputFmtSuccessfulUserExportFile, len(export.Identifiers), "User Opaque Identifiers", "YAML", filename)

	return nil
}

// StorageUserIdentifiersImportRunE is the RunE for the authelia storage user identifiers import command.
func (ctx *CmdCtx) StorageUserIdentifiersImportRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	return runStorageUserIdentifiersImport(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, args[0])
}

func runStorageUserIdentifiersImport(ctx context.Context, w io.Writer, store storage.Provider, filename string) (err error) {
	var (
		stat os.FileInfo
		data []byte
	)

	if stat, err = os.Stat(filename); err != nil {
		return fmt.Errorf("must specify a file that exists but '%s' had an error opening it: %w", filename, err)
	}

	if stat.IsDir() {
		return fmt.Errorf("must specify a file that exists but '%s' is a directory", filename)
	}

	if data, err = os.ReadFile(filename); err != nil {
		return err
	}

	export := &model.UserOpaqueIdentifiersExport{}

	if err = yaml.Unmarshal(data, export); err != nil {
		return err
	}

	if len(export.Identifiers) == 0 {
		return fmt.Errorf("can't import a YAML file without User Opaque Identifiers data")
	}

	for _, opaqueID := range export.Identifiers {
		if err = store.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(w, cliOutputFmtSuccessfulUserImportFile, len(export.Identifiers), "User Opaque Identifiers", "YAML", filename)

	return nil
}

// StorageUserIdentifiersGenerateRunE is the RunE for the authelia storage user identifiers generate command.
func (ctx *CmdCtx) StorageUserIdentifiersGenerateRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		users, services, sectors []string
	)

	if users, services, sectors, err = flagsGetUserIdentifiersGenerateOptions(cmd.Flags()); err != nil {
		return err
	}

	return runStorageUserIdentifiersGenerate(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, users, services, sectors)
}

func runStorageUserIdentifiersGenerate(ctx context.Context, w io.Writer, store storage.Provider, users, services, sectors []string) (err error) {
	if len(users) == 0 {
		return fmt.Errorf("must supply at least one user")
	}

	if !utils.IsStringSliceContainsAll(services, validIdentifierServices) {
		return fmt.Errorf("one or more the service names '%s' is invalid, the valid values are: '%s'", strings.Join(services, "', '"), strings.Join(validIdentifierServices, "', '"))
	}

	var identifiers []model.UserOpaqueIdentifier

	if identifiers, err = store.LoadUserOpaqueIdentifiers(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("can't load the existing identifiers: %w", err)
	}

	if len(sectors) == 0 {
		sectors = append(sectors, "")
	}

	var added, duplicates int

	for _, service := range services {
		for _, sector := range sectors {
			for _, username := range users {
				identifier := model.UserOpaqueIdentifier{
					Service:  service,
					SectorID: sector,
					Username: username,
				}

				if containsIdentifier(identifier, identifiers) {
					duplicates++

					continue
				}

				identifier.Identifier, err = uuid.NewRandom()
				if err != nil {
					return fmt.Errorf("failed to generate a uuid: %w", err)
				}

				if err = store.SaveUserOpaqueIdentifier(ctx, identifier); err != nil {
					return fmt.Errorf("failed to save identifier: %w", err)
				}

				added++
			}
		}
	}

	_, _ = fmt.Fprintf(w, "Successfully generated and added opaque identifiers:\n")
	_, _ = fmt.Fprintf(w, "\tUsers: '%s'\n", strings.Join(users, "', '"))
	_, _ = fmt.Fprintf(w, "\tSectors: '%s'\n", strings.Join(sectors, "', '"))
	_, _ = fmt.Fprintf(w, "\tServices: '%s'\n", strings.Join(services, "', '"))

	if duplicates != 0 {
		_, _ = fmt.Fprintf(w, "\tSkipped Duplicates: %d\n", duplicates)
	}

	_, _ = fmt.Fprintf(w, "\tTotal: %d", added)

	return nil
}

// StorageUserIdentifiersAddRunE is the RunE for the authelia storage user identifiers add command.
func (ctx *CmdCtx) StorageUserIdentifiersAddRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		service, sector, identifier string
	)

	if service, err = cmd.Flags().GetString(cmdFlagNameService); err != nil {
		return err
	}

	if sector, err = cmd.Flags().GetString(cmdFlagNameSector); err != nil {
		return err
	}

	if identifier, err = cmd.Flags().GetString(cmdFlagNameIdentifier); err != nil {
		return err
	}

	return runStorageUserIdentifiersAdd(ctx, cmd.OutOrStdout(), ctx.providers.StorageProvider, service, sector, args[0], identifier, cmd.Flags().Changed(cmdFlagNameIdentifier))
}

func runStorageUserIdentifiersAdd(ctx context.Context, w io.Writer, store storage.Provider, service, sector, username, identifier string, useIdentifier bool) (err error) {
	if service == "" {
		service = identifierServiceOpenIDConnect
	} else if !utils.IsStringInSlice(service, validIdentifierServices) {
		return fmt.Errorf("the service name '%s' is invalid, the valid values are: '%s'", service, strings.Join(validIdentifierServices, "', '"))
	}

	opaqueID := model.UserOpaqueIdentifier{
		Service:  service,
		Username: username,
		SectorID: sector,
	}

	if useIdentifier {
		if opaqueID.Identifier, err = uuid.Parse(identifier); err != nil {
			return fmt.Errorf("the identifier provided '%s' is invalid as it must be a version 4 UUID but parsing it had an error: %w", identifier, err)
		}

		if opaqueID.Identifier.Version() != 4 {
			return fmt.Errorf("the identifier provided '%s' is a version %d UUID but only version 4 UUID's accepted as identifiers", identifier, opaqueID.Identifier.Version())
		}
	} else {
		if opaqueID.Identifier, err = uuid.NewRandom(); err != nil {
			return err
		}
	}

	if err = store.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Added User Opaque Identifier:\n\tService: %s\n\tSector: %s\n\tUsername: %s\n\tIdentifier: %s\n\n", opaqueID.Service, opaqueID.SectorID, opaqueID.Username, opaqueID.Identifier)

	return nil
}
