package authentication

// This file is used to generate mocks. You can generate all mocks using the
// command `go generate github.com/authelia/authelia/v4/internal/authentication`.

//go:generate mockgen -package authentication -destination ldap_client_mock.go -mock_names Client=MockLDAPClient github.com/go-ldap/ldap/v3 Client
//go:generate mockgen -package authentication -destination ldap_client_factory_mock.go -mock_names LDAPClientFactory=MockLDAPClientFactory github.com/authelia/authelia/v4/internal/authentication LDAPClientFactory
