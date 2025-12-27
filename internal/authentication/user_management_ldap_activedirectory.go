package authentication

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

type ActiveDirectoryUserManagement struct {
	provider *LDAPUserProvider
}

func (a *ActiveDirectoryUserManagement) UpdateUser(username string, userData *UserDetailsExtended) (err error) {
	// TODO: implement me.
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) DeleteUser(username string) (err error) {
	// TODO implement me.
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) GetRequiredFields() []string {
	return []string{
		"username",
		"password",
		"full_name",
		"last_name",
	}
}

func (a *ActiveDirectoryUserManagement) GetSupportedFields() []string {
	return []string{
		"username",
		"password",
		"full_name",
		"first_name",
		"last_name",
		"email",
		"emails",
		"groups",
		"object_class",
		"extra",
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

// GetFieldMetadata describes the fields that are required to create new users for the Active Directory Backend.
func (a *ActiveDirectoryUserManagement) GetFieldMetadata() map[string]FieldMetadata {
	return map[string]FieldMetadata{
		"username": {
			DisplayName: "Username",
			Description: "Unique identifier for the user (maps to sAMAccountName attribute)",
			Type:        "string",
			MaxLength:   64, MaxLength: 64, // AD sAMAccountName limit.
		},
		"password": {
			DisplayName: "Password",
			Description: "User's password",
			Type:        "password",
		},
		"full_name": {
			DisplayName: "Full Name",
			Description: "Full name or display name (maps to cn attribute)",
			Type:        "string",
		},
		"first_name": {
			DisplayName: "First Name",
			Description: "User's first/given name (maps to givenName attribute)",
			Type:        "string",
		},
		"last_name": {
			DisplayName: "Last Name",
			Description: "User's last/family name (maps to sn attribute)",
			Type:        "string",
		},
		"email": {
			DisplayName: "Email Address",
			Description: "Primary email address (maps to mail attribute)",
			Type:        "email",
		},
		"emails": {
			DisplayName: "Additional Email Addresses",
			Description: "Additional email addresses for the user",
			Type:        "array",
		},
		"groups": {
			DisplayName: "Groups",
			Description: "Groups the user should be added to",
			Type:        "array",
		},
		"object_class": {
			DisplayName: "Object Classes",
			Description: "LDAP object classes for the user",
			Type:        "array",
		},
		"extra": {
			DisplayName: "Additional Attributes",
			Description: "Additional LDAP attributes as key-value pairs",
			Type:        "object",
		},
	}
}

func (a *ActiveDirectoryUserManagement) ValidateUserData(userData *UserDetailsExtended) error {
	if userData.Username == "" {
		return fmt.Errorf("username required")
	}

	if userData.CommonName == "" {
		// Try to build it from other fields.
		//nolint:gocritic
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

	// RFC2307bis requires these object classes.
	addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "inetOrgPerson"})

	addRequest.Attribute("uid", []string{userData.Username})
	addRequest.Attribute("cn", []string{userData.CommonName})
	addRequest.Attribute("sn", []string{userData.FamilyName})
	addRequest.Attribute("userPassword", []string{userData.Password})

	// Optional attributes.
	if userData.GivenName != "" {
		addRequest.Attribute("givenName", []string{userData.GivenName})
	}

	if len(userData.Emails) > 0 {
		addRequest.Attribute("mail", []string{userData.Emails[0]})
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
