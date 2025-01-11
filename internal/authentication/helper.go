package authentication

import (
	"net/mail"

	"github.com/authelia/authelia/v4/internal/model"
)

// parses email address, and fail if it has a invalid format.
func parseEmail(email string) (string, error) {
	address, err := mail.ParseAddress(email)

	if err != nil {
		return "", err
	}

	return address.Address, nil
}

func userModelToUserDetailsExtended(model model.User) UserDetailsExtended {
	return UserDetailsExtended{
		UserDetails: UserDetails{
			Username:    model.Username,
			DisplayName: model.DisplayName,
			Emails:      []string{model.Email},
			Groups:      model.Groups,
		},
		Disabled: model.Disabled,
	}
}
