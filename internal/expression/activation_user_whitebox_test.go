package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserDetailerActivationWhiteBox(t *testing.T) {
	activation := &UserDetailerActivation{}

	assert.Nil(t, activation.Parent())
	assert.Nil(t, activation.address())
}

func TestUserDetailerActivationResolveNameWithParent(t *testing.T) {
	parent := NewMapActivation(nil, map[string]any{
		"custom_from_parent": "parent_value",
	})

	activation := &UserDetailerActivation{
		parent: parent,
		detailer: &UserAttributeResolverDetailer{
			UserDetailer: &testDetailer{},
		},
	}

	testCases := []struct {
		name          string
		resolve       string
		expectedValue any
		expectedFound bool
	}{
		{
			"ShouldResolveKnownAttribute",
			AttributeUserUsername,
			"testuser",
			true,
		},
		{
			"ShouldFallBackToParentForUnknownAttribute",
			"custom_from_parent",
			"parent_value",
			true,
		},
		{
			"ShouldReturnNotFoundForMissingAttribute",
			"nonexistent",
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, found := activation.ResolveName(tc.resolve)

			assert.Equal(t, tc.expectedValue, actual)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

type testDetailer struct{}

func (d *testDetailer) GetUsername() string { return "testuser" }

func (d *testDetailer) GetGroups() []string { return nil }

func (d *testDetailer) GetDisplayName() string { return "" }

func (d *testDetailer) GetEmails() []string { return nil }

func (d *testDetailer) GetGivenName() string { return "" }

func (d *testDetailer) GetFamilyName() string { return "" }

func (d *testDetailer) GetMiddleName() string { return "" }

func (d *testDetailer) GetNickname() string { return "" }

func (d *testDetailer) GetProfile() string { return "" }

func (d *testDetailer) GetPicture() string { return "" }

func (d *testDetailer) GetWebsite() string { return "" }

func (d *testDetailer) GetGender() string { return "" }

func (d *testDetailer) GetBirthdate() string { return "" }

func (d *testDetailer) GetZoneInfo() string { return "" }

func (d *testDetailer) GetLocale() string { return "" }

func (d *testDetailer) GetPhoneNumber() string { return "" }

func (d *testDetailer) GetPhoneExtension() string { return "" }

func (d *testDetailer) GetPhoneNumberRFC3966() string { return "" }

func (d *testDetailer) GetStreetAddress() string { return "" }

func (d *testDetailer) GetLocality() string { return "" }

func (d *testDetailer) GetRegion() string { return "" }

func (d *testDetailer) GetPostalCode() string { return "" }

func (d *testDetailer) GetCountry() string { return "" }

func (d *testDetailer) GetExtra() map[string]any { return nil }
