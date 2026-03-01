package authentication

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestNewLDAPUserProvider(t *testing.T) {
	provider := NewLDAPUserProvider(schema.AuthenticationBackend{LDAP: &schema.AuthenticationBackendLDAP{}}, nil)

	assert.NotNil(t, provider)
}

func TestShouldCreateRawConnectionWhenSchemeIsLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Timeout:  time.Second * 20,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		config,
		false,
		factory)

	dialURL := mockDialer.EXPECT().DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 20))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	gomock.InOrder(dialURL, setTimeout, NewRootDSESearchRequest(mockClient, nil), clientBind)

	_, err := provider.factory.GetClient()

	require.NoError(t, err)
}

func TestShouldHandleRootDSESearchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Timeout:  time.Second * 20,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		config,
		false,
		factory)

	dialURL := mockDialer.EXPECT().DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 20))

	search := mockClient.EXPECT().Search(ldapNewSearchRequestRootDSE()).Return(nil, fmt.Errorf("failed to search"))

	bind := mockClient.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil)

	gomock.InOrder(dialURL, setTimeout, search, bind, mockClient.EXPECT().Close())

	client, err := provider.factory.GetClient()

	require.NoError(t, err)

	require.NoError(t, client.Close())
}

func TestShouldHandleBindError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Timeout:  time.Second * 20,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(
		config,
		false,
		factory)

	dialURL := mockDialer.EXPECT().DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 20))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(fmt.Errorf("failed to bind"))

	gomock.InOrder(dialURL, setTimeout, NewRootDSESearchRequest(mockClient, nil), clientBind, mockClient.EXPECT().Close())

	_, err := provider.factory.GetClient()

	assert.EqualError(t, err, "error occurred performing bind: failed to bind")
}

func TestShouldCreateTLSConnectionWhenSchemeIsLDAPS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPSAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	dialURL := mockDialer.EXPECT().DialURL("ldaps://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	gomock.InOrder(dialURL, setTimeout, NewRootDSESearchRequest(mockClient, nil), clientBind)

	_, err := provider.factory.GetClient()

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

	config := &schema.AuthenticationBackendLDAP{
		Address:      testLDAPSAddress,
		GroupsFilter: "(|(member={dn})(uid={username})(uid={input}))",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

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

func TestResolveGroupsFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPSAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	testCases := []struct {
		name     string
		have     schema.AuthenticationBackendLDAP
		input    string
		profile  *ldapUserProfile
		expected string
	}{
		{
			"ShouldResolveEmptyFilter",
			schema.AuthenticationBackendLDAP{},
			"",
			&ldapUserProfile{},
			"",
		},
		{
			"ShouldResolveMemberOfRDNFilter",
			schema.AuthenticationBackendLDAP{
				GroupsFilter: "(|{memberof:rdn})",
				Attributes: schema.AuthenticationBackendLDAPAttributes{
					DistinguishedName: "distinguishedName",
					GroupName:         "cn",
					MemberOf:          "memberOf",
					Username:          "uid",
					Mail:              "mail",
					DisplayName:       "displayName",
				},
			},
			"",
			&ldapUserProfile{
				MemberOf: []string{"CN=abc,DC=example,DC=com", "CN=xyz,DC=example,DC=com"},
			},
			"(|(CN=abc)(CN=xyz))",
		},
		{
			"ShouldResolveMemberOfDNFilter",
			schema.AuthenticationBackendLDAP{
				GroupsFilter: "(|{memberof:dn})",
				Attributes: schema.AuthenticationBackendLDAPAttributes{
					DistinguishedName: "distinguishedName",
					GroupName:         "cn",
					MemberOf:          "memberOf",
					Username:          "uid",
					Mail:              "mail",
					DisplayName:       "displayName",
				},
			},
			"",
			&ldapUserProfile{
				MemberOf: []string{"CN=abc,DC=example,DC=com", "CN=xyz,DC=example,DC=com"},
			},
			"(|(distinguishedName=CN=abc,DC=example,DC=com)(distinguishedName=CN=xyz,DC=example,DC=com))",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewLDAPUserProviderWithFactory(&tc.have, false, factory)

			assert.Equal(t, tc.expected, provider.resolveGroupsFilter("", tc.profile))
		})
	}
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

func TestShouldCheckLDAPEpochFilters(t *testing.T) {
	type have struct {
		users string
		attr  schema.AuthenticationBackendLDAPAttributes
	}

	type expected struct {
		dtgeneralized bool
		dtmsftnt      bool
		dtunix        bool
	}

	testCases := []struct {
		name     string
		have     have
		expected expected
	}{
		{
			"ShouldNotEnableAny",
			have{},
			expected{},
		},
		{
			"ShouldNotEnableMSFTNT",
			have{
				users: "(abc={date-time:microsoft-nt})",
			},
			expected{
				dtmsftnt: true,
			},
		},
		{
			"ShouldNotEnableUnix",
			have{
				users: "(abc={date-time:unix})",
			},
			expected{
				dtunix: true,
			},
		},
		{
			"ShouldNotEnableGeneralized",
			have{
				users: "(abc={date-time:generalized})",
			},
			expected{
				dtgeneralized: true,
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPClientFactory(ctrl)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewLDAPUserProviderWithFactory(
				&schema.AuthenticationBackendLDAP{
					UsersFilter: tc.have.users,
					Attributes:  tc.have.attr,
					BaseDN:      "dc=example,dc=com",
				},
				false,
				mockFactory)

			assert.Equal(t, tc.expected.dtgeneralized, provider.usersFilterReplacementDateTimeGeneralized)
			assert.Equal(t, tc.expected.dtmsftnt, provider.usersFilterReplacementDateTimeMicrosoftNTTimeEpoch)
			assert.Equal(t, tc.expected.dtunix, provider.usersFilterReplacementDateTimeUnixEpoch)
		})
	}
}

func TestShouldReturnCheckServerSearchErrorPooled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:     testLDAPAddress,
		User:        "cn=admin,dc=example,dc=com",
		UsersFilter: "(|({username_attribute}={input})({mail_attribute}={input}))",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		Password:          "password",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		Pooling:           schema.AuthenticationBackendLDAPPooling{Count: 1},
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)
	mockClient := NewMockLDAPClient(ctrl)
	mockClientSecond := NewMockLDAPClient(ctrl)

	factory := NewPooledLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(nil, errors.New("could not perform the search"))

	clientBindSecond := mockClientSecond.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	gomock.InOrder(
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil),
		mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		NewRootDSESearchRequest(mockClient, nil),
		clientBind,
		mockClient.EXPECT().IsClosing().Return(false),
		searchOIDs,
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClientSecond, nil),
		mockClientSecond.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		mockClientSecond.EXPECT().Search(gomock.Any()).Return(&ldap.SearchResult{}, nil),
		clientBindSecond,
		mockClientSecond.EXPECT().IsClosing().Return(false),
		mockClientSecond.EXPECT().Close().Return(nil),
	)

	assert.NoError(t, provider.StartupCheck())
	assert.NoError(t, provider.Close())
}

func TestShouldPermitRootDSEFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:     testLDAPAddress,
		User:        "cn=admin,dc=example,dc=com",
		UsersFilter: "(|({username_attribute}={input})({mail_attribute}={input}))",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		Password:          "password",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	search := NewRootDSESearchRequest(mockClient, fmt.Errorf("failed"))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	gomock.InOrder(dialURL, setTimeout, search, clientBind, clientClose)

	assert.NoError(t, provider.StartupCheck())
}

func TestShouldPermitRootDSEFailurePooled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:     testLDAPAddress,
		User:        "cn=admin,dc=example,dc=com",
		UsersFilter: "(|({username_attribute}={input})({mail_attribute}={input}))",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		Password:          "password",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		Pooling:           schema.AuthenticationBackendLDAPPooling{Count: 1},
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)
	mockClient := NewMockLDAPClient(ctrl)

	factory := NewPooledLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	search := NewExtendedSearchRequestMatcher("(objectClass=*)", "",
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, false,
		[]string{
			ldapObjectClassAttribute,
			ldapSupportedLDAPVersionAttribute,
			ldapSupportedExtensionAttribute,
			ldapSupportedControlAttribute,
			ldapSupportedFeaturesAttribute,
			ldapSupportedSASLMechanismsAttribute,
			ldapVendorNameAttribute,
			ldapVendorVersionAttribute,
			ldapDomainFunctionalityAttribute,
			ldapForestFunctionalityAttribute,
		})

	gomock.InOrder(
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil),
		mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		NewRootDSESearchRequest(mockClient, nil),
		clientBind,
		mockClient.EXPECT().IsClosing().Return(false),
		mockClient.EXPECT().
			Search(search).
			Return(&ldap.SearchResult{Entries: []*ldap.Entry{{}}}, nil),
		mockClient.EXPECT().IsClosing().Return(false),
		mockClient.EXPECT().Close().Return(nil),
	)

	assert.NoError(t, provider.StartupCheck())
	assert.NoError(t, provider.Close())
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
		&schema.AuthenticationBackendLDAP{
			Address:     testLDAPAddress,
			User:        "cn=admin,dc=example,dc=com",
			UsersFilter: "(|({username_attribute}={input})({mail_attribute}={input}))",
			Attributes: schema.AuthenticationBackendLDAPAttributes{
				Username:    "uid",
				Mail:        "mail",
				DisplayName: "displayName",
				MemberOf:    "memberOf",
			},
			Password:          "password",
			AdditionalUsersDN: "ou=users",
			BaseDN:            "dc=example,dc=com",
			PermitReferrals:   true,
		},
		false,
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

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "mail",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "(&({username_attribute}={input})(objectClass=inetOrgPerson))",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	assert.Equal(t, []string{"mail", "displayName", "memberOf"}, provider.usersAttributes)

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	clientBind := mockClient.EXPECT().
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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
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

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "uid",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=inetOrgPerson))",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	assert.Equal(t, []string{"uid", "mail", "memberOf"}, provider.usersAttributes)

	clientBind := mockClient.EXPECT().
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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
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

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=inetOrgPerson))",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	assert.Equal(t, []string{"uid", "mail", "displayName", "memberOf"}, provider.usersAttributes)

	clientBind := mockClient.EXPECT().
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
						{
							Name:   "memberOf",
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
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
		&schema.AuthenticationBackendLDAP{
			Address:     testLDAPAddress,
			User:        "cn=admin,dc=example,dc=com",
			UsersFilter: "(&({username_attribute}={input})(&(objectCategory=person)(objectClass=user)))",
			Password:    "password",
			Attributes: schema.AuthenticationBackendLDAPAttributes{
				Username:    "uid",
				Mail:        "mail",
				DisplayName: "displayName",
				MemberOf:    "memberOf",
			},
			AdditionalUsersDN: "ou=users",
			BaseDN:            "dc=example,dc=com",
			PermitReferrals:   true,
		},
		false,
		mockFactory)

	assert.Equal(t, []string{"uid", "mail", "displayName", "memberOf"}, provider.usersAttributes)

	assert.True(t, provider.usersFilterReplacementInput)

	mockClient.EXPECT().
		Search(NewSearchRequestMatcher("(&(uid=john)(&(objectCategory=person)(objectClass=user)))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := provider.getUserProfile(mockClient, "john")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

//nolint:unparam
func createSearchResultWithAttributes(attributes ...*ldap.EntryAttribute) *ldap.SearchResult {
	return &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{Attributes: attributes},
		},
	}
}

func createGroupSearchResultModeFilter(name string, groupNames ...string) *ldap.SearchResult {
	result := &ldap.SearchResult{
		Entries: make([]*ldap.Entry, len(groupNames)),
	}

	for i, groupName := range groupNames {
		result.Entries[i] = &ldap.Entry{Attributes: []*ldap.EntryAttribute{{Name: name, Values: []string{groupName}}}}
	}

	return result
}

func createGroupSearchResultModeFilterWithDN(name string, groupNames []string, groupDNs []string) *ldap.SearchResult {
	if len(groupNames) != len(groupDNs) {
		panic("input sizes mismatch")
	}

	result := &ldap.SearchResult{
		Entries: make([]*ldap.Entry, len(groupNames)),
	}

	for i, groupName := range groupNames {
		result.Entries[i] = &ldap.Entry{
			DN:         groupDNs[i],
			Attributes: []*ldap.EntryAttribute{{Name: name, Values: []string{groupName}}},
		}
	}

	return result
}

func TestShouldNotCrashWhenGroupsAreNotRetrievedFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestLDAPUserProvider_GetDetailsExtended_ShouldPass(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDialer := NewMockLDAPClientDialer(ctrl)

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:       "uid",
			Mail:           "mail",
			DisplayName:    "displayName",
			MemberOf:       "memberOf",
			StreetAddress:  "street",
			FamilyName:     "sn",
			MiddleName:     "middle",
			GivenName:      "givenName",
			Nickname:       "nickname",
			Gender:         "gender",
			Birthdate:      "birthDate",
			Website:        "website",
			Profile:        "profile",
			Picture:        "picture",
			ZoneInfo:       "zoneinfo",
			Locale:         "locale",
			PhoneNumber:    "phone",
			PhoneExtension: "ext",
			Locality:       "locality",
			Region:         "region",
			PostalCode:     "postCode",
			Country:        "c",
			Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
				"exampleStr": {
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: false,
						ValueType:   ValueTypeString,
					},
				},
				"exampleStrMV": {
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: true,
						ValueType:   ValueTypeString,
					},
				},
				"exampleInt": {
					Name: "exampleIntChangedAttributeName",
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: false,
						ValueType:   ValueTypeInteger,
					},
				},
				"exampleIntMV": {
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: true,
						ValueType:   ValueTypeInteger,
					},
				},
				"exampleBool": {
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: false,
						ValueType:   ValueTypeBoolean,
					},
				},
				"exampleBoolMV": {
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: true,
						ValueType:   ValueTypeBoolean,
					},
				},
				"exampleEmptyStringInt": {
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: false,
						ValueType:   ValueTypeInteger,
					},
				},
				"exampleEmptyStringBoolean": {
					AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
						MultiValued: false,
						ValueType:   ValueTypeBoolean,
					},
				},
			},
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
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
						{
							Name:   "memberOf",
							Values: []string{},
						},
						{
							Name:   "street",
							Values: []string{"123 Banksia Ln"},
						},
						{
							Name:   "sn",
							Values: []string{"Smith"},
						},
						{
							Name:   "middle",
							Values: []string{"Jacob"},
						},
						{
							Name:   "givenName",
							Values: []string{"John"},
						},
						{
							Name:   "nickname",
							Values: []string{"Johny"},
						},
						{
							Name:   "gender",
							Values: []string{"male"},
						},
						{
							Name:   "birthDate",
							Values: []string{"2/2/2021"},
						},
						{
							Name:   "website",
							Values: []string{"https://authelia.com"},
						},
						{
							Name:   "profile",
							Values: []string{"https://authelia.com/profile/jsmith.html"},
						},
						{
							Name:   "picture",
							Values: []string{"https://authelia.com/picture/jsmith.jpg"},
						},
						{
							Name:   "zoneinfo",
							Values: []string{"Australia/Melbourne"},
						},
						{
							Name:   "locale",
							Values: []string{"en-AU"},
						},
						{
							Name:   "phone",
							Values: []string{"+1 (604) 555-1234"},
						},
						{
							Name:   "ext",
							Values: []string{"5678"},
						},
						{
							Name:   "locality",
							Values: []string{"Melbourne"},
						},
						{
							Name:   "region",
							Values: []string{"Victoria"},
						},
						{
							Name:   "postCode",
							Values: []string{"2000"},
						},
						{
							Name:   "c",
							Values: []string{"Australia"},
						},
						{
							Name:   "exampleStr",
							Values: []string{"abc"},
						},
						{
							Name:   "exampleStrMV",
							Values: []string{"abc", "123"},
						},
						{
							Name:   "exampleInt",
							Values: []string{"123"},
						},
						{
							Name:   "exampleIntMV",
							Values: []string{"1023879012731.5", "123"},
						},
						{
							Name:   "exampleBool",
							Values: []string{"true"},
						},
						{
							Name:   "exampleBoolMV",
							Values: []string{"true", "false"},
						},
						{
							Name:   "exampleEmptyStringInt",
							Values: []string{""},
						},
						{
							Name:   "exampleEmptyStringBoolean",
							Values: []string{""},
						},
					},
				},
			},
		}, nil)

	searchGroup := mockClient.EXPECT().Search(gomock.Any()).Return(createSearchResultWithAttributes(), nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, searchProfile, searchGroup, connClose)

	enAU := language.MustParse("en-AU")
	website, _ := url.Parse("https://authelia.com")
	profile, _ := url.Parse("https://authelia.com/profile/jsmith.html")
	picture, _ := url.Parse("https://authelia.com/picture/jsmith.jpg")

	details, err := provider.GetDetailsExtended("john")
	assert.Equal(t,
		&UserDetailsExtended{
			GivenName:      "John",
			FamilyName:     "Smith",
			MiddleName:     "Jacob",
			Nickname:       "Johny",
			Profile:        profile,
			Picture:        picture,
			Website:        website,
			Gender:         "male",
			Birthdate:      "2/2/2021",
			ZoneInfo:       "Australia/Melbourne",
			Locale:         &enAU,
			PhoneNumber:    "+1 (604) 555-1234",
			PhoneExtension: "5678",
			Address: &UserDetailsAddress{
				StreetAddress: "123 Banksia Ln",
				Locality:      "Melbourne",
				Region:        "Victoria",
				PostalCode:    "2000",
				Country:       "Australia",
			},
			Extra: map[string]any{
				"exampleStr":                     "abc",
				"exampleStrMV":                   []any{"abc", "123"},
				"exampleIntChangedAttributeName": float64(123),
				"exampleIntMV":                   []any{1023879012731.5, float64(123)},
				"exampleBool":                    true,
				"exampleBoolMV":                  []any{true, false},
			},
			UserDetails: &UserDetails{Username: "john", DisplayName: "", Emails: []string{}, Groups: []string(nil)},
		}, details)

	assert.NoError(t, err)
}

func TestLDAPUserProvider_GetDetailsExtended_ShouldParseError(t *testing.T) {
	testCases := []struct {
		name        string
		valueType   string
		multiValued bool
		err         string
	}{
		{
			"ShouldHandleBadInteger",
			ValueTypeInteger,
			false,
			"cannot parse 'example' with value 'abc' as integer: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			"ShouldHandleBadIntegerMV",
			ValueTypeInteger,
			true,
			"cannot parse 'example' with value 'abc' as integer: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			"ShouldHandleBadBoolean",
			ValueTypeBoolean,
			false,
			"cannot parse 'example' with value 'abc' as boolean: strconv.ParseBool: parsing \"abc\": invalid syntax",
		},
		{
			"ShouldHandleBadBooleanMV",
			ValueTypeBoolean,
			true,
			"cannot parse 'example' with value 'abc' as boolean: strconv.ParseBool: parsing \"abc\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDialer := NewMockLDAPClientDialer(ctrl)

			config := &schema.AuthenticationBackendLDAP{
				Address:  testLDAPAddress,
				User:     "cn=admin,dc=example,dc=com",
				Password: "password",
				Attributes: schema.AuthenticationBackendLDAPAttributes{
					Username:      "uid",
					Mail:          "mail",
					DisplayName:   "displayName",
					MemberOf:      "memberOf",
					StreetAddress: "street",
					Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
						"example": {
							AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
								MultiValued: tc.multiValued,
								ValueType:   tc.valueType,
							},
						},
					},
				},
				UsersFilter:       "uid={input}",
				AdditionalUsersDN: "ou=users",
				BaseDN:            "dc=example,dc=com",
				PermitReferrals:   true,
			}

			factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

			provider := NewLDAPUserProviderWithFactory(config, false, factory)

			mockClient := NewMockLDAPClient(ctrl)

			dialURL := mockDialer.EXPECT().
				DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
				Return(mockClient, nil)

			setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

			dseSearch := NewRootDSESearchRequest(mockClient, nil)

			connBind := mockClient.EXPECT().
				Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
				Return(nil)

			connClose := mockClient.EXPECT().Close()

			searchProfile := mockClient.EXPECT().
				Search(gomock.Any()).
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
								{
									Name:   "memberOf",
									Values: []string{},
								},
								{
									Name:   "street",
									Values: []string{"123 Banksia Ln"},
								},
								{
									Name:   "example",
									Values: []string{"abc"},
								},
							},
						},
					},
				}, nil)

			gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, searchProfile, connClose)

			details, err := provider.GetDetailsExtended("john")
			assert.Nil(t, details)
			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestLDAPUserProvider_GetDetailsExtended_ShouldErrorBadPictureURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDialer := NewMockLDAPClientDialer(ctrl)

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			Picture:     "photoURL",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
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
						{
							Name:   "memberOf",
							Values: []string{},
						},
						{
							Name:   "photoURL",
							Values: []string{"bad_+URL"},
						},
					},
				},
			},
		}, nil)

	searchGroup := mockClient.EXPECT().Search(gomock.Any()).Return(createSearchResultWithAttributes(), nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, searchProfile, searchGroup, connClose)

	details, err := provider.GetDetailsExtended("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "error occurred parsing user details for 'john': failed to parse the picture attribute 'photoURL' with value 'bad_+URL': parse \"bad_+URL\": invalid URI for request")
}

func TestLDAPUserProvider_GetDetailsExtended_ShouldErrorBadProfileURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDialer := NewMockLDAPClientDialer(ctrl)

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			Profile:     "profile",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
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
						{
							Name:   "memberOf",
							Values: []string{},
						},
						{
							Name:   "profile",
							Values: []string{"bad_+URL"},
						},
					},
				},
			},
		}, nil)

	searchGroup := mockClient.EXPECT().Search(gomock.Any()).Return(createSearchResultWithAttributes(), nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, searchProfile, searchGroup, connClose)

	details, err := provider.GetDetailsExtended("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "error occurred parsing user details for 'john': failed to parse the profile attribute 'profile' with value 'bad_+URL': parse \"bad_+URL\": invalid URI for request")
}

func TestLDAPUserProvider_GetDetailsExtended_ShouldErrorBadWebsiteURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDialer := NewMockLDAPClientDialer(ctrl)

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			Website:     "www",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
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
						{
							Name:   "memberOf",
							Values: []string{},
						},
						{
							Name:   "www",
							Values: []string{"bad_+URL"},
						},
					},
				},
			},
		}, nil)

	searchGroup := mockClient.EXPECT().Search(gomock.Any()).Return(createSearchResultWithAttributes(), nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, searchProfile, searchGroup, connClose)

	details, err := provider.GetDetailsExtended("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "error occurred parsing user details for 'john': failed to parse the website attribute 'www' with value 'bad_+URL': parse \"bad_+URL\": invalid URI for request")
}

func TestLDAPUserProvider_GetDetailsExtended_ShouldErrorBadLocale(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDialer := NewMockLDAPClientDialer(ctrl)

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			Locale:      "locale",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
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
						{
							Name:   "memberOf",
							Values: []string{},
						},
						{
							Name:   "locale",
							Values: []string{"ba12390-n2m3jkn123&!@#!_+"},
						},
					},
				},
			},
		}, nil)

	searchGroup := mockClient.EXPECT().Search(gomock.Any()).Return(createSearchResultWithAttributes(), nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, searchProfile, searchGroup, connClose)

	details, err := provider.GetDetailsExtended("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "error occurred parsing user details for 'john': failed to parse the locale attribute 'locale' with value 'ba12390-n2m3jkn123&!@#!_+': language: tag is not well-formed")
}

func TestLDAPUserProvider_GetDetails_ShouldReturnOnUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(nil, fmt.Errorf("failed to search"))

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, clientClose)

	details, err := provider.GetDetails("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "cannot find user DN of user 'john'. Cause: failed to search")
}

func TestLDAPUserProvider_GetDetailsExtendedShouldReturnOnBindError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDialer := NewMockLDAPClientDialer(ctrl)

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(fmt.Errorf("bad bind"))

	connClose := mockClient.EXPECT().Close().Return(nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, connClose)

	details, err := provider.GetDetailsExtended("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "error occurred performing bind: bad bind")
}

func TestLDAPUserProvider_GetDetailsExtendedShouldReturnOnDialError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDialer := NewMockLDAPClientDialer(ctrl)

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(nil, fmt.Errorf("failed to dial"))

	gomock.InOrder(dialURL)

	details, err := provider.GetDetailsExtended("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "error occurred dialing address: failed to dial")
}

func TestLDAPUserProvider_GetDetailsExtendedShouldReturnOnUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	factory := NewStandardLDAPClientFactory(config, nil, mockDialer)
	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	dialURL := mockDialer.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	connBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(nil, fmt.Errorf("failed to search"))

	gomock.InOrder(dialURL, setTimeout, dseSearch, connBind, searchProfile, connClose)

	details, err := provider.GetDetailsExtended("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "cannot find user DN of user 'john'. Cause: failed to search")
}

func TestLDAPUserProvider_GetDetails_ShouldReturnOnGroupsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(nil, fmt.Errorf("failed to search groups"))

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")

	assert.Nil(t, details)
	assert.EqualError(t, err, "unable to retrieve groups of user 'john'. Cause: failed to search groups")
}

func TestShouldNotCrashWhenEmailsAreNotRetrievedFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "displayName",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createGroupSearchResultModeFilter(provider.config.Attributes.GroupName, "group1", "group2"), nil)

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{})
	assert.Equal(t, details.Username, "john")
}

func TestShouldUnauthenticatedBind(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "displayName",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		UnauthenticatedBind(gomock.Eq("cn=admin,dc=example,dc=com")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createGroupSearchResultModeFilter(provider.config.Attributes.GroupName, "group1", "group2"), nil)

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{})
	assert.Equal(t, details.Username, "john")
}

func TestShouldReturnUsernameFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createGroupSearchResultModeFilter(provider.config.Attributes.GroupName, "group1", "group2"), nil)

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnUsernameFromLDAPSearchModeMemberOfRDN(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		GroupSearchMode:   "memberof",
		UsersFilter:       "uid={input}",
		GroupsFilter:      "(|{memberof:rdn})",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "DC=example,DC=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	requestGroups := ldap.NewSearchRequest(
		provider.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, "(|(CN=admins)(CN=users))", provider.groupsAttributes, nil,
	)

	// This ensures the filtering works correctly in the following ways:
	// Item 1 (0th element), has the wrong case.
	// Item 2 (1st element), has the wrong DN.
	searchGroups := mockClient.EXPECT().
		Search(requestGroups).
		Return(createGroupSearchResultModeFilterWithDN(provider.config.Attributes.GroupName, []string{"admins", "notadmins", "users"}, []string{"CN=ADMINS,OU=groups,DC=example,DC=com", "CN=notadmins,OU=wronggroups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com"}), nil)

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
						{
							Name:   "memberOf",
							Values: []string{"CN=admins,OU=groups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"admins", "users"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnUsernameFromLDAPSearchModeMemberOfDN(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "CN=Administrator,CN=Users,DC=example,DC=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
			GroupName:         "cn",
		},
		GroupSearchMode:   "memberof",
		UsersFilter:       "sAMAccountName={input}",
		GroupsFilter:      "(|{memberof:dn})",
		AdditionalUsersDN: "CN=users",
		BaseDN:            "DC=example,DC=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("CN=Administrator,CN=Users,DC=example,DC=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	requestGroups := ldap.NewSearchRequest(
		provider.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, "(|(distinguishedName=CN=admins,OU=groups,DC=example,DC=com)(distinguishedName=CN=users,OU=groups,DC=example,DC=com))", provider.groupsAttributes, nil,
	)

	searchGroups := mockClient.EXPECT().
		Search(requestGroups).
		Return(createGroupSearchResultModeFilterWithDN(provider.config.Attributes.GroupName, []string{"admins", "admins", "users"}, []string{"CN=admins,OU=groups,DC=example,DC=com", "CN=admins,OU=wronggroups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com"}), nil)

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
							Name:   "sAMAccountName",
							Values: []string{"John"},
						},
						{
							Name:   "memberOf",
							Values: []string{"CN=admins,OU=groups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"admins", "users"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnErrSearchMemberOf(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "CN=Administrator,CN=Users,DC=example,DC=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
			GroupName:         "cn",
		},
		GroupSearchMode:   "memberof",
		UsersFilter:       "sAMAccountName={input}",
		GroupsFilter:      "(|{memberof:dn})",
		AdditionalUsersDN: "CN=users",
		BaseDN:            "DC=example,DC=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("CN=Administrator,CN=Users,DC=example,DC=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	requestGroups := ldap.NewSearchRequest(
		provider.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, "(|(distinguishedName=CN=admins,OU=groups,DC=example,DC=com)(distinguishedName=CN=users,OU=groups,DC=example,DC=com))", provider.groupsAttributes, nil,
	)

	searchGroups := mockClient.EXPECT().
		Search(requestGroups).
		Return(nil, fmt.Errorf("failed to get groups"))

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
							Name:   "sAMAccountName",
							Values: []string{"John"},
						},
						{
							Name:   "memberOf",
							Values: []string{"CN=admins,OU=groups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "unable to retrieve groups of user 'john'. Cause: failed to get groups")
}

func TestShouldReturnErrUnknownSearchMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "CN=Administrator,CN=Users,DC=example,DC=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
			GroupName:         "cn",
		},
		GroupSearchMode:   "bad",
		UsersFilter:       "sAMAccountName={input}",
		GroupsFilter:      "(|{memberof:dn})",
		AdditionalUsersDN: "CN=users",
		BaseDN:            "DC=example,DC=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("CN=Administrator,CN=Users,DC=example,DC=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
							Name:   "sAMAccountName",
							Values: []string{"John"},
						},
						{
							Name:   "memberOf",
							Values: []string{"CN=admins,OU=groups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, clientClose)

	details, err := provider.GetDetails("john")
	assert.Nil(t, details)

	assert.EqualError(t, err, "could not perform group search with mode 'bad' as it's unknown")
}

func TestShouldSkipEmptyAttributesSearchModeMemberOf(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "CN=Administrator,CN=Users,DC=example,DC=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
			GroupName:         "cn",
		},
		GroupSearchMode:   "memberof",
		UsersFilter:       "sAMAccountName={input}",
		GroupsFilter:      "(|{memberof:dn})",
		AdditionalUsersDN: "CN=users",
		BaseDN:            "DC=example,DC=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("CN=Administrator,CN=Users,DC=example,DC=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
							Name:   "sAMAccountName",
							Values: []string{"John"},
						},
						{
							Name:   "memberOf",
							Values: []string{"CN=admins,OU=groups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com", "CN=multi,OU=groups,DC=example,DC=com"},
						},
					},
				},
			},
		}, nil)

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN:         "uid=grp,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{},
				},
				{
					DN: "CN=users,OU=groups,DC=example,DC=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "cn",
							Values: []string{},
						},
					},
				},
				{
					DN: "CN=admins,OU=groups,DC=example,DC=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "cn",
							Values: []string{""},
						},
					},
				},
				{
					DN: "CN=multi,OU=groups,DC=example,DC=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "cn",
							Values: []string{"a", "b"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")

	assert.NoError(t, err)
	assert.NotNil(t, details)
}

func TestShouldSkipEmptyAttributesSearchModeFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "CN=Administrator,CN=Users,DC=example,DC=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
			GroupName:         "cn",
		},
		GroupSearchMode:   "filter",
		UsersFilter:       "sAMAccountName={input}",
		GroupsFilter:      "(|{memberof:dn})",
		AdditionalUsersDN: "CN=users",
		BaseDN:            "DC=example,DC=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("CN=Administrator,CN=Users,DC=example,DC=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
							Name:   "sAMAccountName",
							Values: []string{"John"},
						},
						{
							Name:   "memberOf",
							Values: []string{"CN=admins,OU=groups,DC=example,DC=com", "CN=users,OU=groups,DC=example,DC=com", "CN=multi,OU=groups,DC=example,DC=com"},
						},
					},
				},
			},
		}, nil)

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN:         "uid=grp,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{},
				},
				{
					DN: "CN=users,OU=groups,DC=example,DC=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "cn",
							Values: []string{},
						},
					},
				},
				{
					DN: "CN=admins,OU=groups,DC=example,DC=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "cn",
							Values: []string{""},
						},
					},
				},
				{
					DN: "CN=multi,OU=groups,DC=example,DC=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "cn",
							Values: []string{"a", "b"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")

	assert.NoError(t, err)
	assert.NotNil(t, details)
}

func TestShouldSkipEmptyGroupsResultMemberOf(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{},
			},
		}, nil)

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnUsernameFromLDAPWithReferralsInErrorAndNoResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createGroupSearchResultModeFilter(provider.config.Attributes.GroupName, "group1", "group2"), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{Referrals: []string{"ldap://192.168.0.1"}}, &ldap.Error{ResultCode: ldap.LDAPResultReferral, Err: errors.New("referral"), Packet: &testBERPacketReferral})

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(mockClientReferral, nil)

	setTimeoutReferral := mockClientReferral.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearchReferral := NewRootDSESearchRequest(mockClientReferral, nil)

	clientBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseReferral := mockClientReferral.EXPECT().Close()

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, dialURLReferral, setTimeoutReferral, dseSearchReferral, clientBindReferral, searchProfileReferral, clientCloseReferral, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldReturnErrorWhenUntypedReferralError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{Referrals: []string{"ldap://192.168.0.1"}}, fmt.Errorf("referral"))

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, clientClose)

	details, err := provider.GetDetails("john")
	assert.Nil(t, details)
	assert.EqualError(t, err, "cannot find user DN of user 'john'. Cause: referral")
}

func TestShouldReturnDialErrDuringReferralSearchUsernameFromLDAPWithReferralsInErrorAndNoResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{Referrals: []string{"ldap://192.168.0.1"}}, &ldap.Error{ResultCode: ldap.LDAPResultReferral, Err: errors.New("referral"), Packet: &testBERPacketReferral})

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(nil, fmt.Errorf("failed to connect"))

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, dialURLReferral, clientClose)

	details, err := provider.GetDetails("john")

	assert.Nil(t, details)
	assert.EqualError(t, err, "cannot find user DN of user 'john'. Cause: error occurred connecting to referred LDAP server 'ldap://192.168.0.1': error occurred dialing address: failed to connect")
}

func TestShouldReturnSearchErrDuringReferralSearchUsernameFromLDAPWithReferralsInErrorAndNoResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{Referrals: []string{"ldap://192.168.0.1"}}, &ldap.Error{ResultCode: ldap.LDAPResultReferral, Err: errors.New("referral"), Packet: &testBERPacketReferral})

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(mockClientReferral, nil)

	setTimeoutReferral := mockClientReferral.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearchReferral := NewRootDSESearchRequest(mockClientReferral, nil)

	clientBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseReferral := mockClientReferral.EXPECT().Close()

	searchProfileReferral := mockClientReferral.EXPECT().
		Search(gomock.Any()).
		Return(nil, fmt.Errorf("not found"))

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, dialURLReferral, setTimeoutReferral, dseSearchReferral, clientBindReferral, searchProfileReferral, clientCloseReferral, clientClose)

	details, err := provider.GetDetails("john")

	assert.Nil(t, details)
	assert.EqualError(t, err, "cannot find user DN of user 'john'. Cause: error occurred performing search on referred LDAP server 'ldap://192.168.0.1': not found")
}

func TestShouldNotReturnUsernameFromLDAPWithReferralsInErrorAndReferralsNotPermitted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   false,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(nil, &ldap.Error{ResultCode: ldap.LDAPResultReferral, Err: errors.New("referral"), Packet: &testBERPacketReferral})

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, clientClose)

	details, err := provider.GetDetails("john")
	assert.EqualError(t, err, "cannot find user DN of user 'john'. Cause: LDAP Result Code 10 \"Referral\": referral")
	assert.Nil(t, details)
}

func TestShouldReturnUsernameFromLDAPWithReferralsErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchGroups := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(createGroupSearchResultModeFilter(provider.config.Attributes.GroupName, "group1", "group2"), nil)

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{Referrals: []string{"ldap://192.168.0.1"}}, &ldap.Error{ResultCode: ldap.LDAPResultReferral, Err: errors.New("referral"), Packet: &testBERPacketReferral})

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(mockClientReferral, nil)

	setTimeoutReferral := mockClientReferral.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearchReferral := NewRootDSESearchRequest(mockClientReferral, nil)

	clientBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseReferral := mockClientReferral.EXPECT().Close()

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, searchProfile, dialURLReferral, setTimeoutReferral, dseSearchReferral, clientBindReferral, searchProfileReferral, clientCloseReferral, searchGroups, clientClose)

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

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   false,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(nil, errors.New("tcp timeout"))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURLOIDs, setTimeout, dseSearch, clientBindOIDs, clientCloseOIDs, dialURL)

	require.NoError(t, provider.StartupCheck())

	assert.EqualError(t, provider.UpdatePassword("john", "password"), "unable to update password. Cause: error occurred dialing address: tcp timeout")
}

func TestShouldNotUpdateUserPasswordGetDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   false,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

	searchProfile := mockClient.EXPECT().
		Search(gomock.Any()).
		Return(nil, &ldap.Error{ResultCode: ldap.LDAPResultProtocolError, Err: errors.New("permission error")})

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, clientClose)

	require.NoError(t, provider.StartupCheck())

	assert.EqualError(t, provider.UpdatePassword("john", "password"), "unable to update password. Cause: cannot find user DN of user 'john'. Cause: LDAP Result Code 2 \"Protocol Error\": permission error")
}

func TestShouldUpdateUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		nil,
	)

	modifyRequest.Replace(ldapAttributeUserPassword, []string{"password"})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{}, nil)

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSAD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{}, nil)

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSADWithReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearchReferral := mockClientReferral.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{Referral: "ldap://192.168.0.1"}, &ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(mockClientReferral, nil)

	setTimeoutReferral := mockClientReferral.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseReferral := mockClientReferral.EXPECT().Close()

	modifyReferral := mockClientReferral.EXPECT().
		Modify(modifyRequest).
		Return(nil)

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, dialURLReferral, setTimeoutReferral, dseSearchReferral, clientBindReferral, modifyReferral, clientCloseReferral, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordMSADWithReferralsButIncorrectErrorType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{Referral: "ldap://192.168.0.1"}, fmt.Errorf("referral"))

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, clientClose)

	require.NoError(t, provider.StartupCheck())
	assert.EqualError(t, provider.UpdatePassword("john", "password"), "unable to update password. Cause: referral")
}

func TestShouldUpdateUserPasswordMSADWithReferralsButIncorrectResultCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{Referral: "ldap://192.168.0.1"}, &ldap.Error{
			ResultCode: ldap.LDAPResultAdminLimitExceeded,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, clientClose)

	require.NoError(t, provider.StartupCheck())

	assert.EqualError(t, provider.UpdatePassword("john", "password"), "unable to update password. Cause: LDAP Result Code 11 \"Admin Limit Exceeded\": error occurred")
}

func TestShouldUpdateUserPasswordMSADWithReferralsWithReferralConnectErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(
			&ldap.ModifyResult{Referral: "ldap://192.168.0.1"},
			&ldap.Error{
				ResultCode: ldap.LDAPResultReferral,
				Err:        errors.New("error occurred"),
				Packet:     &testBERPacketReferral,
			},
		)

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(nil, errors.New("tcp timeout"))

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, dialURLReferral, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred connecting to referred LDAP server 'ldap://192.168.0.1': error occurred dialing address: tcp timeout. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordMSADWithReferralsWithReferralModifyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearchReferral := mockClientReferral.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(
			&ldap.ModifyResult{Referral: "ldap://192.168.0.1"},
			&ldap.Error{
				ResultCode: ldap.LDAPResultReferral,
				Err:        errors.New("error occurred"),
				Packet:     &testBERPacketReferral,
			},
		)

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(mockClientReferral, nil)

	setTimeoutReferral := mockClientReferral.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseReferral := mockClientReferral.EXPECT().Close()

	modifyReferral := mockClientReferral.EXPECT().
		Modify(modifyRequest).
		Return(&ldap.Error{
			ResultCode: ldap.LDAPResultBusy,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, dialURLReferral, setTimeoutReferral, dseSearchReferral, clientBindReferral, modifyReferral, clientCloseReferral, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred performing modify on referred LDAP server 'ldap://192.168.0.1': LDAP Result Code 51 \"Busy\": error occurred. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordMSADWithoutReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   false,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	pwdEncoded, _ := encodingUTF16LittleEndian.NewEncoder().String(fmt.Sprintf("\"%s\"", "password"))
	modifyRequest.Replace(ldapAttributeUnicodePwd, []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{Referral: "ldap://192.168.0.1"}, &ldap.Error{
			ResultCode: ldap.LDAPResultReferral,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, modify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearchReferral := mockClientReferral.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(mockClientReferral, nil)

	setTimeoutReferral := mockClientReferral.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseReferral := mockClientReferral.EXPECT().Close()

	passwdModifyReferral := mockClientReferral.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(&ldap.PasswordModifyResult{}, nil)

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, dialURLReferral, setTimeoutReferral, dseSearchReferral, clientBindReferral, passwdModifyReferral, clientCloseReferral, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsButBadResultCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
			ResultCode: ldap.LDAPResultAdminLimitExceeded,
			Err:        errors.New("error occurred"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	assert.EqualError(t, provider.UpdatePassword("john", "password"), "unable to update password. Cause: LDAP Result Code 11 \"Admin Limit Exceeded\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsButBadErrorType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		}, fmt.Errorf("referral"))

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	assert.EqualError(t, provider.UpdatePassword("john", "password"), "unable to update password. Cause: referral")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithoutReferrals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   false,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsReferralConnectErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(nil, errors.New("tcp timeout"))

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, dialURLReferral, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred connecting to referred LDAP server 'ldap://192.168.0.1': error occurred dialing address: tcp timeout. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordPasswdModifyExtensionWithReferralsReferralPasswordModifyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		PermitReferrals:   true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)
	mockClientReferral := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	dseSearchReferral := mockClientReferral.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDExtensionPwdModify},
						},
						{
							Name:   ldapSupportedControlAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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

	dialURLReferral := mockDialer.EXPECT().DialURL("ldap://192.168.0.1", gomock.Any()).Return(mockClientReferral, nil)

	setTimeoutReferral := mockClientReferral.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindReferral := mockClientReferral.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseReferral := mockClientReferral.EXPECT().Close()

	passwdModifyReferral := mockClientReferral.EXPECT().
		PasswordModify(pwdModifyRequest).
		Return(nil, &ldap.Error{
			ResultCode: ldap.LDAPResultBusy,
			Err:        errors.New("too busy"),
			Packet:     &testBERPacketReferral,
		})

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, dialURLReferral, setTimeoutReferral, dseSearchReferral, clientBindReferral, passwdModifyReferral, clientCloseReferral, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.EqualError(t, err, "unable to update password. Cause: error occurred performing password modify on referred LDAP server 'ldap://192.168.0.1': LDAP Result Code 51 \"Busy\": too busy. Original Error: LDAP Result Code 10 \"Referral\": error occurred")
}

func TestShouldUpdateUserPasswordActiveDirectoryWithServerPolicyHints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
		},
		UsersFilter:       "cn={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, _ := utf16.NewEncoder().String("\"password\"")

	modifyRequest := ldap.NewModifyRequest(
		"cn=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHints}},
	)

	modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{}, nil)

	gomock.InOrder(dialURLOIDs, setTimeoutOIDs, searchOIDs, clientBindOIDs, clientCloseOIDs, dialURL, setTimeout, dseSearch, clientBind, searchProfile, passwdModify, clientClose)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	assert.NoError(t, err)
}

func TestShouldUpdateUserPasswordActiveDirectoryWithServerPolicyHintsDeprecated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
		},
		UsersFilter:       "cn={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, _ := utf16.NewEncoder().String("\"password\"")

	modifyRequest := ldap.NewModifyRequest(
		"cn=test,dc=example,dc=com",
		[]ldap.Control{&controlMsftServerPolicyHints{ldapOIDControlMsftServerPolicyHintsDeprecated}},
	)

	modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{}, nil)

	gomock.InOrder(
		dialURLOIDs,
		setTimeoutOIDs,
		searchOIDs,
		clientBindOIDs,
		clientCloseOIDs,
		dialURL,
		setTimeout,
		dseSearch,
		clientBind,
		searchProfile,
		passwdModify,
		clientClose,
	)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordActiveDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "activedirectory",
		Address:        testLDAPAddress,
		User:           "cn=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			DistinguishedName: "distinguishedName",
			Username:          "sAMAccountName",
			Mail:              "mail",
			DisplayName:       "displayName",
			MemberOf:          "memberOf",
		},
		UsersFilter:       "cn={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, _ := utf16.NewEncoder().String("\"password\"")

	modifyRequest := ldap.NewModifyRequest(
		"cn=test,dc=example,dc=com",
		nil,
	)

	modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	dseSearch := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{}, nil)

	gomock.InOrder(
		dialURLOIDs,
		setTimeoutOIDs,
		searchOIDs,
		clientBindOIDs,
		clientCloseOIDs,
		dialURL,
		setTimeout,
		dseSearch,
		clientBind,
		searchProfile,
		passwdModify,
		clientClose,
	)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldUpdateUserPasswordBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Implementation: "custom",
		Address:        testLDAPAddress,
		User:           "uid=admin,dc=example,dc=com",
		Password:       "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	modifyRequest := ldap.NewModifyRequest(
		"uid=test,dc=example,dc=com",
		nil,
	)

	modifyRequest.Replace("userPassword", []string{"password"})

	dialURLOIDs := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeoutOIDs := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	clientBindOIDs := mockClient.EXPECT().
		Bind(gomock.Eq("uid=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientCloseOIDs := mockClient.EXPECT().Close()

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("uid=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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
		ModifyWithResult(modifyRequest).
		Return(&ldap.ModifyResult{}, nil)

	dseOIDSearch := NewRootDSESearchRequest(mockClient, nil)

	searchOIDs := mockClient.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute})).
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

	gomock.InOrder(
		dialURLOIDs,
		setTimeoutOIDs,
		dseOIDSearch,
		clientBindOIDs,
		clientCloseOIDs,
		dialURL,
		setTimeout,
		searchOIDs,
		clientBind,
		searchProfile,
		passwdModify,
		clientClose,
	)

	err := provider.StartupCheck()
	require.NoError(t, err)

	err = provider.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldReturnErrorWhenMultipleUsernameAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' has 2 values for for attribute 'uid' but the attribute must be a single value attribute")
}

func TestShouldReturnErrorWhenZeroUsernameAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' must have value for attribute 'uid'")
}

func TestShouldReturnErrorWhenUsernameAttributeNotReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' must have value for attribute 'uid'")
}

func TestShouldReturnErrorWhenMultipleUsersFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "(|(uid={input})(uid=*))",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "there were 2 users found when searching for 'john' but there should only be 1")
}

func TestShouldReturnErrorWhenNoDN(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "(|(uid={input})(uid=*))",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
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

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, search)

	client, err := provider.factory.GetClient()
	assert.NoError(t, err)

	profile, err := provider.getUserProfile(client, "john")

	assert.Nil(t, profile)
	assert.EqualError(t, err, "user 'john' must have a distinguished name but the result returned an empty distinguished name")
}

func TestShouldCheckValidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	mockClientUser := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	dseSearchUser := NewRootDSESearchRequest(mockClientUser, nil)

	gomock.InOrder(
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil),
		mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		dseSearch,
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
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClientUser, nil),
		mockClientUser.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		dseSearchUser,
		mockClientUser.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockClientUser.EXPECT().Close(),
		mockClient.EXPECT().Close(),
	)

	valid, err := provider.CheckUserPassword("john", "password")

	assert.True(t, valid)
	require.NoError(t, err)
}

func TestShouldNotCheckValidUserPasswordWithConnectError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(&ldap.Error{ResultCode: ldap.LDAPResultInvalidCredentials, Err: errors.New("invalid username or password")})

	gomock.InOrder(dialURL, setTimeout, dseSearch, clientBind, mockClient.EXPECT().Close())

	valid, err := provider.CheckUserPassword("john", "password")

	assert.False(t, valid)
	assert.EqualError(t, err, "error occurred performing bind: LDAP Result Code 49 \"Invalid Credentials\": invalid username or password")
}

func TestShouldNotCheckValidUserPasswordWithGetProfileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	gomock.InOrder(
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil),
		mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		dseSearch,
		mockClient.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockClient.EXPECT().
			Search(gomock.Any()).
			Return(nil, &ldap.Error{ResultCode: ldap.LDAPResultBusy, Err: errors.New("directory server busy")}),
		mockClient.EXPECT().Close(),
	)

	valid, err := provider.CheckUserPassword("john", "password")

	assert.False(t, valid)
	assert.EqualError(t, err, "cannot find user DN of user 'john'. Cause: LDAP Result Code 51 \"Busy\": directory server busy")
}

func TestShouldCheckInvalidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	mockUserClient := NewMockLDAPClient(ctrl)

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	dseUserSearch := NewRootDSESearchRequest(mockUserClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	gomock.InOrder(
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil),
		mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		dseSearch,
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
		mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockUserClient, nil),
		mockUserClient.EXPECT().SetTimeout(gomock.Eq(time.Second*0)),
		dseUserSearch,
		mockUserClient.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(errors.New("invalid username or password")),
		mockUserClient.EXPECT().Close(),
		mockClient.EXPECT().Close(),
	)

	valid, err := provider.CheckUserPassword("john", "password")

	assert.False(t, valid)
	require.EqualError(t, err, "authentication failed. Cause: error occurred performing bind: invalid username or password")
}

func TestShouldCallStartTLSWhenEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		TLS:               schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.TLS,
		StartTLS:          true,
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	connStartTLS := mockClient.EXPECT().
		StartTLS(gomock.Any())

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	clientClose := mockClient.EXPECT().Close()

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

	gomock.InOrder(dialURL, setTimeout, dseSearch, connStartTLS, clientBind, searchProfile, searchGroups, clientClose)

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
		&schema.AuthenticationBackendLDAP{
			Address:  testLDAPAddress,
			User:     "cn=admin,dc=example,dc=com",
			Password: "password",
			Attributes: schema.AuthenticationBackendLDAPAttributes{
				Username:    "uid",
				Mail:        "mail",
				DisplayName: "displayName",
				MemberOf:    "memberOf",
			},
			UsersFilter:        "(&(|({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpiresM>={date-time:microsoft-nt})(accountExpiresU>={date-time:unix})(accountExpiresG>={date-time:generalized})))",
			GroupsFilter:       "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))",
			AdditionalUsersDN:  "ou=users",
			AdditionalGroupsDN: "ou=groups",
			BaseDN:             "dc=example,dc=com",
			StartTLS:           true,
		},
		false,
		mockFactory)

	provider.clock = clock.NewFixed(time.Unix(1670250519, 0))

	assert.True(t, provider.groupsFilterReplacementInput)
	assert.True(t, provider.groupsFilterReplacementUsername)
	assert.True(t, provider.groupsFilterReplacementDN)

	assert.True(t, provider.usersFilterReplacementInput)
	assert.True(t, provider.usersFilterReplacementDateTimeGeneralized)
	assert.True(t, provider.usersFilterReplacementDateTimeMicrosoftNTTimeEpoch)

	assert.Equal(t, "(&(|(uid={input})(mail={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpiresM>={date-time:microsoft-nt})(accountExpiresU>={date-time:unix})(accountExpiresG>={date-time:generalized})))", provider.config.UsersFilter)
	assert.Equal(t, "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))", provider.config.GroupsFilter)
	assert.Equal(t, "ou=users,dc=example,dc=com", provider.usersBaseDN)
	assert.Equal(t, "ou=groups,dc=example,dc=com", provider.groupsBaseDN)

	assert.Equal(t, "(&(|(uid=test@example.com)(mail=test@example.com))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpiresM>=133147241190000000)(accountExpiresU>=1670250519)(accountExpiresG>=20221205142839.0Z)))", provider.resolveUsersFilter("test@example.com"))
	assert.Equal(t, "(&(|(member=cn=admin,dc=example,dc=com)(member=test@example.com)(member=test))(objectClass=group))", provider.resolveGroupsFilter("test@example.com", &ldapUserProfile{Username: "test", DN: "cn=admin,dc=example,dc=com"}))
}

func TestShouldCallStartTLSWithInsecureSkipVerifyWhenSkipVerifyTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		Attributes: schema.AuthenticationBackendLDAPAttributes{
			Username:    "uid",
			Mail:        "mail",
			DisplayName: "displayName",
			MemberOf:    "memberOf",
			GroupName:   "cn",
		},
		UsersFilter:       "uid={input}",
		AdditionalUsersDN: "ou=users",
		BaseDN:            "dc=example,dc=com",
		StartTLS:          true,
		TLS: &schema.TLS{
			SkipVerify: true,
		},
	}

	mockDialer := NewMockLDAPClientDialer(ctrl)

	mockClient := NewMockLDAPClient(ctrl)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	assert.False(t, provider.groupsFilterReplacementInput)
	assert.False(t, provider.groupsFilterReplacementUsername)
	assert.False(t, provider.groupsFilterReplacementDN)

	dialURL := mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(mockClient, nil)

	setTimeout := mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

	dseSearch := NewRootDSESearchRequest(mockClient, nil)

	clientBind := mockClient.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	connStartTLS := mockClient.EXPECT().
		StartTLS(gomock.Not(gomock.Nil()))

	clientClose := mockClient.EXPECT().Close()

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
						{
							Name:   "memberOf",
							Values: []string{"CN=example,DC=corp,DC=com"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(dialURL, setTimeout, dseSearch, connStartTLS, clientBind, searchProfile, searchGroups, clientClose)

	details, err := provider.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestLDAPUserProviderChangePasswordErrors(t *testing.T) {
	testCases := []struct {
		name            string
		setupMocks      func(ctrl *gomock.Controller) (*MockLDAPClientDialer, *MockLDAPClient)
		username        string
		oldPassword     string
		newPassword     string
		expectedError   error
		expectedLogMsg  string
		expectedLogType logrus.Level
	}{
		{
			name: "ShouldFailWhenClientError",
			setupMocks: func(ctrl *gomock.Controller) (*MockLDAPClientDialer, *MockLDAPClient) {
				mockDialer := NewMockLDAPClientDialer(ctrl)
				mockClient := NewMockLDAPClient(ctrl)

				mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).
					Return(nil, errors.New("connection error"))

				return mockDialer, mockClient
			},
			username:        "john",
			oldPassword:     "oldpass",
			newPassword:     "newpass",
			expectedError:   fmt.Errorf("unable to update password for user 'john'. Cause: error occurred dialing address: connection error"),
			expectedLogMsg:  "",
			expectedLogType: 0,
		},
		{
			name: "ShouldFailWhenGetUserProfileError",
			setupMocks: func(ctrl *gomock.Controller) (*MockLDAPClientDialer, *MockLDAPClient) {
				mockDialer := NewMockLDAPClientDialer(ctrl)
				mockClient := NewMockLDAPClient(ctrl)

				mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).
					Return(mockClient, nil)

				mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

				NewRootDSESearchRequest(mockClient, nil)

				mockClient.EXPECT().
					Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
					Return(nil)

				mockClient.EXPECT().
					Search(gomock.Any()).
					Return(nil, errors.New("search error"))

				mockClient.EXPECT().Close()

				return mockDialer, mockClient
			},
			username:        "john",
			oldPassword:     "oldpass",
			newPassword:     "newpass",
			expectedError:   fmt.Errorf("unable to update password for user 'john'. Cause: cannot find user DN of user 'john'. Cause: search error"),
			expectedLogMsg:  "",
			expectedLogType: 0,
		},
		{
			name: "ShouldFailWithInvalidCredentials",
			setupMocks: func(ctrl *gomock.Controller) (*MockLDAPClientDialer, *MockLDAPClient) {
				mockDialer := NewMockLDAPClientDialer(ctrl)
				mockClient := NewMockLDAPClient(ctrl)

				mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).
					Return(mockClient, nil)

				mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

				NewRootDSESearchRequest(mockClient, nil)

				mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).
					Return(mockClient, nil)

				NewRootDSESearchRequest(mockClient, nil)

				mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

				mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).
					Return(mockClient, nil)

				NewRootDSESearchRequest(mockClient, nil)

				mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0))

				mockClient.EXPECT().Bind("cn=admin,dc=example,dc=com", "password").
					Return(nil)

				mockClient.EXPECT().Bind("cn=admin,dc=example,dc=com", "password").
					Return(nil)

				mockClient.EXPECT().Search(gomock.Any()).
					Return(&ldap.SearchResult{
						Entries: []*ldap.Entry{
							{
								DN: "uid=john,ou=users,dc=example,dc=com",
								Attributes: []*ldap.EntryAttribute{
									{
										Name:   "uid",
										Values: []string{"john"},
									},
								},
							},
						},
					}, nil)

				mockClient.EXPECT().Search(gomock.Any()).
					Return(&ldap.SearchResult{
						Entries: []*ldap.Entry{
							{
								DN: "uid=john,ou=users,dc=example,dc=com",
								Attributes: []*ldap.EntryAttribute{
									{
										Name:   "uid",
										Values: []string{"john"},
									},
								},
							},
						},
					}, nil)

				mockClient.EXPECT().Bind("uid=john,ou=users,dc=example,dc=com", "oldpass").
					Return(ldap.NewError(ldap.LDAPResultInvalidCredentials, errors.New("invalid credentials")))

				mockClient.EXPECT().Close().Times(3)

				return mockDialer, mockClient
			},
			username:        "john",
			oldPassword:     "oldpass",
			newPassword:     "newpass",
			expectedError:   ErrIncorrectPassword,
			expectedLogMsg:  "",
			expectedLogType: 0,
		},
		{
			name: "ShouldFailWhenSamePassword",
			setupMocks: func(ctrl *gomock.Controller) (*MockLDAPClientDialer, *MockLDAPClient) {
				mockDialer := NewMockLDAPClientDialer(ctrl)
				mockClient := NewMockLDAPClient(ctrl)

				mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).
					Return(mockClient, nil).Times(3)

				mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0)).Times(3)

				NewRootDSESearchRequest(mockClient, nil).Times(3)

				mockClient.EXPECT().Bind("cn=admin,dc=example,dc=com", "password").
					Return(nil).Times(2)

				mockClient.EXPECT().Search(gomock.Any()).
					Return(&ldap.SearchResult{
						Entries: []*ldap.Entry{
							{
								DN: "uid=john,ou=users,dc=example,dc=com",
								Attributes: []*ldap.EntryAttribute{
									{Name: "uid", Values: []string{"john"}},
								},
							},
						},
					}, nil).Times(2)

				mockClient.EXPECT().Bind("uid=john,ou=users,dc=example,dc=com", "samepass").
					Return(nil)

				mockClient.EXPECT().Close().Times(3)

				return mockDialer, mockClient
			},
			username:        "john",
			oldPassword:     "samepass",
			newPassword:     "samepass",
			expectedError:   ErrPasswordWeak,
			expectedLogMsg:  "",
			expectedLogType: 0,
		},
		{
			name: "ShouldFailOnModifyError",
			setupMocks: func(ctrl *gomock.Controller) (*MockLDAPClientDialer, *MockLDAPClient) {
				mockDialer := NewMockLDAPClientDialer(ctrl)
				mockClient := NewMockLDAPClient(ctrl)

				mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).
					Return(mockClient, nil).Times(3)

				mockClient.EXPECT().SetTimeout(gomock.Eq(time.Second * 0)).Times(3)

				request := ldapNewSearchRequestRootDSE()
				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{
						{
							DN: "",
							Attributes: []*ldap.EntryAttribute{
								{
									Name:   ldapSupportedExtensionAttribute,
									Values: []string{ldapOIDExtensionTLS, ldapOIDExtensionWhoAmI},
								},
								{
									Name:   ldapSupportedControlAttribute,
									Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
								},
							},
						},
					},
				}

				mockClient.EXPECT().Search(request).Return(result, nil).Times(3)

				mockClient.EXPECT().Bind("cn=admin,dc=example,dc=com", "password").
					Return(nil).Times(2)

				mockClient.EXPECT().Search(gomock.Any()).
					Return(&ldap.SearchResult{
						Entries: []*ldap.Entry{
							{
								DN: "uid=john,ou=users,dc=example,dc=com",
								Attributes: []*ldap.EntryAttribute{
									{Name: "uid", Values: []string{"john"}},
								},
							},
						},
					}, nil).Times(2)

				mockClient.EXPECT().Bind("uid=john,ou=users,dc=example,dc=com", "oldpass").
					Return(nil)

				mockClient.EXPECT().ModifyWithResult(gomock.Any()).
					Return(nil, ldap.NewError(ldap.LDAPResultConstraintViolation, errors.New("password too weak")))

				mockClient.EXPECT().Close().Times(3)

				return mockDialer, mockClient
			},
			username:        "john",
			oldPassword:     "oldpass",
			newPassword:     "newpass",
			expectedError:   fmt.Errorf("your supplied password does not meet the password policy requirements: LDAP Result Code 19 \"Constraint Violation\": password too weak"),
			expectedLogMsg:  "",
			expectedLogType: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// can't use mocks package due to circular import (mocks -> authentication, authentication(test) -> mocks).
			logger := logrus.New()
			hook := test.NewLocal(logger)
			logger.SetLevel(logrus.TraceLevel)

			mockDialer, _ := tc.setupMocks(ctrl)

			config := &schema.AuthenticationBackendLDAP{
				Address:  testLDAPAddress,
				User:     "cn=admin,dc=example,dc=com",
				Password: "password",
				Attributes: schema.AuthenticationBackendLDAPAttributes{
					Username:    "uid",
					Mail:        "mail",
					DisplayName: "displayName",
					MemberOf:    "memberOf",
				},
				UsersFilter:       "uid={input}",
				AdditionalUsersDN: "ou=users",
				BaseDN:            "dc=example,dc=com",
			}

			provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

			err := provider.ChangePassword(tc.username, tc.oldPassword, tc.newPassword)
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedLogMsg != "" {
				entry := hook.LastEntry()
				require.NotNil(t, entry)
				assert.Equal(t, tc.expectedLogType, logger.Level)
				assert.Contains(t, entry.Message, tc.expectedLogMsg)
			}
		})
	}
}

func NewRootDSESearchRequest(mockClient *MockLDAPClient, err any) *gomock.Call {
	request := ldapNewSearchRequestRootDSE()
	result := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN: "",
				Attributes: []*ldap.EntryAttribute{
					{
						Name:   ldapSupportedExtensionAttribute,
						Values: []string{ldapOIDExtensionPwdModify, ldapOIDExtensionTLS, ldapOIDExtensionWhoAmI},
					},
					{
						Name:   ldapSupportedControlAttribute,
						Values: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
					},
				},
			},
		},
	}

	return mockClient.EXPECT().Search(request).Return(result, err)
}
