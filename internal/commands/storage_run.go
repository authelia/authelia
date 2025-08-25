package commands

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/png"
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
			_ = ctx.providers.StorageProvider.Close()
		}()

		if err = ctx.CheckSchema(); err != nil {
			return storageWrapCheckSchemaErr(err)
		}

		if err = ctx.providers.StorageProvider.DeleteCachedData(ctx, name); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(os.Stdout, "Successfully deleted cached %s data.\n", description)

		return nil
	}
}

func (ctx *CmdCtx) StorageCacheMDS3StatusRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if !ctx.config.WebAuthn.Metadata.Enabled {
		return fmt.Errorf("webauthn metadata is disabled")
	}

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	provider, err := webauthn.NewMetaDataProvider(ctx.config, ctx.providers.StorageProvider)
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

	_, _ = fmt.Fprintf(os.Stdout, "WebAuthn MDS3 Cache Status:\n\n\tValid: %t\n\tInitialized: %t\n\tOutdated: %t\n", valid, initialized, outdated)

	if initialized {
		_, _ = fmt.Fprintf(os.Stdout, "\tVersion: %d\n", mds.Parsed.Number)

		if !outdated {
			_, _ = fmt.Fprintf(os.Stdout, "\tNext Update: %s\n", mds.Parsed.NextUpdate.Format("January 2, 2006"))
		}
	}

	return nil
}

func (ctx *CmdCtx) StorageCacheMDS3DumpRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if !ctx.config.WebAuthn.Metadata.Enabled {
		return fmt.Errorf("webauthn metadata is disabled")
	}

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	provider, err := webauthn.NewMetaDataProvider(ctx.config, ctx.providers.StorageProvider)
	if err != nil {
		return err
	}

	var (
		file *os.File
		mds  *metadata.Metadata
		data []byte
		path string
	)

	if path, err = cmd.Flags().GetString("path"); err != nil {
		return err
	}

	if mds, data, err = provider.LoadCache(ctx); err != nil {
		return err
	} else if mds == nil {
		return fmt.Errorf("error dumping metadata: no metadata is in the cache")
	}

	if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return err
	}

	defer file.Close()

	if _, err = file.Write(data); err != nil {
		return fmt.Errorf("error writing data to file: %w", err)
	}

	_ = file.Sync()

	_, _ = fmt.Fprintf(os.Stdout, "Successfully dumped WebAuthn MDS3 data with version %d from cache to file '%s'.\n", mds.Parsed.Number, path)

	return nil
}

//nolint:gocyclo
func (ctx *CmdCtx) StorageCacheMDS3UpdateRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if !ctx.config.WebAuthn.Metadata.Enabled {
		return fmt.Errorf("webauthn metadata is disabled")
	}

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	provider, err := webauthn.NewMetaDataProvider(ctx.config, ctx.providers.StorageProvider)
	if err != nil {
		return err
	}

	var (
		mds   *metadata.Metadata
		data  []byte
		path  string
		force bool
	)

	if force, err = cmd.Flags().GetBool(cmdFlagNameForce); err != nil {
		return err
	}

	if path, err = cmd.Flags().GetString(cmdFlagNamePath); err != nil {
		return err
	}

	if mds, _, err = provider.LoadCache(ctx); err != nil {
		return err
	} else if mds != nil && !force && !provider.Outdated() {
		_, _ = fmt.Fprintf(os.Stdout, "WebAuthn MDS3 cache data with version %d due for update on %s does not require an update.\n", mds.Parsed.Number, mds.Parsed.NextUpdate.Format("January 2, 2006"))

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
		_, _ = fmt.Fprintf(os.Stdout, "Provided WebAuthn MDS3 data with version %d was due for update on %s and can't be used.\n", mds.Parsed.Number, mds.Parsed.NextUpdate.Format("January 2, 2006"))
	}

	if err = provider.SaveCache(ctx, data); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "WebAuthn MDS3 cache data updated to version %d and is due for update on %s.\n", mds.Parsed.Number, mds.Parsed.NextUpdate.Format("January 2, 2006"))

	return nil
}

func (ctx *CmdCtx) StorageSchemaEncryptionCheckRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		verbose bool
		result  storage.EncryptionValidationResult
	)

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if verbose, err = cmd.Flags().GetBool(cmdFlagNameVerbose); err != nil {
		return err
	}

	if result, err = ctx.providers.StorageProvider.SchemaEncryptionCheckKey(ctx, verbose); err != nil {
		switch {
		case errors.Is(err, storage.ErrSchemaEncryptionVersionUnsupported):
			fmt.Printf("Storage Encryption Key Validation: FAILURE\n\n\tCause: The schema version doesn't support encryption.\n")
		default:
			fmt.Printf("Storage Encryption Key Validation: UNKNOWN\n\n\tCause: %v.\n", err)
		}
	} else {
		if result.Success() {
			fmt.Println("Storage Encryption Key Validation: SUCCESS")
		} else {
			fmt.Printf("Storage Encryption Key Validation: FAILURE\n\n\tCause: %v.\n", storage.ErrSchemaEncryptionInvalidKey)
		}

		if verbose {
			fmt.Printf("\nTables:")

			tables := make([]string, 0, len(result.Tables))

			for name := range result.Tables {
				tables = append(tables, name)
			}

			sort.Strings(tables)

			for _, name := range tables {
				table := result.Tables[name]

				fmt.Printf("\n\n\tTable (%s): %s\n\t\tInvalid Rows: %d\n\t\tTotal Rows: %d", name, table.ResultDescriptor(), table.Invalid, table.Total)
			}

			fmt.Printf("\n")
		}
	}

	return nil
}

// StorageSchemaEncryptionChangeKeyRunE is the RunE for the authelia storage encryption change-key command.
func (ctx *CmdCtx) StorageSchemaEncryptionChangeKeyRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		key     string
		version int
	)

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if version, err = ctx.providers.StorageProvider.SchemaVersion(ctx); err != nil {
		return err
	}

	if version <= 0 {
		return errors.New("schema version must be at least version 1 to change the encryption key")
	}

	useFlag := cmd.Flags().Changed(cmdFlagNameNewEncryptionKey)
	if useFlag {
		if key, err = cmd.Flags().GetString(cmdFlagNameNewEncryptionKey); err != nil {
			return err
		}
	}

	if !useFlag || key == "" {
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

	if err = ctx.providers.StorageProvider.SchemaEncryptionChangeKey(ctx, key); err != nil {
		return err
	}

	fmt.Println("Completed the encryption key change. Please adjust your configuration to use the new key.")

	return nil
}

// StorageMigrateHistoryRunE is the RunE for the authelia storage migrate history command.
func (ctx *CmdCtx) StorageMigrateHistoryRunE(_ *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		version    int
		migrations []model.Migration
	)

	if version, err = ctx.providers.StorageProvider.SchemaVersion(ctx); err != nil {
		return err
	}

	if version <= 0 {
		fmt.Println("No migration history is available for schemas that not version 1 or above.")
		return
	}

	if migrations, err = ctx.providers.StorageProvider.SchemaMigrationHistory(ctx); err != nil {
		return err
	}

	if len(migrations) == 0 {
		return errors.New("no migration history found which may indicate a broken schema")
	}

	fmt.Printf("Migration History:\n\n")

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(w, "ID\tDate\tBefore\tAfter\tAuthelia Version")

	for _, m := range migrations {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%s\n", m.ID, m.Applied.Format("2006-01-02 15:04:05 -0700"), m.Before, m.After, m.Version)
	}

	return w.Flush()
}

// NewStorageMigrateListRunE creates the RunE for the authelia storage migrate list command.
func (ctx *CmdCtx) NewStorageMigrateListRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			_ = ctx.providers.StorageProvider.Close()
		}()

		var (
			migrations   []model.SchemaMigration
			directionStr string
		)

		if up {
			migrations, err = ctx.providers.StorageProvider.SchemaMigrationsUp(ctx, 0)
			directionStr = "Up"
		} else {
			migrations, err = ctx.providers.StorageProvider.SchemaMigrationsDown(ctx, 0)
			directionStr = "Down"
		}

		if err != nil && !errors.Is(err, storage.ErrNoAvailableMigrations) && !errors.Is(err, storage.ErrMigrateCurrentVersionSameAsTarget) {
			return err
		}

		if len(migrations) == 0 {
			fmt.Printf("Storage Schema Migration List (%s)\n\nNo Migrations Available\n", directionStr)
		} else {
			fmt.Printf("Storage Schema Migration List (%s)\n\n", directionStr)

			w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)

			_, _ = fmt.Fprintln(w, "Version\tDescription")

			for _, migration := range migrations {
				_, _ = fmt.Fprintf(w, "%d\t%s\n", migration.Version, migration.Name)
			}

			return w.Flush()
		}

		return nil
	}
}

// NewStorageMigrationRunE creates the RunE for the authelia storage migrate command.
func (ctx *CmdCtx) NewStorageMigrationRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			_ = ctx.providers.StorageProvider.Close()
		}()

		var (
			target int
		)

		if target, err = cmd.Flags().GetInt(cmdFlagNameTarget); err != nil {
			return err
		}

		switch {
		case up:
			switch cmd.Flags().Changed(cmdFlagNameTarget) {
			case true:
				return ctx.providers.StorageProvider.SchemaMigrate(ctx, true, target)
			default:
				return ctx.providers.StorageProvider.SchemaMigrate(ctx, true, storage.SchemaLatest)
			}
		default:
			if !cmd.Flags().Changed(cmdFlagNameTarget) {
				return errors.New("you must set a target version")
			}

			var confirmed bool

			if confirmed, err = termReadConfirmation(cmd.Flags(), cmdFlagNameDestroyData, "Schema Down Migrations may DESTROY data, type 'DESTROY' and press return to continue: ", "DESTROY"); err != nil {
				return err
			}

			if !confirmed {
				return errors.New("cancelling down migration due to user not accepting data destruction")
			}

			return ctx.providers.StorageProvider.SchemaMigrate(ctx, false, target)
		}
	}
}

// StorageSchemaInfoRunE is the RunE for the authelia storage schema info command.
func (ctx *CmdCtx) StorageSchemaInfoRunE(_ *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		upgradeStr, tablesStr string

		tables          []string
		version, latest int
	)

	if version, err = ctx.providers.StorageProvider.SchemaVersion(ctx); err != nil && err.Error() != "unknown schema state" {
		return err
	}

	if tables, err = ctx.providers.StorageProvider.SchemaTables(ctx); err != nil {
		return err
	}

	if len(tables) == 0 {
		tablesStr = "N/A"
	} else {
		tablesStr = strings.Join(tables, ", ")
	}

	if latest, err = ctx.providers.StorageProvider.SchemaLatestVersion(); err != nil {
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

	switch result, err = ctx.providers.StorageProvider.SchemaEncryptionCheckKey(ctx, false); {
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

	fmt.Printf("Schema Version: %s\nSchema Upgrade Available: %s\nSchema Tables: %s\nSchema Encryption Key: %s\n", storage.SchemaVersionToString(version), upgradeStr, tablesStr, encryption)

	return nil
}

func (ctx *CmdCtx) StorageBansListRunE(use string) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			_ = ctx.providers.StorageProvider.Close()
		}()

		if err = ctx.CheckSchema(); err != nil {
			return storageWrapCheckSchemaErr(err)
		}

		switch use {
		case cmdUseIP:
			var results []model.BannedIP

			limit := 10
			count := 0

			for page := 0; true; page++ {
				var bans []model.BannedIP

				if bans, err = ctx.providers.StorageProvider.LoadBannedIPs(context.Background(), limit, page); err != nil {
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
				fmt.Printf("No results.\n")

				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

			_, _ = fmt.Fprintln(w, "ID\tIP\tExpires\tSource\tReason")

			for _, ban := range results {
				_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", ban.ID, ban.IP, regulation.FormatExpiresShort(ban.Expires), ban.Source, ban.Reason.String)
			}

			return w.Flush()
		case cmdUseUser:
			var results []model.BannedUser

			limit := 10
			count := 0

			for page := 0; true; page++ {
				var bans []model.BannedUser

				if bans, err = ctx.providers.StorageProvider.LoadBannedUsers(context.Background(), limit, page); err != nil {
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
				fmt.Printf("No results.\n")

				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

			_, _ = fmt.Fprintln(w, "ID\tUsername\tExpires\tSource\tReason")

			for _, ban := range results {
				_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", ban.ID, ban.Username, regulation.FormatExpiresShort(ban.Expires), ban.Source, ban.Reason.String)
			}

			return w.Flush()
		default:
			return fmt.Errorf("unknown command %q", use)
		}
	}
}

//nolint:gocyclo
func (ctx *CmdCtx) StorageBansRevokeRunE(use string) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			_ = ctx.providers.StorageProvider.Close()
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
			ip := net.ParseIP(target)

			var bans []model.BannedIP

			if id == 0 {
				if bans, err = ctx.providers.StorageProvider.LoadBannedIP(ctx, model.NewIP(ip)); err != nil {
					return err
				}
			} else {
				var ban model.BannedIP

				if ban, err = ctx.providers.StorageProvider.LoadBannedIPByID(ctx, id); err != nil {
					return err
				}

				bans = []model.BannedIP{ban}
			}

			for _, ban := range bans {
				if ban.Revoked {
					fmt.Printf("SKIPPED\tIP ban with id '%d' for '%s is already revoked.\n", ban.ID, ban.IP)
				} else {
					if err = ctx.providers.StorageProvider.RevokeBannedIP(ctx, ban.ID, time.Now()); err != nil {
						fmt.Printf("ERROR\tIP ban with id '%d' for '%s' had error when being revoked: %+v\n", ban.ID, ban.IP, err)
					} else {
						fmt.Printf("REVOKED\tIP ban with id '%d' for '%s' has been revoked\n", ban.ID, ban.IP)
					}
				}
			}
		case cmdUseUser:
			var bans []model.BannedUser

			if id == 0 {
				if bans, err = ctx.providers.StorageProvider.LoadBannedUser(ctx, target); err != nil {
					return err
				}
			} else {
				var ban model.BannedUser

				if ban, err = ctx.providers.StorageProvider.LoadBannedUserByID(ctx, id); err != nil {
					return err
				}

				bans = []model.BannedUser{ban}
			}

			for _, ban := range bans {
				if ban.Revoked {
					fmt.Printf("SKIPPED\tUser ban with id '%d' for '%s is already revoked.\n", ban.ID, ban.Username)
				} else {
					if err = ctx.providers.StorageProvider.RevokeBannedUser(ctx, ban.ID, time.Now()); err != nil {
						fmt.Printf("ERROR\tUser ban with id '%d' for '%s' had error when being revoked: %+v\n", ban.ID, ban.Username, err)
					} else {
						fmt.Printf("REVOKED\tUser ban with id '%d' for '%s' has been revoked\n", ban.ID, ban.Username)
					}
				}
			}
		default:
			return fmt.Errorf("unknown command %q", use)
		}

		return nil
	}
}

//nolint:gocyclo
func (ctx *CmdCtx) StorageBansAddRunE(use string) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			_ = ctx.providers.StorageProvider.Close()
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

			if err = ctx.providers.StorageProvider.SaveBannedIP(ctx, ban); err != nil {
				return err
			}

			return nil
		case cmdUseUser:
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

			if err = ctx.providers.StorageProvider.SaveBannedUser(ctx, ban); err != nil {
				return err
			}

			return nil
		default:
			return fmt.Errorf("unknown command %q", use)
		}
	}
}

func (ctx *CmdCtx) StorageUserWebAuthnExportRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
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

	limit := 10
	count := 0

	var (
		credentials []model.WebAuthnCredential
	)

	export := &model.WebAuthnCredentialExport{
		WebAuthnCredentials: []model.WebAuthnCredential{},
	}

	for page := 0; true; page++ {
		if credentials, err = ctx.providers.StorageProvider.LoadWebAuthnCredentials(ctx, limit, page); err != nil {
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

	if err = exportYAMLWithJSONSchema("export.webauthn", filename, export); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	fmt.Printf(cliOutputFmtSuccessfulUserExportFile, count, "WebAuthn credentials", "YAML", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserWebAuthnImportRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		filename string

		stat os.FileInfo
		data []byte
	)

	filename = args[0]

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

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	for _, credential := range export.WebAuthnCredentials {
		if err = ctx.providers.StorageProvider.SaveWebAuthnCredential(ctx, credential); err != nil {
			return err
		}
	}

	fmt.Printf(cliOutputFmtSuccessfulUserImportFile, len(export.WebAuthnCredentials), "WebAuthn credentials", "YAML", filename)

	return nil
}

// StorageUserWebAuthnListRunE is the RunE for the authelia storage user webauthn list command.
func (ctx *CmdCtx) StorageUserWebAuthnListRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if len(args) == 0 || args[0] == "" {
		return ctx.StorageUserWebAuthnListAllRunE(cmd, args)
	}

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var credentials []model.WebAuthnCredential

	user := args[0]

	credentials, err = ctx.providers.StorageProvider.LoadWebAuthnCredentialsByUsername(ctx, "", user)

	switch {
	case len(credentials) == 0 || (err != nil && errors.Is(err, storage.ErrNoWebAuthnCredential)):
		return fmt.Errorf("user '%s' has no WebAuthn credentials", user)
	case err != nil:
		return fmt.Errorf("can't list credentials for user '%s': %w", user, err)
	default:
		fmt.Printf("WebAuthn Credentials for user '%s':\n\n", user)

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)

		_, _ = fmt.Fprintln(w, "ID\tKID\tDescription")

		for _, credential := range credentials {
			_, _ = fmt.Fprintf(w, "%d\t%s\t%s\n", credential.ID, credential.KID, credential.Description)
		}

		return w.Flush()
	}
}

// StorageUserWebAuthnListAllRunE is the RunE for the authelia storage user webauthn list command when no args are specified.
func (ctx *CmdCtx) StorageUserWebAuthnListAllRunE(_ *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var credentials []model.WebAuthnCredential

	limit := 10

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(w, "ID\tRPID\tKID\tDescription\tUsername")

	for page := 0; true; page++ {
		if credentials, err = ctx.providers.StorageProvider.LoadWebAuthnCredentials(ctx, limit, page); err != nil {
			return fmt.Errorf("failed to list credentials: %w", err)
		}

		if page == 0 && len(credentials) == 0 {
			return errors.New("no WebAuthn credentials in database")
		}

		for _, credential := range credentials {
			_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", credential.ID, credential.RPID, credential.KID, credential.Description, credential.Username)
		}

		if len(credentials) < limit {
			break
		}
	}

	fmt.Printf("WebAuthn Credentials:\n\n")

	return w.Flush()
}

// StorageUserWebAuthnVerifyRunE is the RunE for the authelia storage user webauthn verify command when no args are specified.
func (ctx *CmdCtx) StorageUserWebAuthnVerifyRunE(_ *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		provider    webauthn.MetaDataProvider
		credentials []model.WebAuthnCredential
	)

	if provider, err = webauthn.NewMetaDataProvider(ctx.config, ctx.providers.StorageProvider); err != nil {
		return err
	}

	limit := 10

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)

	_, _ = fmt.Fprintln(w, "ID\tRPID\tKID\tUsername\tAAGUID\tStatement\tBackup\tMDS")

	for page := 0; true; page++ {
		if credentials, err = ctx.providers.StorageProvider.LoadWebAuthnCredentials(ctx, limit, page); err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}

		if page == 0 && len(credentials) == 0 {
			return errors.New("no WebAuthn credentials in database")
		}

		for _, credential := range credentials {
			result := webauthn.VerifyCredential(&ctx.config.WebAuthn, &credential, provider)

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

			_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", credential.ID, credential.RPID, credential.KID, credential.Username, strAAGUID, strStatement, strBackup, strMDS)
		}

		if len(credentials) < limit {
			break
		}
	}

	fmt.Printf("WebAuthn Credential Verifications:\n\n")

	return w.Flush()
}

// StorageUserWebAuthnDeleteRunE is the RunE for the authelia storage user webauthn delete command.
func (ctx *CmdCtx) StorageUserWebAuthnDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
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

	if byKID {
		if err = ctx.providers.StorageProvider.DeleteWebAuthnCredential(ctx, kid); err != nil {
			return fmt.Errorf("failed to delete WebAuthn credential with kid '%s': %w", kid, err)
		}

		fmt.Printf("Successfully deleted WebAuthn credential with key id '%s'\n", kid)
	} else {
		err = ctx.providers.StorageProvider.DeleteWebAuthnCredentialByUsername(ctx, user, description)

		if all {
			if err != nil {
				return fmt.Errorf("failed to delete all WebAuthn credentials with username '%s': %w", user, err)
			}

			fmt.Printf("Successfully deleted all WebAuthn credentials for user '%s'\n", user)
		} else {
			if err != nil {
				return fmt.Errorf("failed to delete WebAuthn credential with username '%s' and description '%s': %w", user, description, err)
			}

			fmt.Printf("Successfully deleted WebAuthn credential with description '%s' for user '%s'\n", description, user)
		}
	}

	return nil
}

// StorageUserTOTPGenerateRunE is the RunE for the authelia storage user totp generate command.
func (ctx *CmdCtx) StorageUserTOTPGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		c                *model.TOTPConfiguration
		force            bool
		filename, secret string
		file             *os.File
		img              image.Image
	)

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if force, filename, secret, err = storageTOTPGenerateRunEOptsFromFlags(cmd.Flags()); err != nil {
		return err
	}

	if _, err = ctx.providers.StorageProvider.LoadTOTPConfiguration(ctx, args[0]); err == nil && !force {
		return fmt.Errorf("%s already has a TOTP configuration, use --force to overwrite", args[0])
	} else if err != nil && !errors.Is(err, storage.ErrNoTOTPConfiguration) {
		return err
	}

	totpProvider := totp.NewTimeBasedProvider(ctx.config.TOTP)

	if c, err = totpProvider.GenerateCustom(totp.NewContext(ctx, &clock.Real{}, &random.Cryptographical{}), args[0], ctx.config.TOTP.DefaultAlgorithm, secret, uint32(ctx.config.TOTP.DefaultDigits), uint(ctx.config.TOTP.DefaultPeriod), uint(ctx.config.TOTP.SecretSize)); err != nil { //nolint:gosec // Validated at runtime.
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

		defer file.Close()

		if img, err = c.Image(256, 256); err != nil {
			return err
		}

		if err = png.Encode(file, img); err != nil {
			return err
		}

		extraInfo = fmt.Sprintf(" and saved it as a PNG image at the path '%s'", filename)
	}

	if err = ctx.providers.StorageProvider.SaveTOTPConfiguration(ctx, *c); err != nil {
		return err
	}

	fmt.Printf("Successfully generated TOTP configuration for user '%s' with URI '%s'%s\n", args[0], c.URI(), extraInfo)

	return nil
}

// StorageUserTOTPDeleteRunE is the RunE for the authelia storage user totp delete command.
func (ctx *CmdCtx) StorageUserTOTPDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
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

	fmt.Printf("Successfully deleted TOTP configuration for user '%s'\n", user)

	return nil
}

const (
	cliOutputFmtSuccessfulUserExportFile = "Successfully exported %d %s as %s to the '%s' file\n"
	cliOutputFmtSuccessfulUserImportFile = "Successfully imported %d %s from the %s file '%s' into the database\n"
)

// StorageUserTOTPExportRunE is the RunE for the authelia storage user totp export command.
func (ctx *CmdCtx) StorageUserTOTPExportRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
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

	limit := 10
	count := 0

	var (
		configs []model.TOTPConfiguration
	)

	export := &model.TOTPConfigurationExport{}

	for page := 0; true; page++ {
		if configs, err = ctx.providers.StorageProvider.LoadTOTPConfigurations(ctx, limit, page); err != nil {
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

	if err = exportYAMLWithJSONSchema("export.totp", filename, export); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	fmt.Printf(cliOutputFmtSuccessfulUserExportFile, count, "TOTP configurations", "YAML", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPImportRunE(_ *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		filename string

		stat os.FileInfo
		data []byte
	)

	filename = args[0]

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

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	for _, config := range export.TOTPConfigurations {
		if err = ctx.providers.StorageProvider.SaveTOTPConfiguration(ctx, config); err != nil {
			return err
		}
	}

	fmt.Printf(cliOutputFmtSuccessfulUserImportFile, len(export.TOTPConfigurations), "TOTP configurations", "YAML", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPExportURIRunE(_ *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		configs []model.TOTPConfiguration
	)

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	limit := 10
	count := 0

	buf := &bytes.Buffer{}

	for page := 0; true; page++ {
		if configs, err = ctx.providers.StorageProvider.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		for _, c := range configs {
			fmt.Fprintf(buf, "%s\n", c.URI())
		}

		l := len(configs)

		count += l

		if l < limit {
			break
		}
	}

	fmt.Print(buf.String())

	fmt.Printf("\n\nSuccessfully exported %d TOTP configurations as TOTP URI's and printed them to the console\n", count)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPExportCSVRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		filename string
		configs  []model.TOTPConfiguration
		buf      *bytes.Buffer
	)

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if filename, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	limit := 10
	count := 0

	buf = &bytes.Buffer{}

	buf.WriteString("issuer,username,algorithm,digits,period,secret\n")

	for page := 0; true; page++ {
		if configs, err = ctx.providers.StorageProvider.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		for _, c := range configs {
			fmt.Fprintf(buf, "%s,%s,%s,%d,%d,%s\n", c.Issuer, c.Username, c.Algorithm, c.Digits, c.Period, string(c.Secret))
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

	fmt.Printf(cliOutputFmtSuccessfulUserExportFile, count, "TOTP configurations", "CSV", filename)

	return nil
}

func (ctx *CmdCtx) StorageUserTOTPExportPNGRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		dir     string
		configs []model.TOTPConfiguration
		img     image.Image
	)

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if dir, err = cmd.Flags().GetString(cmdFlagNameDirectory); err != nil {
		return err
	}

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

	var file *os.File

	for page := 0; true; page++ {
		if configs, err = ctx.providers.StorageProvider.LoadTOTPConfigurations(ctx, limit, page); err != nil {
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

	fmt.Printf("Successfully exported %d TOTP configuration as QR codes in PNG format to the '%s' directory\n", count, dir)

	return nil
}

// StorageUserIdentifiersExportRunE is the RunE for the authelia storage user identifiers export command.
func (ctx *CmdCtx) StorageUserIdentifiersExportRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
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

	export := &model.UserOpaqueIdentifiersExport{
		Identifiers: nil,
	}

	if export.Identifiers, err = ctx.providers.StorageProvider.LoadUserOpaqueIdentifiers(ctx); err != nil {
		return err
	}

	if len(export.Identifiers) == 0 {
		return fmt.Errorf("no data to export")
	}

	if err = exportYAMLWithJSONSchema("export.identifiers", filename, export); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", filename, err)
	}

	fmt.Printf(cliOutputFmtSuccessfulUserExportFile, len(export.Identifiers), "User Opaque Identifiers", "YAML", filename)

	return nil
}

// StorageUserIdentifiersImportRunE is the RunE for the authelia storage user identifiers import command.
func (ctx *CmdCtx) StorageUserIdentifiersImportRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		filename string

		stat os.FileInfo
		data []byte
	)

	filename = args[0]

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

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	for _, opaqueID := range export.Identifiers {
		if err = ctx.providers.StorageProvider.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
			return err
		}
	}

	fmt.Printf(cliOutputFmtSuccessfulUserImportFile, len(export.Identifiers), "User Opaque Identifiers", "YAML", filename)

	return nil
}

// StorageUserIdentifiersGenerateRunE is the RunE for the authelia storage user identifiers generate command.
func (ctx *CmdCtx) StorageUserIdentifiersGenerateRunE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		users, services, sectors []string
	)

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	identifiers, err := ctx.providers.StorageProvider.LoadUserOpaqueIdentifiers(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("can't load the existing identifiers: %w", err)
	}

	if users, services, sectors, err = flagsGetUserIdentifiersGenerateOptions(cmd.Flags()); err != nil {
		return err
	}

	if len(users) == 0 {
		return fmt.Errorf("must supply at least one user")
	}

	if len(sectors) == 0 {
		sectors = append(sectors, "")
	}

	if !utils.IsStringSliceContainsAll(services, validIdentifierServices) {
		return fmt.Errorf("one or more the service names '%s' is invalid, the valid values are: '%s'", strings.Join(services, "', '"), strings.Join(validIdentifierServices, "', '"))
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

				if err = ctx.providers.StorageProvider.SaveUserOpaqueIdentifier(ctx, identifier); err != nil {
					return fmt.Errorf("failed to save identifier: %w", err)
				}

				added++
			}
		}
	}

	fmt.Printf("Successfully generated and added opaque identifiers:\n")
	fmt.Printf("\tUsers: '%s'\n", strings.Join(users, "', '"))
	fmt.Printf("\tSectors: '%s'\n", strings.Join(sectors, "', '"))
	fmt.Printf("\tServices: '%s'\n", strings.Join(services, "', '"))

	if duplicates != 0 {
		fmt.Printf("\tSkipped Duplicates: %d\n", duplicates)
	}

	fmt.Printf("\tTotal: %d", added)

	return nil
}

// StorageUserIdentifiersAddRunE is the RunE for the authelia storage user identifiers add command.
func (ctx *CmdCtx) StorageUserIdentifiersAddRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	var (
		service, sector string
	)

	if service, err = cmd.Flags().GetString(cmdFlagNameService); err != nil {
		return err
	}

	if service == "" {
		service = identifierServiceOpenIDConnect
	} else if !utils.IsStringInSlice(service, validIdentifierServices) {
		return fmt.Errorf("the service name '%s' is invalid, the valid values are: '%s'", service, strings.Join(validIdentifierServices, "', '"))
	}

	if sector, err = cmd.Flags().GetString(cmdFlagNameSector); err != nil {
		return err
	}

	opaqueID := model.UserOpaqueIdentifier{
		Service:  service,
		Username: args[0],
		SectorID: sector,
	}

	if cmd.Flags().Changed(cmdFlagNameIdentifier) {
		var identifierStr string

		if identifierStr, err = cmd.Flags().GetString(cmdFlagNameIdentifier); err != nil {
			return err
		}

		if opaqueID.Identifier, err = uuid.Parse(identifierStr); err != nil {
			return fmt.Errorf("the identifier provided '%s' is invalid as it must be a version 4 UUID but parsing it had an error: %w", identifierStr, err)
		}

		if opaqueID.Identifier.Version() != 4 {
			return fmt.Errorf("the identifier providerd '%s' is a version %d UUID but only version 4 UUID's accepted as identifiers", identifierStr, opaqueID.Identifier.Version())
		}
	} else {
		if opaqueID.Identifier, err = uuid.NewRandom(); err != nil {
			return err
		}
	}

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if err = ctx.providers.StorageProvider.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
		return err
	}

	fmt.Printf("Added User Opaque Identifier:\n\tService: %s\n\tSector: %s\n\tUsername: %s\n\tIdentifier: %s\n\n", opaqueID.Service, opaqueID.SectorID, opaqueID.Username, opaqueID.Identifier)

	return nil
}
