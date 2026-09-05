package handlers

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2ConsentGET(t *testing.T) {
	t.Run("ShouldHandleMissingIdentifiers", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error determining the type of consent request to handle", nil)
	})

	t.Run("ShouldHandleMalformedFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?flow_id=not-a-uuid")

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing flow ID", regexpAnyError)
	})

	t.Run("ShouldHandleUnknownFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", uuid.Must(uuid.NewRandom())))

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching consent session during the Consent Flow stage of the Authorization Flow", "sql: no rows in result set")
	})

	t.Run("ShouldHandleUnknownClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", consent.ChallengeID))

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching client configuration during the Consent Flow stage of the Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleAlreadyRespondedSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
		consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", consent.ChallengeID))

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred performing consent during the Consent FLow stage of the Authorization Flow as the consent session has already been responded to", nil)
	})

	t.Run("ShouldHandleInsufficientAuthenticationLevel", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.AuthorizationPolicy = "two_factor"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", consent.ChallengeID))

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred performing consent during the Consent FLow stage of the Authorization Flow as the user is not sufficiently authenticated", nil)
	})

	t.Run("ShouldHandleExpiredSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
		consent.ExpiresAt = mock.Ctx.GetClock().Now().Add(-time.Minute)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", consent.ChallengeID))

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred performing consent during the Consent FLow stage of the Authorization Flow as the consent session has already been granted or is expired", nil)
	})

	t.Run("ShouldHandleMalformedForm", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
		consent.Form = "1238y12978y189gb128g1287g12807g128702g38172%1"

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", consent.ChallengeID))

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred getting form from consent session", `invalid URL escape "%1"`)
	})

	t.Run("ShouldReturnConsentInformation", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", consent.ChallengeID))

		OAuth2ConsentGET(mock.Ctx)

		body := map[string]any{}

		mock.GetResponseData(t, &body)

		assert.Equal(t, testOIDCAuthorizationCodeID, body["client_id"])
		assert.Contains(t, body, "scopes")
	})
}

func TestOAuth2ConsentPOST(t *testing.T) {
	t.Run("ShouldHandleMalformedBody", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.Ctx.Request.SetBodyString("not json")

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred unmarshaling consent request body", regexpAnyError)
	})

	t.Run("ShouldHandleMissingFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.Ctx.Request.SetBodyString(`{"client_id":"authorization-code","consent":true}`)

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Request is missing the required field 'flow_id' from the JSON body", nil)
	})

	t.Run("ShouldHandleMalformedFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.Ctx.Request.SetBodyString(`{"flow_id":"not-a-uuid","client_id":"authorization-code","consent":true}`)

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing flow ID as a UUID", regexpAnyError)
	})

	t.Run("ShouldHandleClientIDMismatch", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"a-different-client","consent":true}`, consent.ChallengeID))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "The client id of the form and the client id of the consent session do not match", nil)
	})

	t.Run("ShouldGrantConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":true}`, consent.ChallengeID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		redirect, err := url.Parse(body.RedirectURI)

		require.NoError(t, err)

		assert.Equal(t, "login.example.com:8080", redirect.Host)
		assert.Equal(t, oidc.EndpointPathAuthorization, redirect.Path)
		assert.Equal(t, consent.ChallengeID.String(), redirect.Query().Get(queryArgConsentID))
	})

	t.Run("ShouldRejectConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), false).
			Return(nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":false}`, consent.ChallengeID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		assert.NotEmpty(t, body.RedirectURI)
	})

	t.Run("ShouldIgnorePreConfigureForNonPreConfiguredClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":true,"pre_configure":true}`, consent.ChallengeID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		assert.NotEmpty(t, body.RedirectURI)

		AssertLogEntryMessageAndError(t, MustGetLogLastSeq(t, mock.Hook, 0), "Ignored saving pre-configuration as it is not permitted by the client configuration", nil)
	})

	t.Run("ShouldSavePreConfiguredConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "pre-configured"
		client.ConsentPreConfiguredDuration = &testOIDCPreConfiguredDuration

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentPreConfiguration(gomock.Any(), gomock.Any()).
			Return(int64(10), nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":true,"pre_configure":true}`, consent.ChallengeID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		assert.NotEmpty(t, body.RedirectURI)
		assert.Equal(t, sql.NullInt64{Int64: 10, Valid: true}, consent.PreConfiguration)
	})

	t.Run("ShouldHandlePreConfiguredConsentSaveError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "pre-configured"
		client.ConsentPreConfiguredDuration = &testOIDCPreConfiguredDuration

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentPreConfiguration(gomock.Any(), gomock.Any()).
			Return(int64(0), fmt.Errorf("failed to save"))

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":true,"pre_configure":true}`, consent.ChallengeID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		AssertLogEntryMessageAndError(t, MustGetLogLastSeq(t, mock.Hook, 0), "Error occurred saving consent pre-configuration to the database", "failed to save")
	})
}

func newTestOIDCConsentSession(t *testing.T, mock *mocks.MockAutheliaCtx, subject uuid.UUID) *model.OAuth2ConsentSession {
	t.Helper()

	form := url.Values{
		oidc.FormParameterClientID:     []string{testOIDCAuthorizationCodeID},
		oidc.FormParameterResponseType: []string{oidc.ResponseTypeAuthorizationCodeFlow},
		oidc.FormParameterRedirectURI:  []string{testOIDCRedirectURI},
		oidc.FormParameterScope:        []string{oidc.ScopeOpenID},
		oidc.FormParameterState:        []string{"abcdefghijklmnopqrstuvwxyz"},
	}

	return &model.OAuth2ConsentSession{
		ID:              1,
		ChallengeID:     uuid.Must(uuid.NewRandom()),
		ClientID:        testOIDCAuthorizationCodeID,
		Subject:         uuid.NullUUID{UUID: subject, Valid: subject != uuid.Nil},
		RequestedAt:     mock.Ctx.GetClock().Now().Add(-time.Minute),
		ExpiresAt:       mock.Ctx.GetClock().Now().Add(time.Hour),
		Form:            form.Encode(),
		RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
	}
}
