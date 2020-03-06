package authentication

import (
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ldap.v3"
)

func TestShouldCreateRawConnectionWhenSchemeIsLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldap := NewLDAPUserProviderWithFactory(schema.LDAPAuthenticationBackendConfiguration{
		URL: "ldap://127.0.0.1:389",
	}, mockFactory)

	mockFactory.EXPECT().
		Dial(gomock.Eq("tcp"), gomock.Eq("127.0.0.1:389")).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	_, err := ldap.connect("cn=admin,dc=example,dc=com", "password")

	require.NoError(t, err)
}

func TestShouldCreateTLSConnectionWhenSchemeIsLDAPS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldap := NewLDAPUserProviderWithFactory(schema.LDAPAuthenticationBackendConfiguration{
		URL: "ldaps://127.0.0.1:389",
	}, mockFactory)

	mockFactory.EXPECT().
		DialTLS(gomock.Eq("tcp"), gomock.Eq("127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	_, err := ldap.connect("cn=admin,dc=example,dc=com", "password")

	require.NoError(t, err)
}

func TestEscapeSpecialChars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	ldap := NewLDAPUserProviderWithFactory(schema.LDAPAuthenticationBackendConfiguration{
		URL: "ldaps://127.0.0.1:389",
	}, mockFactory)

	// No escape
	assert.Equal(t, "xyz", ldap.ldapEscape("xyz"))

	// Escape
	assert.Equal(t, "test\\,abc", ldap.ldapEscape("test,abc"))
	assert.Equal(t, "test\\\\abc", ldap.ldapEscape("test\\abc"))
	assert.Equal(t, "test\\#abc", ldap.ldapEscape("test#abc"))
	assert.Equal(t, "test\\+abc", ldap.ldapEscape("test+abc"))
	assert.Equal(t, "test\\<abc", ldap.ldapEscape("test<abc"))
	assert.Equal(t, "test\\>abc", ldap.ldapEscape("test>abc"))
	assert.Equal(t, "test\\;abc", ldap.ldapEscape("test;abc"))
	assert.Equal(t, "test\\\"abc", ldap.ldapEscape("test\"abc"))
	assert.Equal(t, "test\\=abc", ldap.ldapEscape("test=abc"))

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

	ldapClient := NewLDAPUserProviderWithFactory(schema.LDAPAuthenticationBackendConfiguration{
		URL:               "ldap://127.0.0.1:389",
		User:              "cn=admin,dc=example,dc=com",
		Password:          "password",
		UsersFilter:       "uid={0}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}, mockFactory)

	mockConn.EXPECT().
		// Here we ensure that the input has been correctly escaped.
		Search(NewSearchRequestMatcher("uid=john\\=abc")).
		Return(&ldap.SearchResult{}, nil)

	ldapClient.getUserProfile(mockConn, "john=abc")
}

func createSearchResultWithAttributes(attributes ...*ldap.EntryAttribute) *ldap.SearchResult {
	return &ldap.SearchResult{
		Entries: []*ldap.Entry{
			&ldap.Entry{Attributes: attributes},
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

	ldapClient := NewLDAPUserProviderWithFactory(schema.LDAPAuthenticationBackendConfiguration{
		URL:               "ldap://127.0.0.1:389",
		User:              "cn=admin,dc=example,dc=com",
		Password:          "password",
		UsernameAttribute: "uid",
		MailAttribute:     "mail",
		UsersFilter:       "uid={0}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}, mockFactory)

	mockFactory.EXPECT().
		Dial(gomock.Eq("tcp"), gomock.Eq("127.0.0.1:389")).
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
				&ldap.Entry{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						&ldap.EntryAttribute{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						&ldap.EntryAttribute{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchGroups, searchProfile)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.Username, "john")
}

func TestShouldNotCrashWhenEmailsAreNotRetrievedFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := NewLDAPUserProviderWithFactory(schema.LDAPAuthenticationBackendConfiguration{
		URL:               "ldap://127.0.0.1:389",
		User:              "cn=admin,dc=example,dc=com",
		Password:          "password",
		UsernameAttribute: "uid",
		UsersFilter:       "uid={0}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}, mockFactory)

	mockFactory.EXPECT().
		Dial(gomock.Eq("tcp"), gomock.Eq("127.0.0.1:389")).
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
				&ldap.Entry{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						&ldap.EntryAttribute{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchGroups, searchProfile)

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

	ldapClient := NewLDAPUserProviderWithFactory(schema.LDAPAuthenticationBackendConfiguration{
		URL:               "ldap://127.0.0.1:389",
		User:              "cn=admin,dc=example,dc=com",
		Password:          "password",
		UsernameAttribute: "uid",
		MailAttribute:     "mail",
		UsersFilter:       "uid={0}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}, mockFactory)

	mockFactory.EXPECT().
		Dial(gomock.Eq("tcp"), gomock.Eq("127.0.0.1:389")).
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
				&ldap.Entry{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						&ldap.EntryAttribute{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						&ldap.EntryAttribute{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchGroups, searchProfile)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.Username, "John")
}
