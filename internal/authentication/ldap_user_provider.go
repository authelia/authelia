package authentication

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/go-ldap/ldap/v3"
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
		logging.Logger().Trace("LDAP client starts a TLS session")
		conn, err := p.connectionFactory.DialTLS("tcp", url.Host, &tls.Config{
			InsecureSkipVerify: p.configuration.SkipVerify,
		})
		if err != nil {
			return nil, err
		}
		newConnection = conn
	} else {
		logging.Logger().Trace("LDAP client starts a session over raw TCP")
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

	profile, err := p.getUserProfile(adminClient, username)
	if err != nil {
		return false, err
	}

	conn, err := p.connect(profile.DN, password)
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

type ldapUserProfile struct {
	DN       string
	Emails   []string
	Username string
}

func (p *LDAPUserProvider) getUserProfile(conn LDAPConnection, username string) (*ldapUserProfile, error) {
	username = p.ldapEscape(username)
	userFilter := fmt.Sprintf("(%s=%s)", p.configuration.UsernameAttribute, username)
	if p.configuration.UsersFilter != "" {
		userFilter = fmt.Sprintf("(&%s%s)", userFilter, p.configuration.UsersFilter)
	}
	baseDN := p.configuration.BaseDN
	if p.configuration.AdditionalUsersDN != "" {
		baseDN = p.configuration.AdditionalUsersDN + "," + baseDN
	}

	attributes := []string{"dn",
		p.configuration.MailAttribute,
		p.configuration.UsernameAttribute}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, userFilter, attributes, nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("Cannot find user DN of user %s. Cause: %s", username, err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("No user %s found", username)
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("Multiple users %s found", username)
	}

	userProfile := ldapUserProfile{
		DN: sr.Entries[0].DN,
	}
	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == p.configuration.MailAttribute {
			userProfile.Emails = attr.Values
		} else if attr.Name == p.configuration.UsernameAttribute {
			if len(attr.Values) != 1 {
				return nil, fmt.Errorf("User %s cannot have multiple value for attribute %s", username, p.configuration.UsernameAttribute)
			}
			userProfile.Username = attr.Values[0]
		}
	}

	if userProfile.DN == "" {
		return nil, fmt.Errorf("No DN has been found for user %s", username)
	}

	return &userProfile, nil
}

func (p *LDAPUserProvider) createGroupsFilter(conn LDAPConnection, username string) (string, error) {
	if strings.Contains(p.configuration.GroupsFilter, "{0}") {
		return strings.Replace(p.configuration.GroupsFilter, "{0}", username, -1), nil
	} else if strings.Contains(p.configuration.GroupsFilter, "{dn}") {
		profile, err := p.getUserProfile(conn, username)
		if err != nil {
			return "", err
		}
		return strings.Replace(p.configuration.GroupsFilter, "{dn}", ldap.EscapeFilter(profile.DN), -1), nil
	} else if strings.Contains(p.configuration.GroupsFilter, "{1}") {
		profile, err := p.getUserProfile(conn, username)
		if err != nil {
			return "", err
		}
		return strings.Replace(p.configuration.GroupsFilter, "{1}", profile.Username, -1), nil
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

	profile, err := p.getUserProfile(conn, username)
	if err != nil {
		return nil, err
	}

	return &UserDetails{
		Username: profile.Username,
		Emails:   profile.Emails,
		Groups:   groups,
	}, nil
}

// UpdatePassword update the password of the given user.
func (p *LDAPUserProvider) UpdatePassword(username string, newPassword string) error {
	client, err := p.connect(p.configuration.User, p.configuration.Password)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	profile, err := p.getUserProfile(client, username)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	modifyRequest := ldap.NewModifyRequest(profile.DN, nil)

	modifyRequest.Replace("userPassword", []string{newPassword})

	err = client.Modify(modifyRequest)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	return nil
}
