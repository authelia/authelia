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

func (a *ActiveDirectoryUserManagement) UpdateUserWithMask(username string, userData *UserDetailsExtended, updateMask []string) (err error) {
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) AddGroup(newGroup string) (err error) {
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) DeleteGroup(group string) (err error) {
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) ListGroups() (groups []string, err error) {
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) DeleteUser(username string) (err error) {
	// TODO implement me.
	panic("implement me")
}

func (a *ActiveDirectoryUserManagement) GetRequiredAttributes() []string {
	return []string{
		"username",
		"password",
		"full_name",
		"last_name",
	}
}

func (a *ActiveDirectoryUserManagement) GetSupportedAttributes() map[string]UserManagementAttributeMetadata {
	return map[string]UserManagementAttributeMetadata{}
}

func (a *ActiveDirectoryUserManagement) GetDefaultObjectClasses() []string {
	return []string{
		"top",
		"person",
		"organizationalPerson",
		"inetOrgPerson",
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

func (a *ActiveDirectoryUserManagement) ValidatePartialUpdate(userData *UserDetailsExtended, updateMask []string) error {
	panic("implement me")
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

	var client LDAPExtendedClient
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
	case client.Features().ControlTypes.MsftPwdPolHints:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints})
	case client.Features().ControlTypes.MsftPwdPolHintsDeprecated:
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
