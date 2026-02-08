package mocks

// This file is used to generate mocks. You can generate all mocks using the
// command `go generate github.com/authelia/authelia/v4/internal/mocks`.

//go:generate mockgen -package mocks -destination user_provider.go -mock_names UserProvider=MockUserProvider github.com/authelia/authelia/v4/internal/authentication UserProvider
//go:generate mockgen -package mocks -destination notifier.go -mock_names Notifier=MockNotifier github.com/authelia/authelia/v4/internal/notification Notifier
//go:generate mockgen -package mocks -destination totp.go -mock_names Provider=MockTOTP github.com/authelia/authelia/v4/internal/totp Provider
//go:generate mockgen -package mocks -destination storage.go -mock_names Provider=MockStorage github.com/authelia/authelia/v4/internal/storage Provider
//go:generate mockgen -package mocks -destination duo_api.go -mock_names Provider=MockDuoProvider github.com/authelia/authelia/v4/internal/duo Provider
//go:generate mockgen -package mocks -destination duo_base_api.go -mock_names BaseProvider=MockDuoBaseProvider github.com/authelia/authelia/v4/internal/duo BaseProvider
//go:generate mockgen -package mocks -destination random.go -mock_names Provider=MockRandom github.com/authelia/authelia/v4/internal/random Provider

// External Mocks.

// Mocks for authelia.com/provider/oauth2.
//go:generate mockgen -package mocks -destination oauth2_client_credentials_grant_storage.go -mock_names Provider=MockClientCredentialsGrantStorage authelia.com/provider/oauth2/handler/oauth2 ClientCredentialsGrantStorage
//go:generate mockgen -package mocks -destination oauth2_token_revocation_storage.go -mock_names Provider=MockTokenRevocationStorage authelia.com/provider/oauth2/handler/oauth2 TokenRevocationStorage
//go:generate mockgen -package mocks -destination oauth2_access_token_strategy.go -mock_names Provider=MockAccessTokenStrategy authelia.com/provider/oauth2/handler/oauth2 AccessTokenStrategy

//go:generate mockgen -package mocks -destination oauth2_pkce_request_storage.go -mock_names Storage=MockPKCERequestStorage authelia.com/provider/oauth2/handler/pkce Storage

//go:generate mockgen -package mocks -destination oauth2_access_requester.go -mock_names Provider=MockAccessRequester authelia.com/provider/oauth2 AccessRequester

//go:generate mockgen -package mocks -destination oauth2_transactional.go -mock_names Provider=MockTransactional authelia.com/provider/oauth2/storage Transactional

//go:generate mockgen -package mocks -destination oauth2_token_introspector.go -mock_names TokenIntrospector=MockTokenIntrospector authelia.com/provider/oauth2 TokenIntrospector
//go:generate mockgen -package mocks -destination oauth2_storage.go -mock_names Storage=MockOAuth2Storage authelia.com/provider/oauth2 Storage
