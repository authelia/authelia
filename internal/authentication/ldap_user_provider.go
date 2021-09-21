package authentication

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// LDAPUserProvider is a UserProvider that connects to LDAP servers like ActiveDirectory, OpenLDAP, OpenDJ, FreeIPA, etc.
type LDAPUserProvider struct {
	configuration     schema.LDAPAuthenticationBackendConfiguration
	tlsConfig         *tls.Config
	dialOpts          []ldap.DialOpt
	logger            *logrus.Logger
	connectionFactory LDAPConnectionFactory

	disableResetPassword bool

	// Automatically detected ldap features.
	supportExtensionPasswdModify bool

	// Dynamically generated users values.
	usersBaseDN                 string
	usersAttributes             []string
	usersFilterReplacementInput bool

	// Dynamically generated groups values.
	groupsBaseDN                    string
	groupsAttributes                []string
	groupsFilterReplacementInput    bool
	groupsFilterReplacementUsername bool
	groupsFilterReplacementDN       bool
}

// NewLDAPUserProvider creates a new instance of LDAPUserProvider.
func NewLDAPUserProvider(configuration schema.AuthenticationBackendConfiguration, certPool *x509.CertPool) (provider *LDAPUserProvider) {
	provider = newLDAPUserProvider(*configuration.LDAP, configuration.DisableResetPassword, certPool, nil)

	return provider
}

func newLDAPUserProvider(configuration schema.LDAPAuthenticationBackendConfiguration, disableResetPassword bool, certPool *x509.CertPool, factory LDAPConnectionFactory) (provider *LDAPUserProvider) {
	if configuration.TLS == nil {
		configuration.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
	}

	tlsConfig := utils.NewTLSConfig(configuration.TLS, tls.VersionTLS12, certPool)

	var dialOpts = []ldap.DialOpt{
		ldap.DialWithDialer(&net.Dialer{Timeout: configuration.Timeout}),
	}

	if tlsConfig != nil {
		dialOpts = append(dialOpts, ldap.DialWithTLSConfig(tlsConfig))
	}

	if factory == nil {
		factory = NewLDAPConnectionFactoryImpl()
	}

	provider = &LDAPUserProvider{
		configuration:        configuration,
		tlsConfig:            tlsConfig,
		dialOpts:             dialOpts,
		logger:               logging.Logger(),
		connectionFactory:    factory,
		disableResetPassword: disableResetPassword,
	}

	provider.parseDynamicUsersConfiguration()
	provider.parseDynamicGroupsConfiguration()

	return provider
}

func (p *LDAPUserProvider) connect(userDN string, password string) (LDAPConnection, error) {
	conn, err := p.connectionFactory.DialURL(p.configuration.URL, p.dialOpts...)
	if err != nil {
		return nil, err
	}

	if p.configuration.StartTLS {
		if err := conn.StartTLS(p.tlsConfig); err != nil {
			return nil, err
		}
	}

	if err := conn.Bind(userDN, password); err != nil {
		return nil, err
	}

	return conn, nil
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(inputUsername string, password string) (bool, error) {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)
	if err != nil {
		return false, err
	}

	userConn, err := p.connect(profile.DN, password)
	if err != nil {
		return false, fmt.Errorf("Authentication of user %s failed. Cause: %s", inputUsername, err)
	}
	defer userConn.Close()

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

func (p *LDAPUserProvider) resolveUsersFilter(inputUsername string) (filter string) {
	filter = p.configuration.UsersFilter

	if p.usersFilterReplacementInput {
		// The {input} placeholder is replaced by the users username input.
		filter = strings.ReplaceAll(filter, ldapPlaceholderInput, p.ldapEscape(inputUsername))
	}

	p.logger.Tracef("Computed user filter is %s", filter)

	return filter
}

func (p *LDAPUserProvider) getUserProfile(conn LDAPConnection, inputUsername string) (*ldapUserProfile, error) {
	userFilter := p.resolveUsersFilter(inputUsername)

	// Search for the given username.
	searchRequest := ldap.NewSearchRequest(
		p.usersBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, userFilter, p.usersAttributes, nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot find user DN of user '%s'. Cause: %w", inputUsername, err)
	}

	if len(sr.Entries) == 0 {
		return nil, ErrUserNotFound
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("multiple users %s found", inputUsername)
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
				return nil, fmt.Errorf("user '%s' cannot have multiple value for attribute '%s'",
					inputUsername, p.configuration.UsernameAttribute)
			}

			userProfile.Username = attr.Values[0]
		}
	}

	if userProfile.DN == "" {
		return nil, fmt.Errorf("no DN has been found for user %s", inputUsername)
	}

	return &userProfile, nil
}

func (p *LDAPUserProvider) resolveGroupsFilter(inputUsername string, profile *ldapUserProfile) (filter string, err error) { //nolint:unparam
	filter = p.configuration.GroupsFilter

	if p.groupsFilterReplacementInput {
		// The {input} placeholder is replaced by the users username input.
		filter = strings.ReplaceAll(p.configuration.GroupsFilter, ldapPlaceholderInput, p.ldapEscape(inputUsername))
	}

	if profile != nil {
		if p.groupsFilterReplacementUsername {
			filter = strings.ReplaceAll(filter, ldapPlaceholderUsername, ldap.EscapeFilter(profile.Username))
		}

		if p.groupsFilterReplacementDN {
			filter = strings.ReplaceAll(filter, ldapPlaceholderDistinguishedName, ldap.EscapeFilter(profile.DN))
		}
	}

	p.logger.Tracef("Computed groups filter is %s", filter)

	return filter, nil
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
		return nil, fmt.Errorf("unable to create group filter for user '%s'. Cause: %w", inputUsername, err)
	}

	// Search for the given username.
	searchGroupRequest := ldap.NewSearchRequest(
		p.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, groupsFilter, p.groupsAttributes, nil,
	)

	sr, err := conn.Search(searchGroupRequest)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve groups of user '%s'. Cause: %w", inputUsername, err)
	}

	groups := make([]string, 0)

	for _, res := range sr.Entries {
		if len(res.Attributes) == 0 {
			p.logger.Warningf("No groups retrieved from LDAP for user %s", inputUsername)
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
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)

	if err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	switch {
	case p.supportExtensionPasswdModify:
		modifyRequest := ldap.NewPasswordModifyRequest(
			profile.DN,
			"",
			newPassword,
		)

		err = conn.PasswordModify(modifyRequest)
	case p.configuration.Implementation == schema.LDAPImplementationActiveDirectory:
		modifyRequest := ldap.NewModifyRequest(profile.DN, nil)
		utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		// The password needs to be enclosed in quotes
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2
		pwdEncoded, _ := utf16.NewEncoder().String(fmt.Sprintf("\"%s\"", newPassword))
		modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

		err = conn.Modify(modifyRequest)
	default:
		modifyRequest := ldap.NewModifyRequest(profile.DN, nil)
		modifyRequest.Replace("userPassword", []string{newPassword})

		err = conn.Modify(modifyRequest)
	}

	if err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	return nil
}
