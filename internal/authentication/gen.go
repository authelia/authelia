package authentication

// This file is used to generate mocks. You can generate all mocks using the
// command `go generate github.com/authelia/authelia/v4/internal/authentication`.

//go:generate mockgen -package authentication -destination ldap_connection_mock.go -mock_names LDAPConnection=MockLDAPConnection github.com/authelia/authelia/v4/internal/authentication LDAPConnection
//go:generate mockgen -package authentication -destination ldap_connection_factory_mock.go -mock_names LDAPConnectionFactory=MockLDAPConnectionFactory github.com/authelia/authelia/v4/internal/authentication LDAPConnectionFactory
