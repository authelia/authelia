package authentication

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// LDAPUserProvider is a provider using a LDAP or AD as a user database.
type LDAPUserProvider struct {
	configuration schema.LDAPAuthenticationBackendConfiguration
	tlsConfig     *tls.Config

	connectionFactory LDAPConnectionFactory
}

// NewLDAPUserProvider creates a new instance of LDAPUserProvider.
func NewLDAPUserProvider(configuration schema.LDAPAuthenticationBackendConfiguration) *LDAPUserProvider {
	minimumTLSVersion, _ := utils.TLSStringToTLSConfigVersion(configuration.MinimumTLSVersion)

	// TODO: RELEASE-4.27.0 Deprecated Completely in this release.
	logger := logging.Logger()
	if strings.Contains(configuration.UsersFilter, "{0}") {
		logger.Warnf("DEPRECATION NOTICE: LDAP Users Filter will no longer support replacing `{0}` in 4.27.0. Please use `{input}` instead.")
		configuration.UsersFilter = strings.ReplaceAll(configuration.UsersFilter, "{0}", "{input}")
	}
	if strings.Contains(configuration.GroupsFilter, "{0}") {
		logger.Warnf("DEPRECATION NOTICE: LDAP Groups Filter will no longer support replacing `{0}` in 4.27.0. Please use `{input}` instead.")
		configuration.GroupsFilter = strings.ReplaceAll(configuration.GroupsFilter, "{0}", "{input}")
	}
	if strings.Contains(configuration.GroupsFilter, "{1}") {
		logger.Warnf("DEPRECATION NOTICE: LDAP Groups Filter will no longer support replacing `{1}` in 4.27.0. Please use `{username}` instead.")
		configuration.GroupsFilter = strings.ReplaceAll(configuration.GroupsFilter, "{1}", "{username}")
	}
	// TODO: RELEASE-4.27.0 Deprecated Completely in this release.

	configuration.UsersFilter = strings.ReplaceAll(configuration.UsersFilter, "{username_attribute}", configuration.UsernameAttribute)
	configuration.UsersFilter = strings.ReplaceAll(configuration.UsersFilter, "{mail_attribute}", configuration.MailAttribute)
	configuration.UsersFilter = strings.ReplaceAll(configuration.UsersFilter, "{display_name_attribute}", configuration.DisplayNameAttribute)

	return &LDAPUserProvider{
		configuration: configuration,
		tlsConfig:     &tls.Config{InsecureSkipVerify: configuration.SkipVerify, MinVersion: minimumTLSVersion}, //nolint:gosec // Disabling InsecureSkipVerify is an informed choice by users.

		connectionFactory: NewLDAPConnectionFactoryImpl(),
	}
}

// NewLDAPUserProviderWithFactory creates a new instance of LDAPUserProvider with existing factory.
func NewLDAPUserProviderWithFactory(configuration schema.LDAPAuthenticationBackendConfiguration, connectionFactory LDAPConnectionFactory) *LDAPUserProvider {
	provider := NewLDAPUserProvider(configuration)
	provider.connectionFactory = connectionFactory

	return provider
}

func (p *LDAPUserProvider) connect(userDN string, password string) (LDAPConnection, error) {
	var newConnection LDAPConnection

	ldapURL, err := url.Parse(p.configuration.URL)

	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL to LDAP: %s", ldapURL)
	}

	if ldapURL.Scheme == "ldaps" {
		logging.Logger().Trace("LDAP client starts a TLS session")

		conn, err := p.connectionFactory.DialTLS("tcp", ldapURL.Host, p.tlsConfig)
		if err != nil {
			return nil, err
		}

		newConnection = conn
	} else {
		logging.Logger().Trace("LDAP client starts a session over raw TCP")
		conn, err := p.connectionFactory.Dial("tcp", ldapURL.Host)
		if err != nil {
			return nil, err
		}
		newConnection = conn
	}

	if p.configuration.StartTLS {
		if err := newConnection.StartTLS(p.tlsConfig); err != nil {
			return nil, err
		}
	}

	if err := newConnection.Bind(userDN, password); err != nil {
		return nil, err
	}

	return newConnection, nil
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(inputUsername string, password string) (bool, error) {
	adminClient, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return false, err
	}
	defer adminClient.Close()

	profile, err := p.getUserProfile(adminClient, inputUsername)
	if err != nil {
		return false, err
	}

	conn, err := p.connect(profile.DN, password)
	if err != nil {
		return false, fmt.Errorf("Authentication of user %s failed. Cause: %s", inputUsername, err)
	}
	defer conn.Close()

	return true, nil
}

func (p *LDAPUserProvider) ldapEscape(inputUsername string) string {
	inputUsername = ldap.EscapeFilter(inputUsername)
	for _, c := range specialLDAPRunes {
		inputUsername = strings.ReplaceAll(inputUsername, string(c), fmt.Sprintf("\\%c", c))
	}

	return inputUsername
}

type ldapUserProfile struct {
	DN          string
	Emails      []string
	DisplayName string
	Username    string
}

func (p *LDAPUserProvider) resolveUsersFilter(userFilter string, inputUsername string) string {
	inputUsername = p.ldapEscape(inputUsername)

	// The {input} placeholder is replaced by the users username input.
	userFilter = strings.ReplaceAll(userFilter, "{input}", inputUsername)

	return userFilter
}

func (p *LDAPUserProvider) getUserProfile(conn LDAPConnection, inputUsername string) (*ldapUserProfile, error) {
	userFilter := p.resolveUsersFilter(p.configuration.UsersFilter, inputUsername)
	logging.Logger().Infof("Computed user filter is %s", userFilter)

	baseDN := p.configuration.BaseDN
	if p.configuration.AdditionalUsersDN != "" {
		baseDN = p.configuration.AdditionalUsersDN + "," + baseDN
	}

	attributes := []string{"dn",
		p.configuration.DisplayNameAttribute,
		p.configuration.MailAttribute,
		p.configuration.UsernameAttribute}

	// Search for the given username.
	searchRequest := ldap.NewSearchRequest(
		baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, userFilter, attributes, nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("Cannot find user DN of user %s. Cause: %s", inputUsername, err)
	}

	if len(sr.Entries) == 0 {
		return nil, ErrUserNotFound
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("Multiple users %s found", inputUsername)
	}

	userProfile := ldapUserProfile{
		DN: sr.Entries[0].DN,
	}

	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == p.configuration.DisplayNameAttribute {
			userProfile.DisplayName = attr.Values[0]
		}

		if attr.Name == p.configuration.MailAttribute {
			userProfile.Emails = attr.Values
		}

		if attr.Name == p.configuration.UsernameAttribute {
			if len(attr.Values) != 1 {
				return nil, fmt.Errorf("User %s cannot have multiple value for attribute %s",
					inputUsername, p.configuration.UsernameAttribute)
			}

			userProfile.Username = attr.Values[0]
		}
	}

	if userProfile.DN == "" {
		return nil, fmt.Errorf("No DN has been found for user %s", inputUsername)
	}

	return &userProfile, nil
}

func (p *LDAPUserProvider) resolveGroupsFilter(inputUsername string, profile *ldapUserProfile) (string, error) { //nolint:unparam
	inputUsername = p.ldapEscape(inputUsername)

	// The {input} placeholder is replaced by the users username input.
	groupFilter := strings.ReplaceAll(p.configuration.GroupsFilter, "{input}", inputUsername)

	if profile != nil {
		groupFilter = strings.ReplaceAll(groupFilter, "{username}", ldap.EscapeFilter(profile.Username))
		groupFilter = strings.ReplaceAll(groupFilter, "{dn}", ldap.EscapeFilter(profile.DN))
	}

	return groupFilter, nil
}

// GetDetails retrieve the groups a user belongs to.
func (p *LDAPUserProvider) GetDetails(inputUsername string) (*UserDetails, error) {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)
	if err != nil {
		return nil, err
	}

	groupsFilter, err := p.resolveGroupsFilter(inputUsername, profile)
	if err != nil {
		return nil, fmt.Errorf("Unable to create group filter for user %s. Cause: %s", inputUsername, err)
	}

	logging.Logger().Tracef("Computed groups filter is %s", groupsFilter)

	groupBaseDN := p.configuration.BaseDN
	if p.configuration.AdditionalGroupsDN != "" {
		groupBaseDN = p.configuration.AdditionalGroupsDN + "," + groupBaseDN
	}

	// Search for the given username.
	searchGroupRequest := ldap.NewSearchRequest(
		groupBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, groupsFilter, []string{p.configuration.GroupNameAttribute}, nil,
	)

	sr, err := conn.Search(searchGroupRequest)

	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve groups of user %s. Cause: %s", inputUsername, err)
	}

	groups := make([]string, 0)

	for _, res := range sr.Entries {
		if len(res.Attributes) == 0 {
			logging.Logger().Warningf("No groups retrieved from LDAP for user %s", inputUsername)
			break
		}
		// Append all values of the document. Normally there should be only one per document.
		groups = append(groups, res.Attributes[0].Values...)
	}

	return &UserDetails{
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		Emails:      profile.Emails,
		Groups:      groups,
	}, nil
}

// UpdatePassword update the password of the given user.
func (p *LDAPUserProvider) UpdatePassword(inputUsername string, newPassword string) error {
	client, err := p.connect(p.configuration.User, p.configuration.Password)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	profile, err := p.getUserProfile(client, inputUsername)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	modifyRequest := ldap.NewModifyRequest(profile.DN, nil)

	switch p.configuration.Implementation {
	case schema.LDAPImplementationActiveDirectory:
		utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		// The password needs to be enclosed in quotes
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2
		pwdEncoded, _ := utf16.NewEncoder().String(fmt.Sprintf("\"%s\"", newPassword))
		modifyRequest.Replace("unicodePwd", []string{pwdEncoded})
	default:
		modifyRequest.Replace("userPassword", []string{newPassword})
	}

	err = client.Modify(modifyRequest)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	return nil
}
