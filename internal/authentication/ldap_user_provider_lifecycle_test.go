// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"errors"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func setTestLDAPStartupCheckRetryDelay(t *testing.T) {
	original := ldapStartupCheckRetryDelay

	ldapStartupCheckRetryDelay = 0

	t.Cleanup(func() {
		ldapStartupCheckRetryDelay = original
	})
}

func newTestLDAPStartupCheckConfig(pooling int) *schema.AuthenticationBackendLDAP {
	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
	}

	if pooling != 0 {
		config.Pooling = schema.AuthenticationBackendLDAPPooling{Enable: true, Count: pooling}
	}

	return config
}

func newTestLDAPStartupCheckClient(ctrl *gomock.Controller) *MockLDAPClient {
	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	return client
}

func TestLDAPUserProviderStartupCheckShouldRetryNetworkErrors(t *testing.T) {
	setTestLDAPStartupCheckRetryDelay(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := newTestLDAPStartupCheckConfig(0)

	dialer := NewMockLDAPClientDialer(ctrl)

	gomock.InOrder(
		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(nil, ldap.NewError(ldap.ErrorNetwork, errors.New("connection refused"))).Times(2),
		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(newTestLDAPStartupCheckClient(ctrl), nil).Times(1),
	)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, dialer))

	assert.NoError(t, provider.StartupCheck())
}

func TestLDAPUserProviderStartupCheckShouldNotRetryOtherErrors(t *testing.T) {
	setTestLDAPStartupCheckRetryDelay(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := newTestLDAPStartupCheckConfig(0)

	dialer := NewMockLDAPClientDialer(ctrl)

	dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(nil, errors.New("bad things")).Times(1)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, dialer))

	assert.EqualError(t, provider.StartupCheck(), "error occurred dialing address: bad things")
}

func TestLDAPUserProviderStartupCheckShouldFailAfterRetriesExhausted(t *testing.T) {
	setTestLDAPStartupCheckRetryDelay(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := newTestLDAPStartupCheckConfig(0)

	dialer := NewMockLDAPClientDialer(ctrl)

	dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(nil, ldap.NewError(ldap.ErrorNetwork, errors.New("connection refused"))).Times(19)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, dialer))

	err := provider.StartupCheck()

	require.Error(t, err)
	assert.True(t, ldap.IsErrorAnyOf(err, ldap.ErrorNetwork))
}

func TestLDAPUserProviderStartupCheckShouldRetryPooledInitialize(t *testing.T) {
	setTestLDAPStartupCheckRetryDelay(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := newTestLDAPStartupCheckConfig(2)

	dialer := NewMockLDAPClientDialer(ctrl)

	gomock.InOrder(
		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(nil, ldap.NewError(ldap.ErrorNetwork, errors.New("connection refused"))).Times(2),
		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(newTestLDAPStartupCheckClient(ctrl), nil).Times(1),
		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(newTestLDAPStartupCheckClient(ctrl), nil).Times(1),
	)

	factory := NewPooledLDAPClientFactory(config, nil, dialer)

	provider := NewLDAPUserProviderWithFactory(config, false, factory)

	assert.NoError(t, provider.StartupCheck())

	pooled, ok := factory.(*PooledLDAPClientFactory)

	require.True(t, ok)
	assert.Equal(t, 2, pooled.size)

	assert.NoError(t, provider.Close())
}

func TestPooledLDAPClientFactoryInitializeShouldBeRetryableAfterTotalFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := newTestLDAPStartupCheckConfig(2)

	dialer := NewMockLDAPClientDialer(ctrl)

	gomock.InOrder(
		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(nil, ldap.NewError(ldap.ErrorNetwork, errors.New("connection refused"))).Times(2),
		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(newTestLDAPStartupCheckClient(ctrl), nil).Times(2),
	)

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)

	err := factory.Initialize()

	require.Error(t, err)
	assert.True(t, ldap.IsErrorAnyOf(err, ldap.ErrorNetwork))
	assert.Nil(t, factory.pool)

	require.NoError(t, factory.Initialize())
	assert.Equal(t, 2, factory.size)

	assert.NoError(t, factory.Close())
}

func TestLDAPUserProviderStartupCheckShouldNotRetryPooledInitializeMixedErrors(t *testing.T) {
	testCases := []struct {
		name         string
		networkFirst bool
	}{
		{"ShouldNotRetryWhenNetworkErrorIsFirst", true},
		{"ShouldNotRetryWhenNetworkErrorIsLast", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setTestLDAPStartupCheckRetryDelay(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config := newTestLDAPStartupCheckConfig(2)

			dialer := NewMockLDAPClientDialer(ctrl)

			client := NewMockLDAPClient(ctrl)

			client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
			client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
			client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(ldap.NewError(ldap.LDAPResultInvalidCredentials, errors.New("invalid credentials"))).Times(1)
			client.EXPECT().Close().Return(nil).Times(1)

			network := dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(nil, ldap.NewError(ldap.ErrorNetwork, errors.New("connection refused"))).Times(1)
			credentials := dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(client, nil).Times(1)

			if tc.networkFirst {
				gomock.InOrder(network, credentials)
			} else {
				gomock.InOrder(credentials, network)
			}

			provider := NewLDAPUserProviderWithFactory(config, false, NewPooledLDAPClientFactory(config, nil, dialer))

			err := provider.StartupCheck()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid credentials")
		})
	}
}
