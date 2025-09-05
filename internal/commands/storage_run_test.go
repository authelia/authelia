package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
		//gitleaks:allo // This is not an actual secret.
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
