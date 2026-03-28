package commands

import (
	"fmt"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestGetStorageProvider(t *testing.T) {
	assert.Nil(t, getStorageProvider(NewCmdCtx()))
}

func TestContainsIdentifier(t *testing.T) {
	identifiers := []model.UserOpaqueIdentifier{
		{Service: "openid", SectorID: "example.com", Username: "john"},
		{Service: "openid", SectorID: "example.org", Username: "harry"},
	}

	testCases := []struct {
		name       string
		identifier model.UserOpaqueIdentifier
		expected   bool
	}{
		{
			"ShouldMatchExactIdentifier",
			model.UserOpaqueIdentifier{Service: "openid", SectorID: "example.com", Username: "john"},
			true,
		},
		{
			"ShouldMatchSecondIdentifier",
			model.UserOpaqueIdentifier{Service: "openid", SectorID: "example.org", Username: "harry"},
			true,
		},
		{
			"ShouldNotMatchDifferentUsername",
			model.UserOpaqueIdentifier{Service: "openid", SectorID: "example.com", Username: "bob"},
			false,
		},
		{
			"ShouldNotMatchDifferentService",
			model.UserOpaqueIdentifier{Service: "other", SectorID: "example.com", Username: "john"},
			false,
		},
		{
			"ShouldNotMatchDifferentSectorID",
			model.UserOpaqueIdentifier{Service: "openid", SectorID: "other.com", Username: "john"},
			false,
		},
		{
			"ShouldNotMatchEmptyIdentifier",
			model.UserOpaqueIdentifier{},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, containsIdentifier(tc.identifier, identifiers))
		})
	}
}

func TestContainsIdentifierEmptySlice(t *testing.T) {
	assert.False(t, containsIdentifier(model.UserOpaqueIdentifier{Service: "openid", SectorID: "example.com", Username: "john"}, nil))
}

func TestStorageWrapCheckSchemaErr(t *testing.T) {
	testCases := []struct {
		name     string
		have     error
		expected string
	}{
		{
			"ShouldWrapIncompatibleError",
			errStorageSchemaIncompatible,
			"command requires the use of a compatible schema version: storage schema incompatible",
		},
		{
			"ShouldWrapOutdatedError",
			errStorageSchemaOutdated,
			"command requires the use of a up to date schema version: storage schema outdated",
		},
		{
			"ShouldWrapWrappedIncompatibleError",
			fmt.Errorf("some context: %w", errStorageSchemaIncompatible),
			"command requires the use of a compatible schema version: some context: storage schema incompatible",
		},
		{
			"ShouldWrapWrappedOutdatedError",
			fmt.Errorf("some context: %w", errStorageSchemaOutdated),
			"command requires the use of a up to date schema version: some context: storage schema outdated",
		},
		{
			"ShouldReturnOtherErrorUnchanged",
			fmt.Errorf("some other error"),
			"some other error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := storageWrapCheckSchemaErr(tc.have)
			assert.EqualError(t, err, tc.expected)
		})
	}
}

func TestStorageTOTPGenerateRunEOptsFromFlags(t *testing.T) {
	testCases := []struct {
		name             string
		setup            func(flags *pflag.FlagSet)
		args             []string
		expectedForce    bool
		expectedFilename string
		expectedSecret   string
		err              string
	}{
		{
			"ShouldReturnDefaults",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.String("path", "", "")
				flags.String("secret", "", "")
			},
			nil,
			false,
			"",
			"",
			"",
		},
		{
			"ShouldReturnForce",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.String("path", "", "")
				flags.String("secret", "", "")
			},
			[]string{"--force"},
			true,
			"",
			"",
			"",
		},
		{
			"ShouldReturnPath",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.String("path", "", "")
				flags.String("secret", "", "")
			},
			[]string{"--path", "/tmp/test.png"},
			false,
			"/tmp/test.png",
			"",
			"",
		},
		{
			"ShouldReturnValidSecret",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.String("path", "", "")
				flags.String("secret", "", "")
			},
			[]string{"--secret", "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXJBSWY3DPEHPK3PXP"},
			false,
			"",
			"JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXJBSWY3DPEHPK3PXP",
			"",
		},
		{
			"ShouldErrSecretTooShort",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.String("path", "", "")
				flags.String("secret", "", "")
			},
			[]string{"--secret", "JBSWY3DP"},
			false,
			"",
			"",
			"decoded length of the base32 secret must have a length of more than 20",
		},
		{
			"ShouldReturnAllFlags",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.String("path", "", "")
				flags.String("secret", "", "")
			},
			[]string{"--force", "--path", "/tmp/test.png", "--secret", "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXJBSWY3DPEHPK3PXP"},
			true,
			"/tmp/test.png",
			"JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXJBSWY3DPEHPK3PXP",
			"",
		},
		{
			"ShouldErrForceWrongType",
			func(flags *pflag.FlagSet) {
				flags.String("force", "", "")
				flags.String("path", "", "")
				flags.String("secret", "", "")
			},
			nil,
			false,
			"",
			"",
			"trying to get bool value of flag of type string",
		},
		{
			"ShouldErrPathWrongType",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.Bool("path", false, "")
				flags.String("secret", "", "")
			},
			nil,
			false,
			"",
			"",
			"trying to get string value of flag of type bool",
		},
		{
			"ShouldErrSecretWrongType",
			func(flags *pflag.FlagSet) {
				flags.Bool("force", false, "")
				flags.String("path", "", "")
				flags.Bool("secret", false, "")
			},
			nil,
			false,
			"",
			"",
			"trying to get string value of flag of type bool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

			tc.setup(flags)

			if tc.args != nil {
				assert.NoError(t, flags.Parse(tc.args))
			}

			force, filename, secret, err := storageTOTPGenerateRunEOptsFromFlags(flags)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedForce, force)
				assert.Equal(t, tc.expectedFilename, filename)
				assert.Equal(t, tc.expectedSecret, secret)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestStorageWebAuthnDeleteRunEOptsFromFlags(t *testing.T) {
	testCases := []struct {
		name                string
		args                []string
		flagArgs            []string
		expectedAll         bool
		expectedByKID       bool
		expectedDescription string
		expectedKID         string
		expectedUser        string
		err                 string
	}{
		{
			"ShouldSucceedWithAllFlag",
			[]string{"john"},
			[]string{"--all"},
			true,
			false,
			"",
			"",
			"john",
			"",
		},
		{
			"ShouldSucceedWithDescriptionFlag",
			[]string{"john"},
			[]string{"--description", "my-key"},
			false,
			false,
			"my-key",
			"",
			"john",
			"",
		},
		{
			"ShouldSucceedWithKIDFlag",
			nil,
			[]string{"--kid", "abc123"},
			false,
			true,
			"",
			"abc123",
			"",
			"",
		},
		{
			"ShouldSucceedWithKIDFlagAndUser",
			[]string{"john"},
			[]string{"--kid", "abc123"},
			false,
			true,
			"",
			"abc123",
			"john",
			"",
		},
		{
			"ShouldErrNoFlags",
			[]string{"john"},
			nil,
			false,
			false,
			"",
			"",
			"",
			"must supply one of the flags --all, --description, or --kid",
		},
		{
			"ShouldErrMultipleFlags",
			[]string{"john"},
			[]string{"--all", "--description", "my-key"},
			false,
			false,
			"",
			"",
			"",
			"must only supply one of the flags --all, --description, and --kid but 2 were specified",
		},
		{
			"ShouldErrAllThreeFlags",
			[]string{"john"},
			[]string{"--all", "--description", "my-key", "--kid", "abc123"},
			false,
			false,
			"",
			"",
			"",
			"must only supply one of the flags --all, --description, and --kid but 3 were specified",
		},
		{
			"ShouldErrAllFlagWithoutUser",
			nil,
			[]string{"--all"},
			false,
			false,
			"",
			"",
			"",
			"must supply the username or the --kid flag",
		},
		{
			"ShouldErrDescriptionFlagWithoutUser",
			nil,
			[]string{"--description", "my-key"},
			false,
			false,
			"",
			"",
			"",
			"must supply the username or the --kid flag",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Bool(cmdFlagNameAll, false, "")
			flags.String(cmdFlagNameDescription, "", "")
			flags.String(cmdFlagNameKeyID, "", "")

			if tc.flagArgs != nil {
				assert.NoError(t, flags.Parse(tc.flagArgs))
			}

			all, byKID, description, kid, user, err := storageWebAuthnDeleteRunEOptsFromFlags(flags, tc.args)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAll, all)
				assert.Equal(t, tc.expectedByKID, byKID)
				assert.Equal(t, tc.expectedDescription, description)
				assert.Equal(t, tc.expectedKID, kid)
				assert.Equal(t, tc.expectedUser, user)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
