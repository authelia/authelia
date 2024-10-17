package authentication

import (
	"net/mail"
)

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string
	DisplayName string
	Emails      []string
	Groups      []string
}

// Addresses returns the Emails []string as []mail.Address formatted with DisplayName as the Name attribute.
func (d UserDetails) Addresses() (addresses []mail.Address) {
	if len(d.Emails) == 0 {
		return nil
	}

	addresses = make([]mail.Address, len(d.Emails))

	for i, email := range d.Emails {
		addresses[i] = mail.Address{
			Name:    d.DisplayName,
			Address: email,
		}
	}

	return addresses
}

func (d UserDetails) GetUsername() (username string) {
	return d.Username
}

func (d UserDetails) GetGroups() (groups []string) {
	return d.Groups
}

func (d UserDetails) GetDisplayName() (name string) {
	return d.DisplayName
}

func (d UserDetails) GetEmails() (emails []string) {
	return d.Emails
}

type ldapUserProfile struct {
	DN          string
	Emails      []string
	DisplayName string
	Username    string
	MemberOf    []string
}

// LDAPSupportedFeatures represents features which a server may support which are implemented in code.
type LDAPSupportedFeatures struct {
	Extensions   LDAPSupportedExtensions
	ControlTypes LDAPSupportedControlTypes
}

// LDAPSupportedExtensions represents extensions which a server may support which are implemented in code.
type LDAPSupportedExtensions struct {
	TLS           bool
	PwdModifyExOp bool
}

// LDAPSupportedControlTypes represents control types which a server may support which are implemented in code.
type LDAPSupportedControlTypes struct {
	MsftPwdPolHints           bool
	MsftPwdPolHintsDeprecated bool
}

// Level is the type representing a level of authentication.
type Level int

const (
	// NotAuthenticated if the user is not authenticated yet.
	NotAuthenticated Level = iota

	// OneFactor if the user has passed first factor only.
	OneFactor

	// TwoFactor if the user has passed two factors.
	TwoFactor
)

// String returns a string representation of an authentication.Level.
func (l Level) String() string {
	switch l {
	case NotAuthenticated:
		return "not_authenticated"
	case OneFactor:
		return "one_factor"
	case TwoFactor:
		return "two_factor"
	default:
		return "invalid"
	}
}
