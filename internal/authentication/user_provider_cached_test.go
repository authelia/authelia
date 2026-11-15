// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestFileUserProviderShouldRetrieveCachedUserDetails(t *testing.T) {
	WithDatabase(t, UserDatabaseContent, func(path string) {
		config := DefaultFileAuthenticationBackendConfiguration
		config.Path = path

		provider := NewFileUserProvider(&config)

		require.NoError(t, provider.StartupCheck())

		details, err := provider.GetDetailsCached("john")

		require.NoError(t, err)
		assert.Equal(t, "john", details.Username)

		extended, err := provider.GetDetailsExtendedCached("john")

		require.NoError(t, err)
		assert.Equal(t, "john", extended.Username)
	})
}

func TestLDAPUserProviderShouldRetrieveCachedUserDetails(t *testing.T) {
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

	mockDialer.EXPECT().DialURL("ldap://127.0.0.1:389", gomock.Any()).Return(nil, errors.New("tcp timeout")).Times(2)

	provider := NewLDAPUserProviderWithFactory(config, false, NewStandardLDAPClientFactory(config, nil, mockDialer))

	details, err := provider.GetDetailsCached("john")

	assert.Nil(t, details)
	assert.ErrorContains(t, err, "tcp timeout")

	extended, err := provider.GetDetailsExtendedCached("john")

	assert.Nil(t, extended)
	assert.ErrorContains(t, err, "tcp timeout")
}
