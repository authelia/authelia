package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestLoadProvidersStorageRunE(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "invalid"), 0000))

	ctx := NewCmdCtx()

	ctx.config.CertificatesDirectory = filepath.Join(dir, "invalid")

	ctx.cconfig = NewCmdCtxConfig()

	err := ctx.LoadProvidersStorageRunE(nil, nil)

	assert.Nil(t, ctx.providers.StorageProvider)
	assert.EqualError(t, err, fmt.Sprintf("had the following errors loading the trusted certificates: could not read certificates from directory open %s: permission denied", filepath.Join(dir, "invalid")))

	ctx.factoryX509SystemCertPool = &TestX509SystemCertPoolFactory{nil, fmt.Errorf("error")}
	ctx.config.CertificatesDirectory = dir

	err = ctx.LoadProvidersStorageRunE(nil, nil)

	assert.Nil(t, ctx.providers.StorageProvider)
	assert.EqualError(t, err, "had the following warnings loading the trusted certificates: could not load system certificate pool which may result in untrusted certificate issues: error")

	ctx.factoryX509SystemCertPool = nil

	ctx.config.CertificatesDirectory = dir

	ctx.cconfig = NewCmdCtxConfig()

	err = ctx.LoadProvidersStorageRunE(nil, nil)

	assert.NoError(t, err)
	assert.Nil(t, ctx.providers.StorageProvider)
}

func TestCmdCtx_ConfigValidateStorageRunE(t *testing.T) {
	dir := t.TempDir()

	ctx := NewCmdCtx()
	ctx.cconfig = NewCmdCtxConfig()

	ctx.cconfig.validator.Push(fmt.Errorf("bad things happened"))
	ctx.cconfig.validator.Push(fmt.Errorf("other things happened"))

	assert.EqualError(t, ctx.ConfigValidateStorageRunE(nil, nil), "bad things happened, other things happened")

	ctx.cconfig = NewCmdCtxConfig()

	assert.EqualError(t, ctx.ConfigValidateStorageRunE(nil, nil), "storage: option 'encryption_key' is required, storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided")

	ctx.cconfig = NewCmdCtxConfig()

	ctx.config.Storage = schema.Storage{
		EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
		Local: &schema.StorageLocal{
			Path: filepath.Join(dir, "db.sqlite3"),
		},
	}

	assert.NoError(t, ctx.ConfigValidateStorageRunE(nil, nil))
}

func TestCmdCtx_StorageBansList(t *testing.T) {
	dir := t.TempDir()

	config := &schema.Configuration{
		Storage: schema.Storage{
			Local: &schema.StorageLocal{
				Path: filepath.Join(dir, "db.sqlite3"),
			},
		},
	}

	store := storage.NewProvider(config, nil)

	require.NoError(t, store.StartupCheck())

	buf := new(bytes.Buffer)

	assert.EqualError(t, runStorageBansList(context.Background(), buf, store, "bad use"), "unknown command \"bad use\"")
	assert.Equal(t, "", buf.String())

	buf.Reset()

	assert.NoError(t, runStorageBansList(context.Background(), buf, store, "ip"))

	assert.Equal(t, "No results.\n", buf.String())
}

func TestRunStorageBansList(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		expected string
		err      string
	}{
		{
			"ShouldListIPBansEmpty",
			"ip",
			"No results.\n",
			"",
		},
		{
			"ShouldListUserBansEmpty",
			"user",
			"No results.\n",
			"",
		},
		{
			"ShouldErrUnknownCommand",
			"bad",
			"",
			"unknown command \"bad\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageBansList(context.Background(), buf, store, tc.use)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, buf.String())
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageCacheDelete(t *testing.T) {
	testCases := []struct {
		name        string
		cacheName   string
		description string
		expected    string
	}{
		{
			"ShouldDeleteMDS3Cache",
			"webauthn_mds3",
			"WebAuthn MDS3",
			"Successfully deleted cached WebAuthn MDS3 data.\n",
		},
		{
			"ShouldDeleteOtherCache",
			"some_cache",
			"some description",
			"Successfully deleted cached some description data.\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageCacheDelete(context.Background(), buf, store, tc.cacheName, tc.description)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestRunStorageSchemaInfo(t *testing.T) {
	t.Run("ShouldShowBasicInfo", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		buf := new(bytes.Buffer)

		err := runStorageSchemaInfo(context.Background(), buf, store)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Schema Version:")
		assert.Contains(t, buf.String(), "Schema Upgrade Available:")
		assert.Contains(t, buf.String(), "Schema Tables:")
		assert.Contains(t, buf.String(), "Schema Encryption Key:")
	})

	t.Run("ShouldShowValidEncryptionKey", func(t *testing.T) {
		store := newTestSQLiteStoreWithEncryptionKey(t)

		buf := new(bytes.Buffer)

		err := runStorageSchemaInfo(context.Background(), buf, store)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Schema Encryption Key: valid")
		assert.Contains(t, buf.String(), "Schema Upgrade Available: no")
	})
}

func TestRunStorageMigrateHistory(t *testing.T) {
	store := newTestSQLiteStore(t)

	buf := new(bytes.Buffer)

	err := runStorageMigrateHistory(context.Background(), buf, store)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Migration History:")
}

func TestRunStorageMigrateList(t *testing.T) {
	testCases := []struct {
		name     string
		up       bool
		expected string
	}{
		{
			"ShouldListUpMigrations",
			true,
			"No Migrations Available",
		},
		{
			"ShouldListDownMigrations",
			false,
			"Storage Schema Migration List (Down)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageMigrateList(context.Background(), buf, store, tc.up)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}
}

func TestRunStorageSchemaEncryptionCheckKey(t *testing.T) {
	testCases := []struct {
		name     string
		verbose  bool
		expected string
	}{
		{
			"ShouldCheckNonVerbose",
			false,
			"Storage Encryption Key Validation:",
		},
		{
			"ShouldCheckVerbose",
			true,
			"Storage Encryption Key Validation:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageSchemaEncryptionCheckKey(context.Background(), buf, store, tc.verbose)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}

	t.Run("ShouldSucceedWithEncryptionNonVerbose", func(t *testing.T) {
		store := newTestSQLiteStoreWithEncryptionKey(t)

		buf := new(bytes.Buffer)

		err := runStorageSchemaEncryptionCheckKey(context.Background(), buf, store, false)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Storage Encryption Key Validation: SUCCESS")
	})

	t.Run("ShouldSucceedWithEncryptionVerbose", func(t *testing.T) {
		store := newTestSQLiteStoreWithEncryptionKey(t)

		buf := new(bytes.Buffer)

		err := runStorageSchemaEncryptionCheckKey(context.Background(), buf, store, true)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Storage Encryption Key Validation: SUCCESS")
		assert.Contains(t, buf.String(), "Tables:")
	})
}

func TestRunStorageSchemaEncryptionChangeKey(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		read     bool
		err      string
		expected string
	}{
		{
			"ShouldErrShortKey",
			"shortkey",
			false,
			"the new encryption key must be at least 20 characters",
			"",
		},
		{
			"ShouldSucceedChangeKey",
			//gitleaks:allow // This is not an actual secret.
			"authelia-new-test-key-not-a-secret-authelia",
			false,
			"",
			"Completed the encryption key change.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			buf := new(bytes.Buffer)

			err := runStorageSchemaEncryptionChangeKey(context.Background(), buf, store, tc.key, tc.read)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageMigration(t *testing.T) {
	t.Run("ShouldErrDownMigrationNoTarget", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		buf := new(bytes.Buffer)

		err := runStorageMigration(context.Background(), buf, store, false, 0, false, false)

		assert.EqualError(t, err, "you must set a target version")
	})

	t.Run("ShouldErrUpMigrationAlreadyLatest", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		buf := new(bytes.Buffer)

		err := runStorageMigration(context.Background(), buf, store, true, storage.SchemaLatest, false, false)

		assert.ErrorContains(t, err, "schema already up to date")
	})

	t.Run("ShouldSucceedDownMigrationWithDestroy", func(t *testing.T) {
		store := newTestSQLiteStoreWithEncryptionKey(t)

		buf := new(bytes.Buffer)

		err := runStorageMigration(context.Background(), buf, store, false, 0, true, true)

		assert.NoError(t, err)
	})
}

func TestRunStorageUserWebAuthnListAll(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldErrNoCredentials",
			false,
			"no WebAuthn credentials in database",
			"",
		},
		{
			"ShouldListCredentials",
			true,
			"",
			"WebAuthn Credentials:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), store, "john", "my-key", []byte("kid-1"))
			}

			buf := new(bytes.Buffer)

			err := runStorageUserWebAuthnListAll(context.Background(), buf, store)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserWebAuthnList(t *testing.T) {
	testCases := []struct {
		name     string
		user     string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldErrNoCredentials",
			"john",
			false,
			"user 'john' has no WebAuthn credentials",
			"",
		},
		{
			"ShouldListCredentials",
			"john",
			true,
			"",
			"WebAuthn Credentials for user 'john':",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), store, "john", "my-key", []byte("kid-1"))
			}

			buf := new(bytes.Buffer)

			err := runStorageUserWebAuthnList(context.Background(), buf, store, tc.user)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserWebAuthnDelete(t *testing.T) {
	testCases := []struct {
		name        string
		seed        bool
		all         bool
		byKID       bool
		description string
		kid         string
		user        string
		err         string
		expected    string
	}{
		{
			"ShouldSucceedDeleteAllForUser",
			false,
			true,
			false,
			"",
			"",
			"john",
			"",
			"Successfully deleted all WebAuthn credentials for user 'john'",
		},
		{
			"ShouldSucceedDeleteByDescription",
			false,
			false,
			false,
			"my-key",
			"",
			"john",
			"",
			"Successfully deleted WebAuthn credential with description 'my-key' for user 'john'",
		},
		{
			"ShouldSucceedDeleteByKID",
			true,
			false,
			true,
			"",
			"a2lkLTE",
			"",
			"",
			"Successfully deleted WebAuthn credential with key id 'a2lkLTE'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), store, "john", "my-key", []byte("kid-1"))
			}

			buf := new(bytes.Buffer)

			err := runStorageUserWebAuthnDelete(context.Background(), buf, store, tc.all, tc.byKID, tc.description, tc.kid, tc.user)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserWebAuthnExport(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldErrNoData",
			false,
			"no data to export",
			"",
		},
		{
			"ShouldSucceedExport",
			true,
			"",
			"Successfully exported 1 WebAuthn credentials as YAML",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), store, "john", "my-key", []byte("kid-1"))
			}

			buf := new(bytes.Buffer)

			filename := filepath.Join(t.TempDir(), "export.yml")

			err := runStorageUserWebAuthnExport(context.Background(), buf, store, filename)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)

				_, statErr := os.Stat(filename)
				assert.NoError(t, statErr)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldSucceedExportImportRoundTrip", func(t *testing.T) {
		store1 := newTestSQLiteStoreWithEncryptionKey(t)

		seedWebAuthnCredential(t, context.Background(), store1, "john", "key-1", []byte("kid-1"))
		seedWebAuthnCredential(t, context.Background(), store1, "harry", "key-2", []byte("kid-2"))

		exportFile := filepath.Join(t.TempDir(), "webauthn-export.yml")
		exportBuf := new(bytes.Buffer)

		require.NoError(t, runStorageUserWebAuthnExport(context.Background(), exportBuf, store1, exportFile))
		assert.Contains(t, exportBuf.String(), "Successfully exported 2 WebAuthn credentials")

		store2 := newTestSQLiteStoreWithEncryptionKey(t)

		importBuf := new(bytes.Buffer)

		require.NoError(t, runStorageUserWebAuthnImport(context.Background(), importBuf, store2, exportFile))
		assert.Contains(t, importBuf.String(), "Successfully imported 2 WebAuthn credentials")

		listBuf := new(bytes.Buffer)

		require.NoError(t, runStorageUserWebAuthnListAll(context.Background(), listBuf, store2))
		assert.Contains(t, listBuf.String(), "john")
		assert.Contains(t, listBuf.String(), "harry")
	})
}

func TestRunStorageUserWebAuthnImport(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T) string
		err   string
	}{
		{
			"ShouldErrFileNotFound",
			func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.yml")
			},
			"must specify a filename that exists but",
		},
		{
			"ShouldErrDirectory",
			func(t *testing.T) string {
				return t.TempDir()
			},
			"is a directory",
		},
		{
			"ShouldErrEmptyYAML",
			func(t *testing.T) string {
				dir := t.TempDir()
				f := filepath.Join(dir, "empty.yml")

				require.NoError(t, os.WriteFile(f, []byte("{}"), 0600))

				return f
			},
			"can't import a YAML file without WebAuthn credentials data",
		},
		{
			"ShouldErrInvalidYAML",
			func(t *testing.T) string {
				dir := t.TempDir()
				f := filepath.Join(dir, "invalid.yml")

				require.NoError(t, os.WriteFile(f, []byte("{{invalid"), 0600))

				return f
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			filename := tc.setup(t)

			err := runStorageUserWebAuthnImport(context.Background(), buf, store, filename)

			assert.Error(t, err)

			if tc.err != "" {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldSucceedImportRoundTrip", func(t *testing.T) {
		store := newTestSQLiteStoreWithEncryptionKey(t)

		seedWebAuthnCredential(t, context.Background(), store, "john", "my-key", []byte("kid-1"))

		exportFile := filepath.Join(t.TempDir(), "export.yml")

		buf := new(bytes.Buffer)

		require.NoError(t, runStorageUserWebAuthnExport(context.Background(), buf, store, exportFile))

		store2 := newTestSQLiteStoreWithEncryptionKey(t)

		buf.Reset()

		err := runStorageUserWebAuthnImport(context.Background(), buf, store2, exportFile)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Successfully imported 1 WebAuthn credentials")
	})
}

func TestRunStorageUserTOTPExportURI(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		expected string
	}{
		{
			"ShouldExportEmpty",
			false,
			"Successfully exported 0 TOTP configurations as TOTP URI's",
		},
		{
			"ShouldExportWithData",
			true,
			"Successfully exported 1 TOTP configurations as TOTP URI's",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), store, "john")
			}

			buf := new(bytes.Buffer)

			err := runStorageUserTOTPExportURI(context.Background(), buf, store)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}
}

func TestRunStorageUserTOTPExportCSV(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldErrBlankFilename",
			"",
			false,
			"must specify a filename to export to",
			"",
		},
		{
			"ShouldErrWhitespaceFilename",
			"   ",
			false,
			"must specify a filename to export to",
			"",
		},
		{
			"ShouldSucceedEmptyExport",
			"auto",
			false,
			"",
			"Successfully exported 0 TOTP configurations as CSV",
		},
		{
			"ShouldSucceedWithData",
			"auto",
			true,
			"",
			"Successfully exported 1 TOTP configurations as CSV",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), store, "john")
			}

			buf := new(bytes.Buffer)

			filename := tc.filename
			if filename == "auto" {
				filename = filepath.Join(t.TempDir(), "export.csv")
			}

			err := runStorageUserTOTPExportCSV(context.Background(), buf, store, filename, 10)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserTOTPExport(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldErrNoData",
			false,
			"no data to export",
			"",
		},
		{
			"ShouldSucceedExport",
			true,
			"",
			"Successfully exported 1 TOTP configurations as YAML",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStoreWithEncryptionKey(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), store, "john")
			}

			buf := new(bytes.Buffer)

			filename := filepath.Join(t.TempDir(), "export.yml")

			err := runStorageUserTOTPExport(context.Background(), buf, store, filename)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserTOTPImport(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T) string
		err   string
	}{
		{
			"ShouldErrFileNotFound",
			func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.yml")
			},
			"must specify a filename that exists but",
		},
		{
			"ShouldErrDirectory",
			func(t *testing.T) string {
				return t.TempDir()
			},
			"is a directory",
		},
		{
			"ShouldErrEmptyYAML",
			func(t *testing.T) string {
				dir := t.TempDir()
				f := filepath.Join(dir, "empty.yml")

				require.NoError(t, os.WriteFile(f, []byte("{}"), 0600))

				return f
			},
			"can't import a YAML file without TOTP configuration data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			filename := tc.setup(t)

			err := runStorageUserTOTPImport(context.Background(), buf, store, filename)

			assert.Error(t, err)
			assert.ErrorContains(t, err, tc.err)
		})
	}

	t.Run("ShouldSucceedImportRoundTrip", func(t *testing.T) {
		store := newTestSQLiteStoreWithEncryptionKey(t)

		seedTOTPConfig(t, context.Background(), store, "john")

		exportFile := filepath.Join(t.TempDir(), "export.yml")

		buf := new(bytes.Buffer)

		require.NoError(t, runStorageUserTOTPExport(context.Background(), buf, store, exportFile))

		store2 := newTestSQLiteStoreWithEncryptionKey(t)

		buf.Reset()

		err := runStorageUserTOTPImport(context.Background(), buf, store2, exportFile)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Successfully imported 1 TOTP configurations")
	})
}

func TestRunStorageUserIdentifiersExport(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, store storage.Provider) string
		err   string
	}{
		{
			"ShouldErrFileExists",
			func(t *testing.T, _ storage.Provider) string {
				dir := t.TempDir()
				f := filepath.Join(dir, "export.yml")

				require.NoError(t, os.WriteFile(f, []byte(""), 0600))

				return f
			},
			"must specify a file that doesn't exist but",
		},
		{
			"ShouldErrNoData",
			func(t *testing.T, _ storage.Provider) string {
				return filepath.Join(t.TempDir(), "export.yml")
			},
			"no data to export",
		},
		{
			"ShouldSucceedExport",
			func(t *testing.T, store storage.Provider) string {
				seedUserOpaqueIdentifier(t, context.Background(), store, "john", "openid", "example.com")

				return filepath.Join(t.TempDir(), "export.yml")
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			filename := tc.setup(t, store)

			err := runStorageUserIdentifiersExport(context.Background(), buf, store, filename)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), "Successfully exported")
			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserIdentifiersImport(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T) string
		err   string
	}{
		{
			"ShouldErrFileNotFound",
			func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.yml")
			},
			"must specify a file that exists but",
		},
		{
			"ShouldErrDirectory",
			func(t *testing.T) string {
				return t.TempDir()
			},
			"is a directory",
		},
		{
			"ShouldErrEmptyYAML",
			func(t *testing.T) string {
				dir := t.TempDir()
				f := filepath.Join(dir, "empty.yml")

				require.NoError(t, os.WriteFile(f, []byte("{}"), 0600))

				return f
			},
			"can't import a YAML file without User Opaque Identifiers data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			filename := tc.setup(t)

			err := runStorageUserIdentifiersImport(context.Background(), buf, store, filename)

			assert.Error(t, err)
			assert.ErrorContains(t, err, tc.err)
		})
	}

	t.Run("ShouldSucceedImportRoundTrip", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		seedUserOpaqueIdentifier(t, context.Background(), store, "john", "openid", "example.com")

		exportFile := filepath.Join(t.TempDir(), "export.yml")

		buf := new(bytes.Buffer)

		require.NoError(t, runStorageUserIdentifiersExport(context.Background(), buf, store, exportFile))

		store2 := newTestSQLiteStore(t)

		buf.Reset()

		err := runStorageUserIdentifiersImport(context.Background(), buf, store2, exportFile)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Successfully imported")
	})
}

func TestRunStorageUserIdentifiersGenerate(t *testing.T) {
	testCases := []struct {
		name     string
		users    []string
		services []string
		sectors  []string
		err      string
		expected string
	}{
		{
			"ShouldErrNoUsers",
			nil,
			[]string{"openid"},
			nil,
			"must supply at least one user",
			"",
		},
		{
			"ShouldErrInvalidService",
			[]string{"john"},
			[]string{"invalid"},
			nil,
			"one or more the service names 'invalid' is invalid",
			"",
		},
		{
			"ShouldGenerateSuccessfully",
			[]string{"john"},
			[]string{"openid"},
			nil,
			"",
			"Successfully generated and added opaque identifiers:",
		},
		{
			"ShouldGenerateMultipleUsers",
			[]string{"john", "harry"},
			[]string{"openid"},
			[]string{"example.com"},
			"",
			"Successfully generated and added opaque identifiers:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageUserIdentifiersGenerate(context.Background(), buf, store, tc.users, tc.services, tc.sectors)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldSkipDuplicates", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		buf := new(bytes.Buffer)

		err := runStorageUserIdentifiersGenerate(context.Background(), buf, store, []string{"john"}, []string{"openid"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Total: 1")

		buf.Reset()

		err = runStorageUserIdentifiersGenerate(context.Background(), buf, store, []string{"john"}, []string{"openid"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Skipped Duplicates: 1")
		assert.Contains(t, buf.String(), "Total: 0")
	})
}

func TestRunStorageUserIdentifiersAdd(t *testing.T) {
	testCases := []struct {
		name          string
		service       string
		sector        string
		username      string
		identifier    string
		useIdentifier bool
		err           string
		expected      string
	}{
		{
			"ShouldSucceedWithDefaults",
			"",
			"",
			"john",
			"",
			false,
			"",
			"Added User Opaque Identifier:",
		},
		{
			"ShouldSucceedWithService",
			"openid",
			"example.com",
			"john",
			"",
			false,
			"",
			"Added User Opaque Identifier:",
		},
		{
			"ShouldErrInvalidService",
			"badservice",
			"",
			"john",
			"",
			false,
			"the service name 'badservice' is invalid",
			"",
		},
		{
			"ShouldSucceedWithIdentifier",
			"openid",
			"",
			"john",
			"fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70",
			true,
			"",
			"Added User Opaque Identifier:",
		},
		{
			"ShouldErrInvalidIdentifier",
			"openid",
			"",
			"john",
			"not-a-uuid",
			true,
			"the identifier provided 'not-a-uuid' is invalid",
			"",
		},
		{
			"ShouldErrNonV4UUID",
			"openid",
			"",
			"john",
			"00000000-0000-1000-8000-000000000000",
			true,
			"is a version 1 UUID but only version 4",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageUserIdentifiersAdd(context.Background(), buf, store, tc.service, tc.sector, tc.username, tc.identifier, tc.useIdentifier)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageBansRevoke(t *testing.T) {
	testCases := []struct {
		name string
		use  string
		err  string
	}{
		{
			"ShouldErrUnknownCommand",
			"bad",
			"unknown command \"bad\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			flags := newFlagSetWithInt("id", 0)

			err := runStorageBansRevoke(context.Background(), buf, flags, nil, store, tc.use)

			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestRunStorageBansRevokeIP(t *testing.T) {
	testCases := []struct {
		name   string
		id     int
		target string
		err    string
	}{
		{
			"ShouldErrNoIDOrIP",
			0,
			"",
			"either the ip or id is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageBansRevokeIP(context.Background(), buf, store, tc.id, tc.target)

			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestRunStorageBansRevokeUser(t *testing.T) {
	testCases := []struct {
		name   string
		id     int
		target string
		err    string
	}{
		{
			"ShouldErrNoIDOrUsername",
			0,
			"",
			"either the username or id is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageBansRevokeUser(context.Background(), buf, store, tc.id, tc.target)

			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestRunStorageBansAddIP(t *testing.T) {
	testCases := []struct {
		name     string
		target   string
		err      string
		expected string
	}{
		{
			"ShouldErrInvalidIP",
			"not-an-ip",
			"invalid IP address: not-an-ip",
			"",
		},
		{
			"ShouldSucceedPermanent",
			"192.168.1.1",
			"",
			"Successfully banned IP '192.168.1.1' permanently.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageBansAddIP(context.Background(), buf, store, tc.target, "", 0, true)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldSucceedAddIPTemporaryWithReason", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		buf := new(bytes.Buffer)

		err := runStorageBansAddIP(context.Background(), buf, store, "192.168.1.1", "malicious", time.Hour, false)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Successfully banned IP '192.168.1.1' until")
	})
}

func TestRunStorageBansAddUser(t *testing.T) {
	testCases := []struct {
		name      string
		target    string
		reason    string
		permanent bool
		expected  string
	}{
		{
			"ShouldSucceedPermanent",
			"john",
			"",
			true,
			"Successfully banned user 'john' permanently.",
		},
		{
			"ShouldSucceedWithReason",
			"john",
			"too many attempts",
			true,
			"Successfully banned user 'john' permanently.",
		},
		{
			"ShouldSucceedTemporary",
			"john",
			"",
			false,
			"Successfully banned user 'john' until",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			buf := new(bytes.Buffer)

			err := runStorageBansAddUser(context.Background(), buf, store, tc.target, tc.reason, time.Hour, tc.permanent)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}
}

func TestRunStorageCacheMDS3Dump(t *testing.T) {
	testCases := []struct {
		name string
		path string
		err  string
	}{
		{
			"ShouldErrBlankPath",
			"",
			"error dumping metadata: path must not be blank",
		},
		{
			"ShouldErrWhitespacePath",
			"   ",
			"error dumping metadata: path must not be blank",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			config := &schema.Configuration{
				WebAuthn: schema.WebAuthn{
					Metadata: schema.WebAuthnMetadata{
						Enabled: true,
					},
				},
			}

			buf := new(bytes.Buffer)

			err := runStorageCacheMDS3Dump(context.Background(), buf, store, config, tc.path)

			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestRunStorageCacheMDS3Status(t *testing.T) {
	testCases := []struct {
		name     string
		enabled  bool
		err      string
		expected string
	}{
		{
			"ShouldErrDisabled",
			false,
			"webauthn metadata is disabled",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			config := &schema.Configuration{
				WebAuthn: schema.WebAuthn{
					Metadata: schema.WebAuthnMetadata{
						Enabled: tc.enabled,
					},
				},
			}

			buf := new(bytes.Buffer)

			err := runStorageCacheMDS3Status(context.Background(), buf, store, config)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageCacheMDS3Update(t *testing.T) {
	testCases := []struct {
		name  string
		force bool
		err   string
	}{
		{
			"ShouldErrDisabled",
			false,
			"webauthn metadata is disabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			config := &schema.Configuration{}

			buf := new(bytes.Buffer)

			err := runStorageCacheMDS3Update(context.Background(), buf, store, config, "", tc.force)

			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestRunStorageUserTOTPExportPNG(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		dirExist bool
		err      string
		expected string
	}{
		{
			"ShouldErrDirExists",
			false,
			true,
			"output directory must not exist",
			"",
		},
		{
			"ShouldSucceedEmpty",
			false,
			false,
			"",
			"Successfully exported 0 TOTP configuration as QR codes in PNG format",
		},
		{
			"ShouldSucceedWithData",
			true,
			false,
			"",
			"Successfully exported 1 TOTP configuration as QR codes in PNG format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), cmdCtx.providers.StorageProvider, "john")
			}

			buf := new(bytes.Buffer)

			dir := filepath.Join(t.TempDir(), "png-export")
			if tc.dirExist {
				require.NoError(t, os.MkdirAll(dir, 0700))
			}

			err := runStorageUserTOTPExportPNG(cmdCtx, buf, cmdCtx.providers.StorageProvider, dir)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserTOTPGenerate(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		seed     bool
		force    bool
		err      string
		expected string
	}{
		{
			"ShouldSucceedGenerate",
			"john",
			false,
			false,
			"",
			"Successfully generated TOTP configuration for user 'john'",
		},
		{
			"ShouldErrAlreadyExists",
			"john",
			true,
			false,
			"john already has a TOTP configuration, use --force to overwrite",
			"",
		},
		{
			"ShouldSucceedForceOverwrite",
			"john",
			true,
			true,
			"",
			"Successfully generated TOTP configuration for user 'john'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), cmdCtx.providers.StorageProvider, tc.username)
			}

			buf := new(bytes.Buffer)

			err := runStorageUserTOTPGenerate(cmdCtx, buf, cmdCtx.providers.StorageProvider, cmdCtx.config, "", tc.username, "", tc.force)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldSucceedGenerateWithPNG", func(t *testing.T) {
		cmdCtx := newTestCmdCtx(t)

		buf := new(bytes.Buffer)

		pngFile := filepath.Join(t.TempDir(), "totp.png")

		err := runStorageUserTOTPGenerate(cmdCtx, buf, cmdCtx.providers.StorageProvider, cmdCtx.config, pngFile, "john", "", false)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Successfully generated TOTP configuration for user 'john'")
		assert.Contains(t, buf.String(), "saved it as a PNG image")

		_, statErr := os.Stat(pngFile)
		assert.NoError(t, statErr)
	})

	t.Run("ShouldErrPNGFileExists", func(t *testing.T) {
		cmdCtx := newTestCmdCtx(t)

		buf := new(bytes.Buffer)

		pngFile := filepath.Join(t.TempDir(), "totp.png")

		require.NoError(t, os.WriteFile(pngFile, []byte(""), 0600))

		err := runStorageUserTOTPGenerate(cmdCtx, buf, cmdCtx.providers.StorageProvider, cmdCtx.config, pngFile, "john", "", false)

		assert.EqualError(t, err, "image output filepath already exists")
	})
}

func TestStorageUserTOTPDeleteRunE(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		user     string
		err      string
		expected string
	}{
		{
			"ShouldErrNotFound",
			false,
			"john",
			"failed to delete TOTP configuration for user 'john':",
			"",
		},
		{
			"ShouldSucceedDelete",
			true,
			"john",
			"",
			"Successfully deleted TOTP configuration for user 'john'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), cmdCtx.providers.StorageProvider, tc.user)
			}

			cmd, buf := newTestCmdWithBuf()

			err := cmdCtx.StorageUserTOTPDeleteRunE(cmd, []string{tc.user})

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageUserTOTPGenerateRunE(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		args     []string
		flags    map[string]string
		err      string
		expected string
	}{
		{
			"ShouldSucceedGenerate",
			false,
			[]string{"john"},
			nil,
			"",
			"Successfully generated TOTP configuration for user 'john'",
		},
		{
			"ShouldErrAlreadyExists",
			true,
			[]string{"john"},
			nil,
			"john already has a TOTP configuration, use --force to overwrite",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), cmdCtx.providers.StorageProvider, "john")
			}

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().Bool("force", false, "")
			cmd.Flags().String("path", "", "")
			cmd.Flags().String("secret", "", "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmdCtx.StorageUserTOTPGenerateRunE(cmd, tc.args)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageUserTOTPExportRunE(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		fileFlag string
		err      string
		expected string
	}{
		{
			"ShouldErrNoData",
			false,
			"auto",
			"no data to export",
			"",
		},
		{
			"ShouldSucceedExport",
			true,
			"auto",
			"",
			"Successfully exported 1 TOTP configurations as YAML",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedTOTPConfig(t, context.Background(), cmdCtx.providers.StorageProvider, "john")
			}

			cmd, buf := newTestCmdWithBuf()

			filename := filepath.Join(t.TempDir(), "export.yml")

			cmd.Flags().String(cmdFlagNameFile, filename, "")

			err := cmdCtx.StorageUserTOTPExportRunE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldErrFileExists", func(t *testing.T) {
		cmdCtx := newTestCmdCtx(t)

		cmd, _ := newTestCmdWithBuf()

		filename := filepath.Join(t.TempDir(), "export.yml")
		require.NoError(t, os.WriteFile(filename, []byte(""), 0600))

		cmd.Flags().String(cmdFlagNameFile, filename, "")

		err := cmdCtx.StorageUserTOTPExportRunE(cmd, nil)

		assert.ErrorContains(t, err, "must specify a file that doesn't exist but")
	})
}

func TestStorageUserTOTPImportRunE(t *testing.T) {
	cmdCtx1 := newTestCmdCtx(t)

	seedTOTPConfig(t, context.Background(), cmdCtx1.providers.StorageProvider, "john")

	exportFile := filepath.Join(t.TempDir(), "export.yml")
	buf := new(bytes.Buffer)

	require.NoError(t, runStorageUserTOTPExport(context.Background(), buf, cmdCtx1.providers.StorageProvider, exportFile))

	cmdCtx2 := newTestCmdCtx(t)

	cmd, buf2 := newTestCmdWithBuf()

	err := cmdCtx2.StorageUserTOTPImportRunE(cmd, []string{exportFile})

	assert.NoError(t, err)
	assert.Contains(t, buf2.String(), "Successfully imported 1 TOTP configurations")
}

func TestStorageUserTOTPExportCSVRunE(t *testing.T) {
	cmdCtx := newTestCmdCtx(t)

	seedTOTPConfig(t, context.Background(), cmdCtx.providers.StorageProvider, "john")

	cmd, buf := newTestCmdWithBuf()

	filename := filepath.Join(t.TempDir(), "export.csv")
	cmd.Flags().String(cmdFlagNameFile, filename, "")

	err := cmdCtx.StorageUserTOTPExportCSVRunE(cmd, nil)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Successfully exported 1 TOTP configurations as CSV")
}

func TestStorageUserTOTPExportPNGRunE(t *testing.T) {
	cmdCtx := newTestCmdCtx(t)

	seedTOTPConfig(t, context.Background(), cmdCtx.providers.StorageProvider, "john")

	cmd, buf := newTestCmdWithBuf()

	dir := filepath.Join(t.TempDir(), "png-export")
	cmd.Flags().String(cmdFlagNameDirectory, dir, "")

	err := cmdCtx.StorageUserTOTPExportPNGRunE(cmd, nil)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Successfully exported 1 TOTP configuration as QR codes in PNG format")

	_, statErr := os.Stat(filepath.Join(dir, "john.png"))
	assert.NoError(t, statErr)
}

func TestStorageUserWebAuthnDeleteRunE(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		flags    map[string]string
		err      string
		expected string
	}{
		{
			"ShouldSucceedDeleteAllForUser",
			[]string{"john"},
			map[string]string{cmdFlagNameAll: "true"},
			"",
			"Successfully deleted all WebAuthn credentials for user 'john'",
		},
		{
			"ShouldSucceedDeleteByDescription",
			[]string{"john"},
			map[string]string{cmdFlagNameDescription: "my-key"},
			"",
			"Successfully deleted WebAuthn credential with description 'my-key' for user 'john'",
		},
		{
			"ShouldSucceedDeleteByKID",
			nil,
			map[string]string{cmdFlagNameKeyID: "abc123"},
			"",
			"Successfully deleted WebAuthn credential with key id 'abc123'",
		},
		{
			"ShouldErrNoFlags",
			[]string{"john"},
			nil,
			"must supply one of the flags --all, --description, or --kid",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			seedWebAuthnCredential(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "my-key", []byte("kid-1"))

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().Bool(cmdFlagNameAll, false, "")
			cmd.Flags().String(cmdFlagNameDescription, "", "")
			cmd.Flags().String(cmdFlagNameKeyID, "", "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmdCtx.StorageUserWebAuthnDeleteRunE(cmd, tc.args)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageUserWebAuthnListAllRunE(t *testing.T) {
	cmdCtx := newTestCmdCtx(t)

	seedWebAuthnCredential(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "my-key", []byte("kid-1"))

	cmd, buf := newTestCmdWithBuf()

	err := cmdCtx.StorageUserWebAuthnListAllRunE(cmd, nil)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "WebAuthn Credentials:")
	assert.Contains(t, buf.String(), "john")
}

func TestStorageUserWebAuthnListRunE(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldListAllWhenNoArgs",
			nil,
			true,
			"",
			"WebAuthn Credentials:",
		},
		{
			"ShouldListAllWhenEmptyArg",
			[]string{""},
			true,
			"",
			"WebAuthn Credentials:",
		},
		{
			"ShouldListForUser",
			[]string{"john"},
			true,
			"",
			"WebAuthn Credentials for user 'john':",
		},
		{
			"ShouldErrNoCredentialsForUser",
			[]string{"nobody"},
			true,
			"user 'nobody' has no WebAuthn credentials",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "my-key", []byte("kid-1"))
			}

			cmd, buf := newTestCmdWithBuf()

			err := cmdCtx.StorageUserWebAuthnListRunE(cmd, tc.args)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageUserIdentifiersAddRunE(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		flags    map[string]string
		err      string
		expected string
	}{
		{
			"ShouldSucceedDefaults",
			[]string{"john"},
			nil,
			"",
			"Added User Opaque Identifier:",
		},
		{
			"ShouldSucceedWithServiceAndSector",
			[]string{"john"},
			map[string]string{cmdFlagNameService: "openid", cmdFlagNameSector: "example.com"},
			"",
			"Added User Opaque Identifier:",
		},
		{
			"ShouldSucceedWithIdentifier",
			[]string{"john"},
			map[string]string{cmdFlagNameIdentifier: "fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70"},
			"",
			"Added User Opaque Identifier:",
		},
		{
			"ShouldErrInvalidService",
			[]string{"john"},
			map[string]string{cmdFlagNameService: "badservice"},
			"the service name 'badservice' is invalid",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().String(cmdFlagNameService, "", "")
			cmd.Flags().String(cmdFlagNameSector, "", "")
			cmd.Flags().String(cmdFlagNameIdentifier, "", "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmdCtx.StorageUserIdentifiersAddRunE(cmd, tc.args)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageUserIdentifiersGenerateRunE(t *testing.T) {
	testCases := []struct {
		name     string
		flags    map[string]string
		err      string
		expected string
	}{
		{
			"ShouldSucceedGenerate",
			map[string]string{
				cmdFlagNameUsers:    "john,harry",
				cmdFlagNameServices: "openid",
				cmdFlagNameSectors:  "example.com",
			},
			"",
			"Successfully generated and added opaque identifiers:",
		},
		{
			"ShouldErrNoUsers",
			map[string]string{
				cmdFlagNameServices: "openid",
			},
			"must supply at least one user",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().StringSlice(cmdFlagNameUsers, nil, "")
			cmd.Flags().StringSlice(cmdFlagNameServices, nil, "")
			cmd.Flags().StringSlice(cmdFlagNameSectors, nil, "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmdCtx.StorageUserIdentifiersGenerateRunE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageUserIdentifiersExportRunE(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldErrNoData",
			false,
			"no data to export",
			"",
		},
		{
			"ShouldSucceedExport",
			true,
			"",
			"Successfully exported",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedUserOpaqueIdentifier(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "openid", "example.com")
			}

			cmd, buf := newTestCmdWithBuf()

			filename := filepath.Join(t.TempDir(), "export.yml")
			cmd.Flags().String(cmdFlagNameFile, filename, "")

			err := cmdCtx.StorageUserIdentifiersExportRunE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageSchemaEncryptionChangeKeyRunE(t *testing.T) {
	testCases := []struct {
		name     string
		flags    map[string]string
		err      string
		expected string
	}{
		{
			"ShouldSucceedChangeKey",
			//gitleaks:allow // This is not an actual secret.
			map[string]string{cmdFlagNameNewEncryptionKey: "authelia-new-test-key-not-a-secret-authelia"},
			"",
			"Completed the encryption key change.",
		},
		{
			"ShouldErrShortKey",
			map[string]string{cmdFlagNameNewEncryptionKey: "shortkey"},
			"the new encryption key must be at least 20 characters",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().String(cmdFlagNameNewEncryptionKey, "", "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmdCtx.StorageSchemaEncryptionChangeKeyRunE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageSchemaEncryptionCheckRunE(t *testing.T) {
	testCases := []struct {
		name     string
		verbose  bool
		expected string
	}{
		{
			"ShouldSucceedNonVerbose",
			false,
			"Storage Encryption Key Validation: SUCCESS",
		},
		{
			"ShouldSucceedVerbose",
			true,
			"Storage Encryption Key Validation: SUCCESS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().Bool(cmdFlagNameVerbose, tc.verbose, "")

			err := cmdCtx.StorageSchemaEncryptionCheckRunE(cmd, nil)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}
}

func TestStorageMigrateHistoryRunE(t *testing.T) {
	cmdCtx := newTestCmdCtx(t)

	cmd, buf := newTestCmdWithBuf()

	err := cmdCtx.StorageMigrateHistoryRunE(cmd, nil)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Migration History:")
	assert.Contains(t, buf.String(), "ID")
	assert.Contains(t, buf.String(), "Date")
	assert.Contains(t, buf.String(), "Authelia Version")
}

func TestNewStorageMigrateListRunE(t *testing.T) {
	testCases := []struct {
		name     string
		up       bool
		expected string
	}{
		{
			"ShouldListUpMigrations",
			true,
			"No Migrations Available",
		},
		{
			"ShouldListDownMigrations",
			false,
			"Storage Schema Migration List (Down)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			cmd, buf := newTestCmdWithBuf()

			runE := cmdCtx.NewStorageMigrateListRunE(tc.up)

			err := runE(cmd, nil)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}
}

func TestNewStorageMigrationRunE(t *testing.T) {
	testCases := []struct {
		name  string
		up    bool
		flags map[string]string
		err   string
	}{
		{
			"ShouldErrDownMigrationNoTarget",
			false,
			nil,
			"you must set a target version",
		},
		{
			"ShouldErrUpMigrationAlreadyLatest",
			true,
			nil,
			"schema already up to date",
		},
		{
			"ShouldErrUpMigrationTargetSameAsCurrent",
			true,
			map[string]string{cmdFlagNameTarget: "24"},
			"schema migration target version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			cmd, _ := newTestCmdWithBuf()
			cmd.Flags().Int(cmdFlagNameTarget, 0, "")
			cmd.Flags().Bool(cmdFlagNameDestroyData, false, "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			runE := cmdCtx.NewStorageMigrationRunE(tc.up)

			err := runE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageSchemaInfoRunE(t *testing.T) {
	testCases := []struct {
		name           string
		useEncryption  bool
		expectedFields []string
	}{
		{
			"ShouldShowSchemaInfoWithEncryption",
			true,
			[]string{"Schema Version:", "Schema Upgrade Available: no", "Schema Tables:", "Schema Encryption Key: valid"},
		},
		{
			"ShouldShowSchemaInfoWithoutEncryption",
			false,
			[]string{"Schema Version:", "Schema Upgrade Available: no", "Schema Tables:", "Schema Encryption Key:"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cmdCtx *CmdCtx

			if tc.useEncryption {
				cmdCtx = newTestCmdCtx(t)
			} else {
				cmdCtx = NewCmdCtx()

				dir := t.TempDir()

				cmdCtx.config.Storage = schema.Storage{
					Local: &schema.StorageLocal{
						Path: filepath.Join(dir, "db.sqlite3"),
					},
				}

				cmdCtx.providers.StorageProvider = storage.NewProvider(cmdCtx.config, nil)

				require.NoError(t, cmdCtx.providers.StorageProvider.StartupCheck())
			}

			cmd, buf := newTestCmdWithBuf()

			err := cmdCtx.StorageSchemaInfoRunE(cmd, nil)

			assert.NoError(t, err)

			for _, expected := range tc.expectedFields {
				assert.Contains(t, buf.String(), expected)
			}
		})
	}
}

func TestStorageUserWebAuthnExportRunE(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		err      string
		expected string
	}{
		{
			"ShouldErrNoData",
			false,
			"no data to export",
			"",
		},
		{
			"ShouldSucceedExport",
			true,
			"",
			"Successfully exported 1 WebAuthn credentials as YAML",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "my-key", []byte("kid-1"))
			}

			cmd, buf := newTestCmdWithBuf()

			filename := filepath.Join(t.TempDir(), "export.yml")
			cmd.Flags().String(cmdFlagNameFile, filename, "")

			err := cmdCtx.StorageUserWebAuthnExportRunE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)

				_, statErr := os.Stat(filename)
				assert.NoError(t, statErr)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldErrFileExists", func(t *testing.T) {
		cmdCtx := newTestCmdCtx(t)

		cmd, _ := newTestCmdWithBuf()

		filename := filepath.Join(t.TempDir(), "export.yml")
		require.NoError(t, os.WriteFile(filename, []byte(""), 0600))

		cmd.Flags().String(cmdFlagNameFile, filename, "")

		err := cmdCtx.StorageUserWebAuthnExportRunE(cmd, nil)

		assert.ErrorContains(t, err, "must specify a file that doesn't exist but")
	})
}

func TestStorageUserWebAuthnImportRunE(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(t *testing.T, store storage.Provider) string
		err      string
		expected string
	}{
		{
			"ShouldSucceedImport",
			func(t *testing.T, store storage.Provider) string {
				seedWebAuthnCredential(t, context.Background(), store, "john", "my-key", []byte("kid-1"))

				exportFile := filepath.Join(t.TempDir(), "export.yml")
				buf := new(bytes.Buffer)

				require.NoError(t, runStorageUserWebAuthnExport(context.Background(), buf, store, exportFile))

				return exportFile
			},
			"",
			"Successfully imported 1 WebAuthn credentials",
		},
		{
			"ShouldErrEmptyYAML",
			func(t *testing.T, _ storage.Provider) string {
				f := filepath.Join(t.TempDir(), "empty.yml")

				require.NoError(t, os.WriteFile(f, []byte("{}"), 0600))

				return f
			},
			"can't import a YAML file without WebAuthn credentials data",
			"",
		},
		{
			"ShouldErrFileNotFound",
			func(t *testing.T, _ storage.Provider) string {
				return filepath.Join(t.TempDir(), "nonexistent.yml")
			},
			"must specify a filename that exists but",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exportCtx := newTestCmdCtx(t)
			importCtx := newTestCmdCtx(t)

			filename := tc.setup(t, exportCtx.providers.StorageProvider)

			cmd, buf := newTestCmdWithBuf()

			err := importCtx.StorageUserWebAuthnImportRunE(cmd, []string{filename})

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageUserWebAuthnVerifyRunE(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		flag     bool
		err      string
		expected string
	}{
		{
			"ShouldErrNoCredentials",
			false,
			true,
			"no WebAuthn credentials in database",
			"",
		},
		{
			"ShouldSucceedVerifyCredentials",
			true,
			true,
			"",
			"WebAuthn Credential Verifications:",
		},
		{
			"ShouldErrorNoFlag",
			false,
			false,
			"flag accessed but not defined: verbose",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "my-key", []byte("kid-1"))
			}

			cmd, buf := newTestCmdWithBuf()

			if tc.flag {
				cmd.Flags().Bool(cmdFlagNameVerbose, false, "")
			}

			err := cmdCtx.StorageUserWebAuthnVerifyRunE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
				assert.Contains(t, buf.String(), "ID")
				assert.Contains(t, buf.String(), "RPID")
				assert.Contains(t, buf.String(), "Username")
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageUserWebAuthnVerify(t *testing.T) {
	testCases := []struct {
		name     string
		seed     bool
		verbose  bool
		err      string
		expected []string
	}{
		{
			"ShouldErrNoCredentials",
			false,
			false,
			"no WebAuthn credentials in database",
			nil,
		},
		{
			"ShouldSucceedVerify",
			true,
			true,
			"",
			[]string{"WebAuthn Credential Verifications:", "john", "example.com"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedWebAuthnCredential(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "my-key", []byte("kid-1"))
			}

			buf := new(bytes.Buffer)

			err := runStorageUserWebAuthnVerify(context.Background(), buf, cmdCtx.providers.StorageProvider, cmdCtx.config, tc.verbose)

			if tc.err == "" {
				assert.NoError(t, err)

				for _, s := range tc.expected {
					assert.Contains(t, buf.String(), s)
				}
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageBansListWithData(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		seedIP   bool
		seedUser bool
		expected []string
	}{
		{
			"ShouldListIPBansWithData",
			"ip",
			true,
			false,
			[]string{"ID", "IP", "Expires", "Source", "Reason", "192.168.1.1", "cli"},
		},
		{
			"ShouldListUserBansWithData",
			"user",
			false,
			true,
			[]string{"ID", "Username", "Expires", "Source", "Reason", "john", "cli"},
		},
		{
			"ShouldListIPBansWithReason",
			"ip",
			true,
			false,
			[]string{"192.168.1.1", "cli"},
		},
		{
			"ShouldListUserBansWithReason",
			"user",
			false,
			true,
			[]string{"john", "cli"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			if tc.seedIP {
				seedBanIP(t, context.Background(), store, "192.168.1.1", "malicious", true)
			}

			if tc.seedUser {
				seedBanUser(t, context.Background(), store, "john", "too many attempts", true)
			}

			buf := new(bytes.Buffer)

			err := runStorageBansList(context.Background(), buf, store, tc.use)

			assert.NoError(t, err)

			for _, s := range tc.expected {
				assert.Contains(t, buf.String(), s)
			}
		})
	}
}

func TestRunStorageBansRevokeIPWithData(t *testing.T) {
	testCases := []struct {
		name     string
		id       int
		target   string
		err      string
		expected string
	}{
		{
			"ShouldErrNoIDOrIP",
			0,
			"",
			"either the ip or id is required",
			"",
		},
		{
			"ShouldSucceedRevokeByIP",
			0,
			"192.168.1.1",
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeByID",
			1,
			"",
			"",
			"SUCCESS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			seedBanIP(t, context.Background(), store, "192.168.1.1", "malicious", true)

			buf := new(bytes.Buffer)

			err := runStorageBansRevokeIP(context.Background(), buf, store, tc.id, tc.target)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldSkipAlreadyRevoked", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		seedBanIP(t, context.Background(), store, "192.168.1.1", "", true)

		buf := new(bytes.Buffer)

		err := runStorageBansRevokeIP(context.Background(), buf, store, 1, "")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "SUCCESS")

		buf.Reset()

		err = runStorageBansRevokeIP(context.Background(), buf, store, 1, "")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "SKIPPED")
		assert.Contains(t, buf.String(), "Ban has already been revoked")
	})
}

func TestRunStorageBansRevokeUserWithData(t *testing.T) {
	testCases := []struct {
		name     string
		id       int
		target   string
		err      string
		expected string
	}{
		{
			"ShouldErrNoIDOrUsername",
			0,
			"",
			"either the username or id is required",
			"",
		},
		{
			"ShouldSucceedRevokeByUsername",
			0,
			"john",
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeByID",
			1,
			"",
			"",
			"SUCCESS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			seedBanUser(t, context.Background(), store, "john", "bad actor", true)

			buf := new(bytes.Buffer)

			err := runStorageBansRevokeUser(context.Background(), buf, store, tc.id, tc.target)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldSkipAlreadyRevoked", func(t *testing.T) {
		store := newTestSQLiteStore(t)

		seedBanUser(t, context.Background(), store, "john", "", true)

		buf := new(bytes.Buffer)

		err := runStorageBansRevokeUser(context.Background(), buf, store, 1, "")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "SUCCESS")

		buf.Reset()

		err = runStorageBansRevokeUser(context.Background(), buf, store, 1, "")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "SKIPPED")
		assert.Contains(t, buf.String(), "Ban has already been revoked")
	})
}

func TestRunStorageBansAdd(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		args     []string
		flags    map[string]string
		err      string
		expected string
	}{
		{
			"ShouldSucceedAddIPPermanent",
			"ip",
			[]string{"192.168.1.1"},
			map[string]string{"permanent": "true", "duration": "1h"},
			"",
			"Successfully banned IP '192.168.1.1' permanently.",
		},
		{
			"ShouldSucceedAddIPTemporary",
			"ip",
			[]string{"10.0.0.1"},
			map[string]string{"duration": "1h"},
			"",
			"Successfully banned IP '10.0.0.1' until",
		},
		{
			"ShouldSucceedAddIPWithReason",
			"ip",
			[]string{"10.0.0.2"},
			map[string]string{"duration": "2h", "reason": "brute force"},
			"",
			"Successfully banned IP '10.0.0.2' until",
		},
		{
			"ShouldSucceedAddUserPermanent",
			"user",
			[]string{"john"},
			map[string]string{"permanent": "true", "duration": "1h"},
			"",
			"Successfully banned user 'john' permanently.",
		},
		{
			"ShouldSucceedAddUserTemporary",
			"user",
			[]string{"john"},
			map[string]string{"duration": "1h"},
			"",
			"Successfully banned user 'john' until",
		},
		{
			"ShouldSucceedAddUserWithReason",
			"user",
			[]string{"john"},
			map[string]string{"duration": "30m", "reason": "too many attempts"},
			"",
			"Successfully banned user 'john' until",
		},
		{
			"ShouldErrInvalidIP",
			"ip",
			[]string{"not-an-ip"},
			map[string]string{"duration": "1h"},
			"invalid IP address: not-an-ip",
			"",
		},
		{
			"ShouldErrInvalidDuration",
			"ip",
			[]string{"192.168.1.1"},
			map[string]string{"duration": "invalid"},
			"failed to parse duration string:",
			"",
		},
		{
			"ShouldErrUnknownCommand",
			"bad",
			[]string{"target"},
			map[string]string{"duration": "1h"},
			"unknown command \"bad\"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Bool("permanent", false, "")
			flags.String("reason", "", "")
			flags.String("duration", "1h", "")

			for k, v := range tc.flags {
				require.NoError(t, flags.Set(k, v))
			}

			buf := new(bytes.Buffer)

			err := runStorageBansAdd(context.Background(), buf, flags, tc.args, store, tc.use)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestRunStorageBansRevokeFull(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		args     []string
		flagID   int
		err      string
		expected string
	}{
		{
			"ShouldSucceedRevokeIPByTarget",
			"ip",
			[]string{"192.168.1.1"},
			0,
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeUserByTarget",
			"user",
			[]string{"john"},
			0,
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeIPByID",
			"ip",
			nil,
			1,
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeUserByID",
			"user",
			nil,
			1,
			"",
			"SUCCESS",
		},
		{
			"ShouldErrUnknownCommand",
			"bad",
			nil,
			1,
			"unknown command \"bad\"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestSQLiteStore(t)

			seedBanIP(t, context.Background(), store, "192.168.1.1", "", true)
			seedBanUser(t, context.Background(), store, "john", "", true)

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Int("id", 0, "")

			if tc.flagID != 0 {
				require.NoError(t, flags.Set("id", fmt.Sprintf("%d", tc.flagID)))
			}

			buf := new(bytes.Buffer)

			err := runStorageBansRevoke(context.Background(), buf, flags, tc.args, store, tc.use)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestStorageBansListRunE(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		seedIP   bool
		seedUser bool
		expected string
	}{
		{
			"ShouldListIPBansEmpty",
			"ip",
			false,
			false,
			"No results.",
		},
		{
			"ShouldListIPBansWithData",
			"ip",
			true,
			false,
			"192.168.1.1",
		},
		{
			"ShouldListUserBansEmpty",
			"user",
			false,
			false,
			"No results.",
		},
		{
			"ShouldListUserBansWithData",
			"user",
			false,
			true,
			"john",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seedIP {
				seedBanIP(t, context.Background(), cmdCtx.providers.StorageProvider, "192.168.1.1", "malicious", true)
			}

			if tc.seedUser {
				seedBanUser(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "brute force", true)
			}

			cmd, buf := newTestCmdWithBuf()

			runE := cmdCtx.StorageBansListRunE(tc.use)

			err := runE(cmd, nil)

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}
}

func TestStorageBansAddRunE(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		args     []string
		flags    map[string]string
		err      string
		expected string
	}{
		{
			"ShouldSucceedAddIPPermanent",
			"ip",
			[]string{"192.168.1.1"},
			map[string]string{"permanent": "true"},
			"",
			"Successfully banned IP '192.168.1.1' permanently.",
		},
		{
			"ShouldSucceedAddUserPermanent",
			"user",
			[]string{"john"},
			map[string]string{"permanent": "true"},
			"",
			"Successfully banned user 'john' permanently.",
		},
		{
			"ShouldSucceedAddIPTemporary",
			"ip",
			[]string{"10.0.0.1"},
			nil,
			"",
			"Successfully banned IP '10.0.0.1' until",
		},
		{
			"ShouldSucceedAddUserTemporary",
			"user",
			[]string{"harry"},
			nil,
			"",
			"Successfully banned user 'harry' until",
		},
		{
			"ShouldErrInvalidIP",
			"ip",
			[]string{"not-an-ip"},
			nil,
			"invalid IP address: not-an-ip",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().Bool("permanent", false, "")
			cmd.Flags().String("reason", "", "")
			cmd.Flags().String("duration", "1h", "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			runE := cmdCtx.StorageBansAddRunE(tc.use)

			err := runE(cmd, tc.args)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageBansRevokeRunE(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		args     []string
		flagID   string
		seedIP   bool
		seedUser bool
		err      string
		expected string
	}{
		{
			"ShouldSucceedRevokeIPByTarget",
			"ip",
			[]string{"192.168.1.1"},
			"",
			true,
			false,
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeUserByTarget",
			"user",
			[]string{"john"},
			"",
			false,
			true,
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeIPByID",
			"ip",
			nil,
			"1",
			true,
			false,
			"",
			"SUCCESS",
		},
		{
			"ShouldSucceedRevokeUserByID",
			"user",
			nil,
			"1",
			false,
			true,
			"",
			"SUCCESS",
		},
		{
			"ShouldErrIPNoIDOrTarget",
			"ip",
			nil,
			"",
			false,
			false,
			"either the ip or id is required",
			"",
		},
		{
			"ShouldErrUserNoIDOrTarget",
			"user",
			nil,
			"",
			false,
			false,
			"either the username or id is required",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seedIP {
				seedBanIP(t, context.Background(), cmdCtx.providers.StorageProvider, "192.168.1.1", "", true)
			}

			if tc.seedUser {
				seedBanUser(t, context.Background(), cmdCtx.providers.StorageProvider, "john", "", true)
			}

			cmd, buf := newTestCmdWithBuf()
			cmd.Flags().Int("id", 0, "")

			if tc.flagID != "" {
				require.NoError(t, cmd.Flags().Set("id", tc.flagID))
			}

			runE := cmdCtx.StorageBansRevokeRunE(tc.use)

			err := runE(cmd, tc.args)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tc.expected)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageBansFullLifecycle(t *testing.T) {
	testCases := []struct {
		name string
		use  string
		args []string
	}{
		{
			"ShouldHandleIPLifecycle",
			"ip",
			[]string{"192.168.1.1"},
		},
		{
			"ShouldHandleUserLifecycle",
			"user",
			[]string{"john"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)
			store := cmdCtx.providers.StorageProvider

			buf := new(bytes.Buffer)

			assert.NoError(t, runStorageBansList(context.Background(), buf, store, tc.use))
			assert.Contains(t, buf.String(), "No results.")

			buf.Reset()

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Bool("permanent", true, "")
			flags.String("reason", "test lifecycle", "")
			flags.String("duration", "1h", "")

			require.NoError(t, flags.Set("permanent", "true"))

			assert.NoError(t, runStorageBansAdd(context.Background(), buf, flags, tc.args, store, tc.use))
			assert.Contains(t, buf.String(), "Successfully banned")

			buf.Reset()

			assert.NoError(t, runStorageBansList(context.Background(), buf, store, tc.use))
			assert.Contains(t, buf.String(), tc.args[0])
			assert.Contains(t, buf.String(), "test lifecycle")

			buf.Reset()

			revokeFlags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			revokeFlags.Int("id", 0, "")

			require.NoError(t, revokeFlags.Set("id", "1"))

			assert.NoError(t, runStorageBansRevoke(context.Background(), buf, revokeFlags, nil, store, tc.use))
			assert.Contains(t, buf.String(), "SUCCESS")

			buf.Reset()

			assert.NoError(t, runStorageBansRevoke(context.Background(), buf, revokeFlags, nil, store, tc.use))
			assert.Contains(t, buf.String(), "SKIPPED")
		})
	}
}

func newTestSQLiteStore(t *testing.T) storage.Provider {
	t.Helper()

	dir := t.TempDir()

	config := &schema.Configuration{
		Storage: schema.Storage{
			Local: &schema.StorageLocal{
				Path: filepath.Join(dir, "db.sqlite3"),
			},
		},
	}

	store := storage.NewProvider(config, nil)

	require.NoError(t, store.StartupCheck())

	return store
}

func newTestSQLiteStoreWithEncryptionKey(t *testing.T) storage.Provider {
	t.Helper()

	dir := t.TempDir()

	config := &schema.Configuration{
		Storage: schema.Storage{
			//gitleaks:allow // This is not an actual secret.
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(dir, "db.sqlite3"),
			},
		},
	}

	store := storage.NewProvider(config, nil)

	require.NoError(t, store.StartupCheck())

	return store
}

func newTestCmdCtx(t *testing.T) *CmdCtx {
	t.Helper()

	dir := t.TempDir()

	ctx := NewCmdCtx()

	ctx.config.Storage = schema.Storage{
		//gitleaks:allow // This is not an actual secret.
		EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
		Local: &schema.StorageLocal{
			Path: filepath.Join(dir, "db.sqlite3"),
		},
	}

	ctx.config.TOTP = schema.DefaultTOTPConfiguration

	ctx.providers.StorageProvider = storage.NewProvider(ctx.config, nil)

	require.NoError(t, ctx.providers.StorageProvider.StartupCheck())

	return ctx
}

func newTestCmdWithBuf() (*cobra.Command, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	cmd := &cobra.Command{
		Use: "test",
	}

	cmd.SetOut(buf)

	return cmd, buf
}

func seedTOTPConfig(t *testing.T, ctx context.Context, store storage.Provider, username string) {
	t.Helper()

	require.NoError(t, store.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
		CreatedAt: time.Now(),
		Username:  username,
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    []byte("JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PX"),
	}))
}

func seedWebAuthnCredential(t *testing.T, ctx context.Context, store storage.Provider, username, description string, kid []byte) {
	t.Helper()

	require.NoError(t, store.SaveWebAuthnCredential(ctx, model.WebAuthnCredential{
		CreatedAt:       time.Now(),
		RPID:            "example.com",
		Username:        username,
		Description:     description,
		KID:             model.NewBase64(kid),
		AttestationType: "none",
		Attachment:      "cross-platform",
		Transport:       "",
		PublicKey:       []byte("fake-public-key"),
	}))
}

func seedUserOpaqueIdentifier(t *testing.T, ctx context.Context, store storage.Provider, username, service, sector string) {
	t.Helper()

	id, err := uuid.NewRandom()
	require.NoError(t, err)

	require.NoError(t, store.SaveUserOpaqueIdentifier(ctx, model.UserOpaqueIdentifier{
		Service:    service,
		SectorID:   sector,
		Username:   username,
		Identifier: id,
	}))
}

//nolint:unparam
func seedBanIP(t *testing.T, ctx context.Context, store storage.Provider, ip string, reason string, permanent bool) {
	t.Helper()

	require.NoError(t, runStorageBansAddIP(ctx, &bytes.Buffer{}, store, ip, reason, time.Hour, permanent))
}

//nolint:unparam
func seedBanUser(t *testing.T, ctx context.Context, store storage.Provider, username string, reason string, permanent bool) {
	t.Helper()

	require.NoError(t, runStorageBansAddUser(ctx, &bytes.Buffer{}, store, username, reason, time.Hour, permanent))
}

func newFlagSetWithInt(name string, value int) *pflag.FlagSet {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int(name, value, "")

	return flags
}
