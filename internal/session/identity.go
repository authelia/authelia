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

// RecoveryAddress returns the mail.Address for the identity.
func (i Identity) RecoveryAddress() mail.Address {
	if len(i.AlternateEmails) > 0 {
		return mail.Address{
			Name:    i.DisplayName,
			Address: i.AlternateEmails[0],
		}
	} else {
		return mail.Address{
			Name:    i.DisplayName,
			Address: i.Email,
		}
	}
}
