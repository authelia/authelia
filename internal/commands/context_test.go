package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestCmdCtx_LoadTrustedCertificates(t *testing.T) {
	dir := t.TempDir()

	ctx := NewCmdCtx()

	ctx.config.CertificatesDirectory = dir

	require.NotNil(t, ctx)

	warns, errs := ctx.LoadTrustedCertificates()

	assert.Empty(t, warns)
	assert.Empty(t, errs)

	ctx.factoryX509SystemCertPool = &TestX509SystemCertPoolFactory{nil, errors.New("error")}

	warns, errs = ctx.LoadTrustedCertificates()

	assert.Empty(t, errs)
	require.Len(t, warns, 1)
	assert.EqualError(t, warns[0], "could not load system certificate pool which may result in untrusted certificate issues: error")
}

func TestNewCmdCtx(t *testing.T) {
	ctx := NewCmdCtx()

	require.NotNil(t, ctx)

	assert.Equal(t, ctx.GetProviders(), ctx.providers)
	assert.Equal(t, ctx.GetClock(), ctx.providers.Clock)
	assert.Equal(t, ctx.GetRandom(), ctx.providers.Random)
	assert.Equal(t, ctx.GetConfiguration(), ctx.config)
	assert.Equal(t, ctx.GetLogger(), ctx.log)
}

func TestNewCmdCtxConfig(t *testing.T) {
	config := NewCmdCtxConfig()

	assert.NotNil(t, config)
}

func TestCmdCtx_LoadTrustedCertificatesRunE(t *testing.T) {
	dir := t.TempDir()

	ctx := NewCmdCtx()

	ctx.config.CertificatesDirectory = dir

	err := ctx.LoadTrustedCertificatesRunE(nil, nil)

	assert.NoError(t, err)
}

func TestCmdCtx_LoadTrustedCertificatesRunEWithErrors(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.pem"), []byte("invalid"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2.pem"), []byte("invalid"), 0600))

	ctx := NewCmdCtx()

	ctx.config.CertificatesDirectory = dir

	err := ctx.LoadTrustedCertificatesRunE(nil, nil)

	assert.EqualError(t, err, "failed to load trusted certificates: could not import certificate file.pem, could not import certificate file2.pem")
}

func TestChainRunE(t *testing.T) {
	oneRun, twoRun, threeRun := false, false, false

	one := func(cmd *cobra.Command, args []string) error {
		oneRun = true
		return nil
	}

	two := func(cmd *cobra.Command, args []string) error {
		twoRun = true
		return fmt.Errorf("error two")
	}

	three := func(cmd *cobra.Command, args []string) error {
		threeRun = true
		return nil
	}

	ctx := NewCmdCtx()

	cmd := ctx.ChainRunE(one, two, three)

	c := &cobra.Command{}
	args := []string{}

	err := cmd(c, args)

	assert.EqualError(t, err, "error two")
	assert.True(t, oneRun)
	assert.True(t, twoRun)
	assert.False(t, threeRun)

	cmd = ctx.ChainRunE(one, three)

	assert.NoError(t, cmd(c, args))

	assert.True(t, oneRun)
	assert.True(t, threeRun)
}

func TestLoadProviders(t *testing.T) {
	ctx := NewCmdCtx()

	warns, errs := ctx.LoadProviders()
	assert.NotNil(t, ctx.providers)
	assert.NotNil(t, ctx.providers.Clock)
	assert.NotNil(t, ctx.providers.Random)

	assert.Empty(t, warns)
	assert.Empty(t, errs)

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.pem"), []byte("invalid"), 0000))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2.pem"), []byte("invalid"), 0000))

	ctx.config.CertificatesDirectory = dir

	warns, errs = ctx.LoadProviders()
	assert.Empty(t, warns)
	require.Len(t, errs, 2)

	assert.EqualError(t, errs[0], fmt.Sprintf("error occurred trying to read certificate: open %s: permission denied", filepath.Join(dir, "file.pem")))
	assert.EqualError(t, errs[1], fmt.Sprintf("error occurred trying to read certificate: open %s: permission denied", filepath.Join(dir, "file2.pem")))
}

func TestCmdCtx_CheckSchema(t *testing.T) {
	dir := t.TempDir()

	ctx := NewCmdCtx()

	assert.NoError(t, ctx.ConfigStorageCommandLineConfigRunE(NewRootCmd(), nil))

	ctx.config.Storage = schema.Storage{
		//gitleaks:allo // This is not an actual secret.
		EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
		Local: &schema.StorageLocal{
			Path: filepath.Join(dir, "db.sqlite3"),
		},
	}

	assert.EqualError(t, ctx.CheckSchemaVersion(), "storage not loaded")
	assert.EqualError(t, ctx.CheckSchema(), "storage not loaded")

	warns, errs := ctx.LoadProviders()

	require.Empty(t, warns)
	require.Empty(t, errs)

	assert.EqualError(t, ctx.CheckSchemaVersion(), "storage schema outdated: version 0 is outdated please migrate to version 22 in order to use this command or use an older binary")

	assert.NoError(t, ctx.providers.StorageProvider.StartupCheck())

	assert.NoError(t, ctx.CheckSchemaVersion())
	assert.NoError(t, ctx.CheckSchema())

	ctx.config.Log.Level = ""
	ctx.config.Log.Format = "not-text"
	assert.NoError(t, ctx.LogConfigure(&cobra.Command{}, nil))
}

func TestHelperConfigSetDefaultsRunE(t *testing.T) {
	ctx := NewCmdCtx()

	cmd := ctx.HelperConfigSetDefaultsRunE(newCryptoHashDefaults())

	assert.NoError(t, cmd(NewRootCmd(), nil))
}

func TestHelperConfigValidateKeysRunE(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.yml"), []byte("certificates_directory: ''\n"), 0600))

	ctx := NewCmdCtx()

	cmd := &cobra.Command{}

	cmd.Flags().StringSliceP(cmdFlagNameConfig, "c", []string{filepath.Join(dir, "file.yml")}, "configuration files or directories to load, for more information run 'authelia -h authelia config'")
	cmd.Flags().StringSlice(cmdFlagNameConfigExpFilters, nil, "list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'")

	assert.NoError(t, ctx.HelperConfigLoadRunE(cmd, nil))
	assert.NoError(t, ctx.HelperConfigValidateKeysRunE(cmd, nil))
	assert.NoError(t, ctx.HelperConfigValidateRunE(cmd, nil))
	assert.NoError(t, ctx.LogConfigure(cmd, nil))

	require.NoError(t, cmd.Flags().Set(cmdFlagNameConfigExpFilters, "expand-env"))
	assert.NoError(t, ctx.HelperConfigLoadRunE(cmd, nil))
}

func TestLogProcessCurrentUserRunE(t *testing.T) {
	ctx := NewCmdCtx()

	log, hook := test.NewNullLogger()
	log.Level = logrus.TraceLevel

	ctx.log = log.WithField("test", "test")

	err := ctx.LogProcessCurrentUserRunE(nil, nil)

	assert.NoError(t, err)

	entry := hook.LastEntry()

	require.NotNil(t, entry)
	assert.Equal(t, "test", entry.Data["test"])
	assert.Contains(t, entry.Data, "username")
	assert.Contains(t, entry.Data, "uid")
	assert.Contains(t, entry.Data, "gid")
	assert.Equal(t, "Process user information", entry.Message)
}

func TestConfigValidateLogRunE(t *testing.T) {
	ctx := NewCmdCtx()

	ctx.cconfig = NewCmdCtxConfig()

	exitCode := 0

	log, hook := test.NewNullLogger()
	log.Level = logrus.TraceLevel
	log.ExitFunc = func(value int) {
		exitCode = value
	}

	ctx.log = log.WithField("test", "test")

	err := ctx.ConfigValidateLogRunE(nil, nil)

	assert.NoError(t, err)

	entry := hook.LastEntry()

	assert.Nil(t, entry)
	assert.Equal(t, 0, exitCode)

	ctx.cconfig.validator.PushWarning(errors.New("warning"))
	ctx.cconfig.validator.Push(errors.New("error"))

	err = ctx.ConfigValidateLogRunE(nil, nil)

	assert.NoError(t, err)

	entry = hook.LastEntry()

	assert.NotNil(t, entry)
	assert.Equal(t, 1, exitCode)
}

func TestConfigValidateSectionPasswordRunE(t *testing.T) {
	ctx := NewCmdCtx()

	ctx.cconfig = NewCmdCtxConfig()

	log, hook := test.NewNullLogger()
	log.Level = logrus.TraceLevel

	ctx.log = log.WithField("test", "test")

	assert.EqualError(t, ctx.ConfigValidateSectionPasswordRunE(nil, nil), "password configuration was not initialized")
	assert.Nil(t, hook.LastEntry())

	ctx.config.AuthenticationBackend.File = &schema.AuthenticationBackendFile{}

	assert.NoError(t, ctx.ConfigValidateSectionPasswordRunE(nil, nil))

	ctx.config.AuthenticationBackend.File.Password.Algorithm = "not-a-real-algorithm"
	ctx.config.AuthenticationBackend.File.Password.Bcrypt.Variant = "not-a-real-variant"

	assert.EqualError(t, ctx.ConfigValidateSectionPasswordRunE(nil, nil), "errors occurred validating the password configuration: authentication_backend: file: password: option 'algorithm' must be one of 'sha2crypt', 'pbkdf2', 'scrypt', 'bcrypt', or 'argon2' but it's configured as 'not-a-real-algorithm', authentication_backend: file: password: bcrypt: option 'variant' must be one of 'standard' or 'sha256' but it's configured as 'not-a-real-variant'")
}

func TestConfigEnsureExistsRunE(t *testing.T) {
	dir := t.TempDir()

	_, err := os.Stat(filepath.Join(dir, "file.yml"))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	ctx := NewCmdCtx()

	ctx.cconfig = NewCmdCtxConfig()

	exitCode := 0

	log, _ := test.NewNullLogger()
	log.Level = logrus.TraceLevel
	log.ExitFunc = func(value int) {
		exitCode = value
	}

	ctx.log = log.WithField("test", "test")

	cmd := &cobra.Command{}
	assert.EqualError(t, ctx.ConfigEnsureExistsRunE(cmd, nil), "flag accessed but not defined: config")

	cmd.Flags().StringSliceP(cmdFlagNameConfig, "c", []string{filepath.Join(dir, "dir", "file.yml")}, "configuration files or directories to load, for more information run 'authelia -h authelia config'")
	cmd.Flags().StringSlice(cmdFlagNameConfigExpFilters, nil, "list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'")

	assert.Equal(t, 0, exitCode)

	assert.NoError(t, ctx.ConfigEnsureExistsRunE(cmd, nil))

	assert.Equal(t, 1, exitCode)

	cmd = &cobra.Command{}

	cmd.Flags().StringSliceP(cmdFlagNameConfig, "c", []string{filepath.Join(dir, "file.yml")}, "configuration files or directories to load, for more information run 'authelia -h authelia config'")
	cmd.Flags().StringSlice(cmdFlagNameConfigExpFilters, nil, "list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'")

	assert.True(t, errors.Is(ctx.ConfigEnsureExistsRunE(cmd, nil), ErrConfigCreated))
	assert.NoError(t, ctx.ConfigEnsureExistsRunE(cmd, nil))

	t.Setenv(cmdFlagEnvNameConfig, filepath.Join(dir, "file.yml"))
	assert.NoError(t, ctx.ConfigEnsureExistsRunE(cmd, nil))

	assert.NoError(t, cmd.Flags().Set(cmdFlagNameConfig, filepath.Join(dir, "file.yml")))
	assert.NoError(t, ctx.ConfigEnsureExistsRunE(cmd, nil))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "dir"), 0000))
}
