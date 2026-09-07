package authentication

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newTestFileUserManagementProvider(t *testing.T, content []byte, extraAttributes map[string]schema.AuthenticationBackendExtraAttribute) *FileUserProvider {
	t.Helper()

	dir := t.TempDir()

	f := filepath.Join(dir, "users_database.yml")

	require.NoError(t, os.WriteFile(f, content, 0600))

	provider := NewFileUserProvider(&schema.AuthenticationBackendFile{
		Path:            f,
		Password:        schema.DefaultPasswordConfig,
		ExtraAttributes: extraAttributes,
	})

	require.NotNil(t, provider)
	require.NoError(t, provider.StartupCheck())

	return provider
}

func TestFileUserManagementUpdateUserWithMaskExtraAttributes(t *testing.T) {
	testCases := []struct {
		name          string
		updateMask    []string
		extra         map[string]any
		expectedExtra map[string]any
	}{
		{
			name:          "ShouldSetNewExtraAttribute",
			updateMask:    []string{"extra.example"},
			extra:         map[string]any{"example": "456"},
			expectedExtra: map[string]any{"example": "456"},
		},
		{
			name:          "ShouldNotModifyExtraWhenFieldMissingFromUpdateData",
			updateMask:    []string{"extra.example"},
			extra:         nil,
			expectedExtra: map[string]any{"example": "123"},
		},
		{
			name:          "ShouldNotModifyExtraWhenKeyAbsentFromExtraMap",
			updateMask:    []string{"extra.example"},
			extra:         map[string]any{"other": "789"},
			expectedExtra: map[string]any{"example": "123"},
		},
		{
			name:          "ShouldRemoveExtraAttributeWhenValueIsEmptyString",
			updateMask:    []string{"extra.example"},
			extra:         map[string]any{"example": ""},
			expectedExtra: map[string]any{},
		},
		{
			name:          "ShouldIgnoreExtraWhenNotInUpdateMask",
			updateMask:    []string{"given_name"},
			extra:         map[string]any{"example": "999"},
			expectedExtra: map[string]any{"example": "123"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestFileUserManagementProvider(t, UserDatabaseContentExtra, map[string]schema.AuthenticationBackendExtraAttribute{
				"example": {ValueType: ValueTypeString},
			})

			userData := &UserDetailsExtended{
				UserDetails: &UserDetails{Username: "john"},
				Extra:       tc.extra,
			}

			require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, tc.updateMask))

			details, err := provider.database.GetUserDetails("john")
			require.NoError(t, err)

			assert.Equal(t, tc.expectedExtra, details.Extra)
		})
	}
}

func TestFileUserManagementUpdateUserWithMaskExtraAttributeConversion(t *testing.T) {
	provider := newTestFileUserManagementProvider(t, UserDatabaseContent, map[string]schema.AuthenticationBackendExtraAttribute{
		"example": {ValueType: ValueTypeInteger},
	})

	userData := &UserDetailsExtended{
		UserDetails: &UserDetails{Username: "john"},
		Extra:       map[string]any{"example": "456"},
	}

	require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{"extra.example"}))

	details, err := provider.database.GetUserDetails("john")
	require.NoError(t, err)

	assert.Equal(t, int64(456), details.Extra["example"])
}

func TestFileUserManagementUpdateUserWithMaskExtraAttributeConversionError(t *testing.T) {
	provider := newTestFileUserManagementProvider(t, UserDatabaseContent, map[string]schema.AuthenticationBackendExtraAttribute{
		"example": {ValueType: ValueTypeInteger},
	})

	userData := &UserDetailsExtended{
		UserDetails: &UserDetails{Username: "john"},
		Extra:       map[string]any{"example": "not-an-integer"},
	}

	err := provider.Management.UpdateUserWithMask("john", userData, []string{"extra.example"})
	assert.ErrorContains(t, err, "failed to convert extra attribute 'example' for user 'john'")
}

func TestFileUserManagementUpdateUserWithMaskNilUserData(t *testing.T) {
	provider := newTestFileUserManagementProvider(t, UserDatabaseContentExtra, map[string]schema.AuthenticationBackendExtraAttribute{
		"example": {ValueType: ValueTypeString},
	})

	assert.EqualError(t, provider.Management.UpdateUserWithMask("john", nil, []string{"extra.example"}), "userData and userData.UserDetails cannot be nil")
}

func TestFileUserManagementUpdateUserWithMaskUnknownUser(t *testing.T) {
	provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

	userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "nonexistent"}}

	err := provider.Management.UpdateUserWithMask("nonexistent", userData, []string{AttributeGivenName})
	assert.ErrorContains(t, err, "unable to retrieve user for update of user 'nonexistent'")
}

func TestFileUserManagementUpdateUserWithMaskDisabledUser(t *testing.T) {
	provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

	userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "dis"}, GivenName: "Updated"}

	err := provider.Management.UpdateUserWithMask("dis", userData, []string{AttributeGivenName})
	assert.EqualError(t, err, "cannot update disabled user 'dis'")
}

func TestFileUserManagementUpdateUserWithMaskSimpleTextFields(t *testing.T) {
	testCases := []struct {
		field string
		value string
		set   func(userData *UserDetailsExtended, value string)
		get   func(details FileUserDatabaseUserDetails) string
	}{
		{AttributeGivenName, "Jane", func(u *UserDetailsExtended, v string) { u.GivenName = v }, func(d FileUserDatabaseUserDetails) string { return d.GivenName }},
		{AttributeFamilyName, "Smith", func(u *UserDetailsExtended, v string) { u.FamilyName = v }, func(d FileUserDatabaseUserDetails) string { return d.FamilyName }},
		{AttributeMiddleName, "Marie", func(u *UserDetailsExtended, v string) { u.MiddleName = v }, func(d FileUserDatabaseUserDetails) string { return d.MiddleName }},
		{AttributeNickname, "Janie", func(u *UserDetailsExtended, v string) { u.Nickname = v }, func(d FileUserDatabaseUserDetails) string { return d.Nickname }},
		{AttributeGender, "female", func(u *UserDetailsExtended, v string) { u.Gender = v }, func(d FileUserDatabaseUserDetails) string { return d.Gender }},
		{AttributeBirthdate, "1990-01-01", func(u *UserDetailsExtended, v string) { u.Birthdate = v }, func(d FileUserDatabaseUserDetails) string { return d.Birthdate }},
		{AttributeZoneInfo, "America/New_York", func(u *UserDetailsExtended, v string) { u.ZoneInfo = v }, func(d FileUserDatabaseUserDetails) string { return d.ZoneInfo }},
		{AttributePhoneNumber, "+15551234567", func(u *UserDetailsExtended, v string) { u.PhoneNumber = v }, func(d FileUserDatabaseUserDetails) string { return d.PhoneNumber }},
		{AttributePhoneExtension, "42", func(u *UserDetailsExtended, v string) { u.PhoneExtension = v }, func(d FileUserDatabaseUserDetails) string { return d.PhoneExtension }},
	}

	for _, tc := range testCases {
		t.Run(tc.field, func(t *testing.T) {
			provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

			userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}
			tc.set(userData, tc.value)

			require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{tc.field}))

			details, err := provider.database.GetUserDetails("john")
			require.NoError(t, err)

			assert.Equal(t, tc.value, tc.get(details))
		})
	}
}

func TestFileUserManagementUpdateUserWithMaskLocale(t *testing.T) {
	provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

	locale := language.MustParse("en-US")

	userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}, Locale: &locale}

	require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeLocale}))

	details, err := provider.database.GetUserDetails("john")
	require.NoError(t, err)
	require.NotNil(t, details.Locale)
	assert.Equal(t, "en-US", details.Locale.String())
}

func TestFileUserManagementUpdateUserWithMaskURLFields(t *testing.T) {
	testCases := []struct {
		field string
		set   func(userData *UserDetailsExtended, value *url.URL)
		get   func(details FileUserDatabaseUserDetails) *url.URL
	}{
		{AttributeProfile, func(u *UserDetailsExtended, v *url.URL) { u.Profile = v }, func(d FileUserDatabaseUserDetails) *url.URL { return d.Profile }},
		{AttributePicture, func(u *UserDetailsExtended, v *url.URL) { u.Picture = v }, func(d FileUserDatabaseUserDetails) *url.URL { return d.Picture }},
		{AttributeWebsite, func(u *UserDetailsExtended, v *url.URL) { u.Website = v }, func(d FileUserDatabaseUserDetails) *url.URL { return d.Website }},
	}

	for _, tc := range testCases {
		t.Run(tc.field, func(t *testing.T) {
			provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

			value, err := url.Parse("https://example.com/" + tc.field)
			require.NoError(t, err)

			userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}}
			tc.set(userData, value)

			require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{tc.field}))

			details, err := provider.database.GetUserDetails("john")
			require.NoError(t, err)
			require.NotNil(t, tc.get(details))
			assert.Equal(t, value.String(), tc.get(details).String())
		})
	}
}

func TestFileUserManagementUpdateUserWithMaskDisplayName(t *testing.T) {
	t.Run("ShouldUpdateWhenNotEmpty", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", DisplayName: "Johnny Doe"}}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeDisplayName}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Equal(t, "Johnny Doe", details.DisplayName)
	})

	t.Run("ShouldNotUpdateWhenEmpty", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", DisplayName: ""}}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeDisplayName}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Equal(t, "John Doe", details.DisplayName)
	})
}

func TestFileUserManagementUpdateUserWithMaskMail(t *testing.T) {
	t.Run("ShouldUpdateWhenEmailsPresent", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Emails: []string{"new@example.com", "other@example.com"}}}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeMail}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Equal(t, "new@example.com", details.Email)
	})

	t.Run("ShouldNotUpdateWhenEmailsEmpty", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Emails: nil}}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeMail}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Equal(t, "john.doe@authelia.com", details.Email)
	})
}

func TestFileUserManagementUpdateUserWithMaskGroups(t *testing.T) {
	t.Run("ShouldUpdateWhenGroupsNotNil", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Groups: []string{"newgroup"}}}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeGroups}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Equal(t, []string{"newgroup"}, details.Groups)
	})

	t.Run("ShouldUpdateToEmptyWhenGroupsIsEmptySlice", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Groups: []string{}}}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeGroups}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Equal(t, []string{}, details.Groups)
	})

	t.Run("ShouldNotUpdateWhenGroupsNil", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john", Groups: nil}}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeGroups}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Equal(t, []string{"admins", "dev"}, details.Groups)
	})
}

func TestFileUserManagementUpdateUserWithMaskAddress(t *testing.T) {
	address := &UserDetailsAddress{
		StreetAddress: "123 Main St",
		Locality:      "Springfield",
		Region:        "IL",
		PostalCode:    "62701",
		Country:       "US",
	}

	t.Run("ShouldSetAllAddressFieldsWhenAddressPresent", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}, Address: address}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeAddress}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		require.NotNil(t, details.Address)
		assert.Equal(t, address.StreetAddress, details.Address.StreetAddress)
		assert.Equal(t, address.Locality, details.Address.Locality)
		assert.Equal(t, address.Region, details.Address.Region)
		assert.Equal(t, address.PostalCode, details.Address.PostalCode)
		assert.Equal(t, address.Country, details.Address.Country)
	})

	t.Run("ShouldNotCreateAddressWhenNil", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}, Address: nil}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{AttributeAddress}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Nil(t, details.Address)
	})
}

func TestFileUserManagementUpdateUserWithMaskAddressSubFields(t *testing.T) {
	address := &UserDetailsAddress{
		StreetAddress: "123 Main St",
		Locality:      "Springfield",
		Region:        "IL",
		PostalCode:    "62701",
		Country:       "US",
	}

	testCases := []struct {
		field string
		get   func(a *FileUserDatabaseUserDetailsAddressModel) string
	}{
		{AttributeAddressStreetAddress, func(a *FileUserDatabaseUserDetailsAddressModel) string { return a.StreetAddress }},
		{AttributeAddressLocality, func(a *FileUserDatabaseUserDetailsAddressModel) string { return a.Locality }},
		{AttributeAddressRegion, func(a *FileUserDatabaseUserDetailsAddressModel) string { return a.Region }},
		{AttributeAddressPostalCode, func(a *FileUserDatabaseUserDetailsAddressModel) string { return a.PostalCode }},
		{AttributeAddressCountry, func(a *FileUserDatabaseUserDetailsAddressModel) string { return a.Country }},
	}

	for _, tc := range testCases {
		t.Run(tc.field, func(t *testing.T) {
			provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

			userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}, Address: address}

			require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{PrefixAttributeAddress + tc.field}))

			details, err := provider.database.GetUserDetails("john")
			require.NoError(t, err)
			require.NotNil(t, details.Address)

			expected := &FileUserDatabaseUserDetailsAddressModel{
				StreetAddress: address.StreetAddress,
				Locality:      address.Locality,
				Region:        address.Region,
				PostalCode:    address.PostalCode,
				Country:       address.Country,
			}

			assert.Equal(t, tc.get(expected), tc.get(details.Address))
		})
	}

	t.Run("ShouldNotCreateAddressWhenNil", func(t *testing.T) {
		provider := newTestFileUserManagementProvider(t, UserDatabaseContent, nil)

		userData := &UserDetailsExtended{UserDetails: &UserDetails{Username: "john"}, Address: nil}

		require.NoError(t, provider.Management.UpdateUserWithMask("john", userData, []string{PrefixAttributeAddress + AttributeAddressStreetAddress}))

		details, err := provider.database.GetUserDetails("john")
		require.NoError(t, err)
		assert.Nil(t, details.Address)
	})
}
