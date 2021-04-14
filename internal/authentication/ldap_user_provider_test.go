package authentication

import (
	"errors"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldCreateRawConnectionWhenSchemeIsLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL: "ldap://127.0.0.1:389",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	_, err := ldapClient.connect("cn=admin,dc=example,dc=com", "password")

	require.NoError(t, err)
}

func TestShouldCreateTLSConnectionWhenSchemeIsLDAPS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL: "ldaps://127.0.0.1:389",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldaps://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	_, err := ldapClient.connect("cn=admin,dc=example,dc=com", "password")

	require.NoError(t, err)
}

func TestEscapeSpecialCharsFromUserInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL: "ldaps://127.0.0.1:389",
		},
		nil,
		mockFactory)

	// No escape
	assert.Equal(t, "xyz", ldapClient.ldapEscape("xyz"))

	// Escape
	assert.Equal(t, "test\\,abc", ldapClient.ldapEscape("test,abc"))
	assert.Equal(t, "test\\5cabc", ldapClient.ldapEscape("test\\abc"))
	assert.Equal(t, "test\\2aabc", ldapClient.ldapEscape("test*abc"))
	assert.Equal(t, "test \\28abc\\29", ldapClient.ldapEscape("test (abc)"))
	assert.Equal(t, "test\\#abc", ldapClient.ldapEscape("test#abc"))
	assert.Equal(t, "test\\+abc", ldapClient.ldapEscape("test+abc"))
	assert.Equal(t, "test\\<abc", ldapClient.ldapEscape("test<abc"))
	assert.Equal(t, "test\\>abc", ldapClient.ldapEscape("test>abc"))
	assert.Equal(t, "test\\;abc", ldapClient.ldapEscape("test;abc"))
	assert.Equal(t, "test\\\"abc", ldapClient.ldapEscape("test\"abc"))
	assert.Equal(t, "test\\=abc", ldapClient.ldapEscape("test=abc"))
	assert.Equal(t, "test\\,\\5c\\28abc\\29", ldapClient.ldapEscape("test,\\(abc)"))
}

func TestEscapeSpecialCharsInGroupsFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:          "ldaps://127.0.0.1:389",
			GroupsFilter: "(|(member={dn})(uid={username})(uid={input}))",
		},
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

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockConn.EXPECT().
		// Here we ensure that the input has been correctly escaped.
		Search(NewSearchRequestMatcher("(|(uid=john\\=abc)(mail=john\\=abc))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := ldapClient.getUserProfile(mockConn, "john=abc")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func TestShouldCombineUsernameFilterAndUsersFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsernameAttribute:    "uid",
			UsersFilter:          "(&({username_attribute}={input})(&(objectCategory=person)(objectClass=user)))",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
		},
		nil,
		mockFactory)

	mockConn.EXPECT().
		Search(NewSearchRequestMatcher("(&(uid=john)(&(objectCategory=person)(objectClass=user)))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := ldapClient.getUserProfile(mockConn, "john")
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

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
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

	gomock.InOrder(searchProfile, searchGroups)

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

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:               "ldap://127.0.0.1:389",
			User:              "cn=admin,dc=example,dc=com",
			Password:          "password",
			UsernameAttribute: "uid",
			UsersFilter:       "uid={input}",
			AdditionalUsersDN: "ou=users",
			BaseDN:            "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)
	searchProfile := mockConn.EXPECT().
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

	gomock.InOrder(searchProfile, searchGroups)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{})
	assert.Equal(t, details.Username, "john")
}

func TestShouldReturnUsernameFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
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

	gomock.InOrder(searchProfile, searchGroups)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldUpdateUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	modifyRequest := ldap.NewModifyRequest("uid=test,dc=example,dc=com", nil)
	modifyRequest.Replace("userPassword", []string{"password"})

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayname",
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
		mockConn.EXPECT().
			Modify(modifyRequest).
			Return(nil),
		mockConn.EXPECT().
			Close(),
	)

	err := ldapClient.UpdatePassword("john", "password")

	require.NoError(t, err)
}

func TestShouldCheckValidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayname",
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
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Close().Times(2),
	)

	valid, err := ldapClient.CheckUserPassword("john", "password")

	assert.True(t, valid)
	require.NoError(t, err)
}

func TestShouldCheckInvalidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayname",
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
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(errors.New("Invalid username or password")),
		mockConn.EXPECT().
			Close(),
	)

	valid, err := ldapClient.CheckUserPassword("john", "password")

	assert.False(t, valid)
	require.EqualError(t, err, "Authentication of user john failed. Cause: Invalid username or password")
}

func TestShouldCallStartTLSWhenEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		StartTLS(ldapClient.tlsConfig)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
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

	gomock.InOrder(searchProfile, searchGroups)

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

	mockFactory := NewMockLDAPConnectionFactory(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "(&(|({username_attribute}={input})({mail_attribute}={input})({display_name_attribute}={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))",
			GroupsFilter:         "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))",
			AdditionalUsersDN:    "ou=users",
			AdditionalGroupsDN:   "ou=groups",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
		},
		nil,
		mockFactory)

	assert.Equal(t, "(&(|(uid={input})(mail={input})(displayname={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))", ldapClient.configuration.UsersFilter)
	assert.Equal(t, "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))", ldapClient.configuration.GroupsFilter)
	assert.Equal(t, "ou=users,dc=example,dc=com", ldapClient.usersBaseDN)
	assert.Equal(t, "ou=groups,dc=example,dc=com", ldapClient.groupsBaseDN)
}

func TestShouldCallStartTLSWithInsecureSkipVerifyWhenSkipVerifyTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
			TLS: &schema.TLSConfig{
				SkipVerify: true,
			},
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		StartTLS(ldapClient.tlsConfig)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
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

	gomock.InOrder(searchProfile, searchGroups)

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

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldaps://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
			TLS: &schema.TLSConfig{
				SkipVerify: true,
			},
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldaps://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		StartTLS(ldapClient.tlsConfig).
		Return(errors.New("LDAP Result Code 200 \"Network Error\": ldap: already encrypted"))

	_, err := ldapClient.GetDetails("john")
	assert.EqualError(t, err, "LDAP Result Code 200 \"Network Error\": ldap: already encrypted")
}
