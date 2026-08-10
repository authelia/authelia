package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
)

func newTestCmd(t *testing.T) (cmd *cobra.Command, stdout, stderr *bytes.Buffer) {
	t.Helper()

	cmd = &cobra.Command{Use: "test"}
	stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	return cmd, stdout, stderr
}

func newTestUsersCmdCtx(t *testing.T) (ctx *CmdCtx, ctrl *gomock.Controller, userProvider *mocks.MockUserProvider, storageProvider *mocks.MockStorage) {
	t.Helper()

	ctrl = gomock.NewController(t)
	userProvider = mocks.NewMockUserProvider(ctrl)
	storageProvider = mocks.NewMockStorage(ctrl)

	ctx = NewCmdCtx()
	ctx.providers.UserProvider = userProvider
	ctx.providers.StorageProvider = storageProvider
	ctx.cconfig = NewCmdCtxConfig()

	return ctx, ctrl, userProvider, storageProvider
}

func TestConfigValidateAdministrationRunE(t *testing.T) {
	t.Run("ShouldReturnExistingValidatorErrors", func(t *testing.T) {
		ctx := NewCmdCtx()
		ctx.cconfig = NewCmdCtxConfig()
		ctx.cconfig.validator.Push(errors.New("existing error"))

		err := ctx.ConfigValidateAdministrationRunE(nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "existing error")
	})

	t.Run("ShouldDefaultAdminGroupAndPass", func(t *testing.T) {
		ctx := NewCmdCtx()
		ctx.cconfig = NewCmdCtxConfig()

		err := ctx.ConfigValidateAdministrationRunE(nil, nil)
		require.NoError(t, err)
		assert.Equal(t, schema.DefaultAdministrationConfiguration.AdminGroup, ctx.config.Administration.AdminGroup)
	})

	t.Run("ShouldAggregateMultipleValidationErrors", func(t *testing.T) {
		ctx := NewCmdCtx()
		ctx.cconfig = NewCmdCtxConfig()
		ctx.cconfig.validator.Push(errors.New("first"))
		ctx.cconfig.validator.Push(errors.New("second"))

		err := ctx.ConfigValidateAdministrationRunE(nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "first")
		assert.Contains(t, err.Error(), "second")
	})
}

func TestConfigValidateUserBackendRunE(t *testing.T) {
	t.Run("ShouldReturnExistingValidatorErrors", func(t *testing.T) {
		ctx := NewCmdCtx()
		ctx.cconfig = NewCmdCtxConfig()
		ctx.cconfig.validator.Push(errors.New("existing error"))

		err := ctx.ConfigValidateUserBackendRunE(nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "existing error")
	})

	t.Run("ShouldErrorWhenNoBackendConfigured", func(t *testing.T) {
		ctx := NewCmdCtx()
		ctx.cconfig = NewCmdCtxConfig()

		err := ctx.ConfigValidateUserBackendRunE(nil, nil)
		require.Error(t, err)
	})
}

func TestLoadProvidersUserBackendRunE(t *testing.T) {
	t.Run("ShouldErrorWhenNoBackendConfigured", func(t *testing.T) {
		ctx := NewCmdCtx()

		err := ctx.LoadProvidersUserBackendRunE(nil, nil)
		require.EqualError(t, err, "user management requires a configured authentication backend (file or ldap)")
	})
}

func newTestSupportedAttributes() map[string]authentication.UserManagementAttributeMetadata {
	return map[string]authentication.UserManagementAttributeMetadata{
		authentication.AttributeUsername:    {Type: authentication.Text, Label: "Username"},
		authentication.AttributePassword:    {Type: authentication.Password, Label: "Password"},
		authentication.AttributeDisplayName: {Type: authentication.Text, Label: "Display Name"},
		authentication.AttributeMail:        {Type: authentication.Text, Label: "Email"},
		authentication.AttributeGroups:      {Type: authentication.Groups, Multiple: true, Label: "Groups"},
		"extra.employee_id":                 {Type: authentication.Text, Label: "Employee ID"},
	}
}

func TestUsersSchemaPrintRunE(t *testing.T) {
	ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
	defer ctrl.Finish()

	userProvider.EXPECT().GetSupportedAttributes().Return(newTestSupportedAttributes())
	userProvider.EXPECT().GetRequiredAttributes().Return([]string{authentication.AttributeUsername, authentication.AttributePassword})

	cmd, stdout, _ := newTestCmd(t)

	err := ctx.UsersSchemaPrintRunE()(cmd, nil)
	require.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "Attribute")
	assert.Contains(t, out, "Label")
	assert.Contains(t, out, "Type")
	assert.Contains(t, out, "Required")
	assert.Contains(t, out, "Multi-Valued")
	assert.Contains(t, out, authentication.AttributeUsername)
	assert.Contains(t, out, "extra.employee_id")

	// Required attributes should be listed before non-required ones.
	assert.Less(t, strings.Index(out, authentication.AttributeUsername), strings.Index(out, authentication.AttributeMail))
}

func TestUsersGetRunE(t *testing.T) {
	t.Run("ShouldReturnUserWithDefaultFields", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		user := &authentication.UserDetailsExtended{
			UserDetails: &authentication.UserDetails{Username: "john", DisplayName: "John Doe", Emails: []string{"john@example.com"}},
			Password:    "secret",
		}

		userProvider.EXPECT().GetRequiredAttributes().Return([]string{authentication.AttributeUsername, authentication.AttributePassword, authentication.AttributeDisplayName})
		userProvider.EXPECT().GetUser("john").Return(user, nil)
		storageProvider.EXPECT().Close().Return(nil)

		cmd, stdout, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		cmd.Flags().StringSlice(cmdFlagNameFields, nil, "")

		err := ctx.UsersGetRunE()(cmd, []string{"john"})
		require.NoError(t, err)

		out := stdout.String()
		assert.Contains(t, out, "john")
		assert.Contains(t, out, "John Doe")
		assert.NotContains(t, out, "secret")
	})

	t.Run("ShouldRejectUnsupportedField", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().GetSupportedAttributes().Return(newTestSupportedAttributes())
		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		cmd.Flags().StringSlice(cmdFlagNameFields, nil, "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFields, "not_a_real_field"))

		err := ctx.UsersGetRunE()(cmd, []string{"john"})
		require.EqualError(t, err, "field 'not_a_real_field' is not supported by the configured authentication backend")
	})

	t.Run("ShouldRejectJSONFormatWithFields", func(t *testing.T) {
		ctx, ctrl, _, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		cmd.Flags().StringSlice(cmdFlagNameFields, nil, "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFormat, cmdFlagValueFormatJSON))
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFields, authentication.AttributeUsername))

		err := ctx.UsersGetRunE()(cmd, []string{"john"})
		require.EqualError(t, err, "flag '--fields' cannot be used with '--format json'")
	})

	t.Run("ShouldWrapProviderError", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().GetRequiredAttributes().Return([]string{authentication.AttributeUsername})
		userProvider.EXPECT().GetUser("ghost").Return(nil, errors.New("not found"))
		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		cmd.Flags().StringSlice(cmdFlagNameFields, nil, "")

		err := ctx.UsersGetRunE()(cmd, []string{"ghost"})
		require.EqualError(t, err, "error occurred retrieving user 'ghost': not found")
	})

	t.Run("ShouldPanicWhenStorageCloseFails", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().GetRequiredAttributes().Return([]string{authentication.AttributeUsername})
		userProvider.EXPECT().GetUser("john").Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}}, nil)
		storageProvider.EXPECT().Close().Return(errors.New("close failed"))

		cmd, _, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		cmd.Flags().StringSlice(cmdFlagNameFields, nil, "")

		assert.Panics(t, func() {
			_ = ctx.UsersGetRunE()(cmd, []string{"john"})
		})
	})
}

func TestUsersListRunE(t *testing.T) {
	t.Run("ShouldListUsersWithDefaultFields", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		users := []authentication.UserDetailsExtended{
			{UserDetails: &authentication.UserDetails{Username: "john", DisplayName: "John Doe"}, Password: "secret"},
			{UserDetails: &authentication.UserDetails{Username: "harry", DisplayName: "Harry Potter"}, Password: "secret"},
		}

		userProvider.EXPECT().GetRequiredAttributes().Return([]string{authentication.AttributeUsername, authentication.AttributePassword, authentication.AttributeDisplayName})
		userProvider.EXPECT().ListUsers().Return(users, nil)
		storageProvider.EXPECT().Close().Return(nil)

		cmd, stdout, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		cmd.Flags().StringSlice(cmdFlagNameFields, nil, "")

		err := ctx.UsersListRunE()(cmd, nil)
		require.NoError(t, err)

		out := stdout.String()
		assert.Contains(t, out, "john")
		assert.Contains(t, out, "harry")
		assert.NotContains(t, out, "secret")
	})

	t.Run("ShouldWrapProviderError", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().GetRequiredAttributes().Return([]string{authentication.AttributeUsername})
		userProvider.EXPECT().ListUsers().Return(nil, errors.New("boom"))
		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		cmd.Flags().StringSlice(cmdFlagNameFields, nil, "")

		err := ctx.UsersListRunE()(cmd, nil)
		require.EqualError(t, err, "error occurred retrieving user list: boom")
	})
}

// newTestFormatUsers builds a new user object for format testing.
func newTestFormatUsers(t *testing.T) []authentication.UserDetailsExtended {
	t.Helper()

	profile, err := url.Parse("https://example.com/john")
	require.NoError(t, err)

	return []authentication.UserDetailsExtended{
		{
			UserDetails: &authentication.UserDetails{Username: "john", DisplayName: "John Doe", Emails: []string{"john@example.com"}, Groups: []string{"admins"}},
			GivenName:   "John",
			FamilyName:  "Doe",
			Profile:     profile,
			Extra:       map[string]any{"employee_id": "123"},
		},
	}
}

func TestFormatUserOutput(t *testing.T) {
	t.Run("ShouldEncodeJSONAndStripEmptyFields", func(t *testing.T) {
		var buf bytes.Buffer

		err := FormatUserOutput(&buf, newTestFormatUsers(t), nil, cmdFlagValueFormatJSON)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"username": "john"`)
		assert.NotContains(t, buf.String(), `"middle_name"`)
	})

	t.Run("ShouldRenderTableWithKnownAndExtraFields", func(t *testing.T) {
		users := newTestFormatUsers(t)

		var buf bytes.Buffer

		fields := []string{
			authentication.AttributeUsername,
			authentication.AttributeDisplayName,
			authentication.AttributeMail,
			authentication.AttributeGroups,
			authentication.AttributeGivenName,
			authentication.AttributeFamilyName,
			authentication.AttributeProfile,
			"extra.employee_id",
			"extra.unknown_field",
		}

		err := FormatUserOutput(&buf, users, fields, cmdFlagValueFormatTable)
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "john")
		assert.Contains(t, out, "John Doe")
		assert.Contains(t, out, "john@example.com")
		assert.Contains(t, out, "admins")
		assert.Contains(t, out, "123")
		assert.Contains(t, out, "https://example.com/john")
	})

	t.Run("ShouldFallBackToGeneratedLabelForUnknownField", func(t *testing.T) {
		var buf bytes.Buffer

		err := FormatUserOutput(&buf, newTestFormatUsers(t), []string{"extra.employee_id"}, cmdFlagValueFormatTable)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Employee Id")
	})
}

func TestAttributeFieldLabel(t *testing.T) {
	testCases := []struct {
		name     string
		field    string
		expected string
	}{
		{name: "ShouldTitleCaseSingleWord", field: "gender", expected: "Gender"},
		{name: "ShouldTitleCaseSnakeCase", field: "phone_number", expected: "Phone Number"},
		{name: "ShouldStripExtraPrefix", field: "extra.employee_id", expected: "Employee Id"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, attributeFieldLabel(tc.field))
		})
	}
}

func TestUsersAddRunEFromJSON(t *testing.T) {
	t.Run("ShouldCreateUserFromJSONStdin", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		payload, err := marshalTestUser(authentication.UserDetailsExtended{
			UserDetails: &authentication.UserDetails{Username: "john", DisplayName: "John Doe", Emails: []string{"john@example.com"}},
			Password:    "apple123",
		})
		require.NoError(t, err)

		userProvider.EXPECT().ValidateUserData(gomock.Any()).Return(nil)
		userProvider.EXPECT().AddUser(gomock.Any()).Return(nil)
		storageProvider.EXPECT().LoadUserMetadataByUsername(gomock.Any(), "john").Return(model.UserInfo{}, errors.New("not found"))
		storageProvider.EXPECT().CreateNewUserMetadata(gomock.Any(), "john").Return(nil)
		storageProvider.EXPECT().Close().Return(nil)

		cmd, stdout, _ := newTestCmd(t)
		cmd.SetIn(strings.NewReader(payload))
		cmd.Flags().StringP(cmdFlagNameFile, "f", "", "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "-"))

		err = ctx.UsersAddRunE()(cmd, nil)
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Successfully created user 'john'.")
	})

	t.Run("ShouldSkipMetadataCreationWhenAlreadyExists", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		payload, err := marshalTestUser(authentication.UserDetailsExtended{
			UserDetails: &authentication.UserDetails{Username: "john"},
			Password:    "apple123",
		})
		require.NoError(t, err)

		userProvider.EXPECT().ValidateUserData(gomock.Any()).Return(nil)
		userProvider.EXPECT().AddUser(gomock.Any()).Return(nil)
		storageProvider.EXPECT().LoadUserMetadataByUsername(gomock.Any(), "john").Return(model.UserInfo{}, nil)
		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.SetIn(strings.NewReader(payload))
		cmd.Flags().StringP(cmdFlagNameFile, "f", "", "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "-"))

		err = ctx.UsersAddRunE()(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("ShouldErrorOnMalformedJSON", func(t *testing.T) {
		ctx, ctrl, _, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.SetIn(strings.NewReader("{not-json"))
		cmd.Flags().StringP(cmdFlagNameFile, "f", "", "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "-"))

		err := ctx.UsersAddRunE()(cmd, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing user JSON")
	})

	t.Run("ShouldErrorOnMissingFile", func(t *testing.T) {
		ctx, ctrl, _, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFile, "f", "", "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "/does/not/exist.json"))

		err := ctx.UsersAddRunE()(cmd, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error reading file")
	})

	t.Run("ShouldTranslateValidationErrors", func(t *testing.T) {
		testCases := []struct {
			name     string
			returned error
			expected string
		}{
			{name: "ShouldRequireUsername", returned: authentication.ErrUsernameIsRequired, expected: "username is required"},
			{name: "ShouldRequireFamilyName", returned: authentication.ErrFamilyNameIsRequired, expected: "family name (last name) is required for this backend"},
			{name: "ShouldWrapOtherErrors", returned: errors.New("something else"), expected: "validation failed for user 'john': something else"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
				defer ctrl.Finish()

				payload, err := marshalTestUser(authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}, Password: "apple123"})
				require.NoError(t, err)

				userProvider.EXPECT().ValidateUserData(gomock.Any()).Return(tc.returned)
				storageProvider.EXPECT().Close().Return(nil)

				cmd, _, _ := newTestCmd(t)
				cmd.SetIn(strings.NewReader(payload))
				cmd.Flags().StringP(cmdFlagNameFile, "f", "", "")
				require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "-"))

				err = ctx.UsersAddRunE()(cmd, nil)
				require.EqualError(t, err, tc.expected)
			})
		}
	})

	t.Run("ShouldRejectPasswordViolatingPolicy", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		ctx.config.PasswordPolicy.Standard.Enabled = true
		ctx.config.PasswordPolicy.Standard.MinLength = 20

		payload, err := marshalTestUser(authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}, Password: "short"})
		require.NoError(t, err)

		userProvider.EXPECT().ValidateUserData(gomock.Any()).Return(nil)
		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.SetIn(strings.NewReader(payload))
		cmd.Flags().StringP(cmdFlagNameFile, "f", "", "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "-"))

		err = ctx.UsersAddRunE()(cmd, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password does not meet policy requirements")
	})

	t.Run("ShouldWrapAddUserError", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		payload, err := marshalTestUser(authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}, Password: "apple123"})
		require.NoError(t, err)

		userProvider.EXPECT().ValidateUserData(gomock.Any()).Return(nil)
		userProvider.EXPECT().AddUser(gomock.Any()).Return(errors.New("backend unavailable"))
		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, _ := newTestCmd(t)
		cmd.SetIn(strings.NewReader(payload))
		cmd.Flags().StringP(cmdFlagNameFile, "f", "", "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "-"))

		err = ctx.UsersAddRunE()(cmd, nil)
		require.EqualError(t, err, "error occurred creating user 'john': backend unavailable")
	})
}

// marshalTestUser builds the JSON payload used by `users add --file -` from a Go value instead of a checked-in
// fixture file.
func marshalTestUser(user authentication.UserDetailsExtended) (string, error) {
	data, err := json.Marshal(&user)
	if err != nil {
		return "", err
	}

	// The 'password' field is deliberately excluded from MarshalJSON (ingest-only), so it must be injected
	// manually to build a representative "add user" payload.
	var m map[string]any

	if err = json.Unmarshal(data, &m); err != nil {
		return "", err
	}

	if user.Password != "" {
		m[authentication.AttributePassword] = user.Password
	}

	out, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// TestUsersAddInteractive covers usersAddInteractive's required-attribute loop and its early return when the
// backend has no optional attributes. usersAddPromptAttribute opens a new bufio.Scanner over cmd.InOrStdin() for
// every prompt, so a test reader that yields more than one line per Read (e.g. strings.Reader) only behaves
// predictably across a single prompt: piping multi-line input into the interactive flow isn't a supported usage
// (`users add --file -` is the supported non-interactive path), so multi-prompt sequencing isn't covered here -
// per-prompt value parsing is already covered exhaustively by TestUsersAddPromptAttribute.
func TestUsersAddInteractive(t *testing.T) {
	ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
	defer ctrl.Finish()

	userProvider.EXPECT().GetRequiredAttributes().Return([]string{authentication.AttributeUsername})
	userProvider.EXPECT().GetSupportedAttributes().Return(map[string]authentication.UserManagementAttributeMetadata{
		authentication.AttributeUsername: {Type: authentication.Text},
	})

	cmd, _, _ := newTestCmd(t)
	cmd.SetIn(strings.NewReader("john\n"))

	user, err := ctx.usersAddInteractive(cmd)
	require.NoError(t, err)
	assert.Equal(t, "john", user.GetUsername())
}

func TestUsersAddPromptAttribute(t *testing.T) {
	t.Run("ShouldPromptGroups", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

		var out bytes.Buffer

		err := usersAddPromptAttribute(strings.NewReader("admins, dev\n"), &out, user, authentication.AttributeGroups, authentication.UserManagementAttributeMetadata{Type: authentication.Groups})
		require.NoError(t, err)
		assert.Equal(t, []string{"admins", "dev"}, user.Groups)
	})

	t.Run("ShouldPromptCheckbox", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

		var out bytes.Buffer

		err := usersAddPromptAttribute(strings.NewReader("yes\n"), &out, user, "extra.test_flag", authentication.UserManagementAttributeMetadata{Type: authentication.Checkbox})
		require.NoError(t, err)
		assert.Equal(t, true, user.Extra["test_flag"])
	})

	t.Run("ShouldPromptNumber", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

		var out bytes.Buffer

		err := usersAddPromptAttribute(strings.NewReader("42\n"), &out, user, "extra.employee_number", authentication.UserManagementAttributeMetadata{Type: authentication.Number})
		require.NoError(t, err)
		assert.Equal(t, int64(42), user.Extra["employee_number"])
	})

	t.Run("ShouldErrorOnInvalidNumber", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

		var out bytes.Buffer

		err := usersAddPromptAttribute(strings.NewReader("not-a-number\n"), &out, user, "extra.employee_number", authentication.UserManagementAttributeMetadata{Type: authentication.Number})
		require.Error(t, err)
	})

	t.Run("ShouldPromptMultiValuedText", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

		var out bytes.Buffer

		err := usersAddPromptAttribute(strings.NewReader("a, b\n"), &out, user, "extra.tags", authentication.UserManagementAttributeMetadata{Type: authentication.Text, Multiple: true})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, user.Extra["tags"])
	})

	t.Run("ShouldNoOpWhenScannerHasNoInput", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

		var out bytes.Buffer

		err := usersAddPromptAttribute(strings.NewReader(""), &out, user, authentication.AttributeDisplayName, authentication.UserManagementAttributeMetadata{Type: authentication.Text})
		require.NoError(t, err)
		assert.Equal(t, "", user.DisplayName)
	})
}

func TestSplitTrimmedCSV(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected []string
	}{
		{name: "ShouldSplitAndTrim", have: "a, b ,c", expected: []string{"a", "b", "c"}},
		{name: "ShouldDropEmptyEntries", have: "a,,b,", expected: []string{"a", "b"}},
		{name: "ShouldReturnEmptySliceForEmptyString", have: "", expected: []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, splitTrimmedCSV(tc.have))
		})
	}
}

func TestUsersSetField(t *testing.T) {
	testCases := []struct {
		name    string
		attr    string
		value   string
		verify  func(t *testing.T, user *authentication.UserDetailsExtended)
		wantErr string
	}{
		{name: "ShouldSetUsername", attr: authentication.AttributeUsername, value: "john", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "john", u.Username)
		}},
		{name: "ShouldSetPassword", attr: authentication.AttributePassword, value: "secret", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "secret", u.Password)
		}},
		{name: "ShouldSetDisplayName", attr: authentication.AttributeDisplayName, value: "John Doe", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "John Doe", u.DisplayName)
		}},
		{name: "ShouldSetMailAsSingleEmail", attr: authentication.AttributeMail, value: "john@example.com", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, []string{"john@example.com"}, u.Emails)
		}},
		{name: "ShouldSetGivenName", attr: authentication.AttributeGivenName, value: "John", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "John", u.GivenName)
		}},
		{name: "ShouldSetFamilyName", attr: authentication.AttributeFamilyName, value: "Doe", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "Doe", u.FamilyName)
		}},
		{name: "ShouldSetMiddleName", attr: authentication.AttributeMiddleName, value: "Q", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "Q", u.MiddleName)
		}},
		{name: "ShouldSetCommonName", attr: authentication.AttributeCommonName, value: "JD", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "JD", u.CommonName)
		}},
		{name: "ShouldSetNickname", attr: authentication.AttributeNickname, value: "Johnny", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "Johnny", u.Nickname)
		}},
		{name: "ShouldSetGender", attr: authentication.AttributeGender, value: "male", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "male", u.Gender)
		}},
		{name: "ShouldSetBirthdate", attr: authentication.AttributeBirthdate, value: "2000-01-01", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "2000-01-01", u.Birthdate)
		}},
		{name: "ShouldSetZoneInfo", attr: authentication.AttributeZoneInfo, value: "America/New_York", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "America/New_York", u.ZoneInfo)
		}},
		{name: "ShouldSetPhoneNumber", attr: authentication.AttributePhoneNumber, value: "+15555550100", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "+15555550100", u.PhoneNumber)
		}},
		{name: "ShouldSetPhoneExtension", attr: authentication.AttributePhoneExtension, value: "123", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "123", u.PhoneExtension)
		}},
		{name: "ShouldSetProfileURL", attr: authentication.AttributeProfile, value: "https://example.com/john", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Profile)
			assert.Equal(t, "https://example.com/john", u.Profile.String())
		}},
		{name: "ShouldSetPictureURL", attr: authentication.AttributePicture, value: "https://example.com/john.png", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Picture)
			assert.Equal(t, "https://example.com/john.png", u.Picture.String())
		}},
		{name: "ShouldSetWebsiteURL", attr: authentication.AttributeWebsite, value: "https://example.com", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Website)
			assert.Equal(t, "https://example.com", u.Website.String())
		}},
		{name: "ShouldSetLocale", attr: authentication.AttributeLocale, value: "en-US", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Locale)
			assert.Equal(t, "en-US", u.Locale.String())
		}},
		{name: "ShouldSetStreetAddress", attr: authentication.AttributeAddressStreetAddress, value: "1 Main St", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Address)
			assert.Equal(t, "1 Main St", u.Address.StreetAddress)
		}},
		{name: "ShouldSetLocality", attr: authentication.AttributeAddressLocality, value: "Anytown", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Address)
			assert.Equal(t, "Anytown", u.Address.Locality)
		}},
		{name: "ShouldSetRegion", attr: authentication.AttributeAddressRegion, value: "CA", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Address)
			assert.Equal(t, "CA", u.Address.Region)
		}},
		{name: "ShouldSetPostalCode", attr: authentication.AttributeAddressPostalCode, value: "90210", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Address)
			assert.Equal(t, "90210", u.Address.PostalCode)
		}},
		{name: "ShouldSetCountry", attr: authentication.AttributeAddressCountry, value: "US", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			require.NotNil(t, u.Address)
			assert.Equal(t, "US", u.Address.Country)
		}},
		{name: "ShouldSetExtraAttribute", attr: "extra.employee_id", value: "123", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Equal(t, "123", u.Extra["employee_id"])
		}},
		{name: "ShouldIgnoreUnknownAttribute", attr: "not_a_real_attribute", value: "x", verify: func(t *testing.T, u *authentication.UserDetailsExtended) {
			assert.Empty(t, u.Extra)
		}},
		{name: "ShouldErrorOnInvalidProfileURL", attr: authentication.AttributeProfile, value: "://bad", wantErr: "invalid URL for attribute 'profile'"},
		{name: "ShouldErrorOnInvalidPictureURL", attr: authentication.AttributePicture, value: "://bad", wantErr: "invalid URL for attribute 'picture'"},
		{name: "ShouldErrorOnInvalidWebsiteURL", attr: authentication.AttributeWebsite, value: "://bad", wantErr: "invalid URL for attribute 'website'"},
		{name: "ShouldErrorOnInvalidLocale", attr: authentication.AttributeLocale, value: "not-a-locale!!", wantErr: "invalid locale for attribute 'locale'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}

			err := usersSetField(user, tc.attr, tc.value)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
			tc.verify(t, user)
		})
	}
}

func TestUsersSetFieldMultiple(t *testing.T) {
	t.Run("ShouldSetGroups", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
		require.NoError(t, usersSetFieldMultiple(user, authentication.AttributeGroups, []string{"admins", "dev"}))
		assert.Equal(t, []string{"admins", "dev"}, user.Groups)
	})

	t.Run("ShouldSetEmails", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
		require.NoError(t, usersSetFieldMultiple(user, authentication.AttributeMail, []string{"a@example.com", "b@example.com"}))
		assert.Equal(t, []string{"a@example.com", "b@example.com"}, user.Emails)
	})

	t.Run("ShouldSetExtraMultiValue", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
		require.NoError(t, usersSetFieldMultiple(user, "extra.tags", []string{"a", "b"}))
		assert.Equal(t, []string{"a", "b"}, user.Extra["tags"])
	})

	t.Run("ShouldIgnoreUnknownAttribute", func(t *testing.T) {
		user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
		require.NoError(t, usersSetFieldMultiple(user, "not_a_real_attribute", []string{"a"}))
		assert.Empty(t, user.Extra)
	})
}

func TestUsersSetFieldBool(t *testing.T) {
	user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
	require.NoError(t, usersSetFieldBool(user, "extra.test_flag", true))
	assert.Equal(t, true, user.Extra["test_flag"])

	user = &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
	require.NoError(t, usersSetFieldBool(user, "not_a_real_attribute", true))
	assert.Empty(t, user.Extra)
}

func TestUsersSetFieldNumber(t *testing.T) {
	user := &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
	require.NoError(t, usersSetFieldNumber(user, "extra.employee_number", 42))
	assert.Equal(t, int64(42), user.Extra["employee_number"])

	user = &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
	require.NoError(t, usersSetFieldNumber(user, "not_a_real_attribute", 42))
	assert.Empty(t, user.Extra)
}

func TestUsersUpdateRunE(t *testing.T) {
	ctx := NewCmdCtx()
	cmd, _, _ := newTestCmd(t)

	err := ctx.UsersUpdateRunE()(cmd, nil)
	require.NoError(t, err)
}

func TestUsersDeleteRunE(t *testing.T) {
	t.Run("ShouldDeleteUserAndAllAssociatedData", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().DeleteUser("john").Return(nil)
		storageProvider.EXPECT().DeleteTOTPConfiguration(gomock.Any(), "john").Return(nil)
		storageProvider.EXPECT().DeleteWebAuthnCredentialByUsername(gomock.Any(), "john", "").Return(nil)
		storageProvider.EXPECT().DeletePreferredDuoDevice(gomock.Any(), "john").Return(nil)
		storageProvider.EXPECT().DeleteUserByUsername(gomock.Any(), "john").Return(nil)
		storageProvider.EXPECT().Close().Return(nil)

		cmd, stdout, _ := newTestCmd(t)

		err := ctx.UsersDeleteRunE()(cmd, []string{"john"})
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Successfully deleted user 'john'.")
	})

	t.Run("ShouldContinueDeletingAfterPartialFailuresAndJoinErrors", func(t *testing.T) {
		ctx, ctrl, userProvider, storageProvider := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().DeleteUser("john").Return(errors.New("auth backend down"))
		storageProvider.EXPECT().DeleteTOTPConfiguration(gomock.Any(), "john").Return(nil)
		storageProvider.EXPECT().DeleteWebAuthnCredentialByUsername(gomock.Any(), "john", "").Return(errors.New("webauthn store down"))
		storageProvider.EXPECT().DeletePreferredDuoDevice(gomock.Any(), "john").Return(nil)
		storageProvider.EXPECT().DeleteUserByUsername(gomock.Any(), "john").Return(nil)
		storageProvider.EXPECT().Close().Return(nil)

		cmd, _, stderr := newTestCmd(t)

		err := ctx.UsersDeleteRunE()(cmd, []string{"john"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auth backend down")
		assert.Contains(t, err.Error(), "webauthn store down")
		assert.Contains(t, stderr.String(), "Error occurred deleting user 'john' from the authentication backend")
		assert.Contains(t, stderr.String(), "Error occurred deleting WebAuthn data for user 'john'")
	})
}

func TestUsersGroupsListRunE(t *testing.T) {
	t.Run("ShouldRenderTable", func(t *testing.T) {
		ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().ListGroups().Return([]string{"admins", "dev"}, nil)

		cmd, stdout, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")

		err := ctx.UsersGroupsListRunE()(cmd, nil)
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "admins")
		assert.Contains(t, stdout.String(), "dev")
	})

	t.Run("ShouldRenderJSON", func(t *testing.T) {
		ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().ListGroups().Return([]string{"admins"}, nil)

		cmd, stdout, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")
		require.NoError(t, cmd.Flags().Set(cmdFlagNameFormat, cmdFlagValueFormatJSON))

		err := ctx.UsersGroupsListRunE()(cmd, nil)
		require.NoError(t, err)
		assert.JSONEq(t, `["admins"]`, stdout.String())
	})

	t.Run("ShouldWrapProviderError", func(t *testing.T) {
		ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().ListGroups().Return(nil, errors.New("boom"))

		cmd, _, _ := newTestCmd(t)
		cmd.Flags().StringP(cmdFlagNameFormat, "f", cmdFlagValueFormatTable, "")

		err := ctx.UsersGroupsListRunE()(cmd, nil)
		require.EqualError(t, err, "error occurred retrieving groups: boom")
	})
}

func TestUsersGroupsAddRunE(t *testing.T) {
	t.Run("ShouldCreateGroup", func(t *testing.T) {
		ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().AddGroup("admins").Return(nil)

		cmd, stdout, _ := newTestCmd(t)

		err := ctx.UsersGroupsAddRunE()(cmd, []string{"admins"})
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Successfully created group 'admins'.")
	})

	t.Run("ShouldWrapProviderError", func(t *testing.T) {
		ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().AddGroup("admins").Return(errors.New("boom"))

		cmd, _, _ := newTestCmd(t)

		err := ctx.UsersGroupsAddRunE()(cmd, []string{"admins"})
		require.EqualError(t, err, "error occurred creating group 'admins': boom")
	})
}

func TestUsersGroupsDeleteRunE(t *testing.T) {
	t.Run("ShouldDeleteGroup", func(t *testing.T) {
		ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().DeleteGroup("admins").Return(nil)

		cmd, stdout, _ := newTestCmd(t)

		err := ctx.UsersGroupsDeleteRunE()(cmd, []string{"admins"})
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Successfully deleted group 'admins'.")
	})

	t.Run("ShouldWrapProviderError", func(t *testing.T) {
		ctx, ctrl, userProvider, _ := newTestUsersCmdCtx(t)
		defer ctrl.Finish()

		userProvider.EXPECT().DeleteGroup("admins").Return(errors.New("boom"))

		cmd, _, _ := newTestCmd(t)

		err := ctx.UsersGroupsDeleteRunE()(cmd, []string{"admins"})
		require.EqualError(t, err, "error occurred deleting group 'admins': boom")
	})
}
