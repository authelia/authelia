package commands

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/totp"
	"github.com/authelia/authelia/v4/internal/utils"
)

func (ctx *CmdCtx) ConfigStorageCommandLineConfigPersistentPreRunE(cmd *cobra.Command, _ []string) (err error) {
	flagsMap := map[string]string{
		cmdFlagNameEncryptionKey: "storage.encryption_key",

		cmdFlagNameSQLite3Path: "storage.local.path",

		cmdFlagNameMySQLHost:     "storage.mysql.host",
		cmdFlagNameMySQLPort:     "storage.mysql.port",
		cmdFlagNameMySQLDatabase: "storage.mysql.database",
		cmdFlagNameMySQLUsername: "storage.mysql.username",
		cmdFlagNameMySQLPassword: "storage.mysql.password",

		cmdFlagNamePostgreSQLHost:       "storage.postgres.host",
		cmdFlagNamePostgreSQLPort:       "storage.postgres.port",
		cmdFlagNamePostgreSQLDatabase:   "storage.postgres.database",
		cmdFlagNamePostgreSQLSchema:     "storage.postgres.schema",
		cmdFlagNamePostgreSQLUsername:   "storage.postgres.username",
		cmdFlagNamePostgreSQLPassword:   "storage.postgres.password",
		"postgres.ssl.mode":             "storage.postgres.ssl.mode",
		"postgres.ssl.root_certificate": "storage.postgres.ssl.root_certificate",
		"postgres.ssl.certificate":      "storage.postgres.ssl.certificate",
		"postgres.ssl.key":              "storage.postgres.ssl.key",

		cmdFlagNamePeriod:     "totp.period",
		cmdFlagNameDigits:     "totp.digits",
		cmdFlagNameAlgorithm:  "totp.algorithm",
		cmdFlagNameIssuer:     "totp.issuer",
		cmdFlagNameSecretSize: "totp.secret_size",
	}

	return ctx.ConfigSetFlagsMapRunE(cmd.Flags(), flagsMap, true, false)
}

func (ctx *CmdCtx) ConfigValidateStoragePersistentPreRunE(_ *cobra.Command, _ []string) (err error) {
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

/*
func (ctx *CmdCtx) StoragePersistentPreRunE(cmd *cobra.Command, _ []string) (err error) {
	var configs []string

	if configs, err = cmd.Flags().GetStringSlice(cmdFlagNameConfig); err != nil {
		return err
	}

	sources := make([]configuration.Source, 0, len(configs)+3)

	if cmd.Flags().Changed(cmdFlagNameConfig) {
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
		cmdFlagNameEncryptionKey: "storage.encryption_key",

		cmdFlagNameSQLite3Path: "storage.local.path",

		cmdFlagNameMySQLHost:     "storage.mysql.host",
		cmdFlagNameMySQLPort:     "storage.mysql.port",
		cmdFlagNameMySQLDatabase: "storage.mysql.database",
		cmdFlagNameMySQLUsername: "storage.mysql.username",
		cmdFlagNameMySQLPassword: "storage.mysql.password",

		cmdFlagNamePostgreSQLHost:       "storage.postgres.host",
		cmdFlagNamePostgreSQLPort:       "storage.postgres.port",
		cmdFlagNamePostgreSQLDatabase:   "storage.postgres.database",
		cmdFlagNamePostgreSQLSchema:     "storage.postgres.schema",
		cmdFlagNamePostgreSQLUsername:   "storage.postgres.username",
		cmdFlagNamePostgreSQLPassword:   "storage.postgres.password",
		"postgres.ssl.mode":             "storage.postgres.ssl.mode",
		"postgres.ssl.root_certificate": "storage.postgres.ssl.root_certificate",
		"postgres.ssl.certificate":      "storage.postgres.ssl.certificate",
		"postgres.ssl.key":              "storage.postgres.ssl.key",

		cmdFlagNamePeriod:     "totp.period",
		cmdFlagNameDigits:     "totp.digits",
		cmdFlagNameAlgorithm:  "totp.algorithm",
		cmdFlagNameIssuer:     "totp.issuer",
		cmdFlagNameSecretSize: "totp.secret_size",
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
}.

*/

func (ctx *CmdCtx) StorageSchemaEncryptionCheckRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		verbose bool
		result  storage.EncryptionValidationResult
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

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

func (ctx *CmdCtx) StorageSchemaEncryptionChangeKeyRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		key     string
		version int
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
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
		if key, err = termReadPasswordStrWithPrompt("Enter New Storage Encryption Key: ", cmdFlagNameNewEncryptionKey); err != nil {
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

func (ctx *CmdCtx) StorageWebauthnImportRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		filename string
		data     []byte
		stat     os.FileInfo
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

	o := &model.WebauthnDeviceExport{}

	if err = yaml.Unmarshal(data, o); err != nil {
		return err
	}

	if len(o.WebauthnDevices) == 0 {
		return fmt.Errorf("can't import a YAML file without webauthn devices data")
	}

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	for _, device := range o.WebauthnDevices {
		if err = ctx.providers.StorageProvider.SaveWebauthnDevice(ctx, device); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully imported %d Webauthn devices into the database\n", len(o.WebauthnDevices))

	return nil
}

func (ctx *CmdCtx) StorageWebauthnExportRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		filename string
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if filename, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	limit := 10

	var (
		devices []model.WebauthnDevice
	)

	o := &model.WebauthnDeviceExport{
		WebauthnDevices: nil,
	}

	for page := 0; true; page++ {
		if devices, err = ctx.providers.StorageProvider.LoadWebauthnDevices(ctx, limit, page); err != nil {
			return err
		}

		o.WebauthnDevices = append(o.WebauthnDevices, devices...)

		if len(devices) < limit {
			break
		}
	}

	var data []byte

	if data, err = yaml.Marshal(o); err != nil {
		return err
	}

	if err = os.WriteFile(filename, data, 0600); err != nil {
		return err
	}

	fmt.Printf("Exported Webauthn devices backup with %d devices to the '%s' file", len(o.WebauthnDevices), filename)

	return nil
}

func (ctx *CmdCtx) StorageWebauthnListRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || args[0] == "" {
		return ctx.StorageWebAuthnListAllRunE(cmd, args)
	}

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var devices []model.WebauthnDevice

	user := args[0]

	devices, err = ctx.providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, user)

	switch {
	case len(devices) == 0 || (err != nil && errors.Is(err, storage.ErrNoWebauthnDevice)):
		return fmt.Errorf("user '%s' has no webauthn devices", user)
	case err != nil:
		return fmt.Errorf("can't list devices for user '%s': %w", user, err)
	default:
		fmt.Printf("Webauthn Devices for user '%s':\n\n", user)
		fmt.Printf("ID\tKID\tDescription\n")

		for _, device := range devices {
			fmt.Printf("%d\t%s\t%s", device.ID, device.KID, device.Description)
		}
	}

	return nil
}

func (ctx *CmdCtx) StorageWebAuthnListAllRunE(_ *cobra.Command, _ []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var devices []model.WebauthnDevice

	limit := 10

	output := strings.Builder{}

	for page := 0; true; page++ {
		if devices, err = ctx.providers.StorageProvider.LoadWebauthnDevices(ctx, limit, page); err != nil {
			return fmt.Errorf("failed to list devices: %w", err)
		}

		if page == 0 && len(devices) == 0 {
			return errors.New("no webauthn devices in database")
		}

		for _, device := range devices {
			output.WriteString(fmt.Sprintf("%d\t%s\t%s\t%s\n", device.ID, device.KID, device.Description, device.Username))
		}

		if len(devices) < limit {
			break
		}
	}

	fmt.Printf("Webauthn Devices:\n\nID\tKID\tDescription\tUsername\n")
	fmt.Println(output.String())

	return nil
}

func (ctx *CmdCtx) StorageWebauthnDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		all, byKID             bool
		description, kid, user string
	)

	if all, byKID, description, kid, user, err = storageWebauthnDeleteRunEOptsFromFlags(cmd.Flags(), args); err != nil {
		return err
	}

	if byKID {
		if err = ctx.providers.StorageProvider.DeleteWebauthnDevice(ctx, kid); err != nil {
			return fmt.Errorf("failed to delete WebAuthn device with kid '%s': %w", kid, err)
		}

		fmt.Printf("Deleted WebAuthn device with kid '%s'", kid)
	} else {
		err = ctx.providers.StorageProvider.DeleteWebauthnDeviceByUsername(ctx, user, description)

		if all {
			if err != nil {
				return fmt.Errorf("failed to delete all WebAuthn devices with username '%s': %w", user, err)
			}

			fmt.Printf("Deleted all WebAuthn devices for user '%s'", user)
		} else {
			if err != nil {
				return fmt.Errorf("failed to delete WebAuthn device with username '%s' and description '%s': %w", user, description, err)
			}

			fmt.Printf("Deleted WebAuthn device with username '%s' and description '%s'", user, description)
		}
	}

	return nil
}

func (ctx *CmdCtx) StorageTOTPGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		c                *model.TOTPConfiguration
		force            bool
		filename, secret string
		file             *os.File
		img              image.Image
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
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

	if c, err = totpProvider.GenerateCustom(args[0], ctx.config.TOTP.Algorithm, secret, ctx.config.TOTP.Digits, ctx.config.TOTP.Period, ctx.config.TOTP.SecretSize); err != nil {
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

	fmt.Printf("Generated TOTP configuration for user '%s' with URI '%s'%s\n", args[0], c.URI(), extraInfo)

	return nil
}

func (ctx *CmdCtx) StorageTOTPDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	user := args[0]

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if _, err = ctx.providers.StorageProvider.LoadTOTPConfiguration(ctx, user); err != nil {
		return fmt.Errorf("can't delete configuration for user '%s': %+v", user, err)
	}

	if err = ctx.providers.StorageProvider.DeleteTOTPConfiguration(ctx, user); err != nil {
		return fmt.Errorf("can't delete configuration for user '%s': %+v", user, err)
	}

	fmt.Printf("Deleted TOTP configuration for user '%s'.", user)

	return nil
}

func (ctx *CmdCtx) StorageTOTPImportRunE(_ *cobra.Command, args []string) (err error) {
	var (
		filename string
		data     []byte
		stat     os.FileInfo
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

	o := &model.TOTPConfigurationExport{}

	if err = yaml.Unmarshal(data, o); err != nil {
		return err
	}

	if len(o.TOTPConfigurations) == 0 {
		return fmt.Errorf("can't import a YAML file without TOTP configuration data")
	}

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	for _, config := range o.TOTPConfigurations {
		if err = ctx.providers.StorageProvider.SaveTOTPConfiguration(ctx, config); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully imported %d TOTP configurations into the database\n", len(o.TOTPConfigurations))

	return nil
}

func (ctx *CmdCtx) StorageTOTPExportRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		filename string
		configs  []model.TOTPConfiguration
		o        *model.TOTPConfigurationExport
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if filename, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	limit := 10
	count := 0

	o = &model.TOTPConfigurationExport{}

	for page := 0; true; page++ {
		if configs, err = ctx.providers.StorageProvider.LoadTOTPConfigurations(ctx, limit, page); err != nil {
			return err
		}

		o.TOTPConfigurations = append(o.TOTPConfigurations, configs...)

		l := len(configs)

		count += l

		if l < limit {
			break
		}
	}

	var data []byte

	if data, err = yaml.Marshal(o); err != nil {
		return err
	}

	if err = os.WriteFile(filename, data, 0600); err != nil {
		return err
	}

	fmt.Printf("Successfully exported %d TOTP configurations as YAML to the '%s' file", count, filename)

	return nil
}

func (ctx *CmdCtx) StorageTOTPExportURIRunE(_ *cobra.Command, _ []string) (err error) {
	var (
		configs []model.TOTPConfiguration
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
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
			buf.WriteString(fmt.Sprintf("%s\n", c.URI()))
		}

		l := len(configs)

		count += l

		if l < limit {
			break
		}
	}

	fmt.Printf("\n\nSuccessfully exported %d TOTP configurations as a TOTP URI and printed them to the console\n", count)

	return nil
}

func (ctx *CmdCtx) StorageTOTPExportCSVRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		filename string
		configs  []model.TOTPConfiguration
		buf      *bytes.Buffer
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
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
			buf.WriteString(fmt.Sprintf("%s,%s,%s,%d,%d,%s\n", c.Issuer, c.Username, c.Algorithm, c.Digits, c.Period, string(c.Secret)))
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

	fmt.Printf("Successfully exported %d TOTP configurations as a CSV to the '%s' file", count, filename)

	return nil
}

func (ctx *CmdCtx) StorageTOTPExportPNGRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		dir     string
		configs []model.TOTPConfiguration
		img     image.Image
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if dir, err = cmd.Flags().GetString(cmdFlagNameDirectory); err != nil {
		return err
	}

	if dir == "" {
		dir = utils.RandomString(8, utils.CharSetAlphaNumeric, false)
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

	fmt.Printf("Successfully exported %d TOTP configuratioun QR codes in PNG format in the '%s' directory\n", count, dir)

	return nil
}

func (ctx *CmdCtx) StorageMigrateHistoryRunE(_ *cobra.Command, _ []string) (err error) {
	var (
		version    int
		migrations []model.Migration
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

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

	fmt.Printf("Migration History:\n\nID\tDate\t\t\t\tBefore\tAfter\tAuthelia Version\n")

	for _, m := range migrations {
		fmt.Printf("%d\t%s\t%d\t%d\t%s\n", m.ID, m.Applied.Format("2006-01-02 15:04:05 -0700"), m.Before, m.After, m.Version)
	}

	return nil
}

func (ctx *CmdCtx) NewStorageMigrateListRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		var (
			migrations   []model.SchemaMigration
			directionStr string
		)

		defer func() {
			_ = ctx.providers.StorageProvider.Close()
		}()

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
			fmt.Printf("Storage Schema Migration List (%s)\n\nVersion\t\tDescription\n", directionStr)

			for _, migration := range migrations {
				fmt.Printf("%d\t\t%s\n", migration.Version, migration.Name)
			}
		}

		return nil
	}
}

func (ctx *CmdCtx) NewStorageMigrationRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		var (
			target int
		)

		defer func() {
			_ = ctx.providers.StorageProvider.Close()
		}()

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

			if err = ctx.StorageMigrateDownConfirmDestroy(cmd); err != nil {
				return err
			}

			return ctx.providers.StorageProvider.SchemaMigrate(ctx, false, target)
		}
	}
}

func (ctx *CmdCtx) StorageMigrateDownConfirmDestroy(cmd *cobra.Command) (err error) {
	var destroy bool

	if destroy, err = cmd.Flags().GetBool(cmdFlagNameDestroyData); err != nil {
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

func (ctx *CmdCtx) StorageSchemaInfoRunE(_ *cobra.Command, _ []string) (err error) {
	var (
		upgradeStr, tablesStr string

		tables          []string
		version, latest int
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

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

func (ctx *CmdCtx) StorageUserIdentifiersExportRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		file string
	)

	if file, err = cmd.Flags().GetString(cmdFlagNameFile); err != nil {
		return err
	}

	_, err = os.Stat(file)

	switch {
	case err == nil:
		return fmt.Errorf("must specify a file that doesn't exist but '%s' exists", file)
	case !os.IsNotExist(err):
		return fmt.Errorf("error occurred opening '%s': %w", file, err)
	}

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	var (
		export model.UserOpaqueIdentifiersExport

		data []byte
	)

	if export.Identifiers, err = ctx.providers.StorageProvider.LoadUserOpaqueIdentifiers(ctx); err != nil {
		return err
	}

	if len(export.Identifiers) == 0 {
		return fmt.Errorf("no data to export")
	}

	if data, err = yaml.Marshal(&export); err != nil {
		return fmt.Errorf("error occurred marshalling data to YAML: %w", err)
	}

	if err = os.WriteFile(file, data, 0600); err != nil {
		return fmt.Errorf("error occurred writing to file '%s': %w", file, err)
	}

	fmt.Printf("Exported %d User Opaque Identifiers to %s\n", len(export.Identifiers), file)

	return nil
}

func (ctx *CmdCtx) StorageUserIdentifiersImportRunE(_ *cobra.Command, args []string) (err error) {
	var (
		filename string
		data     []byte
		stat     os.FileInfo
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

	o := &model.UserOpaqueIdentifiersExport{}

	if err = yaml.Unmarshal(data, o); err != nil {
		return err
	}

	if len(o.Identifiers) == 0 {
		return fmt.Errorf("can't import a YAML file without user opaque identifiers data")
	}

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	for _, opaqueID := range o.Identifiers {
		if err = ctx.providers.StorageProvider.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully imported %d opaque idnentifiers into the database\n", len(o.Identifiers))

	return nil
}

//nolint:gocyclo
func (ctx *CmdCtx) StorageUserIdentifiersGenerateRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		users, services, sectors []string
	)

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	identifiers, err := ctx.providers.StorageProvider.LoadUserOpaqueIdentifiers(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("can't load the existing identifiers: %w", err)
	}

	if users, err = cmd.Flags().GetStringSlice(cmdFlagNameUsers); err != nil {
		return err
	}

	if services, err = cmd.Flags().GetStringSlice(cmdFlagNameServices); err != nil {
		return err
	}

	if sectors, err = cmd.Flags().GetStringSlice(cmdFlagNameSectors); err != nil {
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

	fmt.Printf("Successfully added %d opaque identifiers and %d duplicates were skipped\n", added, duplicates)

	return nil
}

func (ctx *CmdCtx) StorageUserIdentifiersAddRunE(cmd *cobra.Command, args []string) (err error) {
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

	defer func() {
		_ = ctx.providers.StorageProvider.Close()
	}()

	if err = ctx.CheckSchemaVersion(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	if err = ctx.providers.StorageProvider.SaveUserOpaqueIdentifier(ctx, opaqueID); err != nil {
		return err
	}

	fmt.Printf("Added User Opaque Identifier:\n\tService: %s\n\tSector: %s\n\tUsername: %s\n\tIdentifier: %s\n\n", opaqueID.Service, opaqueID.SectorID, opaqueID.Username, opaqueID.Identifier)

	return nil
}
