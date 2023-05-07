package authentication

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strconv"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// LDAPUserProvider is a UserProvider that connects to LDAP servers like ActiveDirectory, OpenLDAP, OpenDJ, FreeIPA, etc.
type LDAPUserProvider struct {
	config    schema.LDAPAuthenticationBackend
	tlsConfig *tls.Config
	dialOpts  []ldap.DialOpt
	log       *logrus.Logger
	factory   LDAPClientFactory

	clock utils.Clock

	disableResetPassword bool

	// Automatically detected LDAP features.
	features LDAPSupportedFeatures

	// Dynamically generated users values.
	usersBaseDN                                        string
	usersAttributes                                    []string
	usersFilterReplacementInput                        bool
	usersFilterReplacementDateTimeGeneralized          bool
	usersFilterReplacementDateTimeUnixEpoch            bool
	usersFilterReplacementDateTimeMicrosoftNTTimeEpoch bool

	// Dynamically generated groups values.
	groupsBaseDN                        string
	groupsAttributes                    []string
	groupsFilterReplacementInput        bool
	groupsFilterReplacementUsername     bool
	groupsFilterReplacementDN           bool
	groupsFilterReplacementsMemberOfDN  bool
	groupsFilterReplacementsMemberOfRDN bool
}

// NewLDAPUserProvider creates a new instance of LDAPUserProvider with the ProductionLDAPClientFactory.
func NewLDAPUserProvider(config schema.AuthenticationBackend, certPool *x509.CertPool) (provider *LDAPUserProvider) {
	provider = NewLDAPUserProviderWithFactory(*config.LDAP, config.PasswordReset.Disable, certPool, NewProductionLDAPClientFactory())

	return provider
}

// NewLDAPUserProviderWithFactory creates a new instance of LDAPUserProvider with the specified LDAPClientFactory.
func NewLDAPUserProviderWithFactory(config schema.LDAPAuthenticationBackend, disableResetPassword bool, certPool *x509.CertPool, factory LDAPClientFactory) (provider *LDAPUserProvider) {
	if config.TLS == nil {
		config.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.TLS
	}

	tlsConfig := utils.NewTLSConfig(config.TLS, certPool)

	var dialOpts = []ldap.DialOpt{
		ldap.DialWithDialer(&net.Dialer{Timeout: config.Timeout}),
	}

	if tlsConfig != nil {
		dialOpts = append(dialOpts, ldap.DialWithTLSConfig(tlsConfig))
	}

	if factory == nil {
		factory = NewProductionLDAPClientFactory()
	}

	provider = &LDAPUserProvider{
		config:               config,
		tlsConfig:            tlsConfig,
		dialOpts:             dialOpts,
		log:                  logging.Logger(),
		factory:              factory,
		disableResetPassword: disableResetPassword,
		clock:                &utils.RealClock{},
	}

	provider.parseDynamicUsersConfiguration()
	provider.parseDynamicGroupsConfiguration()
	provider.parseDynamicConfiguration()

	return provider
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(username string, password string) (valid bool, err error) {
	var (
		client, clientUser LDAPClient
		profile            *ldapUserProfile
	)

	if client, err = p.connect(); err != nil {
		return false, err
	}

	defer client.Close()

	if profile, err = p.getUserProfile(client, username); err != nil {
		return false, err
	}

	if clientUser, err = p.connectCustom(p.config.Address.String(), profile.DN, password, p.config.StartTLS, p.dialOpts...); err != nil {
		return false, fmt.Errorf("authentication failed. Cause: %w", err)
	}

	defer clientUser.Close()

	return true, nil
}

// GetDetails retrieve the groups a user belongs to.
func (p *LDAPUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var (
		client  LDAPClient
		profile *ldapUserProfile
	)

	if client, err = p.connect(); err != nil {
		return nil, err
	}

	defer client.Close()

	if profile, err = p.getUserProfile(client, username); err != nil {
		return nil, err
	}

	var (
		groups []string
	)

	if groups, err = p.getUserGroups(client, username, profile); err != nil {
		return nil, err
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
		client  LDAPClient
		profile *ldapUserProfile
	)

	if client, err = p.connect(); err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	defer client.Close()

	if profile, err = p.getUserProfile(client, username); err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	var controls []ldap.Control

	switch {
	case p.features.ControlTypes.MsftPwdPolHints:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints})
	case p.features.ControlTypes.MsftPwdPolHintsDeprecated:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHintsDeprecated})
	}

	switch {
	case p.features.Extensions.PwdModifyExOp:
		pwdModifyRequest := ldap.NewPasswordModifyRequest(
			profile.DN,
			"",
			password,
		)

		err = p.pwdModify(client, pwdModifyRequest)
	case p.config.Implementation == schema.LDAPImplementationActiveDirectory:
		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)
		// The password needs to be enclosed in quotes
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2
		pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", password))
		modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

		err = p.modify(client, modifyRequest)
	default:
		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)
		modifyRequest.Replace(ldapAttributeUserPassword, []string{password})

		err = p.modify(client, modifyRequest)
	}

	if err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	return nil
}

func (p *LDAPUserProvider) connect() (client LDAPClient, err error) {
	return p.connectCustom(p.config.Address.String(), p.config.User, p.config.Password, p.config.StartTLS, p.dialOpts...)
}

func (p *LDAPUserProvider) connectCustom(url, username, password string, startTLS bool, opts ...ldap.DialOpt) (client LDAPClient, err error) {
	if client, err = p.factory.DialURL(url, opts...); err != nil {
		return nil, fmt.Errorf("dial failed with error: %w", err)
	}

	if startTLS {
		if err = client.StartTLS(p.tlsConfig); err != nil {
			client.Close()

			return nil, fmt.Errorf("starttls failed with error: %w", err)
		}
	}

	if password == "" {
		err = client.UnauthenticatedBind(username)
	} else {
		err = client.Bind(username, password)
	}

	if err != nil {
		client.Close()

		return nil, fmt.Errorf("bind failed with error: %w", err)
	}

	return client, nil
}

func (p *LDAPUserProvider) search(client LDAPClient, request *ldap.SearchRequest) (result *ldap.SearchResult, err error) {
	if result, err = client.Search(request); err != nil {
		if referral, ok := p.getReferral(err); ok {
			if result == nil {
				result = &ldap.SearchResult{
					Referrals: []string{referral},
				}
			} else {
				result.Referrals = append(result.Referrals, referral)
			}
		} else {
			return nil, err
		}
	}

	if !p.config.PermitReferrals || len(result.Referrals) == 0 {
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	if err = p.searchReferrals(request, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *LDAPUserProvider) searchReferral(referral string, request *ldap.SearchRequest, searchResult *ldap.SearchResult) (err error) {
	var (
		client LDAPClient
		result *ldap.SearchResult
	)

	if client, err = p.connectCustom(referral, p.config.User, p.config.Password, p.config.StartTLS, p.dialOpts...); err != nil {
		return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %w", referral, err)
	}

	defer client.Close()

	if result, err = client.Search(request); err != nil {
		return fmt.Errorf("error occurred performing search on referred LDAP server '%s': %w", referral, err)
	}

	for i := 0; i < len(result.Entries); i++ {
		if !ldapEntriesContainsEntry(result.Entries[i], searchResult.Entries) {
			searchResult.Entries = append(searchResult.Entries, result.Entries[i])
		}
	}

	return nil
}

func (p *LDAPUserProvider) searchReferrals(request *ldap.SearchRequest, result *ldap.SearchResult) (err error) {
	for i := 0; i < len(result.Referrals); i++ {
		if err = p.searchReferral(result.Referrals[i], request, result); err != nil {
			return err
		}
	}

	return nil
}

func (p *LDAPUserProvider) getUserProfile(client LDAPClient, username string) (profile *ldapUserProfile, err error) {
	// Search for the given username.
	request := ldap.NewSearchRequest(
		p.usersBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, p.resolveUsersFilter(username), p.usersAttributes, nil,
	)

	p.log.
		WithField("base_dn", request.BaseDN).
		WithField("filter", request.Filter).
		WithField("attr", request.Attributes).
		WithField("scope", request.Scope).
		WithField("deref", request.DerefAliases).
		Trace("Performing user search")

	var result *ldap.SearchResult

	if result, err = p.search(client, request); err != nil {
		return nil, fmt.Errorf("cannot find user DN of user '%s'. Cause: %w", username, err)
	}

	if len(result.Entries) == 0 {
		return nil, ErrUserNotFound
	}

	if len(result.Entries) > 1 {
		return nil, fmt.Errorf("there were %d users found when searching for '%s' but there should only be 1", len(result.Entries), username)
	}

	return p.getUserProfileResultToProfile(username, result)
}

//nolint:gocyclo // Not overly complex.
func (p *LDAPUserProvider) getUserProfileResultToProfile(username string, result *ldap.SearchResult) (profile *ldapUserProfile, err error) {
	userProfile := ldapUserProfile{
		DN: result.Entries[0].DN,
	}

	for _, attr := range result.Entries[0].Attributes {
		attrs := len(attr.Values)

		switch attr.Name {
		case p.config.Attributes.Username:
			switch attrs {
			case 1:
				userProfile.Username = attr.Values[0]

				if attr.Name == p.config.Attributes.DisplayName && userProfile.DisplayName == "" {
					userProfile.DisplayName = attr.Values[0]
				}

				if attr.Name == p.config.Attributes.Mail && len(userProfile.Emails) == 0 {
					userProfile.Emails = []string{attr.Values[0]}
				}
			case 0:
				return nil, fmt.Errorf("user '%s' must have value for attribute '%s'",
					username, p.config.Attributes.Username)
			default:
				return nil, fmt.Errorf("user '%s' has %d values for for attribute '%s' but the attribute must be a single value attribute",
					username, attrs, p.config.Attributes.Username)
			}
		case p.config.Attributes.Mail:
			if attrs == 0 {
				continue
			}

			userProfile.Emails = attr.Values
		case p.config.Attributes.DisplayName:
			if attrs == 0 {
				continue
			}

			userProfile.DisplayName = attr.Values[0]
		case p.config.Attributes.MemberOf:
			if attrs == 0 {
				continue
			}

			userProfile.MemberOf = attr.Values
		}
	}

	if userProfile.Username == "" {
		return nil, fmt.Errorf("user '%s' must have value for attribute '%s'",
			username, p.config.Attributes.Username)
	}

	if userProfile.DN == "" {
		return nil, fmt.Errorf("user '%s' must have a distinguished name but the result returned an empty distinguished name", username)
	}

	return &userProfile, nil
}

func (p *LDAPUserProvider) getUserGroups(client LDAPClient, username string, profile *ldapUserProfile) (groups []string, err error) {
	request := ldap.NewSearchRequest(
		p.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, p.resolveGroupsFilter(username, profile), p.groupsAttributes, nil,
	)

	p.log.
		WithField("base_dn", request.BaseDN).
		WithField("filter", request.Filter).
		WithField("attributes", request.Attributes).
		WithField("scope", request.Scope).
		WithField("deref", request.DerefAliases).
		WithField("mode", p.config.GroupSearchMode).
		Trace("Performing group search")

	switch p.config.GroupSearchMode {
	case "", "filter":
		return p.getUserGroupsRequestFilter(client, username, profile, request)
	case "memberof":
		return p.getUserGroupsRequestMemberOf(client, username, profile, request)
	default:
		return nil, fmt.Errorf("could not perform group search with mode '%s' as it's unknown", p.config.GroupSearchMode)
	}
}

func (p *LDAPUserProvider) getUserGroupsRequestFilter(client LDAPClient, username string, _ *ldapUserProfile, request *ldap.SearchRequest) (groups []string, err error) {
	var result *ldap.SearchResult

	if result, err = p.search(client, request); err != nil {
		return nil, fmt.Errorf("unable to retrieve groups of user '%s'. Cause: %w", username, err)
	}

	for _, entry := range result.Entries {
		if group := p.getUserGroupFromEntry(entry); len(group) != 0 {
			groups = append(groups, group)
		}
	}

	return groups, nil
}

func (p *LDAPUserProvider) getUserGroupsRequestMemberOf(client LDAPClient, username string, profile *ldapUserProfile, request *ldap.SearchRequest) (groups []string, err error) {
	var result *ldap.SearchResult

	if result, err = p.search(client, request); err != nil {
		return nil, fmt.Errorf("unable to retrieve groups of user '%s'. Cause: %w", username, err)
	}

	for _, entry := range result.Entries {
		if len(entry.Attributes) == 0 {
			p.log.
				WithField("dn", entry.DN).
				WithField("attributes", request.Attributes).
				WithField("mode", "memberof").
				Trace("Skipping Group as the server did not return any requested attributes")

			continue
		}

		if !utils.IsStringInSliceFold(entry.DN, profile.MemberOf) {
			p.log.
				WithField("dn", entry.DN).
				WithField("mode", "memberof").
				Trace("Skipping Group as it doesn't match the users memberof entries")

			continue
		}

		if group := p.getUserGroupFromEntry(entry); len(group) != 0 {
			groups = append(groups, group)
		}
	}

	return groups, nil
}

func (p *LDAPUserProvider) getUserGroupFromEntry(entry *ldap.Entry) string {
attributes:
	for _, attr := range entry.Attributes {
		switch attr.Name {
		case p.config.Attributes.GroupName:
			switch len(attr.Values) {
			case 0:
				p.log.
					WithField("dn", entry.DN).
					WithField("attribute", attr.Name).
					Trace("Group skipped as the server returned a null attribute")
			case 1:
				switch len(attr.Values[0]) {
				case 0:
					p.log.
						WithField("dn", entry.DN).
						WithField("attribute", attr.Name).
						Trace("Skipping group as the configured group name attribute had no value")

				default:
					return attr.Values[0]
				}
			default:
				p.log.
					WithField("dn", entry.DN).
					WithField("attribute", attr.Name).
					Trace("Group skipped as the server returned a multi-valued attribute but it should be a single-valued attribute")
			}

			break attributes
		}
	}

	return ""
}

func (p *LDAPUserProvider) resolveUsersFilter(input string) (filter string) {
	filter = p.config.UsersFilter

	if p.usersFilterReplacementInput {
		// The {input} placeholder is replaced by the username input.
		filter = strings.ReplaceAll(filter, ldapPlaceholderInput, ldapEscape(input))
	}

	if p.usersFilterReplacementDateTimeGeneralized {
		filter = strings.ReplaceAll(filter, ldapPlaceholderDateTimeGeneralized, p.clock.Now().UTC().Format(ldapGeneralizedTimeDateTimeFormat))
	}

	if p.usersFilterReplacementDateTimeUnixEpoch {
		filter = strings.ReplaceAll(filter, ldapPlaceholderDateTimeUnixEpoch, strconv.Itoa(int(p.clock.Now().Unix())))
	}

	if p.usersFilterReplacementDateTimeMicrosoftNTTimeEpoch {
		filter = strings.ReplaceAll(filter, ldapPlaceholderDateTimeMicrosoftNTTimeEpoch, strconv.Itoa(int(utils.UnixNanoTimeToMicrosoftNTEpoch(p.clock.Now().UnixNano()))))
	}

	p.log.Tracef("Detected user filter is %s", filter)

	return filter
}

func (p *LDAPUserProvider) resolveGroupsFilter(input string, profile *ldapUserProfile) (filter string) {
	filter = p.config.GroupsFilter

	if p.groupsFilterReplacementInput {
		// The {input} placeholder is replaced by the users username input.
		filter = strings.ReplaceAll(p.config.GroupsFilter, ldapPlaceholderInput, ldapEscape(input))
	}

	if profile != nil {
		if p.groupsFilterReplacementUsername {
			filter = strings.ReplaceAll(filter, ldapPlaceholderUsername, ldap.EscapeFilter(profile.Username))
		}

		if p.groupsFilterReplacementDN {
			filter = strings.ReplaceAll(filter, ldapPlaceholderDistinguishedName, ldap.EscapeFilter(profile.DN))
		}
	}

	if p.groupsFilterReplacementsMemberOfDN {
		sep := fmt.Sprintf(")(%s=", p.config.Attributes.DistinguishedName)
		values := make([]string, len(profile.MemberOf))

		for i, memberof := range profile.MemberOf {
			values[i] = ldap.EscapeFilter(memberof)
		}

		filter = strings.ReplaceAll(filter, ldapPlaceholderMemberOfDistinguishedName, fmt.Sprintf("(%s=%s)", p.config.Attributes.DistinguishedName, strings.Join(values, sep)))
	}

	if p.groupsFilterReplacementsMemberOfRDN {
		values := make([]string, len(profile.MemberOf))

		for i, memberof := range profile.MemberOf {
			values[i] = ldap.EscapeFilter(strings.SplitN(memberof, ",", 2)[0])
		}

		filter = strings.ReplaceAll(filter, ldapPlaceholderMemberOfRelativeDistinguishedName, fmt.Sprintf("(%s)", strings.Join(values, ")(")))
	}

	p.log.Tracef("Computed groups filter is %s", filter)

	return filter
}

func (p *LDAPUserProvider) modify(client LDAPClient, modifyRequest *ldap.ModifyRequest) (err error) {
	if err = client.Modify(modifyRequest); err != nil {
		var (
			referral string
			ok       bool
		)

		if referral, ok = p.getReferral(err); !ok {
			return err
		}

		p.log.Debugf("Attempting Modify on referred URL %s", referral)

		var (
			clientRef LDAPClient
			errRef    error
		)

		if clientRef, errRef = p.connectCustom(referral, p.config.User, p.config.Password, p.config.StartTLS, p.dialOpts...); errRef != nil {
			return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %+v. Original Error: %w", referral, errRef, err)
		}

		defer clientRef.Close()

		if errRef = clientRef.Modify(modifyRequest); errRef != nil {
			return fmt.Errorf("error occurred performing modify on referred LDAP server '%s': %+v. Original Error: %w", referral, errRef, err)
		}

		return nil
	}

	return nil
}

func (p *LDAPUserProvider) pwdModify(client LDAPClient, pwdModifyRequest *ldap.PasswordModifyRequest) (err error) {
	if _, err = client.PasswordModify(pwdModifyRequest); err != nil {
		var (
			referral string
			ok       bool
		)

		if referral, ok = p.getReferral(err); !ok {
			return err
		}

		p.log.Debugf("Attempting PwdModify ExOp (1.3.6.1.4.1.4203.1.11.1) on referred URL %s", referral)

		var (
			clientRef LDAPClient
			errRef    error
		)

		if clientRef, errRef = p.connectCustom(referral, p.config.User, p.config.Password, p.config.StartTLS, p.dialOpts...); errRef != nil {
			return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %+v. Original Error: %w", referral, errRef, err)
		}

		defer clientRef.Close()

		if _, errRef = clientRef.PasswordModify(pwdModifyRequest); errRef != nil {
			return fmt.Errorf("error occurred performing password modify on referred LDAP server '%s': %+v. Original Error: %w", referral, errRef, err)
		}

		return nil
	}

	return nil
}

func (p *LDAPUserProvider) getReferral(err error) (referral string, ok bool) {
	if !p.config.PermitReferrals {
		return "", false
	}

	return ldapGetReferral(err)
}
