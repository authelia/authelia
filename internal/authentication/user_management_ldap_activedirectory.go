package authentication

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

type ActiveDirectoryUserManagement struct {
	provider *LDAPUserProvider
}

func (a *ActiveDirectoryUserManagement) UpdateUser(username string, userData *UserDetailsExtended) (err error) {
	//TODO implement me
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) DeleteUser(username string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) GetRequiredFields() []string {
	return []string{
		"Username",
		"Password",
		"CommonName",
		"FamilyName",
	}
}

func (a *ActiveDirectoryUserManagement) GetSupportedFields() []string {
	return []string{
		"Username",
		"Password",
		"CommonName",
		"GivenName",
		"FamilyName",
		"Email",
		"Emails",
		"Groups",
		"ObjectClass",
		"Extra",
	}
}

func (a *ActiveDirectoryUserManagement) GetDefaultObjectClasses() []string {
	return []string{
		"top",
		"person",
		"organizationalPerson",
		"inetOrgPerson",
	}
}

// GetFieldMetadata describes the fields that are required to create new users for the RFC2307bis Backend.
func (a *ActiveDirectoryUserManagement) GetFieldMetadata() map[string]FieldMetadata {
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

func (a *ActiveDirectoryUserManagement) ValidateUserData(userData *UserDetailsExtended) error {
	if userData.Username == "" {
		return fmt.Errorf("username required")
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

func (a *ActiveDirectoryUserManagement) ModifyUser(username string, userData *UserDetailsExtended) error {
	return nil
}

func (a *ActiveDirectoryUserManagement) AddUser(userData *UserDetailsExtended) (err error) {
	if userData == nil || userData.UserDetails == nil {
		return fmt.Errorf("userData and userData.UserDetails cannot be nil")
	}

	if err = a.ValidateUserData(userData); err != nil {
		return fmt.Errorf("validation failed for user '%s': %w", userData.Username, err)
	}

	if userData.Password == "" {
		return fmt.Errorf("password is required to create user '%s'", userData.Username)
	}

	var client ldap.Client
	if client, err = a.provider.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to create user '%s': %w", userData.Username, err)
	}

	defer func() {
		if err := a.provider.factory.ReleaseClient(client); err != nil {
			a.provider.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	userDN := fmt.Sprintf("%s=%s,%s", a.provider.config.Attributes.Username, ldap.EscapeFilter(userData.Username), a.provider.usersBaseDN)

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

	var controls []ldap.Control

	switch {
	case a.provider.features.ControlTypes.MsftPwdPolHints:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints})
	case a.provider.features.ControlTypes.MsftPwdPolHintsDeprecated:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHintsDeprecated})
	}

	// some switch for different implementations.
	if len(controls) > 0 {
		addRequest.Controls = controls
	}

	if err = client.Add(addRequest); err != nil {
		return fmt.Errorf("unable to add user '%s': %w", userData.Username, err)
	}

	return nil
}
