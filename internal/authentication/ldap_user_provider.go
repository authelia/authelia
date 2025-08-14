package authentication

import (
	"crypto/x509"
	"errors"
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

	Management UserManagementProvider

	clock clock.Provider

	disableResetPassword bool

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

	switch config.Implementation {
	case "activedirectory":
		provider.Management = &ActiveDirectoryUserManagement{
			provider: provider,
		}
	case "rfc2307bis":
	default:
		provider.Management = &RFC2307bisUserManagement{
			provider: provider,
		}
	}

	return provider
}

func (p *LDAPUserProvider) UpdateUser(username string, userData *UserDetailsExtended) error {
	return p.Management.UpdateUser(username, userData)
}

func (p *LDAPUserProvider) AddUser(userData *UserDetailsExtended) error {
	return p.Management.AddUser(userData)
}

func (p *LDAPUserProvider) DeleteUser(username string) error {
	return p.Management.DeleteUser(username)
}

func (p *LDAPUserProvider) GetSupportedFields() []string {
	return p.Management.GetSupportedFields()
}

func (p *LDAPUserProvider) GetRequiredFields() []string {
	return p.Management.GetRequiredFields()
}
func (p *LDAPUserProvider) GetFieldMetadata() map[string]FieldMetadata {
	return p.Management.GetFieldMetadata()
}

func (p *LDAPUserProvider) ValidateUserData(userData *UserDetailsExtended) error {
	return p.Management.ValidateUserData(userData)
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(username string, password string) (valid bool, err error) {
	var (
		client, uclient LDAPExtendedClient
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

// GetDetails retrieve the users basic information.
func (p *LDAPUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var (
		client  LDAPExtendedClient
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

func (p *LDAPUserProvider) GetUser(username string) (details *UserDetailsExtended, err error) {
	return p.GetDetailsExtended(username)
}

// GetDetailsExtended retrieves the UserDetailsExtended values.
func (p *LDAPUserProvider) GetDetailsExtended(username string) (details *UserDetailsExtended, err error) {
	var (
		client  LDAPExtendedClient
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

func (p *LDAPUserProvider) ListUsers() (users []UserDetailsExtended, err error) {
	var client ldap.Client

	if client, err = p.factory.GetClient(); err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}

	defer func() {
		if err := p.factory.ReleaseClient(client); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	request := ldap.NewSearchRequest(
		p.usersBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=inetOrgPerson)",
		p.usersAttributesExtended,
		nil,
	)

	p.log.
		WithField("base_dn", request.BaseDN).
		WithField("filter", request.Filter).
		WithField("attr", request.Attributes).
		WithField("scope", request.Scope).
		WithField("deref", request.DerefAliases).
		Trace("Performing search for all users (extended)")

	var result *ldap.SearchResult

	if result, err = p.search(client, request); err != nil {
		return nil, fmt.Errorf("failed to search for users: %w", err)
	}

	users = make([]UserDetailsExtended, 0, len(result.Entries))

	for _, entry := range result.Entries {
		profile, err := p.getUserProfileResultToProfileExtended(entry.GetAttributeValue(p.config.Attributes.Username), entry)
		if err != nil {
			p.log.WithError(err).Warnf("Failed to process user entry: %s", entry.DN)
			continue
		}

		groups, err := p.getUserGroups(client, profile.Username, profile.ldapUserProfile)
		if err != nil {
			p.log.WithError(err).Warnf("Failed to get groups for user: %s", profile.Username)
		}

		// Build UserDetailsExtended similar to GetDetailsExtended method.
		userDetails := &UserDetailsExtended{
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
			if uri, err = parseAttributeURI(profile.Username, "profile", p.config.Attributes.Profile, profile.Profile); err != nil {
				p.log.WithError(err).Warnf("Failed to parse profile URL for user: %s", profile.Username)
			} else {
				userDetails.Profile = uri
			}
		}

		if profile.Picture != "" {
			if uri, err = parseAttributeURI(profile.Username, "picture", p.config.Attributes.Picture, profile.Picture); err != nil {
				p.log.WithError(err).Warnf("Failed to parse picture URL for user: %s", profile.Username)
			} else {
				userDetails.Picture = uri
			}
		}

		if profile.Website != "" {
			if uri, err = parseAttributeURI(profile.Username, "website", p.config.Attributes.Website, profile.Website); err != nil {
				p.log.WithError(err).Warnf("Failed to parse website URL for user: %s", profile.Username)
			} else {
				userDetails.Website = uri
			}
		}

		if profile.Locale != "" {
			if locale, err = language.Parse(profile.Locale); err != nil {
				p.log.WithError(err).Warnf("Failed to parse locale for user '%s': %s", profile.Username, profile.Locale)
			} else {
				userDetails.Locale = &locale
			}
		}

		users = append(users, *userDetails)
	}

	return users, nil
}

// UpdatePassword update the password of the given user.
func (p *LDAPUserProvider) UpdatePassword(username, password string) (err error) {
	var (
		client  LDAPExtendedClient
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

	if err = p.setPassword(client, profile, username, "", password); err != nil {
		return fmt.Errorf("unable to update password. Cause: %w", err)
	}

	return nil
}

// ChangePassword is used to change a user's password but requires their old password to be successfully verified.
func (p *LDAPUserProvider) ChangePassword(username, oldPassword string, newPassword string) (err error) {
	var (
		client  LDAPExtendedClient
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

	if err = p.setPassword(client, profile, username, oldPassword, newPassword); err != nil {
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

// getGroupDN is a helper function to get the DN of a group given its name.
// TODO: Use this method :)
//
//nolint:unused
func (p *LDAPUserProvider) getGroupDN(client ldap.Client, groupName string) (string, error) {
	searchRequest := ldap.NewSearchRequest(
		p.groupsBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(&(objectClass=group)(%s=%s))", p.config.Attributes.GroupName, ldap.EscapeFilter(groupName)),
		[]string{"dn"},
		nil,
	)

	result, err := p.search(client, searchRequest)
	if err != nil {
		return "", fmt.Errorf("error searching for group '%s': %w", groupName, err)
	}

	if len(result.Entries) == 0 {
		return "", fmt.Errorf("group '%s' not found", groupName)
	}

	return result.Entries[0].DN, nil
}

func (p *LDAPUserProvider) setPassword(client LDAPExtendedClient, profile *ldapUserProfile, username, oldPassword, newPassword string) (err error) {
	var controls []ldap.Control

	switch {
	case client.Features().ControlTypes.MsftPwdPolHints:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints})
	case client.Features().ControlTypes.MsftPwdPolHintsDeprecated:
		controls = append(controls, &controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHintsDeprecated})
	}

	switch {
	case p.config.Implementation == schema.LDAPImplementationActiveDirectory:
		var value string

		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)

		if value, err = encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", newPassword)); err != nil {
			return fmt.Errorf("error occurred encoding new password for user '%s': %w", username, err)
		}

		modifyRequest.Replace(ldapAttributeUnicodePwd, []string{value})

		return p.modify(client, modifyRequest)
	case client.Features().Extensions.PwdModify:
		pwdModifyRequest := ldap.NewPasswordModifyRequest(
			profile.DN,
			oldPassword,
			newPassword,
		)

		return p.pwdModify(client, pwdModifyRequest)
	default:
		modifyRequest := ldap.NewModifyRequest(profile.DN, controls)
		modifyRequest.Replace(ldapAttributeUserPassword, []string{newPassword})

		return p.modify(client, modifyRequest)
	}
}

func (p *LDAPUserProvider) search(client LDAPExtendedClient, request *ldap.SearchRequest) (result *ldap.SearchResult, err error) {
	if result, err = client.Search(request); err != nil {
		var e *ldap.Error

		if !errors.As(err, &e) {
			return nil, err
		}

		switch e.ResultCode {
		case ldap.LDAPResultReferral:
			if !p.config.PermitReferrals {
				return nil, err
			}
		default:
			return nil, err
		}
	}

	if !p.config.PermitReferrals || result == nil || len(result.Referrals) == 0 {
		return result, nil
	}

	if err = p.searchReferrals(request, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *LDAPUserProvider) searchReferral(referral string, request *ldap.SearchRequest, searchResult *ldap.SearchResult) (err error) {
	var (
		client LDAPExtendedClient
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

func (p *LDAPUserProvider) getUserProfile(client LDAPExtendedClient, username string) (profile *ldapUserProfile, err error) {
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

func (p *LDAPUserProvider) getUserProfileExtended(client LDAPExtendedClient, username string) (profile *ldapUserProfileExtended, err error) {
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

func (p *LDAPUserProvider) getUserGroups(client LDAPExtendedClient, username string, profile *ldapUserProfile) (groups []string, err error) {
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

func (p *LDAPUserProvider) getUserGroupsRequestFilter(client LDAPExtendedClient, username string, _ *ldapUserProfile, request *ldap.SearchRequest) (groups []string, err error) {
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

func (p *LDAPUserProvider) getUserGroupsRequestMemberOf(client LDAPExtendedClient, username string, profile *ldapUserProfile, request *ldap.SearchRequest) (groups []string, err error) {
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

type resolveUsersFilterOpts struct {
	escape bool
}

func (p *LDAPUserProvider) resolveUsersFilter(input string, opts ...func(options *resolveUsersFilterOpts)) (filter string) {
	options := &resolveUsersFilterOpts{}

	for _, opt := range opts {
		opt(options)
	}

	filter = p.config.UsersFilter

	if p.usersFilterReplacementInput {
		if options.escape {
			// The {input} placeholder is replaced by the username input.
			filter = strings.ReplaceAll(filter, ldapPlaceholderInput, ldapEscape(input))
		} else {
			filter = strings.ReplaceAll(filter, ldapPlaceholderInput, input)
		}
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

func (p *LDAPUserProvider) modify(client LDAPExtendedClient, modifyRequest *ldap.ModifyRequest) (err error) {
	var result *ldap.ModifyResult
	if result, err = client.ModifyWithResult(modifyRequest); err != nil {
		var e *ldap.Error

		if !errors.As(err, &e) {
			return err
		}

		switch e.ResultCode {
		case ldap.LDAPResultReferral:
			if !p.config.PermitReferrals {
				return err
			}
		default:
			return err
		}
	}

	if !p.config.PermitReferrals || result == nil || len(result.Referral) == 0 {
		return nil
	}

	p.log.Debugf("Attempting Modify on referred URL %s", result.Referral)

	var (
		clientRef LDAPExtendedClient
		errRef    error
	)
	if clientRef, errRef = p.factory.GetClient(WithAddress(result.Referral)); errRef != nil {
		return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %+v. Original Error: %w", result.Referral, errRef, err)
	}

	defer func() {
		if err := p.factory.ReleaseClient(clientRef); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	if errRef = clientRef.Modify(modifyRequest); errRef != nil {
		return fmt.Errorf("error occurred performing modify on referred LDAP server '%s': %+v. Original Error: %w", result.Referral, errRef, err)
	}

	return nil
}

func (p *LDAPUserProvider) pwdModify(client LDAPExtendedClient, pwdModifyRequest *ldap.PasswordModifyRequest) (err error) {
	var result *ldap.PasswordModifyResult
	if result, err = client.PasswordModify(pwdModifyRequest); err != nil {
		var e *ldap.Error

		if !errors.As(err, &e) {
			return err
		}

		switch e.ResultCode {
		case ldap.LDAPResultReferral:
			if !p.config.PermitReferrals {
				return err
			}
		default:
			return err
		}
	}

	if !p.config.PermitReferrals || result == nil || len(result.Referral) == 0 {
		return nil
	}

	p.log.Debugf("Attempting PwdModify ExOp (1.3.6.1.4.1.4203.1.11.1) on referred URL %s", result.Referral)

	var (
		clientRef LDAPExtendedClient
		errRef    error
	)
	if clientRef, errRef = p.factory.GetClient(WithAddress(result.Referral)); errRef != nil {
		return fmt.Errorf("error occurred connecting to referred LDAP server '%s': %+v. Original Error: %w", result.Referral, errRef, err)
	}

	defer func() {
		if err := p.factory.ReleaseClient(clientRef); err != nil {
			p.log.WithError(err).Warn("Error occurred releasing the LDAP client")
		}
	}()

	if _, errRef = clientRef.PasswordModify(pwdModifyRequest); errRef != nil {
		return fmt.Errorf("error occurred performing password modify on referred LDAP server '%s': %+v. Original Error: %w", result.Referral, errRef, err)
	}

	return nil
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
