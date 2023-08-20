package mocks

// This file is used to generate mocks. You can generate all mocks using the
// command `go generate github.com/authelia/authelia/v4/internal/mocks`.

//go:generate mockgen -package mocks -destination user_provider.go -mock_names UserProvider=MockUserProvider github.com/authelia/authelia/v4/internal/authentication UserProvider
//go:generate mockgen -package mocks -destination notifier.go -mock_names Notifier=MockNotifier github.com/authelia/authelia/v4/internal/notification Notifier
//go:generate mockgen -package mocks -destination totp.go -mock_names Provider=MockTOTP github.com/authelia/authelia/v4/internal/totp Provider
//go:generate mockgen -package mocks -destination storage.go -mock_names Provider=MockStorage github.com/authelia/authelia/v4/internal/storage Provider
//go:generate mockgen -package mocks -destination duo_api.go -mock_names API=MockAPI github.com/authelia/authelia/v4/internal/duo API
//go:generate mockgen -package mocks -destination random.go -mock_names Provider=MockRandom github.com/authelia/authelia/v4/internal/random Provider

// Fosite Mocks.
//go:generate mockgen -package mocks -destination fosite_client_credentials_grant_storage.go -mock_names Provider=MockClientCredentialsGrantStorage github.com/ory/fosite/handler/oauth2 ClientCredentialsGrantStorage
//go:generate mockgen -package mocks -destination fosite_token_revocation_storage.go -mock_names Provider=MockTokenRevocationStorage github.com/ory/fosite/handler/oauth2 TokenRevocationStorage
//go:generate mockgen -package mocks -destination fosite_access_token_strategy.go -mock_names Provider=MockAccessTokenStrategy github.com/ory/fosite/handler/oauth2 AccessTokenStrategy

//go:generate mockgen -package mocks -destination fosite_pkce_request_storage.go -mock_names Provider=MockPKCERequestStorage github.com/ory/fosite/handler/pkce PKCERequestStorage

//go:generate mockgen -package mocks -destination fosite_access_requester.go -mock_names Provider=MockAccessRequester github.com/ory/fosite AccessRequester

//go:generate mockgen -package mocks -destination fosite_transactional.go -mock_names Provider=MockTransactional github.com/ory/fosite/storage Transactional

//go:generate mockgen -package mocks -destination fosite_token_introspector.go -mock_names TokenIntrospector=MockTokenIntrospector github.com/ory/fosite TokenIntrospector
