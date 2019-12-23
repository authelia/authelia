package authentication

import (
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
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
