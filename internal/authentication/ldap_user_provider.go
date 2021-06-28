package authentication

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// LDAPUserProvider is a provider using a LDAP or AD as a user database.
type LDAPUserProvider struct {
	configuration     schema.LDAPAuthenticationBackendConfiguration
	tlsConfig         *tls.Config
	dialOpts          ldap.DialOpt
	logger            *logrus.Logger
	connectionFactory LDAPConnectionFactory
	usersBaseDN       string
	groupsBaseDN      string
}

// NewLDAPUserProvider creates a new instance of LDAPUserProvider.
func NewLDAPUserProvider(configuration schema.LDAPAuthenticationBackendConfiguration, certPool *x509.CertPool) *LDAPUserProvider {
	if configuration.TLS == nil {
		configuration.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
	}

	tlsConfig := utils.NewTLSConfig(configuration.TLS, tls.VersionTLS12, certPool)

	var dialOpts ldap.DialOpt

	if tlsConfig != nil {
		dialOpts = ldap.DialWithTLSConfig(tlsConfig)
	}

	provider := &LDAPUserProvider{
		configuration:     configuration,
		tlsConfig:         tlsConfig,
		dialOpts:          dialOpts,
		logger:            logging.Logger(),
		connectionFactory: NewLDAPConnectionFactoryImpl(),
	}

	provider.parseDynamicConfiguration()

	return provider
}

// NewLDAPUserProviderWithFactory creates a new instance of LDAPUserProvider with existing factory.
func NewLDAPUserProviderWithFactory(configuration schema.LDAPAuthenticationBackendConfiguration, certPool *x509.CertPool, connectionFactory LDAPConnectionFactory) *LDAPUserProvider {
	provider := NewLDAPUserProvider(configuration, certPool)
	provider.connectionFactory = connectionFactory

	return provider
}

func (p *LDAPUserProvider) parseDynamicConfiguration() {
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{username_attribute}", p.configuration.UsernameAttribute)
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{mail_attribute}", p.configuration.MailAttribute)
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{display_name_attribute}", p.configuration.DisplayNameAttribute)

	p.logger.Tracef("Dynamically generated users filter is %s", p.configuration.UsersFilter)

	if p.configuration.AdditionalUsersDN != "" {
		p.usersBaseDN = p.configuration.AdditionalUsersDN + "," + p.configuration.BaseDN
	} else {
		p.usersBaseDN = p.configuration.BaseDN
	}

	p.logger.Tracef("Dynamically generated users BaseDN is %s", p.usersBaseDN)

	if p.configuration.AdditionalGroupsDN != "" {
		p.groupsBaseDN = ldap.EscapeFilter(p.configuration.AdditionalGroupsDN + "," + p.configuration.BaseDN)
	} else {
		p.groupsBaseDN = p.configuration.BaseDN
	}

	p.logger.Tracef("Dynamically generated groups BaseDN is %s", p.groupsBaseDN)
}

func (p *LDAPUserProvider) connect(userDN string, password string) (LDAPConnection, error) {
	conn, err := p.connectionFactory.DialURL(p.configuration.URL, p.dialOpts)
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

func (p *LDAPUserProvider) resolveUsersFilter(userFilter string, inputUsername string) string {
	inputUsername = p.ldapEscape(inputUsername)

	// The {input} placeholder is replaced by the users username input.
	userFilter = strings.ReplaceAll(userFilter, "{input}", inputUsername)

	p.logger.Tracef("Computed user filter is %s", userFilter)

	return userFilter
}

func (p *LDAPUserProvider) getUserProfile(conn LDAPConnection, inputUsername string) (*ldapUserProfile, error) {
	userFilter := p.resolveUsersFilter(p.configuration.UsersFilter, inputUsername)

	attributes := []string{"dn",
		p.configuration.DisplayNameAttribute,
		p.configuration.MailAttribute,
		p.configuration.UsernameAttribute}

	// Search for the given username.
	searchRequest := ldap.NewSearchRequest(
		p.usersBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
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

	p.logger.Tracef("Computed groups filter is %s", groupFilter)

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

	// Search for the given username.
	searchGroupRequest := ldap.NewSearchRequest(
		p.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, groupsFilter, []string{p.configuration.GroupNameAttribute}, nil,
	)

	sr, err := conn.Search(searchGroupRequest)

	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve groups of user %s. Cause: %s", inputUsername, err)
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
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	switch p.configuration.Implementation {
	case schema.LDAPImplementationActiveDirectory:
		modifyRequest := ldap.NewModifyRequest(profile.DN, nil)
		utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		// The password needs to be enclosed in quotes
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2
		pwdEncoded, _ := utf16.NewEncoder().String(fmt.Sprintf("\"%s\"", newPassword))
		modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

		err = conn.Modify(modifyRequest)
	default:
		modifyRequest := ldap.NewPasswordModifyRequest(
			profile.DN,
			"",
			newPassword,
		)

		err = conn.PasswordModify(modifyRequest)
	}

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	return nil
}
