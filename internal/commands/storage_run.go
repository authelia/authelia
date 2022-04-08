package commands

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/totp"
	"github.com/authelia/authelia/v4/internal/utils"
)

func storagePersistentPreRunE(cmd *cobra.Command, _ []string) (err error) {
	var configs []string

	if configs, err = cmd.Flags().GetStringSlice("config"); err != nil {
		return err
	}

	sources := make([]configuration.Source, 0, len(configs)+3)

	if cmd.Flags().Changed("config") {
		for _, configFile := range configs {
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				return fmt.Errorf("could not load the provided configuration file %s: %w", configFile, err)
			}

			sources = append(sources, configuration.NewYAMLFileSource(configFile))
		}
	} else if _, err := os.Stat(configs[0]); err == nil {
		sources = append(sources, configuration.NewYAMLFileSource(configs[0]))
	}

	mapping := map[string]string{
		"encryption-key": "storage.encryption_key",
		"sqlite.path":    "storage.local.path",

		"mysql.host":     "storage.mysql.host",
		"mysql.port":     "storage.mysql.port",
		"mysql.database": "storage.mysql.database",
		"mysql.username": "storage.mysql.username",
		"mysql.password": "storage.mysql.password",

		"postgres.host":                 "storage.postgres.host",
		"postgres.port":                 "storage.postgres.port",
		"postgres.database":             "storage.postgres.database",
		"postgres.schema":               "storage.postgres.schema",
		"postgres.username":             "storage.postgres.username",
		"postgres.password":             "storage.postgres.password",
		"postgres.ssl.mode":             "storage.postgres.ssl.mode",
		"postgres.ssl.root_certificate": "storage.postgres.ssl.root_certificate",
		"postgres.ssl.certificate":      "storage.postgres.ssl.certificate",
		"postgres.ssl.key":              "storage.postgres.ssl.key",

		"period":      "totp.period",
		"digits":      "totp.digits",
		"algorithm":   "totp.algorithm",
		"issuer":      "totp.issuer",
		"secret-size": "totp.secret_size",
	}

	sources = append(sources, configuration.NewEnvironmentSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter))
	sources = append(sources, configuration.NewSecretsSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter))
	sources = append(sources, configuration.NewCommandLineSourceWithMapping(cmd.Flags(), mapping, true, false))

	val := schema.NewStructValidator()

	config = &schema.Configuration{}

	if _, err = configuration.LoadAdvanced(val, "", &config, sources...); err != nil {
		return err
	}

	if val.HasErrors() {
		var finalErr error

		for i, err := range val.Errors() {
			if i == 0 {
				finalErr = err
				continue
			}

			finalErr = fmt.Errorf("%w, %v", finalErr, err)
		}

		return finalErr
	}

	validator.ValidateStorage(config.Storage, val)

	validator.ValidateTOTP(config, val)

	if val.HasErrors() {
		var finalErr error

		for i, err := range val.Errors() {
			if i == 0 {
				finalErr = err
				continue
			}

			finalErr = fmt.Errorf("%w, %v", finalErr, err)
		}

		return finalErr
	}

	return nil
}

func storageSchemaEncryptionCheckRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider storage.Provider
		verbose  bool

		ctx = context.Background()
	)

	provider = getStorageProvider()

	defer func() {
		_ = provider.Close()
	}()

	if verbose, err = cmd.Flags().GetBool("verbose"); err != nil {
		return err
	}

	if err = provider.SchemaEncryptionCheckKey(ctx, verbose); err != nil {
		switch {
		case errors.Is(err, storage.ErrSchemaEncryptionVersionUnsupported):
			fmt.Printf("Could not check encryption key for validity. The schema version doesn't support encryption.\n")
		case errors.Is(err, storage.ErrSchemaEncryptionInvalidKey):
			fmt.Printf("Encryption key validation: failed.\n\nError: %v.\n", err)
		default:
			fmt.Printf("Could not check encryption key for validity.\n\nError: %v.\n", err)
		}
	} else {
		fmt.Println("Encryption key validation: success.")
	}

	return nil
}

func storageSchemaEncryptionChangeKeyRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider storage.Provider
		key      string
		version  int

		ctx = context.Background()
	)

	provider = getStorageProvider()

	defer func() {
		_ = provider.Close()
	}()

	if err = checkStorageSchemaUpToDate(ctx, provider); err != nil {
		return err
	}

	if version, err = provider.SchemaVersion(ctx); err != nil {
		return err
	}

	if version <= 0 {
		return errors.New("schema version must be at least version 1 to change the encryption key")
	}

	key, err = cmd.Flags().GetString("new-encryption-key")

	switch {
	case err != nil:
		return err
	case key == "":
		return errors.New("you must set the --new-encryption-key flag")
	case len(key) < 20:
		return errors.New("the new encryption key must be at least 20 characters")
	}

	if err = provider.SchemaEncryptionChangeKey(ctx, key); err != nil {
		return err
	}

	fmt.Println("Completed the encryption key change. Please adjust your configuration to use the new key.")

	return nil
}

func storageTOTPGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider         storage.Provider
		ctx              = context.Background()
		c                *model.TOTPConfiguration
		force            bool
		filename, secret string
		file             *os.File
		img              image.Image
	)

	provider = getStorageProvider()

	defer func() {
		_ = provider.Close()
	}()

	if force, filename, secret, err = storageTOTPGenerateRunEOptsFromFlags(cmd.Flags()); err != nil {
		return err
	}

	if _, err = provider.LoadTOTPConfiguration(ctx, args[0]); err == nil && !force {
		return fmt.Errorf("%s already has a TOTP configuration, use --force to overwrite", args[0])
	} else if err != nil && !errors.Is(err, storage.ErrNoTOTPConfiguration) {
		return err
	}

	totpProvider := totp.NewTimeBasedProvider(config.TOTP)

	if c, err = totpProvider.GenerateCustom(args[0], config.TOTP.Algorithm, secret, config.TOTP.Digits, config.TOTP.Period, config.TOTP.SecretSize); err != nil {
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

	if err = provider.SaveTOTPConfiguration(ctx, *c); err != nil {
		return err
	}

	fmt.Printf("Generated TOTP configuration for user '%s' with URI '%s'%s\n", args[0], c.URI(), extraInfo)

	return nil
}

func storageTOTPGenerateRunEOptsFromFlags(flags *pflag.FlagSet) (force bool, filename, secret string, err error) {
	if force, err = flags.GetBool("force"); err != nil {
		return force, filename, secret, err
	}

	if filename, err = flags.GetString("path"); err != nil {
		return force, filename, secret, err
	}

	if secret, err = flags.GetString("secret"); err != nil {
		return force, filename, secret, err
	}

	secretLength := base32.StdEncoding.WithPadding(base32.NoPadding).DecodedLen(len(secret))
	if secret != "" && secretLength < schema.TOTPSecretSizeMinimum {
		return force, filename, secret, fmt.Errorf("decoded length of the base32 secret must have "+
			"a length of more than %d but '%s' has a decoded length of %d", schema.TOTPSecretSizeMinimum, secret, secretLength)
	}

	return force, filename, secret, nil
}

func storageTOTPDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider storage.Provider
		ctx      = context.Background()
	)

	user := args[0]

	provider = getStorageProvider()

	defer func() {
		_ = provider.Close()
	}()

	if _, err = provider.LoadTOTPConfiguration(ctx, user); err != nil {
		return fmt.Errorf("can't delete configuration for user '%s': %+v", user, err)
	}

	if err = provider.DeleteTOTPConfiguration(ctx, user); err != nil {
		return fmt.Errorf("can't delete configuration for user '%s': %+v", user, err)
	}

	fmt.Printf("Deleted TOTP configuration for user '%s'.", user)

	return nil
}

func storageTOTPExportRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider       storage.Provider
		format, dir    string
		configurations []model.TOTPConfiguration
		img            image.Image

		ctx = context.Background()
	)

	provider = getStorageProvider()

	defer func() {
		_ = provider.Close()
	}()

	if err = checkStorageSchemaUpToDate(ctx, provider); err != nil {
		return err
	}

	if format, dir, err = storageTOTPExportGetConfigFromFlags(cmd); err != nil {
		return err
	}

	limit := 10

	for page := 0; true; page++ {
		if configurations, err = provider.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		if page == 0 && format == storageExportFormatCSV {
			fmt.Printf("issuer,username,algorithm,digits,period,secret\n")
		}

		for _, c := range configurations {
			switch format {
			case storageExportFormatCSV:
				fmt.Printf("%s,%s,%s,%d,%d,%s\n", c.Issuer, c.Username, c.Algorithm, c.Digits, c.Period, string(c.Secret))
			case storageExportFormatURI:
				fmt.Println(c.URI())
			case storageExportFormatPNG:
				file, _ := os.Create(filepath.Join(dir, fmt.Sprintf("%s.png", c.Username)))

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
		}

		if len(configurations) < limit {
			break
		}
	}

	if format == storageExportFormatPNG {
		fmt.Printf("Exported TOTP QR codes in PNG format in the '%s' directory\n", dir)
	}

	return nil
}

func storageTOTPExportGetConfigFromFlags(cmd *cobra.Command) (format, dir string, err error) {
	if format, err = cmd.Flags().GetString("format"); err != nil {
		return "", "", err
	}

	if dir, err = cmd.Flags().GetString("dir"); err != nil {
		return "", "", err
	}

	switch format {
	case storageExportFormatCSV, storageExportFormatURI:
		break
	case storageExportFormatPNG:
		if dir == "" {
			dir = utils.RandomString(8, utils.AlphaNumericCharacters, false)
		}

		if _, err = os.Stat(dir); !os.IsNotExist(err) {
			return "", "", errors.New("output directory must not exist")
		}

		if err = os.MkdirAll(dir, 0700); err != nil {
			return "", "", err
		}
	default:
		return "", "", errors.New("format must be csv, uri, or png")
	}

	return format, dir, nil
}

func storageMigrateHistoryRunE(_ *cobra.Command, _ []string) (err error) {
	var (
		provider   storage.Provider
		version    int
		migrations []model.Migration

		ctx = context.Background()
	)

	provider = getStorageProvider()
	if provider == nil {
		return errNoStorageProvider
	}

	defer func() {
		_ = provider.Close()
	}()

	if version, err = provider.SchemaVersion(ctx); err != nil {
		return err
	}

	if version <= 0 {
		fmt.Println("No migration history is available for schemas that not version 1 or above.")
		return
	}

	if migrations, err = provider.SchemaMigrationHistory(ctx); err != nil {
		return err
	}

	if len(migrations) == 0 {
		return errors.New("no migration history found which may indicate a broken schema")
	}

	fmt.Printf("Migration History:\n\nID\tDate\t\t\t\tBefore\tAfter\tAuthelia Version\n")

	for _, m := range migrations {
		fmt.Printf("%d\t%s\t%d\t%d\t%s\n", m.ID, m.Applied.Format("2006-01-02 15:04:05 -0700"), m.Before, m.After, m.Version)
	}

	return nil
}

func newStorageMigrateListRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		var (
			provider     storage.Provider
			ctx          = context.Background()
			migrations   []model.SchemaMigration
			directionStr string
		)

		provider = getStorageProvider()

		defer func() {
			_ = provider.Close()
		}()

		if up {
			migrations, err = provider.SchemaMigrationsUp(ctx, 0)
			directionStr = "Up"
		} else {
			migrations, err = provider.SchemaMigrationsDown(ctx, 0)
			directionStr = "Down"
		}

		if err != nil && !errors.Is(err, storage.ErrNoAvailableMigrations) && !errors.Is(err, storage.ErrMigrateCurrentVersionSameAsTarget) {
			return err
		}

		if len(migrations) == 0 {
			fmt.Printf("Storage Schema Migration List (%s)\n\nNo Migrations Available\n", directionStr)
		} else {
			fmt.Printf("Storage Schema Migration List (%s)\n\nVersion\t\tDescription\n", directionStr)

			for _, migration := range migrations {
				fmt.Printf("%d\t\t%s\n", migration.Version, migration.Name)
			}
		}

		return nil
	}
}

func newStorageMigrationRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		var (
			provider storage.Provider
			target   int
			pre1     bool

			ctx = context.Background()
		)

		provider = getStorageProvider()

		defer func() {
			_ = provider.Close()
		}()

		if target, err = cmd.Flags().GetInt("target"); err != nil {
			return err
		}

		switch {
		case up:
			switch cmd.Flags().Changed("target") {
			case true:
				return provider.SchemaMigrate(ctx, true, target)
			default:
				return provider.SchemaMigrate(ctx, true, storage.SchemaLatest)
			}
		default:
			if pre1, err = cmd.Flags().GetBool("pre1"); err != nil {
				return err
			}

			if !cmd.Flags().Changed("target") && !pre1 {
				return errors.New("must set target")
			}

			if err = storageMigrateDownConfirmDestroy(cmd); err != nil {
				return err
			}

			switch {
			case pre1:
				return provider.SchemaMigrate(ctx, false, -1)
			default:
				return provider.SchemaMigrate(ctx, false, target)
			}
		}
	}
}

func storageMigrateDownConfirmDestroy(cmd *cobra.Command) (err error) {
	var destroy bool

	if destroy, err = cmd.Flags().GetBool("destroy-data"); err != nil {
		return err
	}

	if !destroy {
		fmt.Printf("Schema Down Migrations may DESTROY data, type 'DESTROY' and press return to continue: ")

		var text string

		_, _ = fmt.Scanln(&text)

		if text != "DESTROY" {
			return errors.New("cancelling down migration due to user not accepting data destruction")
		}
	}

	return nil
}

func storageSchemaInfoRunE(_ *cobra.Command, _ []string) (err error) {
	var (
		upgradeStr, tablesStr string

		provider        storage.Provider
		tables          []string
		version, latest int

		ctx = context.Background()
	)

	provider = getStorageProvider()

	defer func() {
		_ = provider.Close()
	}()

	if version, err = provider.SchemaVersion(ctx); err != nil && err.Error() != "unknown schema state" {
		return err
	}

	if tables, err = provider.SchemaTables(ctx); err != nil {
		return err
	}

	if len(tables) == 0 {
		tablesStr = "N/A"
	} else {
		tablesStr = strings.Join(tables, ", ")
	}

	if latest, err = provider.SchemaLatestVersion(); err != nil {
		return err
	}

	if latest > version {
		upgradeStr = fmt.Sprintf("yes - version %d", latest)
	} else {
		upgradeStr = "no"
	}

	var encryption string

	if err = provider.SchemaEncryptionCheckKey(ctx, false); err != nil {
		if errors.Is(err, storage.ErrSchemaEncryptionVersionUnsupported) {
			encryption = "unsupported (schema version)"
		} else {
			encryption = "invalid"
		}
	} else {
		encryption = "valid"
	}

	fmt.Printf("Schema Version: %s\nSchema Upgrade Available: %s\nSchema Tables: %s\nSchema Encryption Key: %s\n", storage.SchemaVersionToString(version), upgradeStr, tablesStr, encryption)

	return nil
}

func checkStorageSchemaUpToDate(ctx context.Context, provider storage.Provider) (err error) {
	var version, latest int

	if version, err = provider.SchemaVersion(ctx); err != nil {
		return err
	}

	if latest, err = provider.SchemaLatestVersion(); err != nil {
		return err
	}

	if version != latest {
		return fmt.Errorf("schema is version %d which is outdated please migrate to version %d in order to use this command or use an older binary", version, latest)
	}

	return nil
}

func storageUserIdentifiersExport(cmd *cobra.Command, _ []string) (err error) {
	var (
		provider storage.Provider

		ctx = context.Background()

		file string
	)

	if file, err = cmd.Flags().GetString("file"); err != nil {
		return err
	}

	_, err = os.Stat(file)

	switch {
	case err == nil:
		return fmt.Errorf("must specify a file that doesn't exist but '%s' exists", file)
	case !os.IsNotExist(err):
		return fmt.Errorf("error occurred opening '%s': %w", file, err)
	}

	provider = getStorageProvider()

	var (
		identifiers []model.UserOpaqueIdentifier
		export      exportUserOpaqueIdentifiers

		data []byte
	)

	if identifiers, err = provider.LoadUserOpaqueIdentifiers(ctx); err != nil {
		return err
	}

	export = exportUserOpaqueIdentifiers{
		Identifiers: identifiers,
	}

	if data, err = yaml.Marshal(&export); err != nil {
		return err
	}

	if err = os.WriteFile(file, data, 0600); err != nil {
		return err
	}

	fmt.Printf("Exported User Opaque Identifiers to %s\n", file)

	return nil
}

func storageUserIdentifiersImport(cmd *cobra.Command, _ []string) (err error) {
	var (
		provider storage.Provider

		ctx = context.Background()

		file string
		stat os.FileInfo
	)

	if file, err = cmd.Flags().GetString("file"); err != nil {
		return err
	}

	if stat, err = os.Stat(file); err != nil {
		return fmt.Errorf("must specify a file that exists but '%s' had an error opening it: %w", file, err)
	}

	if stat.IsDir() {
		return fmt.Errorf("must specify a file that exists but '%s' is a directory", file)
	}

	var (
		data   []byte
		export exportUserOpaqueIdentifiers
	)

	if data, err = os.ReadFile(file); err != nil {
		return err
	}

	if err = yaml.Unmarshal(data, &export); err != nil {
		return err
	}

	if len(export.Identifiers) == 0 {
		return fmt.Errorf("can't import a file with no data")
	}

	provider = getStorageProvider()

	for _, opaqueID := range export.Identifiers {
		if err = provider.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
			return err
		}
	}

	fmt.Printf("Imported User Opaque Identifiers from %s\n", file)

	return nil
}

func storageUserIdentifiersAdd(cmd *cobra.Command, args []string) (err error) {
	var (
		provider storage.Provider

		ctx = context.Background()

		service, sector string
	)

	if service, err = cmd.Flags().GetString("service"); err != nil {
		return err
	}

	if sector, err = cmd.Flags().GetString("sector"); err != nil {
		return err
	}

	opaqueID := model.UserOpaqueIdentifier{
		Service:  service,
		Username: args[0],
		SectorID: sector,
	}

	if cmd.Flags().Changed("identifier") {
		var identifierStr string

		if identifierStr, err = cmd.Flags().GetString("identifier"); err != nil {
			return err
		}

		if opaqueID.Identifier, err = uuid.Parse(identifierStr); err != nil {
			return err
		}

		if opaqueID.Identifier.Version() != 4 {
			return fmt.Errorf("UUID version must be 4 but you provided one with version %d", opaqueID.Identifier.Version())
		}
	} else {
		if opaqueID.Identifier, err = uuid.NewRandom(); err != nil {
			return err
		}
	}

	provider = getStorageProvider()

	if err = provider.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
		return err
	}

	fmt.Printf("Added User Opaque Identifier; service: %s, sector: %s, username: %s, identifier: %s\n", opaqueID.Service, opaqueID.SectorID, opaqueID.Username, opaqueID.Identifier)

	return nil
}
