package authentication

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// LDAPUserProvider is a UserProvider that connects to LDAP servers like ActiveDirectory, OpenLDAP, OpenDJ, FreeIPA, etc.
type LDAPUserProvider struct {
	config            schema.LDAPAuthenticationBackendConfiguration
	tlsConfig         *tls.Config
	dialOpts          []ldap.DialOpt
	log               *logrus.Logger
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
func NewLDAPUserProvider(config schema.AuthenticationBackendConfiguration, certPool *x509.CertPool) (provider *LDAPUserProvider) {
	provider = newLDAPUserProvider(*config.LDAP, config.DisableResetPassword, certPool, nil)

	return provider
}

func newLDAPUserProvider(config schema.LDAPAuthenticationBackendConfiguration, disableResetPassword bool, certPool *x509.CertPool, factory LDAPConnectionFactory) (provider *LDAPUserProvider) {
	if config.TLS == nil {
		config.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
	}

	tlsConfig := utils.NewTLSConfig(config.TLS, tls.VersionTLS12, certPool)

	var dialOpts = []ldap.DialOpt{
		ldap.DialWithDialer(&net.Dialer{Timeout: config.Timeout}),
	}

	if tlsConfig != nil {
		dialOpts = append(dialOpts, ldap.DialWithTLSConfig(tlsConfig))
	}

	if factory == nil {
		factory = NewProductionLDAPConnectionFactory()
	}

	provider = &LDAPUserProvider{
		config:               config,
		tlsConfig:            tlsConfig,
		dialOpts:             dialOpts,
		log:                  logging.Logger(),
		connectionFactory:    factory,
		disableResetPassword: disableResetPassword,
	}

	provider.parseDynamicUsersConfiguration()
	provider.parseDynamicGroupsConfiguration()

	return provider
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(inputUsername string, password string) (valid bool, err error) {
	var (
		conn, connUser LDAPConnection
		profile        *ldapUserProfile
	)

	if conn, err = p.connect(); err != nil {
		return false, err
	}

	defer conn.Close()

	if profile, err = p.getUserProfile(conn, inputUsername); err != nil {
		return false, err
	}

	if connUser, err = p.connectCustom(p.config.URL, profile.DN, password, p.dialOpts...); err != nil {
		return false, fmt.Errorf("authentication failed. Cause: %w", err)
	}

	defer connUser.Close()

	return true, nil
}

// GetDetails retrieve the groups a user belongs to.
func (p *LDAPUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var (
		conn    LDAPConnection
		profile *ldapUserProfile
	)

	if conn, err = p.connect(); err != nil {
		return nil, err
	}

	defer conn.Close()

	if profile, err = p.getUserProfile(conn, username); err != nil {
		return nil, err
	}

	var (
		filter        string
		searchRequest *ldap.SearchRequest
		searchResult  *ldap.SearchResult
	)

	if filter, err = p.resolveGroupsFilter(username, profile); err != nil {
		return nil, fmt.Errorf("unable to create group filter for user '%s'. Cause: %w", username, err)
	}

	// Search for the users groups.
	searchRequest = ldap.NewSearchRequest(
		p.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, filter, p.groupsAttributes, nil,
	)

	if searchResult, err = conn.Search(searchRequest); err != nil {
		return nil, fmt.Errorf("unable to retrieve groups of user '%s'. Cause: %w", username, err)
	}

	groups := make([]string, 0)

	for _, res := range searchResult.Entries {
		if len(res.Attributes) == 0 {
			p.log.Warningf("No groups retrieved from LDAP for user %s", username)
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
func (p *LDAPUserProvider) UpdatePassword(username, password string) (err error) {
	var (
		conn    LDAPConnection
		profile *ldapUserProfile
	)

	if conn, err = p.connect(); err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	defer conn.Close()

	if profile, err = p.getUserProfile(conn, username); err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	var controls []ldap.Control

	switch {
	case p.supportExtensionPasswdModify:
		modifyRequest := ldap.NewPasswordModifyRequest(
			profile.DN,
			"",
			password,
		)

		err = p.pwdModify(conn, modifyRequest)
	case p.config.Implementation == schema.LDAPImplementationActiveDirectory:
		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)
		// The password needs to be enclosed in quotes
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2
		pwdEncoded, _ := utf16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", password))
		modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

		err = conn.Modify(modifyRequest)
	default:
		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)
		modifyRequest.Replace("userPassword", []string{password})

		err = conn.Modify(modifyRequest)
	}

	if err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	return nil
}

func (p *LDAPUserProvider) connect() (LDAPConnection, error) {
	return p.connectCustom(p.config.URL, p.config.User, p.config.Password, p.dialOpts...)
}

func (p *LDAPUserProvider) connectCustom(url, userDN, password string, opts ...ldap.DialOpt) (conn LDAPConnection, err error) {
	if conn, err = p.connectionFactory.DialURL(url, opts...); err != nil {
		return nil, fmt.Errorf("dial failed with error: %w", err)
	}

	if p.config.StartTLS {
		if err = conn.StartTLS(p.tlsConfig); err != nil {
			return nil, fmt.Errorf("starttls failed with error: %w", err)
		}
	}

	if err = conn.Bind(userDN, password); err != nil {
		return nil, fmt.Errorf("bind failed with error: %w", err)
	}

	return conn, nil
}

func (p *LDAPUserProvider) resolveUsersFilter(inputUsername string) (filter string) {
	filter = p.config.UsersFilter

	if p.usersFilterReplacementInput {
		// The {input} placeholder is replaced by the username input.
		filter = strings.ReplaceAll(filter, ldapPlaceholderInput, ldapEscape(inputUsername))
	}

	p.log.Tracef("Detected user filter is %s", filter)

	return filter
}

func (p *LDAPUserProvider) getUserProfile(conn LDAPConnection, inputUsername string) (profile *ldapUserProfile, err error) {
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
		if attr.Name == p.config.DisplayNameAttribute {
			userProfile.DisplayName = attr.Values[0]
		}

		if attr.Name == p.config.MailAttribute {
			userProfile.Emails = attr.Values
		}

		if attr.Name == p.config.UsernameAttribute {
			if len(attr.Values) != 1 {
				return nil, fmt.Errorf("user '%s' cannot have multiple value for attribute '%s'",
					inputUsername, p.config.UsernameAttribute)
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
	filter = p.config.GroupsFilter

	if p.groupsFilterReplacementInput {
		// The {input} placeholder is replaced by the users username input.
		filter = strings.ReplaceAll(p.config.GroupsFilter, ldapPlaceholderInput, ldapEscape(inputUsername))
	}

	if profile != nil {
		if p.groupsFilterReplacementUsername {
			filter = strings.ReplaceAll(filter, ldapPlaceholderUsername, ldap.EscapeFilter(profile.Username))
		}

		if p.groupsFilterReplacementDN {
			filter = strings.ReplaceAll(filter, ldapPlaceholderDistinguishedName, ldap.EscapeFilter(profile.DN))
		}
	}

	p.log.Tracef("Computed groups filter is %s", filter)

	return filter, nil
}

func (p *LDAPUserProvider) pwdModify(conn LDAPConnection, mr *ldap.PasswordModifyRequest) (err error) {
	var result *ldap.PasswordModifyResult

	if result, err = conn.PasswordModify(mr); err != nil {
		lerr, ok := err.(*ldap.Error)
		if !ok || lerr.ResultCode != ldap.LDAPResultReferral || !p.config.PermitReferrals {
			return err
		}

		p.log.Debugf("Attempting PwdModify ExOp (1.3.6.1.4.1.4203.1.11.1) on referred URL %s", result.Referral)

		connReferral, errReferral := p.connectCustom(result.Referral, p.config.User, p.config.Password, p.dialOpts...)
		if errReferral != nil {
			p.log.Errorf("Failed to connect during password modify request (referred to %s): %v", result.Referral, errReferral)

			return err
		}

		_, err = connReferral.PasswordModify(mr)

		connReferral.Close()
	}

	return err
}

func ldapEscape(inputUsername string) string {
	inputUsername = ldap.EscapeFilter(inputUsername)
	for _, c := range specialLDAPRunes {
		inputUsername = strings.ReplaceAll(inputUsername, string(c), fmt.Sprintf("\\%c", c))
	}

	return inputUsername
}
