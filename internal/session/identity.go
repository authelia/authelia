package session

import (
	"net/mail"
)

// Address returns the mail.Address for the identity.
func (i Identity) Address() mail.Address {
	return mail.Address{
		Name:    i.DisplayName,
		Address: i.Email,
	}
}
