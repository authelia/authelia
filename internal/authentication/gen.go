package authentication

// This file is used to generate mocks. You can generate all mocks using the
// command `go generate github.com/authelia/authelia/v4/internal/authentication`.

//go:generate mockgen -package authentication -destination ldap_client_mock_test.go -mock_names LDAPExtendedClient=MockLDAPClient github.com/authelia/authelia/v4/internal/authentication LDAPExtendedClient
//go:generate mockgen -package authentication -destination ldap_client_dialer_mock_test.go -mock_names LDAPClientDialer=MockLDAPClientDialer github.com/authelia/authelia/v4/internal/authentication LDAPClientDialer
//go:generate mockgen -package authentication -destination ldap_client_factory_mock_test.go -mock_names LDAPClientFactory=MockLDAPClientFactory github.com/authelia/authelia/v4/internal/authentication LDAPClientFactory
//go:generate mockgen -package authentication -destination file_user_provider_database_mock_test.go -mock_names FileUserProviderDatabase=MockFileUserDatabase github.com/authelia/authelia/v4/internal/authentication FileUserProviderDatabase
//go:generate mockgen -package authentication -destination file_user_provider_hash_mock_test.go -mock_names Hash=MockHash github.com/go-crypt/crypt/algorithm Hash
