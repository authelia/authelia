package authentication

import (
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserDetails_Addresses(t *testing.T) {
	details := &UserDetails{}

	assert.Equal(t, []mail.Address(nil), details.Addresses())

	details = &UserDetails{
		DisplayName: "Example",
		Emails:      []string{"abc@123.com"},
	}

	assert.Equal(t, []mail.Address{{Name: "Example", Address: "abc@123.com"}}, details.Addresses())

	details = &UserDetails{
		DisplayName: "Example",
		Emails:      []string{"abc@123.com", "two@apple.com"},
	}

	assert.Equal(t, []mail.Address{{Name: "Example", Address: "abc@123.com"}, {Name: "Example", Address: "two@apple.com"}}, details.Addresses())

	details = &UserDetails{
		DisplayName: "",
		Emails:      []string{"abc@123.com"},
	}

	assert.Equal(t, []mail.Address{{Address: "abc@123.com"}}, details.Addresses())
}

func TestLevel_String(t *testing.T) {
	assert.Equal(t, "one_factor", OneFactor.String())
	assert.Equal(t, "two_factor", TwoFactor.String())
	assert.Equal(t, "not_authenticated", NotAuthenticated.String())
	assert.Equal(t, "invalid", Level(-1).String())
}
