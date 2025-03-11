package authentication

import (
	"crypto/x509"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// LDAPUserProvider is a UserProvider that connects to LDAP servers like ActiveDirectory, OpenLDAP, OpenDJ, FreeIPA, etc.
type LDAPUserProvider struct {
	config  *schema.AuthenticationBackendLDAP
	log     *logrus.Logger
	factory LDAPClientFactory

	clock clock.Provider

	disableResetPassword bool

	// Automatically detected LDAP features.
	features LDAPSupportedFeatures

	// Dynamically generated users values.
	usersBaseDN                                        string
	usersAttributes                                    []string
	usersAttributesExtended                            []string
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

// NewLDAPUserProvider creates a new instance of LDAPUserProvider with the StandardLDAPClientFactory.
func NewLDAPUserProvider(config schema.AuthenticationBackend, certs *x509.CertPool) (provider *LDAPUserProvider) {
	if config.LDAP.TLS == nil {
		config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.TLS
	}

	var factory LDAPClientFactory

	if config.LDAP.Pooling.Enable {
		factory = NewPooledLDAPClientFactory(config.LDAP, certs, nil)
	} else {
		factory = NewStandardLDAPClientFactory(config.LDAP, certs, nil)
	}

	return NewLDAPUserProviderWithFactory(config.LDAP, config.PasswordReset.Disable, factory)
}

// NewLDAPUserProviderWithFactory creates a new instance of LDAPUserProvider with the specified LDAPClientFactory.
func NewLDAPUserProviderWithFactory(config *schema.AuthenticationBackendLDAP, disableResetPassword bool, factory LDAPClientFactory) (provider *LDAPUserProvider) {
	provider = &LDAPUserProvider{
		config:               config,
		log:                  logging.Logger(),
		factory:              factory,
		disableResetPassword: disableResetPassword,
		clock:                clock.New(),
	}

	provider.parseDynamicUsersConfiguration()
	provider.parseDynamicGroupsConfiguration()

	return provider
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(username string, password string) (valid bool, err error) {
	var (
		client, uclient ldap.Client
		profile         *ldapUserProfile
	)

	if client, err = p.factory.GetClient(); err != nil {
		return false, err
	}

	defer func() {
		if err := p.factory.ReleaseClient(client); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	if profile, err = p.getUserProfile(client, username); err != nil {
		return false, err
	}

	if uclient, err = p.factory.GetClient(WithUsername(profile.DN), WithPassword(password)); err != nil {
		return false, fmt.Errorf("authentication failed. Cause: %w", err)
	}

	defer func() {
		if err := p.factory.ReleaseClient(uclient); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	return true, nil
}

// GetDetails retrieve the groups a user belongs to.
func (p *LDAPUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var (
		client  ldap.Client
		profile *ldapUserProfile
	)

	if client, err = p.factory.GetClient(); err != nil {
		return nil, err
	}

	defer func() {
		if err := p.factory.ReleaseClient(client); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

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

// GetDetailsExtended retrieves the UserDetailsExtended values.
func (p *LDAPUserProvider) GetDetailsExtended(username string) (details *UserDetailsExtended, err error) {
	var (
		client  ldap.Client
		profile *ldapUserProfileExtended
	)

	if client, err = p.factory.GetClient(); err != nil {
		return nil, err
	}

	defer func() {
		if err := p.factory.ReleaseClient(client); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	if profile, err = p.getUserProfileExtended(client, username); err != nil {
		return nil, err
	}

	var (
		groups []string
	)

	if groups, err = p.getUserGroups(client, username, profile.ldapUserProfile); err != nil {
		return nil, err
	}

	details = &UserDetailsExtended{
		GivenName:      profile.GivenName,
		FamilyName:     profile.FamilyName,
		MiddleName:     profile.MiddleName,
		Nickname:       profile.Nickname,
		Gender:         profile.Gender,
		Birthdate:      profile.Birthdate,
		ZoneInfo:       profile.ZoneInfo,
		PhoneNumber:    profile.PhoneNumber,
		PhoneExtension: profile.PhoneExtension,
		Address:        profile.Address,
		UserDetails: &UserDetails{
			Username:    profile.Username,
			DisplayName: profile.DisplayName,
			Emails:      profile.Emails,
			Groups:      groups,
		},
		Extra: profile.Extra,
	}

	var (
		uri    *url.URL
		locale language.Tag
	)

	if profile.Profile != "" {
		if uri, err = parseAttributeURI(username, "profile", p.config.Attributes.Profile, profile.Profile); err != nil {
			return nil, err
		} else {
			details.Profile = uri
		}
	}

	if profile.Picture != "" {
		if uri, err = parseAttributeURI(username, "picture", p.config.Attributes.Picture, profile.Picture); err != nil {
			return nil, err
		} else {
			details.Picture = uri
		}
	}

	if profile.Website != "" {
		if uri, err = parseAttributeURI(username, "website", p.config.Attributes.Website, profile.Website); err != nil {
			return nil, err
		} else {
			details.Website = uri
		}
	}

	if profile.Locale != "" {
		if locale, err = language.Parse(profile.Locale); err != nil {
			return nil, fmt.Errorf("error occurred parsing user details for '%s': failed to parse the locale attribute '%s' with value '%s': %w", username, p.config.Attributes.Locale, profile.Locale, err)
		} else {
			details.Locale = &locale
		}
	}

	return details, nil
}

// UpdatePassword update the password of the given user.
func (p *LDAPUserProvider) UpdatePassword(username, password string) (err error) {
	var (
		client  ldap.Client
		profile *ldapUserProfile
	)

	if client, err = p.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	defer func() {
		if err := p.factory.ReleaseClient(client); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

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

// ChangePassword is used to change a user's password but requires their old password to be successfully verified.
//
//nolint:gocyclo
func (p *LDAPUserProvider) ChangePassword(username, oldPassword string, newPassword string) (err error) {
	var (
		client  ldap.Client
		profile *ldapUserProfile
	)

	if client, err = p.factory.GetClient(); err != nil {
		return fmt.Errorf("unable to update password for user '%s'. Cause: %w", username, err)
	}

	defer func() {
		if err := p.factory.ReleaseClient(client); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	if profile, err = p.getUserProfile(client, username); err != nil {
		return fmt.Errorf("unable to update password for user '%s'. Cause: %w", username, err)
	}

	var controls []ldap.Control

	switch {
	case p.features.ControlTypes.MsftPwdPolHints:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints})
	case p.features.ControlTypes.MsftPwdPolHintsDeprecated:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHintsDeprecated})
	}

	userPasswordOk, err := p.CheckUserPassword(username, oldPassword)

	if err != nil {
		errorCode := getLDAPResultCode(err)
		if errorCode == ldap.LDAPResultInvalidCredentials {
			return ErrIncorrectPassword
		} else {
			return err
		}
	}

	if !userPasswordOk {
		return ErrIncorrectPassword
	}

	if oldPassword == newPassword {
		return ErrPasswordWeak
	}

	switch {
	case p.features.Extensions.PwdModifyExOp:
		pwdModifyRequest := ldap.NewPasswordModifyRequest(
			profile.DN,
			oldPassword,
			newPassword,
		)

		err = p.pwdModify(client, pwdModifyRequest)
	case p.config.Implementation == schema.LDAPImplementationActiveDirectory:
		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)
		// The password needs to be enclosed in quotes
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2
		pwdEncoded, err := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", newPassword))
		if err != nil {
			return fmt.Errorf("failed to encode new password for user '%s'. Cause: %w", username, err)
		}

		modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

		//nolint
		err = p.modify(client, modifyRequest)
	default:
		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)
		modifyRequest.Replace(ldapAttributeUserPassword, []string{newPassword})

		err = p.modify(client, modifyRequest)
	}

	//TODO: Better inform users regarding password reuse/password history.
	if err != nil {
		if errorCode := getLDAPResultCode(err); errorCode != -1 {
			switch errorCode {
			case ldap.LDAPResultInvalidCredentials,
				ldap.LDAPResultInappropriateAuthentication:
				return fmt.Errorf("%w: %v", ErrIncorrectPassword, err)
			case ldap.LDAPResultConstraintViolation,
				ldap.LDAPResultObjectClassViolation,
				ldap.ErrorEmptyPassword,
				ldap.LDAPResultUnwillingToPerform:
				return fmt.Errorf("%w: %v", ErrPasswordWeak, err)
			default:
				return fmt.Errorf("%w: %v", ErrOperationFailed, err)
			}
		}

		return fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return nil
}

func (p *LDAPUserProvider) search(client ldap.Client, request *ldap.SearchRequest) (result *ldap.SearchResult, err error) {
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
		return result, nil
	}

	if err = p.searchReferrals(request, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *LDAPUserProvider) searchReferral(referral string, request *ldap.SearchRequest, searchResult *ldap.SearchResult) (err error) {
	var (
		client ldap.Client
		result *ldap.SearchResult
	)

	if client, err = p.factory.GetClient(WithAddress(referral)); err != nil {
		return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %w", referral, err)
	}

	defer func() {
		if err := p.factory.ReleaseClient(client); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

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

func (p *LDAPUserProvider) getUserProfile(client ldap.Client, username string) (profile *ldapUserProfile, err error) {
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

	return p.getUserProfileResultToProfile(username, result.Entries[0])
}

func (p *LDAPUserProvider) getUserProfileResultToProfile(username string, entry *ldap.Entry) (profile *ldapUserProfile, err error) {
	userProfile := ldapUserProfile{
		DN:          entry.DN,
		Emails:      getValuesFromEntry(entry, p.config.Attributes.Mail),
		DisplayName: getValueFromEntry(entry, p.config.Attributes.DisplayName),
		MemberOf:    getValuesFromEntry(entry, p.config.Attributes.MemberOf),
	}

	attrUsername := getValuesFromEntry(entry, p.config.Attributes.Username)

	switch n := len(attrUsername); n {
	case 1:
		userProfile.Username = attrUsername[0]

		if p.config.Attributes.Username == p.config.Attributes.DisplayName && userProfile.DisplayName == "" {
			userProfile.DisplayName = attrUsername[0]
		}

		if p.config.Attributes.Username == p.config.Attributes.Mail && len(userProfile.Emails) == 0 {
			userProfile.Emails = []string{attrUsername[0]}
		}
	case 0:
		return nil, fmt.Errorf("user '%s' must have value for attribute '%s'",
			username, p.config.Attributes.Username)
	default:
		return nil, fmt.Errorf("user '%s' has %d values for for attribute '%s' but the attribute must be a single value attribute",
			username, n, p.config.Attributes.Username)
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

func (p *LDAPUserProvider) getUserProfileExtended(client ldap.Client, username string) (profile *ldapUserProfileExtended, err error) {
	// Search for the given username.
	request := ldap.NewSearchRequest(
		p.usersBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, p.resolveUsersFilter(username), p.usersAttributesExtended, nil,
	)

	p.log.
		WithField("base_dn", request.BaseDN).
		WithField("filter", request.Filter).
		WithField("attr", request.Attributes).
		WithField("scope", request.Scope).
		WithField("deref", request.DerefAliases).
		Trace("Performing extended user search")

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

	return p.getUserProfileResultToProfileExtended(username, result.Entries[0])
}

func (p *LDAPUserProvider) getUserProfileResultToProfileExtended(username string, entry *ldap.Entry) (profile *ldapUserProfileExtended, err error) {
	base, err := p.getUserProfileResultToProfile(username, entry)
	if err != nil {
		return nil, err
	}

	userProfile := ldapUserProfileExtended{
		GivenName:       getValueFromEntry(entry, p.config.Attributes.GivenName),
		FamilyName:      getValueFromEntry(entry, p.config.Attributes.FamilyName),
		MiddleName:      getValueFromEntry(entry, p.config.Attributes.MiddleName),
		Nickname:        getValueFromEntry(entry, p.config.Attributes.Nickname),
		Profile:         getValueFromEntry(entry, p.config.Attributes.Profile),
		Picture:         getValueFromEntry(entry, p.config.Attributes.Picture),
		Website:         getValueFromEntry(entry, p.config.Attributes.Website),
		Gender:          getValueFromEntry(entry, p.config.Attributes.Gender),
		Birthdate:       getValueFromEntry(entry, p.config.Attributes.Birthdate),
		ZoneInfo:        getValueFromEntry(entry, p.config.Attributes.ZoneInfo),
		Locale:          getValueFromEntry(entry, p.config.Attributes.Locale),
		PhoneNumber:     getValueFromEntry(entry, p.config.Attributes.PhoneNumber),
		PhoneExtension:  getValueFromEntry(entry, p.config.Attributes.PhoneExtension),
		Extra:           map[string]any{},
		ldapUserProfile: base,
	}

	street, locality, region, postcode, country := getValueFromEntry(entry, p.config.Attributes.StreetAddress), getValueFromEntry(entry, p.config.Attributes.Locality), getValueFromEntry(entry, p.config.Attributes.Region), getValueFromEntry(entry, p.config.Attributes.PostalCode), getValueFromEntry(entry, p.config.Attributes.Country)

	if street != "" || locality != "" || region != "" || postcode != "" || country != "" {
		userProfile.Address = &UserDetailsAddress{
			StreetAddress: street,
			Locality:      locality,
			Region:        region,
			PostalCode:    postcode,
			Country:       country,
		}
	}

	var attr any

	for attribute, properties := range p.config.Attributes.Extra {
		if attr, err = getExtraValueFromEntry(entry, attribute, properties); err != nil {
			return nil, err
		}

		if attr == nil {
			continue
		}

		if len(properties.Name) == 0 {
			userProfile.Extra[attribute] = attr
		} else {
			userProfile.Extra[properties.Name] = attr
		}
	}

	return &userProfile, nil
}

func (p *LDAPUserProvider) getUserGroups(client ldap.Client, username string, profile *ldapUserProfile) (groups []string, err error) {
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

func (p *LDAPUserProvider) getUserGroupsRequestFilter(client ldap.Client, username string, _ *ldapUserProfile, request *ldap.SearchRequest) (groups []string, err error) {
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

func (p *LDAPUserProvider) getUserGroupsRequestMemberOf(client ldap.Client, username string, profile *ldapUserProfile, request *ldap.SearchRequest) (groups []string, err error) {
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
		switch strings.ToLower(attr.Name) {
		case strings.ToLower(p.config.Attributes.GroupName):
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
		filter = strings.ReplaceAll(filter, ldapPlaceholderDateTimeMicrosoftNTTimeEpoch, strconv.FormatUint(utils.UnixNanoTimeToMicrosoftNTEpoch(p.clock.Now().UnixNano()), 10))
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

func (p *LDAPUserProvider) modify(client ldap.Client, modifyRequest *ldap.ModifyRequest) (err error) {
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
			clientRef ldap.Client
			errRef    error
		)

		if clientRef, errRef = p.factory.GetClient(WithAddress(referral)); errRef != nil {
			return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %+v. Original Error: %w", referral, errRef, err)
		}

		defer func() {
			if err := p.factory.ReleaseClient(clientRef); err != nil {
				p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
			}
		}()

		if errRef = clientRef.Modify(modifyRequest); errRef != nil {
			return fmt.Errorf("error occurred performing modify on referred LDAP server '%s': %+v. Original Error: %w", referral, errRef, err)
		}

		return nil
	}

	return nil
}

func (p *LDAPUserProvider) pwdModify(client ldap.Client, pwdModifyRequest *ldap.PasswordModifyRequest) (err error) {
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
			clientRef ldap.Client
			errRef    error
		)

		if clientRef, errRef = p.factory.GetClient(WithAddress(referral)); errRef != nil {
			return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %+v. Original Error: %w", referral, errRef, err)
		}

		defer func() {
			if err := p.factory.ReleaseClient(clientRef); err != nil {
				p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
			}
		}()

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

func parseAttributeURI(username, attributeName, attribute, value string) (uri *url.URL, err error) {
	if uri, err = url.ParseRequestURI(value); err == nil {
		if uri.Scheme != "http" && uri.Scheme != "https" {
			err = fmt.Errorf("invalid URL scheme '%s' for profile attribute", uri.Scheme)
		}
	}

	if err != nil {
		if attributeName == "" {
			return nil, fmt.Errorf("error occurred parsing user details for '%s': failed to parse the %s attribute with value '%s': %w", username, attribute, value, err)
		}

		return nil, fmt.Errorf("error occurred parsing user details for '%s': failed to parse the %s attribute '%s' with value '%s': %w", username, attributeName, attribute, value, err)
	}

	return uri, nil
}
