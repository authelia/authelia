package commands

import (
	"encoding/base64"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/storage"
)

var MigrateCmd *cobra.Command

func init() {
	MigrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "helper function to migrate from v3 to v4",
	}
	MigrateCmd.AddCommand(MigrateLocalCmd, MigrateMongoCmd)
}

// TOTPSecretsV3 one entry of TOTP secrets in v3
type TOTPSecretsV3 struct {
	UserID string `json:"userId"`
	Secret struct {
		Base32 string `json:"base32"`
	} `json:"secret"`
}

// U2FDeviceHandleV3 one entry of U2F device handle in v3
type U2FDeviceHandleV3 struct {
	UserID       string `json:"userId"`
	Registration struct {
		KeyHandle string `json:"keyHandle"`
		PublicKey string `json:"publicKey"`
	} `json:"registration"`
}

// PreferencesV3 one entry of preferences in v3
type PreferencesV3 struct {
	UserID string `json:"userId"`
	Method string `json:"method"`
}

// AuthenticationTraceV3 one authentication trace in v3
type AuthenticationTraceV3 struct {
	UserID     string `json:"userId"`
	Successful bool   `json:"isAuthenticationSuccessful"`
	Date       struct {
		Date int64 `json:"$$date"`
	} `json:"date"`
}

func decodeWebsafeBase64(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, "_", "/")
	s = strings.ReplaceAll(s, "-", "+")

	for len(s)%4 != 0 {
		s += "="
	}

	return base64.StdEncoding.DecodeString(s)
}

func createDBProvider(configurationPath string) storage.Provider {
	config, _ := configuration.Read(configurationPath)

	var dbProvider storage.Provider
	if config.Storage.Local != nil {
		dbProvider = storage.NewSQLiteProvider(config.Storage.Local.Path)
	} else if config.Storage.MySQL != nil {
		dbProvider = storage.NewMySQLProvider(*config.Storage.MySQL)
	} else if config.Storage.PostgreSQL != nil {
		dbProvider = storage.NewPostgreSQLProvider(*config.Storage.PostgreSQL)
	}

	return dbProvider
}
