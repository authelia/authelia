package commands

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/models"
	"github.com/authelia/authelia/internal/storage"
)

var configurationPath string
var localDatabasePath string

// MigrateLocalCmd migration command
var MigrateLocalCmd = &cobra.Command{
	Use:   "localdb",
	Short: "Migrate data from v3 local database into database configured in v4 configuration file",
	Run:   migrateLocal,
}

func init() {
	MigrateLocalCmd.PersistentFlags().StringVarP(&localDatabasePath, "db-path", "p", "", "The path to the v3 local database")
	MigrateLocalCmd.MarkPersistentFlagRequired("db-path")

	MigrateLocalCmd.PersistentFlags().StringVarP(&configurationPath, "config", "c", "", "The configuration file of Authelia v4")
	MigrateLocalCmd.MarkPersistentFlagRequired("config")
}

// migrateLocal data from v3 to v4
func migrateLocal(cmd *cobra.Command, args []string) {
	dbProvider := createDBProvider(configurationPath)

	migrateLocalTOTPSecret(dbProvider)
	migrateLocalU2FSecret(dbProvider)
	migrateLocalPreferences(dbProvider)
	migrateLocalAuthenticationTraces(dbProvider)
	// We don't need to migrate identity tokens

	log.Println("Migration done!")
}

func migrateLocalTOTPSecret(dbProvider storage.Provider) {
	file, err := os.Open(path.Join(localDatabasePath, "totp_secrets"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		data := scanner.Text()

		entry := TOTPSecretsV3{}
		json.Unmarshal([]byte(data), &entry)
		err := dbProvider.SaveTOTPSecret(entry.UserID, entry.Secret.Base32)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func migrateLocalU2FSecret(dbProvider storage.Provider) {
	file, err := os.Open(path.Join(localDatabasePath, "u2f_registrations"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		data := scanner.Text()

		entry := U2FDeviceHandleV3{}
		json.Unmarshal([]byte(data), &entry)

		kH, err := decodeWebsafeBase64(entry.Registration.KeyHandle)

		if err != nil {
			log.Fatal(err)
		}

		pK, err := decodeWebsafeBase64(entry.Registration.PublicKey)

		if err != nil {
			log.Fatal(err)
		}

		err = dbProvider.SaveU2FDeviceHandle(entry.UserID, kH, pK)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func migrateLocalPreferences(dbProvider storage.Provider) {
	file, err := os.Open(path.Join(localDatabasePath, "prefered_2fa_method"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		data := scanner.Text()

		entry := PreferencesV3{}
		json.Unmarshal([]byte(data), &entry)
		err := dbProvider.SavePreferred2FAMethod(entry.UserID, entry.Method)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func migrateLocalAuthenticationTraces(dbProvider storage.Provider) {
	file, err := os.Open(path.Join(localDatabasePath, "authentication_traces"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		data := scanner.Text()

		entry := AuthenticationTraceV3{}
		json.Unmarshal([]byte(data), &entry)

		attempt := models.AuthenticationAttempt{
			Username:   entry.UserID,
			Successful: entry.Successful,
			Time:       time.Unix(entry.Date.Date/1000.0, 0),
		}
		err := dbProvider.AppendAuthenticationLog(attempt)

		if err != nil {
			log.Fatal(err)
		}
	}
}
