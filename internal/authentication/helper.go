package authentication

import (
	"net/mail"
)

// parses email address, and fail if it has a invalid format.
func parseEmail(email string) (string, error) {
	address, err := mail.ParseAddress(email)

	if err != nil {
		return "", err
	}

	return address.Address, nil
}
