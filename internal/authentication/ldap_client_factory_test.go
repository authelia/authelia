package authentication

import (
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestPooledLDAPClientFactoryShouldNotDialPerRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 2},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	for range config.Pooling.Count {
		client := NewMockLDAPClient(ctrl)

		client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
		client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
		client.EXPECT().Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).Return(nil).Times(1)
		client.EXPECT().IsClosing().Return(false).AnyTimes()
		client.EXPECT().Close().Return(nil).AnyTimes()

		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(client, nil).Times(1)
	}

	factory := NewPooledLDAPClientFactory(config, nil, dialer)

	require.NoError(t, factory.Initialize())

	for range 10 {
		acquired, err := factory.GetClient(WithPermitUnauthenticatedBind(config.PermitUnauthenticatedBind))

		require.NoError(t, err)
		require.IsType(t, &PooledLDAPClient{}, acquired)
		require.NoError(t, factory.ReleaseClient(acquired))
	}

	require.NoError(t, factory.Close())
}

func TestPooledLDAPClientFactoryGetClient(t *testing.T) {
	testCases := []struct {
		Name   string
		Opts   []LDAPClientFactoryOption
		Pooled bool
	}{
		{
			Name:   "ShouldPoolWithoutOptions",
			Opts:   nil,
			Pooled: true,
		},
		{
			Name:   "ShouldPoolWithPermitUnauthenticatedBind",
			Opts:   []LDAPClientFactoryOption{WithPermitUnauthenticatedBind(false)},
			Pooled: true,
		},
		{
			Name:   "ShouldPoolWithPermitUnauthenticatedBindEnabled",
			Opts:   []LDAPClientFactoryOption{WithPermitUnauthenticatedBind(true)},
			Pooled: true,
		},
		{
			Name:   "ShouldPoolWithMatchingAddress",
			Opts:   []LDAPClientFactoryOption{WithAddress(testLDAPURL)},
			Pooled: true,
		},
		{
			Name:   "ShouldNotPoolWithReferralAddress",
			Opts:   []LDAPClientFactoryOption{WithAddress("ldap://127.0.0.1:390"), WithPermitUnauthenticatedBind(false)},
			Pooled: false,
		},
		{
			Name:   "ShouldNotPoolWithUserCredentials",
			Opts:   []LDAPClientFactoryOption{WithUsername("uid=john,ou=users,dc=example,dc=com"), WithPassword("userpassword")},
			Pooled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config := &schema.AuthenticationBackendLDAP{
				Address:  testLDAPAddress,
				User:     "cn=admin,dc=example,dc=com",
				Password: "password",
				BaseDN:   "dc=example,dc=com",
				Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 1},
			}

			dialer := NewMockLDAPClientDialer(ctrl)
			client := NewMockLDAPClient(ctrl)

			dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(client, nil).Times(1)
			client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
			client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
			client.EXPECT().Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).Return(nil).Times(1)
			client.EXPECT().IsClosing().Return(false).AnyTimes()
			client.EXPECT().Close().Return(nil).AnyTimes()

			factory := NewPooledLDAPClientFactory(config, nil, dialer)

			require.NoError(t, factory.Initialize())

			if !tc.Pooled {
				unpooled := NewMockLDAPClient(ctrl)

				dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(unpooled, nil).Times(1)
				unpooled.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
				unpooled.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
				unpooled.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				unpooled.EXPECT().Close().Return(nil).Times(1)
			}

			acquired, err := factory.GetClient(tc.Opts...)

			require.NoError(t, err)
			require.NotNil(t, acquired)

			_, pooled := acquired.(*PooledLDAPClient)

			assert.Equal(t, tc.Pooled, pooled)

			require.NoError(t, factory.ReleaseClient(acquired))
			require.NoError(t, factory.Close())
		})
	}
}

func testLDAPRootDSESearchResult() *ldap.SearchResult {
	return &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN: "",
				Attributes: []*ldap.EntryAttribute{
					{Name: ldapSupportedLDAPVersionAttribute, Values: []string{"3"}},
				},
			},
		},
	}
}
