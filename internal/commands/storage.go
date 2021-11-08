package commands

import (
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

// NewStorageCmd returns a new storage *cobra.Command.
func NewStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "storage",
		Short:             "Perform storage actions like migrations, re-encryption, etc",
		Args:              cobra.NoArgs,
		RunE:              storageRunE,
		PersistentPreRunE: storagePersistentPreRunE,
	}

	cmd.PersistentFlags().StringSliceP("config", "c", []string{"config.yml"}, "configuration file to load for the storage migration")

	cmd.PersistentFlags().StringP("encryption-key", "k", "", "the SQL encryption key")

	cmd.PersistentFlags().String("sqlite.path", "", "the SQLite database path")

	cmd.PersistentFlags().String("mysql.host", "", "the MySQL hostname")
	cmd.PersistentFlags().Int("mysql.port", 3306, "the MySQL port")
	cmd.PersistentFlags().String("mysql.database", "authelia", "the MySQL database name")
	cmd.PersistentFlags().String("mysql.username", "authelia", "the MySQL username")
	cmd.PersistentFlags().String("mysql.password", "", "the MySQL password")

	cmd.PersistentFlags().String("postgres.host", "", "the PostgreSQL hostname")
	cmd.PersistentFlags().Int("postgres.port", 5432, "the PostgreSQL port")
	cmd.PersistentFlags().String("postgres.database", "authelia", "the PostgreSQL database name")
	cmd.PersistentFlags().String("postgres.username", "authelia", "the PostgreSQL username")
	cmd.PersistentFlags().String("postgres.password", "", "the PostgreSQL password")

	cmd.AddCommand(
		newMigrateStorageCmd(),
		newSchemaInfoStorageCmd(),
	)

	return cmd
}

func storageRunE(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("%+v\n", config.Storage)

	return nil
}

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

	validator.ValidateStorage(config.Storage, val)
	// TODO CHECK this.

	return nil
}

func newSchemaInfoStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "schema-info",
		Short: "Show the storage information",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				provider   storage.Provider
				upgradeStr string
				tablesStr  string
			)

			provider, err = getStorageProvider()
			if err != nil {
				return err
			}

			version, err := provider.SchemaVersion()
			if err != nil && err.Error() != "unknown schema state" {
				return err
			}

			tables, err := provider.SchemaTables()
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
		},
	}

	return cmd
}

// NewMigrationCmd returns a new Migration Cmd.
func newMigrateStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "migrate",
		Short: "Perform or list migrations",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		newMigrateStorageUpCmd(), newMigrateStorageDownCmd(),
		newMigrateStorageListUpCmd(), newMigrateStorageListDownCmd(),
	)

	return cmd
}

func newMigrateStorageListUpCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "list-up",
		Short: "List the up migrations available",
		Args:  cobra.NoArgs,
		RunE:  newListMigrationsRunE(true),
	}

	return cmd
}

func newMigrateStorageListDownCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "list-down",
		Short: "List the down migrations available",
		Args:  cobra.NoArgs,
		RunE:  newListMigrationsRunE(false),
	}

	return cmd
}

func newListMigrationsRunE(up bool) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		var (
			provider     storage.Provider
			migrations   []storage.SchemaMigration
			directionStr string
		)

		provider, err = getStorageProvider()
		if err != nil {
			return err
		}

		if up {
			migrations, err = provider.SchemaMigrationsUp(0)
			directionStr = "Up"
		} else {
			migrations, err = provider.SchemaMigrationsDown(0)
			directionStr = "Down"
		}

		if err != nil {
			if err.Error() == "cannot migrate to the same version as prior" {
				fmt.Printf("No %s migrations found\n", directionStr)

				return nil
			}

			return err
		}

		fmt.Printf("Storage Schema Migration List (%s)\n\nVersion\tDescription\n", directionStr)

		for _, migration := range migrations {
			fmt.Printf("%d\t%s\n", migration.Version, migration.Name)
		}

		return nil
	}
}

func newMigrateStorageUpCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "up",
		Short: "Perform a migration up",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				provider storage.Provider
			)

			provider, err = getStorageProvider()
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("target") {
				latest, err := provider.SchemaLatestVersion()
				if err != nil {
					return err
				}

				version, err := provider.SchemaVersion()
				if err != nil {
					return err
				}

				if latest == version {
					return errors.New("already on latest version")
				}

				err = provider.SchemaMigrateLatest()
				if err != nil {
					return err
				}
			} else {
				target, err := cmd.Flags().GetInt("target")
				if err != nil {
					return err
				}

				latest, err := provider.SchemaLatestVersion()
				if err != nil {
					return err
				}

				if target < latest {
					return fmt.Errorf("can't migrate up to version %d as it's a higher version than the latest version available %d", target, latest)
				}

				version, err := provider.SchemaVersion()
				if err != nil {
					return err
				}

				if version == target {
					return fmt.Errorf("can't migrate up to version %d as it's the same as the current version", target)
				}

				err = provider.SchemaMigrate(target)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().IntP("target", "t", 0, "sets the version to migrate to, by default this is the latest version")

	return cmd
}

func newMigrateStorageDownCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "down",
		Short: "Perform a migration down",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				provider storage.Provider
			)

			provider, err = getStorageProvider()
			if err != nil {
				return err
			}

			pre1, err := cmd.Flags().GetBool("pre1")
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("target") && !pre1 {
				return fmt.Errorf("you must set the --target flag in order to migrate down")
			}

			version, err := provider.SchemaVersion()
			if err != nil {
				return err
			}

			if version == 0 {
				return fmt.Errorf("can't migrate down when the current version is 0")
			}

			target, err := cmd.Flags().GetInt("target")
			if err != nil {
				return err
			}

			if target < 0 {
				return fmt.Errorf("can't migrate to the non-existant version %d", target)
			}

			if version == target || pre1 && version == -1 {
				return fmt.Errorf("can't down migrate to version %d as it's the same as the current version", target)
			}

			fmt.Printf("Schema Down Migrations may DESTROY data, type 'DESTROY' and press return to continue: ")
			var text string

			_, _ = fmt.Scanln(&text)

			if text != "DESTROY" {
				return errors.New("cancelling down migration due to user not accepting data destruction")
			}

			if pre1 {
				err = provider.SchemaMigrate(-1)
				if err != nil {
					return err
				}
			} else {
				err = provider.SchemaMigrate(target)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().IntP("target", "t", 0, "sets the version to migrate to")
	cmd.Flags().Bool("pre1", false, "sets pre1 as the version to migrate to")
	return cmd
}
