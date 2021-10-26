package commands

import (
	"github.com/spf13/cobra"
)

func NewStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "storage",
		Short: "Perform storage actions like migrations, re-encryption, etc",
		Args:  cobra.NoArgs,
	}

	cmd.PersistentFlags().StringP("config", "c", "", "configuration file to load for the storage migration")
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
	)

	return cmd
}

// NewMigrationCmd returns a new Migration Cmd.
func newMigrateStorageCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "migrate",
		Short: "Perform migrations",
		Args:  cobra.NoArgs,
		RunE:  migrateStorageRunE,
	}

	return cmd
}

func migrateStorageRunE(cmd *cobra.Command, args []string) (err error) {
	return nil
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
