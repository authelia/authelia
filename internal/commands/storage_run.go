package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/storage"
)

func storagePersistentPreRunE(cmd *cobra.Command, args []string) (err error) {
	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
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
	} else {
		if _, err := os.Stat(configs[0]); err == nil {
			sources = append(sources, configuration.NewYAMLFileSource(configs[0]))
		}
	}

	mapping := map[string]string{
		"encryption-key":    "storage.encryption_key",
		"sqlite.path":       "storage.local.path",
		"mysql.host":        "storage.mysql.host",
		"mysql.port":        "storage.mysql.port",
		"mysql.database":    "storage.mysql.database",
		"mysql.username":    "storage.mysql.username",
		"mysql.password":    "storage.mysql.password",
		"postgres.host":     "storage.postgres.host",
		"postgres.port":     "storage.postgres.port",
		"postgres.database": "storage.postgres.database",
		"postgres.username": "storage.postgres.username",
		"postgres.password": "storage.postgres.password",
		"postgres.schema":   "storage.postgres.schema",
	}

	sources = append(sources, configuration.NewEnvironmentSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter))
	sources = append(sources, configuration.NewSecretsSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter))
	sources = append(sources, configuration.NewCommandLineSourceWithMapping(cmd.Flags(), mapping, true, false))

	val := schema.NewStructValidator()

	config = &schema.Configuration{}

	_, err = configuration.LoadAdvanced(val, "storage", &config.Storage, sources...)
	if err != nil {
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

func storageSchemaEncryptionChangeKeyRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider storage.Provider
		ctx      = context.Background()
	)

	provider, err = getStorageProvider()
	if err != nil {
		return err
	}

	version, err := provider.SchemaVersion(ctx)
	if err != nil {
		return err
	}

	if version <= 0 {
		return errors.New("schema version must be at least version 1 to change the encryption key")
	}

	key, err := cmd.Flags().GetString("new-encryption-key")
	if err != nil {
		return err
	}

	if key == "" {
		return errors.New("you must set the --new-encryption-key flag")
	}

	if len(key) < 20 {
		return errors.New("the encryption key must be at least 20 characters")
	}

	return provider.SchemaEncryptionChangeKey(ctx, key)
}

func storageMigrateHistoryRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider storage.Provider
		ctx      = context.Background()
	)

	provider, err = getStorageProvider()
	if err != nil {
		return err
	}

	migrations, err := provider.SchemaMigrationHistory(ctx)
	if err != nil {
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
			migrations   []storage.SchemaMigration
			directionStr string
		)

		provider, err = getStorageProvider()
		if err != nil {
			return err
		}

		if up {
			migrations, err = provider.SchemaMigrationsUp(ctx, 0)
			directionStr = "Up"
		} else {
			migrations, err = provider.SchemaMigrationsDown(ctx, 0)
			directionStr = "Down"
		}

		if err != nil {
			if err.Error() == "cannot migrate to the same version as prior" {
				fmt.Printf("No %s migrations found\n", directionStr)

				return nil
			}

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
			ctx      = context.Background()
		)

		provider, err = getStorageProvider()
		if err != nil {
			return err
		}

		target, err := cmd.Flags().GetInt("target")
		if err != nil {
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
			if !cmd.Flags().Changed("target") {
				return errors.New("must set target")
			}

			if err = storageMigrateDownConfirmDestroy(cmd); err != nil {
				return err
			}

			pre1, err := cmd.Flags().GetBool("pre1")
			if err != nil {
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
	destroy, err := cmd.Flags().GetBool("destroy-data")
	if err != nil {
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

func storageSchemaInfoRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		provider   storage.Provider
		ctx        = context.Background()
		upgradeStr string
		tablesStr  string
	)

	provider, err = getStorageProvider()
	if err != nil {
		return err
	}

	version, err := provider.SchemaVersion(ctx)
	if err != nil && err.Error() != "unknown schema state" {
		return err
	}

	tables, err := provider.SchemaTables(ctx)
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		tablesStr = "N/A"
	} else {
		tablesStr = strings.Join(tables, ", ")
	}

	latest, err := provider.SchemaLatestVersion()
	if err != nil {
		return err
	}

	if latest > version {
		upgradeStr = fmt.Sprintf("yes - version %d", latest)
	} else {
		upgradeStr = "no"
	}

	fmt.Printf("Schema Version: %s\nSchema Upgrade Available: %s\nSchema Tables: %s\n", storage.SchemaVersionToString(version), upgradeStr, tablesStr)

	return nil
}
