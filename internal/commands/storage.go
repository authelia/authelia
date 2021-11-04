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
	cmd.PersistentFlags().StringP("provider", "t", "", "the SQL provider to use (sqlite, postgres, mysql)")
	cmd.PersistentFlags().StringP("host", "H", "", "the SQL hostname")
	cmd.PersistentFlags().IntP("port", "P", 0, "The SQL port")
	cmd.PersistentFlags().StringP("database", "d", "", "the SQL database name")
	cmd.PersistentFlags().StringP("username", "u", "", "the SQL username")
	cmd.PersistentFlags().StringP("password", "p", "", "the SQL password")
	cmd.PersistentFlags().StringP("encryption-key", "k", "", "the SQL encryption key")
	cmd.PersistentFlags().StringP("path", "f", "", "the SQLite path")

	cmd.AddCommand(
		newMigrateStorageCmd(),
		newReEncryptStorageCmd(),
		newSchemaInfoStorageCmd(),
	)

	return cmd
}

func storageRunE(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("%+v\n", config.Storage)

	return nil
}

func storagePersistentPreRunE(cmd *cobra.Command, args []string) (err error) {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	if provider == "" {
		return errors.New("you must specify a provider")
	}

	switch provider {
	case "":
		return errors.New("you must specify a provider")
	case "sqlite", "sqlite3":
		provider = "local"
	case "mariadb":
		provider = "mysql"
	case "postgresql":
		provider = "postgres"
	case "mysql", "postgres", "local":
		break
	default:
		return fmt.Errorf("unknown storage provider which should be one of sqlite, mysql, or postgres: %s", provider)
	}

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

	sources = append(sources, configuration.NewEnvironmentSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter))
	sources = append(sources, configuration.NewSecretsSource(configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter))
	sources = append(sources, configuration.NewCommandLineSourceWithPrefixes(cmd.Flags(), ".", []string{fmt.Sprintf("storage.%s", provider), "storage"}))

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
		Use:   "info",
		Short: "Show the storage information",
		RunE: func(cmd *cobra.Command, args []string) error {
			var storageProvider storage.Provider

			switch {
			case config.Storage.PostgreSQL != nil:
				storageProvider = storage.NewPostgreSQLProvider(*config.Storage.PostgreSQL)
			case config.Storage.MySQL != nil:
				storageProvider = storage.NewMySQLProvider(*config.Storage.MySQL)
			case config.Storage.Local != nil:
				storageProvider = storage.NewSQLiteProvider(config.Storage.Local.Path)
			}

			version, err := storageProvider.SchemaVersion()
			if err != nil {
				return err
			}

			tables, err := storageProvider.SchemaTables()
			if err != nil {
				return err
			}

			var versionStr string
			switch version {
			case -1:
				versionStr = "pre1"
			case 0:
				versionStr = "N/A"
			default:
				versionStr = string(version)
			}

			tablesStr := strings.Join(tables, ", ")
			if tablesStr == "" {
				tablesStr = "N/A"
			}

			fmt.Printf("Schema Version: %s\n", versionStr)
			fmt.Printf("Schema Tables: %s\n", tablesStr)

			return nil
		},
	}

	return cmd
}

// NewMigrationCmd returns a new Migration Cmd.
func newMigrateStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "migrate",
		Short: "Perform migrations",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newMigrateStorageUpCmd(), newMigrateStorageDownCmd())

	return cmd
}

func newMigrateStorageUpCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "up",
		Short: "Perform an up migration",
		Args:  cobra.NoArgs,
	}

	return cmd
}

func newMigrateStorageDownCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "up",
		Short: "Perform a down migration",
		Args:  cobra.NoArgs,
	}

	return cmd
}
func newReEncryptStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "re-encrypt",
		Short: "Encrypts secure data in the database with a new key",
		Args:  cobra.NoArgs,
		RunE:  reEncryptStorageRunE,
	}

	return cmd
}

func reEncryptStorageRunE(cmd *cobra.Command, args []string) (err error) {
	if cmd.PersistentFlags().Changed("config") {

	}

	return nil
}

func storageConfigurationPreRunE() {

}
