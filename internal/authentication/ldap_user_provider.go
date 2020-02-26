package authentication

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"gopkg.in/ldap.v3"
)

// LDAPUserProvider is a provider using a LDAP or AD as a user database.
type LDAPUserProvider struct {
	configuration schema.LDAPAuthenticationBackendConfiguration

	connectionFactory LDAPConnectionFactory
}

// NewLDAPUserProvider creates a new instance of LDAPUserProvider.
func NewLDAPUserProvider(configuration schema.LDAPAuthenticationBackendConfiguration) *LDAPUserProvider {
	return &LDAPUserProvider{
		configuration:     configuration,
		connectionFactory: NewLDAPConnectionFactoryImpl(),
	}
}

func NewLDAPUserProviderWithFactory(configuration schema.LDAPAuthenticationBackendConfiguration,
	connectionFactory LDAPConnectionFactory) *LDAPUserProvider {
	return &LDAPUserProvider{
		configuration:     configuration,
		connectionFactory: connectionFactory,
	}
}

func (p *LDAPUserProvider) connect(userDN string, password string) (LDAPConnection, error) {
	var newConnection LDAPConnection

	url, err := url.Parse(p.configuration.URL)

	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL to LDAP: %s", url)
	}

	if url.Scheme == "ldaps" {
		logging.Logger().Debug("LDAP client starts a TLS session")
		conn, err := p.connectionFactory.DialTLS("tcp", url.Host, &tls.Config{
			InsecureSkipVerify: p.configuration.SkipVerify,
		})
		if err != nil {
			return nil, err
		}
		newConnection = conn
	} else {
		logging.Logger().Debug("LDAP client starts a session over raw TCP")
		conn, err := p.connectionFactory.Dial("tcp", url.Host)
		if err != nil {
			return nil, err
		}
		newConnection = conn
	}

	if err := newConnection.Bind(userDN, password); err != nil {
		return nil, err
	}
	return newConnection, nil
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

// OWASP recommends to escape some special characters
// https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/LDAP_Injection_Prevention_Cheat_Sheet.md
const SpecialLDAPRunes = "\\,#+<>;\"="

func (p *LDAPUserProvider) ldapEscape(input string) string {
	for _, c := range SpecialLDAPRunes {
		input = strings.ReplaceAll(input, string(c), fmt.Sprintf("\\%c", c))
	}
	return input
}

func (p *LDAPUserProvider) getUserAttribute(conn LDAPConnection, username string, attribute string) ([]string, error) {
	client, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	username = p.ldapEscape(username)
	userFilter := strings.Replace(p.configuration.UsersFilter, "{0}", username, -1)
	baseDN := p.configuration.BaseDN
	if p.configuration.AdditionalUsersDN != "" {
		baseDN = p.configuration.AdditionalUsersDN + "," + baseDN
	}

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

func (p *LDAPUserProvider) getUserDN(conn LDAPConnection, username string) (string, error) {
	values, err := p.getUserAttribute(conn, username, "dn")

	if err != nil {
		return "", err
	}

	if len(values) != 1 {
		return "", fmt.Errorf("DN attribute of user %s must be set", username)
	}

	return values[0], nil
}

func (p *LDAPUserProvider) getUserUID(conn LDAPConnection, username string) (string, error) {
	values, err := p.getUserAttribute(conn, username, "uid")

	if err != nil {
		return "", err
	}

	if len(values) != 1 {
		return "", fmt.Errorf("UID attribute of user %s must be set", username)
	}

	return values[0], nil
}

func (p *LDAPUserProvider) createGroupsFilter(conn LDAPConnection, username string) (string, error) {
	if strings.Contains(p.configuration.GroupsFilter, "{0}") {
		return strings.Replace(p.configuration.GroupsFilter, "{0}", username, -1), nil
	} else if strings.Contains(p.configuration.GroupsFilter, "{dn}") {
		userDN, err := p.getUserDN(conn, username)
		if err != nil {
			return "", err
		}
		return strings.Replace(p.configuration.GroupsFilter, "{dn}", userDN, -1), nil
	} else if strings.Contains(p.configuration.GroupsFilter, "{uid}") {
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

	groupBaseDN := p.configuration.BaseDN
	if p.configuration.AdditionalGroupsDN != "" {
		groupBaseDN = p.configuration.AdditionalGroupsDN + "," + groupBaseDN
	}

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
		if len(res.Attributes) == 0 {
			logging.Logger().Warningf("No groups retrieved from LDAP for user %s", username)
			break
		}
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
		if len(res.Attributes) == 0 {
			logging.Logger().Warningf("No email retrieved from LDAP for user %s", username)
			break
		}
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
