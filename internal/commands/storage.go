package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/storage"
)

// NewStorageCmd returns a new storage *cobra.Command.
func NewStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "storage",
		Short:             "Perform storage actions like migrations, re-encryption, etc",
		Args:              cobra.NoArgs,
		PersistentPreRunE: storagePersistentPreRunE,
	}

	cmd.PersistentFlags().StringSliceP("config", "c", []string{"config.yml"}, "configuration file to load for the storage migration")

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
		newStorageMigrateCmd(),
		newStorageSchemaInfoCmd(),
		storageTestE(),
	)

	return cmd
}

func storageTestE() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "test",
		Short: "test",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				provider storage.Provider
			)

			provider, err = getStorageProvider()
			if err != nil {
				return err
			}

			totp, err := provider.LoadTOTPConfiguration(context.Background(), args[0])
			if err != nil {
				return err
			}

			fmt.Printf("id: %d, period: %d, digits: %d, username: %s, algo: %s, secret: %s", totp.ID, totp.Period, totp.Digits, totp.Username, totp.Algorithm, totp.Secret)

			return nil
		},
	}

	return cmd
}

func newStorageSchemaInfoCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "schema-info",
		Short: "Show the storage information",
		RunE:  storageSchemaInfoRunE,
	}

	return cmd
}

// NewMigrationCmd returns a new Migration Cmd.
func newStorageMigrateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "migrate",
		Short: "Perform or list migrations",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		newStorageMigrateUpCmd(), newStorageMigrateDownCmd(),
		newStorageMigrateListUpCmd(), newStorageMigrateListDownCmd(),
		newStorageMigrateHistoryCmd(),
	)

	return cmd
}

func newStorageMigrateHistoryCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "history",
		Short: "Show migration history",
		Args:  cobra.NoArgs,
		RunE:  storageMigrateHistoryRunE,
	}

	return cmd
}

func newStorageMigrateListUpCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "list-up",
		Short: "List the up migrations available",
		Args:  cobra.NoArgs,
		RunE:  newStorageMigrateListRunE(true),
	}

	return cmd
}

func newStorageMigrateListDownCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "list-down",
		Short: "List the down migrations available",
		Args:  cobra.NoArgs,
		RunE:  newStorageMigrateListRunE(false),
	}

	return cmd
}

func newStorageMigrateUpCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   storageMigrateDirectionUp,
		Short: "Perform a migration up",
		Args:  cobra.NoArgs,
		RunE:  storageMigrateUpRunE,
	}

	cmd.Flags().IntP("target", "t", 0, "sets the version to migrate to, by default this is the latest version")

	return cmd
}

func newStorageMigrateDownCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   storageMigrateDirectionDown,
		Short: "Perform a migration down",
		Args:  cobra.NoArgs,
		RunE:  storageMigrateDownCmdE,
	}

	cmd.Flags().IntP("target", "t", 0, "sets the version to migrate to")
	cmd.Flags().Bool("pre1", false, "sets pre1 as the version to migrate to")
	cmd.Flags().Bool("destroy-data", false, "confirms you want to destroy data with this migration")

	return cmd
}
