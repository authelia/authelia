package authentication

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

func (p *LDAPUserProvider) getRFC2307bisRequiredFields() []string {
	return []string{
		"Username",
		"Password",
		"CommonName",
		"FamilyName",
	}
}

func (p *LDAPUserProvider) getRFC2307bisSupportedFields() []string {
	return []string{
		"Username",
		"Password",
		"CommonName",
		"GivenName",
		"FamilyName",
		"Email",
		"Emails",
		"Groups",
		"DN",
		"ObjectClass",
		"Extended",
		"BackendAttributes",
	}
}

func (p *LDAPUserProvider) getRFC2307bisDefaultObjectClasses() []string {
	return []string{
		"top",
		"person",
		"organizationalPerson",
		"inetOrgPerson",
	}
}

// getRFC2307bisFieldMetadata describes the fields that are required to create new users for the RFC2307bis Backend.
func (p *LDAPUserProvider) getRFC2307bisFieldMetadata() map[string]FieldMetadata {
	return map[string]FieldMetadata{
		"Username": {
			Required:    true,
			DisplayName: "Username",
			Description: "Unique identifier for the user (maps to uid attribute)",
			Type:        "string",
			MaxLength:   100,
		},
		"Password": {
			Required:    true,
			DisplayName: "Password",
			Description: "User's password",
			Type:        "password",
		},
		"CommonName": {
			Required:    true,
			DisplayName: "Common Name",
			Description: "Full name or display name (maps to cn attribute)",
			Type:        "string",
		},
		"GivenName": {
			Required:    false,
			DisplayName: "First Name",
			Description: "User's first/given name",
			Type:        "string",
		},
		"FamilyName": {
			Required:    true,
			DisplayName: "Last Name",
			Description: "User's last/family name (maps to sn attribute)",
			Type:        "string",
		},
		"Email": {
			Required:    false,
			DisplayName: "Email Address",
			Description: "Primary email address",
			Type:        "email",
		},
		"Groups": {
			Required:    false,
			DisplayName: "Groups",
			Description: "Groups the user should be added to",
			Type:        "array",
		},
	}
}

func (p *LDAPUserProvider) validateRFC2307bisUserData(userData *UserDetailsExtended) error {
	if userData.Username == "" {
		return fmt.Errorf("username required")
	}
	if userData.Password == "" {
		return fmt.Errorf("password required")
	}
	if userData.CommonName == "" {
		// Try to build it from other fields
		if userData.DisplayName != "" {
			userData.CommonName = userData.DisplayName
		} else if userData.GivenName != "" && userData.FamilyName != "" {
			userData.CommonName = userData.GivenName + " " + userData.FamilyName
		} else {
			return fmt.Errorf("commonName (cn) required for RFC2307bis")
		}
	}
	if userData.FamilyName == "" {
		return fmt.Errorf("familyName (sn) required for RFC2307bis")
	}
	return nil
}

func (p *LDAPUserProvider) createRFC2307bisAddRequest(userData *UserDetailsExtended) (*ldap.AddRequest, error) {
	userDN := fmt.Sprintf("%s=%s,%s", p.config.Attributes.Username, ldap.EscapeFilter(userData.Username), p.usersBaseDN)

	addRequest := ldap.NewAddRequest(userDN, nil)

	// RFC2307bis requires these object classes
	addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "inetOrgPerson"})

	addRequest.Attribute("uid", []string{userData.UserDetails.Username})
	addRequest.Attribute("cn", []string{userData.CommonName})
	addRequest.Attribute("sn", []string{userData.FamilyName})
	addRequest.Attribute("userPassword", []string{userData.Password})

	// Optional attributes
	if userData.GivenName != "" {
		addRequest.Attribute("givenName", []string{userData.GivenName})
	}
	if len(userData.UserDetails.Emails) > 0 {
		addRequest.Attribute("mail", []string{userData.UserDetails.Emails[0]})
	}

	return addRequest, nil
}
