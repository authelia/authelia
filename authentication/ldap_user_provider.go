package authentication

import (
	"fmt"
	"strings"

	"github.com/clems4ever/authelia/configuration/schema"
	"gopkg.in/ldap.v3"
)

// LDAPUserProvider is a provider using a LDAP or AD as a user database.
type LDAPUserProvider struct {
	configuration schema.LDAPAuthenticationBackendConfiguration
}

func (p *LDAPUserProvider) connect(userDN string, password string) (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", p.configuration.URL)
	if err != nil {
		return nil, err
	}

	err = conn.Bind(userDN, password)

	if err != nil {
		return nil, err
	}
	return conn, nil
}

// NewLDAPUserProvider creates a new instance of LDAPUserProvider.
func NewLDAPUserProvider(configuration schema.LDAPAuthenticationBackendConfiguration) *LDAPUserProvider {
	return &LDAPUserProvider{configuration}
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(username string, password string) (bool, error) {
	adminClient, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return false, err
	}
	defer adminClient.Close()

	userDN, err := p.getUserDN(adminClient, username)
	if err != nil {
		return false, err
	}

	conn, err := p.connect(userDN, password)
	if err != nil {
		return false, fmt.Errorf("Authentication of user %s failed. Cause: %s", username, err)
	}
	defer conn.Close()

	return true, nil
}

func (p *LDAPUserProvider) getUserAttribute(conn *ldap.Conn, username string, attribute string) ([]string, error) {
	client, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	userFilter := strings.Replace(p.configuration.UsersFilter, "{0}", username, -1)
	baseDN := p.configuration.AdditionalUsersDN + "," + p.configuration.BaseDN

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, userFilter, []string{attribute}, nil,
	)

	sr, err := client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("Cannot find user DN of user %s. Cause: %s", username, err)
	}

	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("No %s found for user %s", attribute, username)
	}

	if attribute == "dn" {
		return []string{sr.Entries[0].DN}, nil
	}

	return sr.Entries[0].Attributes[0].Values, nil
}

func (p *LDAPUserProvider) getUserDN(conn *ldap.Conn, username string) (string, error) {
	values, err := p.getUserAttribute(conn, username, "dn")

	if err != nil {
		return "", err
	}

	if len(values) != 1 {
		return "", fmt.Errorf("DN attribute of user %s must be set", username)
	}

	return values[0], nil
}

func (p *LDAPUserProvider) getUserUID(conn *ldap.Conn, username string) (string, error) {
	values, err := p.getUserAttribute(conn, username, "uid")

	if err != nil {
		return "", err
	}

	if len(values) != 1 {
		return "", fmt.Errorf("UID attribute of user %s must be set", username)
	}

	return values[0], nil
}

func (p *LDAPUserProvider) createGroupsFilter(conn *ldap.Conn, username string) (string, error) {
	if strings.Index(p.configuration.GroupsFilter, "{0}") >= 0 {
		return strings.Replace(p.configuration.GroupsFilter, "{0}", username, -1), nil
	} else if strings.Index(p.configuration.GroupsFilter, "{dn}") >= 0 {
		userDN, err := p.getUserDN(conn, username)
		if err != nil {
			return "", err
		}
		return strings.Replace(p.configuration.GroupsFilter, "{dn}", userDN, -1), nil
	} else if strings.Index(p.configuration.GroupsFilter, "{uid}") >= 0 {
		userUID, err := p.getUserUID(conn, username)
		if err != nil {
			return "", err
		}
		return strings.Replace(p.configuration.GroupsFilter, "{uid}", userUID, -1), nil
	}
	return p.configuration.GroupsFilter, nil
}

// GetDetails retrieve the groups a user belongs to.
func (p *LDAPUserProvider) GetDetails(username string) (*UserDetails, error) {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	groupsFilter, err := p.createGroupsFilter(conn, username)
	if err != nil {
		return nil, fmt.Errorf("Unable to create group filter for user %s. Cause: %s", username, err)
	}

	groupBaseDN := fmt.Sprintf("%s,%s", p.configuration.AdditionalGroupsDN, p.configuration.BaseDN)

	// Search for the given username
	searchGroupRequest := ldap.NewSearchRequest(
		groupBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, groupsFilter, []string{p.configuration.GroupNameAttribute}, nil,
	)

	sr, err := conn.Search(searchGroupRequest)

	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve groups of user %s. Cause: %s", username, err)
	}

	groups := make([]string, 0)

	for _, res := range sr.Entries {
		// append all values of the document. Normally there should be only one per document.
		groups = append(groups, res.Attributes[0].Values...)
	}

	userDN, err := p.getUserDN(conn, username)

	if err != nil {
		return nil, err
	}

	searchEmailRequest := ldap.NewSearchRequest(
		userDN, ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		0, 0, false, "(cn=*)", []string{p.configuration.MailAttribute}, nil,
	)

	sr, err = conn.Search(searchEmailRequest)

	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve email of user %s. Cause: %s", username, err)
	}

	emails := make([]string, 0)

	for _, res := range sr.Entries {
		// append all values of the document. Normally there should be only one per document.
		emails = append(emails, res.Attributes[0].Values...)
	}

	return &UserDetails{
		Emails: emails,
		Groups: groups,
	}, nil
}

// UpdatePassword update the password of the given user.
func (p *LDAPUserProvider) UpdatePassword(username string, newPassword string) error {
	client, err := p.connect(p.configuration.User, p.configuration.Password)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	userDN, err := p.getUserDN(client, username)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	modifyRequest := ldap.NewModifyRequest(userDN, nil)

	modifyRequest.Replace("userPassword", []string{newPassword})

	err = client.Modify(modifyRequest)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	return nil
}
