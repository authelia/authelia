package authentication

import (
	"errors"
	"fmt"
	"testing"

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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:      "ldap://127.0.0.1:389",
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

	_, err := ldapClient.connect()

	require.NoError(t, err)
}

func TestShouldCreateTLSConnectionWhenSchemeIsLDAPS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:      "ldaps://127.0.0.1:389",
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

	_, err := ldapClient.connect()

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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:          "ldaps://127.0.0.1:389",
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

	filter, _ := ldapClient.resolveGroupsFilter("john", &profile)
	assert.Equal(t, "(|(member=cn=john \\28external\\29,dc=example,dc=com)(uid=john)(uid=john))", filter)

	filter, _ = ldapClient.resolveGroupsFilter("john#=(abc,def)", &profile)
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

func (e *ExtendedSearchRequestMatcher) Matches(x interface{}) bool {
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	assert.NoError(t, err)

	assert.True(t, ldapClient.supportExtensionPasswdModify)
}

func TestShouldCheckLDAPServerControlTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsControlType, ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
						},
					},
				},
			},
		}, nil)

	connClose := mockClient.EXPECT().Close()

	gomock.InOrder(dialURL, connBind, searchOIDs, connClose)

	err := ldapClient.StartupCheck()
	assert.NoError(t, err)

	assert.True(t, ldapClient.supportControlTypeMicrosoftServerPolicyHints)
	assert.True(t, ldapClient.supportControlTypeMicrosoftServerPolicyHintsDeprecated)
}

func TestShouldNotEnablePasswdModifyExtensionOrControlTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	err := ldapClient.StartupCheck()
	assert.NoError(t, err)

	assert.False(t, ldapClient.supportExtensionPasswdModify)
	assert.False(t, ldapClient.supportControlTypeMicrosoftServerPolicyHints)
	assert.False(t, ldapClient.supportControlTypeMicrosoftServerPolicyHintsDeprecated)
}

func TestShouldReturnCheckServerConnectError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	err := ldapClient.StartupCheck()
	assert.EqualError(t, err, "dial failed with error: could not connect")

	assert.False(t, ldapClient.supportExtensionPasswdModify)
}

func TestShouldReturnCheckServerSearchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	err := ldapClient.StartupCheck()
	assert.EqualError(t, err, "could not perform the search")

	assert.False(t, ldapClient.supportExtensionPasswdModify)
}

type SearchRequestMatcher struct {
	expected string
}

func NewSearchRequestMatcher(expected string) *SearchRequestMatcher {
	return &SearchRequestMatcher{expected}
}

func (srm *SearchRequestMatcher) Matches(x interface{}) bool {
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	mockClient.EXPECT().
		// Here we ensure that the input has been correctly escaped.
		Search(NewSearchRequestMatcher("(|(uid=john\\=abc)(mail=john\\=abc))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := ldapClient.getUserProfile(mockClient, "john=abc")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func TestShouldCombineUsernameFilterAndUsersFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsernameAttribute:    "uid",
			UsersFilter:          "(&({username_attribute}={input})(&(objectCategory=person)(objectClass=user)))",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
		},
		false,
		nil,
		mockFactory)

	assert.True(t, ldapClient.usersFilterReplacementInput)

	mockClient.EXPECT().
		Search(NewSearchRequestMatcher("(&(uid=john)(&(objectCategory=person)(objectClass=user)))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := ldapClient.getUserProfile(mockClient, "john")
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	details, err := ldapClient.GetDetails("john")
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:               "ldap://127.0.0.1:389",
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

	details, err := ldapClient.GetDetails("john")
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	details, err := ldapClient.GetDetails("john")
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: dial failed with error: tcp timeout")
}

func TestShouldNotUpdateUserPasswordGetDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: cannot find user DN of user 'john'. Cause: LDAP Result Code 2 \"Protocol Error\": permission error")
}

func TestShouldUpdateUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSAD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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
		[]ldap.Control{&controlMicrosoftServerPolicyHints{ldapOIDMicrosoftServerPolicyHintsControlType}},
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsControlType, ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSADWithReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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
		[]ldap.Control{&controlMicrosoftServerPolicyHints{ldapOIDMicrosoftServerPolicyHintsControlType}},
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsControlType, ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSADWithReferralsWithReferralConnectErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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
		[]ldap.Control{&controlMicrosoftServerPolicyHints{ldapOIDMicrosoftServerPolicyHintsControlType}},
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsControlType, ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred connecting to referred LDAP server 'ldap://192.168.0.1': dial failed with error: tcp timeout. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordMSADWithReferralsWithReferralModifyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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
		[]ldap.Control{&controlMicrosoftServerPolicyHints{ldapOIDMicrosoftServerPolicyHintsControlType}},
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsControlType, ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred performing modify on referred LDAP server 'ldap://192.168.0.1': LDAP Result Code 51 \"Busy\": error occurred. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordMSADWithoutReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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
		[]ldap.Control{&controlMicrosoftServerPolicyHints{ldapOIDMicrosoftServerPolicyHintsControlType}},
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsControlType, ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithoutReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsReferralConnectErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred connecting to referred LDAP server 'ldap://192.168.0.1': dial failed with error: tcp timeout. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsReferralPasswordModifyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
							Values: []string{ldapOIDPasswdModifyExtension},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred performing password modify on referred LDAP server 'ldap://192.168.0.1': LDAP Result Code 51 \"Busy\": too busy. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordActiveDirectoryWithServerPolicyHints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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
		[]ldap.Control{&controlMicrosoftServerPolicyHints{ldapOIDMicrosoftServerPolicyHintsControlType}},
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsControlType, ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordActiveDirectoryWithServerPolicyHintsDeprecated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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
		[]ldap.Control{&controlMicrosoftServerPolicyHints{ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType}},
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
							Values: []string{ldapOIDMicrosoftServerPolicyHintsDeprecatedControlType},
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordActiveDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "activedirectory",
			URL:                  "ldap://127.0.0.1:389",
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			Implementation:       "custom",
			URL:                  "ldap://127.0.0.1:389",
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

	err := ldapClient.StartupCheck()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldCheckValidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	valid, err := ldapClient.CheckUserPassword("john", "password")

	assert.True(t, valid)
	require.NoError(t, err)
}

func TestShouldCheckInvalidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
		mockClient.EXPECT().Close(),
	)

	valid, err := ldapClient.CheckUserPassword("john", "password")

	assert.False(t, valid)
	require.EqualError(t, err, "authentication failed. Cause: bind failed with error: invalid username or password")
}

func TestShouldCallStartTLSWhenEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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
		StartTLS(ldapClient.tlsConfig)

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

	details, err := ldapClient.GetDetails("john")
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayName",
			UsersFilter:          "(&(|({username_attribute}={input})({mail_attribute}={input})({display_name_attribute}={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))",
			GroupsFilter:         "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))",
			AdditionalUsersDN:    "ou=users",
			AdditionalGroupsDN:   "ou=groups",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
		},
		false,
		nil,
		mockFactory)

	assert.True(t, ldapClient.groupsFilterReplacementInput)
	assert.True(t, ldapClient.groupsFilterReplacementUsername)
	assert.True(t, ldapClient.groupsFilterReplacementDN)

	assert.True(t, ldapClient.usersFilterReplacementInput)

	assert.Equal(t, "(&(|(uid={input})(mail={input})(displayName={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))", ldapClient.config.UsersFilter)
	assert.Equal(t, "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))", ldapClient.config.GroupsFilter)
	assert.Equal(t, "ou=users,dc=example,dc=com", ldapClient.usersBaseDN)
	assert.Equal(t, "ou=groups,dc=example,dc=com", ldapClient.groupsBaseDN)
}

func TestShouldCallStartTLSWithInsecureSkipVerifyWhenSkipVerifyTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
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

	assert.False(t, ldapClient.groupsFilterReplacementInput)
	assert.False(t, ldapClient.groupsFilterReplacementUsername)
	assert.False(t, ldapClient.groupsFilterReplacementDN)

	dialURL := mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connStartTLS := mockClient.EXPECT().
		StartTLS(ldapClient.tlsConfig)

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

	details, err := ldapClient.GetDetails("john")
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

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldaps://127.0.0.1:389",
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
		StartTLS(ldapClient.tlsConfig).
		Return(errors.New("LDAP Result Code 200 \"Network Error\": ldap: already encrypted"))

	gomock.InOrder(dialURL, connStartTLS)

	_, err := ldapClient.GetDetails("john")
	assert.EqualError(t, err, "starttls failed with error: LDAP Result Code 200 \"Network Error\": ldap: already encrypted")
}
