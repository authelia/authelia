package authentication

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/unicode"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldCreateRawConnectionWhenSchemeIsLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:  testLDAPAddress,
			User:     "cn=admin,dc=example,dc=com",
			Password: "password",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	gomock.InOrder(dialURL, connBind)

	_, err := provider.connect()

	require.NoError(t, err)
}

func TestShouldCreateTLSConnectionWhenSchemeIsLDAPS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:  testLDAPSAddress,
			User:     "cn=admin,dc=example,dc=com",
			Password: "password",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldaps://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	gomock.InOrder(dialURL, connBind)

	_, err := provider.connect()

	require.NoError(t, err)
}

func TestEscapeSpecialCharsFromUserInput(t *testing.T) {
	// No escape.
	assert.Equal(t, "xyz", ldapEscape("xyz"))

	// Escape.
	assert.Equal(t, "test\\,abc", ldapEscape("test,abc"))
	assert.Equal(t, "test\\5cabc", ldapEscape("test\\abc"))
	assert.Equal(t, "test\\2aabc", ldapEscape("test*abc"))
	assert.Equal(t, "test \\28abc\\29", ldapEscape("test (abc)"))
	assert.Equal(t, "test\\#abc", ldapEscape("test#abc"))
	assert.Equal(t, "test\\+abc", ldapEscape("test+abc"))
	assert.Equal(t, "test\\<abc", ldapEscape("test<abc"))
	assert.Equal(t, "test\\>abc", ldapEscape("test>abc"))
	assert.Equal(t, "test\\;abc", ldapEscape("test;abc"))
	assert.Equal(t, "test\\\"abc", ldapEscape("test\"abc"))
	assert.Equal(t, "test\\=abc", ldapEscape("test=abc"))
	assert.Equal(t, "test\\,\\5c\\28abc\\29", ldapEscape("test,\\(abc)"))
}

func TestEscapeSpecialCharsInGroupsFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:      testLDAPSAddress,
			GroupsFilter: "(|(member={dn})(uid={username})(uid={input}))",
		},
		false,
		nil,
		mockFactory)

	profile := ldapUserProfile{
		DN:          "cn=john (external),dc=example,dc=com",
		Username:    "john",
		DisplayName: "John Doe",
		Emails:      []string{"john.doe@authelia.com"},
	}

	filter := provider.resolveGroupsFilter("john", &profile)
	assert.Equal(t, "(|(member=cn=john \\28external\\29,dc=example,dc=com)(uid=john)(uid=john))", filter)

	filter = provider.resolveGroupsFilter("john#=(abc,def)", &profile)
	assert.Equal(t, "(|(member=cn=john \\28external\\29,dc=example,dc=com)(uid=john)(uid=john\\#\\=\\28abc\\,def\\29))", filter)
}

type ExtendedSearchRequestMatcher struct {
	filter       string
	baseDN       string
	scope        int
	derefAliases int
	typesOnly    bool
	attributes   []string
}

func NewExtendedSearchRequestMatcher(filter, base string, scope, derefAliases int, typesOnly bool, attributes []string) *ExtendedSearchRequestMatcher {
	return &ExtendedSearchRequestMatcher{filter, base, scope, derefAliases, typesOnly, attributes}
}

func (e *ExtendedSearchRequestMatcher) Matches(x any) bool {
	sr := x.(*ldap.SearchRequest)

	if e.filter != sr.Filter || e.baseDN != sr.BaseDN || e.scope != sr.Scope || e.derefAliases != sr.DerefAliases ||
		e.typesOnly != sr.TypesOnly || utils.IsStringSlicesDifferent(e.attributes, sr.Attributes) {
		return false
	}

	return true
}

func (e *ExtendedSearchRequestMatcher) String() string {
	return fmt.Sprintf("baseDN: %s, filter %s", e.baseDN, e.filter)
}

func TestShouldCheckLDAPServerExtensions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp, ldapOIDExtensionTLS},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connClose := mockClient.EXPECT().Close()

	gomock.InOrder(dialURL, connBind, searchOIDs, connClose)

	err := provider.StartupCheck()
	assert.NoError(t, err)

	assert.True(t, provider.features.Extensions.PwdModifyExOp)
	assert.True(t, provider.features.Extensions.TLS)

	assert.False(t, provider.features.ControlTypes.MsftPwdPolHints)
	assert.False(t, provider.features.ControlTypes.MsftPwdPolHintsDeprecated)
}

func TestShouldNotCheckLDAPServerExtensionsWhenRootDSEReturnsMoreThanOneEntry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp, ldapOIDExtensionTLS},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
				{},
			},
		}, nil)

	connClose := mockClient.EXPECT().Close()

	gomock.InOrder(dialURL, connBind, searchOIDs, connClose)

	err := provider.StartupCheck()
	assert.NoError(t, err)

	assert.False(t, provider.features.Extensions.PwdModifyExOp)
	assert.False(t, provider.features.Extensions.TLS)

	assert.False(t, provider.features.ControlTypes.MsftPwdPolHints)
	assert.False(t, provider.features.ControlTypes.MsftPwdPolHintsDeprecated)
}

func TestShouldCheckLDAPServerControlTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connClose := mockClient.EXPECT().Close()

	gomock.InOrder(dialURL, connBind, searchOIDs, connClose)

	err := provider.StartupCheck()
	assert.NoError(t, err)

	assert.False(t, provider.features.Extensions.PwdModifyExOp)
	assert.False(t, provider.features.Extensions.TLS)

	assert.True(t, provider.features.ControlTypes.MsftPwdPolHints)
	assert.True(t, provider.features.ControlTypes.MsftPwdPolHintsDeprecated)
}

func TestShouldNotEnablePasswdModifyExtensionOrControlTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connClose := mockClient.EXPECT().Close()

	gomock.InOrder(dialURL, connBind, searchOIDs, connClose)

	err := provider.StartupCheck()
	assert.NoError(t, err)

	assert.False(t, provider.features.Extensions.PwdModifyExOp)
	assert.False(t, provider.features.Extensions.TLS)

	assert.False(t, provider.features.ControlTypes.MsftPwdPolHints)
	assert.False(t, provider.features.ControlTypes.MsftPwdPolHintsDeprecated)
}

func TestShouldReturnCheckServerConnectError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, errors.New("could not connect"))

	err := provider.StartupCheck()
	assert.EqualError(t, err, "dial failed with error: could not connect")

	assert.False(t, provider.features.Extensions.PwdModifyExOp)
}

func TestShouldReturnCheckServerSearchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(nil, errors.New("could not perform the search"))

	connClose := mockClient.EXPECT().Close()

	gomock.InOrder(dialURL, connBind, searchOIDs, connClose)

	err := provider.StartupCheck()
	assert.EqualError(t, err, "error occurred during RootDSE search: could not perform the search")

	assert.False(t, provider.features.Extensions.PwdModifyExOp)
}

type SearchRequestMatcher struct {
	expected string
}

func NewSearchRequestMatcher(expected string) *SearchRequestMatcher {
	return &SearchRequestMatcher{expected}
}

func (srm *SearchRequestMatcher) Matches(x any) bool {
	sr := x.(*ldap.SearchRequest)
	return sr.Filter == srm.expected
}

func (srm *SearchRequestMatcher) String() string {
	return ""
}

func TestShouldEscapeUserInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	mockClient.EXPECT().
		// Here we ensure that the input has been correctly escaped.
		Search(NewSearchRequestMatcher("(|(uid=john\\=abc)(mail=john\\=abc))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := provider.getUserProfile(mockClient, "john=abc")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func TestShouldReturnEmailWhenAttributeSameAsUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "mail",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "(&({username_attribute}={input})(objectClass=inetOrgPerson))",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	assert.Equal(t, []string{"mail", "displayName"}, provider.usersAttributes)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(NewSearchRequestMatcher("(&(mail=john@example.com)(objectClass=inetOrgPerson))")).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=john,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "mail",
							Values: []string{"john@example.com"},
						},
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john@example.com")

	assert.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, "uid=john,dc=example,dc=com", profile.DN)
	assert.Equal(t, "john@example.com", profile.Username)
	assert.Equal(t, "John Doe", profile.DisplayName)

	require.Len(t, profile.Emails, 1)
	assert.Equal(t, "john@example.com", profile.Emails[0])
}

func TestShouldReturnUsernameAndBlankDisplayNameWhenAttributesTheSame(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "uid",
			UsersFilter:          "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=inetOrgPerson))",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	assert.Equal(t, []string{"uid", "mail"}, provider.usersAttributes)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(NewSearchRequestMatcher("(&(|(uid=john@example.com)(mail=john@example.com))(objectClass=inetOrgPerson))")).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=john,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "uid",
							Values: []string{"john"},
						},
						{
							Name:   "mail",
							Values: []string{"john@example.com"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john@example.com")

	assert.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, "uid=john,dc=example,dc=com", profile.DN)
	assert.Equal(t, "john", profile.Username)
	assert.Equal(t, "john", profile.DisplayName)

	require.Len(t, profile.Emails, 1)
	assert.Equal(t, "john@example.com", profile.Emails[0])
}

func TestShouldReturnBlankEmailAndDisplayNameWhenAttrsLenZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=inetOrgPerson))",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	assert.Equal(t, []string{"uid", "mail", "displayName"}, provider.usersAttributes)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(NewSearchRequestMatcher("(&(|(uid=john@example.com)(mail=john@example.com))(objectClass=inetOrgPerson))")).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=john,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "uid",
							Values: []string{"john"},
						},
						{
							Name:   "mail",
							Values: []string{},
						},
						{
							Name:   "displayName",
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john@example.com")

	assert.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, "uid=john,dc=example,dc=com", profile.DN)
	assert.Equal(t, "john", profile.Username)
	assert.Equal(t, "", profile.DisplayName)

	assert.Len(t, profile.Emails, 0)
}

func TestShouldCombineUsernameFilterAndUsersFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			UsernameAttribute:    "uid",
			UsersFilter:          "(&({username_attribute}={input})(&(objectCategory=person)(objectClass=user)))",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	assert.Equal(t, []string{"uid", "mail", "displayName"}, provider.usersAttributes)

	assert.True(t, provider.usersFilterReplacementInput)

	mockClient.EXPECT().
		Search(NewSearchRequestMatcher("(&(uid=john)(&(objectCategory=person)(objectClass=user)))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := provider.getUserProfile(mockClient, "john")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func createSearchResultWithAttributes(attributes ...*ldap.EntryAttribute) *ldap.SearchResult {
	return &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{Attributes: attributes},
		},
	}
}

func createSearchResultWithAttributeValues(values ...string) *ldap.SearchResult {
	return createSearchResultWithAttributes(&ldap.EntryAttribute{
		Values: values,
	})
}

func TestShouldNotCrashWhenGroupsAreNotRetrievedFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connBind, searchProfile, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestShouldNotCrashWhenEmailsAreNotRetrievedFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:           testLDAPAddress,
			User:              "cn=admin,dc=example,dc=com",
			Password:          "password",
			UsernameAttribute: "uid",
			UsersFilter:       "uid={input}",
			AdditionalUsersDN: "ou=users",
			BaseDN:            "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connBind, searchProfile, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{})
	assert.Equal(t, details.Username, "john")
}

func TestShouldReturnUsernameFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connBind, searchProfile, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnUsernameFromLDAPWithReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries:   []*ldap.Entry{},
			Referrals: []string{"ldap://192.168.2.1"},
		}, nil)

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.2.1"), gomock.Any()).
		Return(mockClientReferral, nil)

	connBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferral := mockClientReferral.EXPECT().Close()

	searchProfileReferral := mockClientReferral.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connBind, searchProfile, dialURLReferral, connBindReferral, searchProfileReferral, connCloseReferral, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnUsernameFromLDAPWithReferralsInErrorAndResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)
	mockClientReferralAlt := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries:   []*ldap.Entry{},
			Referrals: []string{"ldap://192.168.2.1"},
		}, &ldap.Error{ResultCode: ldap.LDAPResultReferral, Err: errors.New("referral"), Packet: &testBERPacketReferral})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.2.1"), gomock.Any()).
		Return(mockClientReferral, nil)

	connBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferral := mockClientReferral.EXPECT().Close()

	searchProfileReferral := mockClientReferral.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	dialURLReferralAlt := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(mockClientReferralAlt, nil)

	connBindReferralAlt := mockClientReferralAlt.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferralAlt := mockClientReferralAlt.EXPECT().Close()

	searchProfileReferralAlt := mockClientReferralAlt.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connBind, searchProfile, dialURLReferral, connBindReferral, searchProfileReferral, connCloseReferral, dialURLReferralAlt, connBindReferralAlt, searchProfileReferralAlt, connCloseReferralAlt, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnUsernameFromLDAPWithReferralsErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{}, &ldap.Error{ResultCode: ldap.LDAPResultReferral, Err: errors.New("referral"), Packet: &testBERPacketReferral})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(mockClientReferral, nil)

	connBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferral := mockClientReferral.EXPECT().Close()

	searchProfileReferral := mockClientReferral.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connBind, searchProfile, dialURLReferral, connBindReferral, searchProfileReferral, connCloseReferral, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldNotUpdateUserPasswordConnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      false,
		},
		false,
		nil,
		mockFactory)

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(nil, errors.New("tcp timeout"))

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: dial failed with error: tcp timeout")
}

func TestShouldNotUpdateUserPasswordGetDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      false,
		},
		false,
		nil,
		mockFactory)

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(nil, &ldap.Error{ResultCode: ldap.LDAPResultProtocolError, Err: errors.New("permission error")})

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: cannot find user DN of user 'john'. Cause: LDAP Result Code 2 \"Protocol Error\": permission error")
}

func TestShouldUpdateUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		nil,
	)

	modifyRequest.Replace(ldapAttributeUserPassword, []string{"password"})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	modify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, modify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSAD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := utf16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	modify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, modify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSADWithReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := utf16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	modify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(&ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(mockClientReferral, nil)

	connBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferral := mockClientReferral.EXPECT().Close()

	modifyReferral := mockClientReferral.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, modify, dialURLReferral, connBindReferral, modifyReferral, connCloseReferral, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSADWithReferralsWithReferralConnectErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := utf16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	modify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(&ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(nil, errors.New("tcp timeout"))

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, modify, dialURLReferral, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred connecting to referred LDAP server 'ldap://192.168.0.1': dial failed with error: tcp timeout. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordMSADWithReferralsWithReferralModifyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := utf16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	modify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(&ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(mockClientReferral, nil)

	connBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferral := mockClientReferral.EXPECT().Close()

	modifyReferral := mockClientReferral.EXPECT().
		Modify(modifyRequest).
		Return(&ldap.Error{
			ResultCode: ldap.LDAPResultBusy,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, modify, dialURLReferral, connBindReferral, modifyReferral, connCloseReferral, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred performing modify on referred LDAP server 'ldap://192.168.0.1': LDAP Result Code 51 \"Busy\": error occurred. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordMSADWithoutReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      false,
		},
		false,
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := utf16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	modify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(&ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, modify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(nil, nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(&ldap.PasswordModifyResult{
			Referral: "ldap://192.168.0.1",
		}, &ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(mockClientReferral, nil)

	connBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferral := mockClientReferral.EXPECT().Close()

	passwdModifyReferral := mockClientReferral.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(&ldap.PasswordModifyResult{}, nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, dialURLReferral, connBindReferral, passwdModifyReferral, connCloseReferral, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithoutReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      false,
		},
		false,
		nil,
		mockFactory)

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(&ldap.PasswordModifyResult{
			Referral: "ldap://192.168.0.1",
		}, &ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsReferralConnectErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(&ldap.PasswordModifyResult{
			Referral: "ldap://192.168.0.1",
		}, &ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(nil, errors.New("tcp timeout"))

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, dialURLReferral, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred connecting to referred LDAP server 'ldap://192.168.0.1': dial failed with error: tcp timeout. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsReferralPasswordModifyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			PermitReferrals:      true,
		},
		false,
		nil,
		mockFactory)

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModifyExOp},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(&ldap.PasswordModifyResult{
			Referral: "ldap://192.168.0.1",
		}, &ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	dialURLReferral := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://192.168.0.1"), gomock.Any()).
		Return(mockClientReferral, nil)

	connBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connCloseReferral := mockClientReferral.EXPECT().Close()

	passwdModifyReferral := mockClientReferral.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(nil, &ldap.Error{
			ResultCode: ldap.LDAPResultBusy,
			Err:        errors.New("too busy"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, dialURLReferral, connBindReferral, passwdModifyReferral, connCloseReferral, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred performing password modify on referred LDAP server 'ldap://192.168.0.1': LDAP Result Code 51 \"Busy\": too busy. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordActiveDirectoryWithServerPolicyHints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "sAMAccountName",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "cn={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, _ := utf16.NewEncoder().String("\"password\"")

	modifyRequest := ldap.NewModifyRequest(
		"cn=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "cn=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "sAMAccountName",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.NoError(t, err)
}

func TestShouldUpdateUserPasswordActiveDirectoryWithServerPolicyHintsDeprecated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "sAMAccountName",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "cn={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, _ := utf16.NewEncoder().String("\"password\"")

	modifyRequest := ldap.NewModifyRequest(
		"cn=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHintsDeprecated}},
	)

	modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{ldapOIDControlMsftServerPolicyHintsDeprecated},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "cn=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "sAMAccountName",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordActiveDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "activedirectory",
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "sAMAccountName",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "cn={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, _ := utf16.NewEncoder().String("\"password\"")

	modifyRequest := ldap.NewModifyRequest(
		"cn=test,dc=example,dc=com",
		nil,
	)

	modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "cn=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "sAMAccountName",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Implementation:       "custom",
			Address:              testLDAPAddress,
			User:                 "uid=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		nil,
	)

	modifyRequest.Replace("userPassword", []string{"password"})

	dialURLOIDs := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("uid=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	connCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("uid=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	passwdModify := mockClient.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, connBindOIDs, searchOIDs, connCloseOIDs, dialURL, connBind, searchProfile, passwdModify, connClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldReturnErrorWhenMultipleUsernameAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John", "Jacob"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' has 2 values for for attribute 'uid' but the attribute must be a single value attribute")
}

func TestShouldReturnErrorWhenZeroUsernameAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' must have value for attribute 'uid'")
}

func TestShouldReturnErrorWhenUsernameAttributeNotReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' must have value for attribute 'uid'")
}

func TestShouldReturnErrorWhenMultipleUsersFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "(|(uid={input})(uid=*))",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
				{
					DN: "uid=sam,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"sam"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "there were 2 users found when searching for 'john' but there should only be 1")
}

func TestShouldReturnErrorWhenNoDN(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "(|(uid={input})(uid=*))",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, bind, search)

	client, err := provider.connect()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' must have a distinguished name but the result returned an empty distinguished name")
}

func TestShouldCheckValidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockClient, nil),
		mockClient.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockClient.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayName",
								Values: []string{"John Doe"},
							},
							{
								Name:   "mail",
								Values: []string{"test@example.com"},
							},
							{
								Name:   "uid",
								Values: []string{"John"},
							},
						},
					},
				},
			}, nil),
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockClient, nil),
		mockClient.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockClient.EXPECT().Close().Times(2),
	)

	valid, err := provider.CheckUserPassword("john", "password")

	assert.True(t, valid)
	require.NoError(t, err)
}

func TestShouldNotCheckValidUserPasswordWithConnectError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	bind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(&ldap.Error{ResultCode: ldap.LDAPResultInvalidCredentials, Err: errors.New("invalid username or password")})

	gomock.InOrder(dialURL, bind, mockClient.EXPECT().Close())

	valid, err := provider.CheckUserPassword("john", "password")

	assert.False(t, valid)
	assert.EqualError(t, err, "bind failed with error: LDAP Result Code 49 \"Invalid Credentials\": invalid username or password")
}

func TestShouldCheckInvalidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		false,
		nil,
		mockFactory)

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockClient, nil),
		mockClient.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockClient.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayName",
								Values: []string{"John Doe"},
							},
							{
								Name:   "mail",
								Values: []string{"test@example.com"},
							},
							{
								Name:   "uid",
								Values: []string{"John"},
							},
						},
					},
				},
			}, nil),
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockClient, nil),
		mockClient.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(errors.New("invalid username or password")),
		mockClient.EXPECT().Close().Times(2),
	)

	valid, err := provider.CheckUserPassword("john", "password")

	assert.False(t, valid)
	require.EqualError(t, err, "authentication failed. Cause: bind failed with error: invalid username or password")
}

func TestShouldCallStartTLSWhenEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connStartTLS := mockClient.EXPECT().
		StartTLS(provider.tlsConfig)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connStartTLS, connBind, searchProfile, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestShouldParseDynamicConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "(&(|({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt})(accountExpires>={date-time:generalized})))",
			GroupsFilter:         "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))",
			AdditionalUsersDN:    "ou=users",
			AdditionalGroupsDN:   "ou=groups",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
		},
		false,
		nil,
		mockFactory)

	clock := &utils.TestingClock{}

	provider.clock = clock

	clock.Set(time.Unix(1670250519, 0))

	assert.True(t, provider.groupsFilterReplacementInput)
	assert.True(t, provider.groupsFilterReplacementUsername)
	assert.True(t, provider.groupsFilterReplacementDN)

	assert.True(t, provider.usersFilterReplacementInput)
	assert.True(t, provider.usersFilterReplacementDateTimeGeneralized)
	assert.True(t, provider.usersFilterReplacementDateTimeMicrosoftNTTimeEpoch)

	assert.Equal(t, "(&(|(uid={input})(mail={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt})(accountExpires>={date-time:generalized})))", provider.config.UsersFilter)
	assert.Equal(t, "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))", provider.config.GroupsFilter)
	assert.Equal(t, "ou=users,dc=example,dc=com", provider.usersBaseDN)
	assert.Equal(t, "ou=groups,dc=example,dc=com", provider.groupsBaseDN)

	assert.Equal(t, "(&(|(uid=test@example.com)(mail=test@example.com))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpires>=133147241190000000)(accountExpires>=20221205142839.0Z)))", provider.resolveUsersFilter("test@example.com"))
	assert.Equal(t, "(&(|(member=cn=admin,dc=example,dc=com)(member=test@example.com)(member=test))(objectClass=group))", provider.resolveGroupsFilter("test@example.com", &ldapUserProfile{Username: "test", DN: "cn=admin,dc=example,dc=com"}))
}

func TestShouldCallStartTLSWithInsecureSkipVerifyWhenSkipVerifyTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
			TLS: &schema.TLSConfig{
				SkipVerify: true,
			},
		},
		false,
		nil,
		mockFactory)

	assert.False(t, provider.groupsFilterReplacementInput)
	assert.False(t, provider.groupsFilterReplacementUsername)
	assert.False(t, provider.groupsFilterReplacementDN)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connStartTLS := mockClient.EXPECT().
		StartTLS(provider.tlsConfig)

	connClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayName",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, connStartTLS, connBind, searchProfile, searchGroups, connClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestShouldReturnLDAPSAlreadySecuredWhenStartTLSAttempted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackend{
			Address:              testLDAPSAddress,
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
			TLS: &schema.TLSConfig{
				SkipVerify: true,
			},
		},
		false,
		nil,
		mockFactory)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldaps://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connStartTLS := mockClient.EXPECT().
		StartTLS(provider.tlsConfig).
		Return(errors.New("LDAP Result Code 200 \"Network Error\": ldap: already encrypted"))

	gomock.InOrder(dialURL, connStartTLS, mockClient.EXPECT().Close())

	_, err := provider.GetDetails("john")
	assert.EqualError(t, err, "starttls failed with error: LDAP Result Code 200 \"Network Error\": ldap: already encrypted")
}
