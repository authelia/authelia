// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

const (
	tRelyingPartyID        = "login.example.com"
	tRelyingPartyHost      = "login.example.com:8080"
	tRelyingPartyOrigin    = "https://login.example.com:8080"
	tRelatedOriginHost     = "auth.example2.com"
	tRelatedOrigin         = "https://auth.example2.com"
	tRelyingPartyChallenge = "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE"
	tRelyingPartyUserID    = "example"
	tWebAuthnKIDLength     = 32
)

var tRelyingPartyAAGUID = uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708"))

type tWebAuthnAuthenticator struct {
	key *ecdsa.PrivateKey
	kid []byte
}

func newTWebAuthnAuthenticator(t *testing.T) *tWebAuthnAuthenticator {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	kid := make([]byte, tWebAuthnKIDLength)

	_, err = rand.Read(kid)
	require.NoError(t, err)

	return &tWebAuthnAuthenticator{key: key, kid: kid}
}

func (a *tWebAuthnAuthenticator) PublicKey(t *testing.T) []byte {
	t.Helper()

	public, err := a.key.PublicKey.ECDH()
	require.NoError(t, err)

	raw := public.Bytes()

	data, err := webauthncbor.Marshal(webauthncose.EC2PublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			KeyType:   int64(webauthncose.EllipticKey),
			Algorithm: int64(webauthncose.AlgES256),
		},
		Curve:  int64(webauthncose.P256),
		XCoord: raw[1:33],
		YCoord: raw[33:],
	})
	require.NoError(t, err)

	return data
}

func (a *tWebAuthnAuthenticator) Credential(t *testing.T, rpid string, attestationType string) model.WebAuthnCredential {
	t.Helper()

	return model.WebAuthnCredential{
		ID:              1,
		CreatedAt:       time.Now(),
		RPID:            rpid,
		Username:        testUsername,
		Description:     "test",
		KID:             model.NewBase64(a.kid),
		AAGUID:          uuid.NullUUID{UUID: tRelyingPartyAAGUID, Valid: true},
		AttestationType: attestationType,
		Attachment:      "cross-platform",
		Transport:       "usb",
		Present:         true,
		Verified:        true,
		PublicKey:       a.PublicKey(t),
	}
}

func (a *tWebAuthnAuthenticator) Attestation(t *testing.T, rpid, origin, challenge string) string {
	t.Helper()

	hash := sha256.Sum256([]byte(rpid))

	authData := append([]byte{}, hash[:]...)
	authData = append(authData, 0x45)
	authData = binary.BigEndian.AppendUint32(authData, 0)
	authData = append(authData, tRelyingPartyAAGUID[:]...)
	authData = binary.BigEndian.AppendUint16(authData, tWebAuthnKIDLength)
	authData = append(authData, a.kid...)
	authData = append(authData, a.PublicKey(t)...)

	attestation, err := webauthncbor.Marshal(struct {
		Format       string         `cbor:"fmt"`
		AttStatement map[string]any `cbor:"attStmt"`
		AuthData     []byte         `cbor:"authData"`
	}{Format: "none", AttStatement: map[string]any{}, AuthData: authData})
	require.NoError(t, err)

	clientData := fmt.Sprintf(`{"type":"webauthn.create","challenge":"%s","origin":"%s","crossOrigin":false}`, challenge, origin)

	id := base64.RawURLEncoding.EncodeToString(a.kid)

	return fmt.Sprintf(`{"id":"%s","rawId":"%s","response":{"attestationObject":"%s","clientDataJSON":"%s","transports":["usb"]},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`,
		id, id, base64.RawURLEncoding.EncodeToString(attestation), base64.RawURLEncoding.EncodeToString([]byte(clientData)))
}

func (a *tWebAuthnAuthenticator) Assertion(t *testing.T, rpid, origin, challenge string, userHandle []byte) string {
	t.Helper()

	hash := sha256.Sum256([]byte(rpid))

	authData := append([]byte{}, hash[:]...)
	authData = append(authData, 0x05)
	authData = binary.BigEndian.AppendUint32(authData, 1)

	clientData := fmt.Appendf(nil, `{"type":"webauthn.get","challenge":"%s","origin":"%s","crossOrigin":false}`, challenge, origin)
	clientDataHash := sha256.Sum256(clientData)

	digest := sha256.Sum256(append(append([]byte{}, authData...), clientDataHash[:]...))

	signature, err := ecdsa.SignASN1(rand.Reader, a.key, digest[:])
	require.NoError(t, err)

	handle := ""

	if userHandle != nil {
		handle = fmt.Sprintf(`,"userHandle":"%s"`, base64.RawURLEncoding.EncodeToString(userHandle))
	}

	id := base64.RawURLEncoding.EncodeToString(a.kid)

	return fmt.Sprintf(`{"response":{"id":"%s","rawId":"%s","response":{"authenticatorData":"%s","clientDataJSON":"%s","signature":"%s"%s},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`,
		id, id, base64.RawURLEncoding.EncodeToString(authData), base64.RawURLEncoding.EncodeToString(clientData), base64.RawURLEncoding.EncodeToString(signature), handle)
}

func tWebAuthnRelyingPartyConfig(modify func(party *schema.WebAuthnRelyingParty), origins ...string) *schema.WebAuthn {
	config := schema.DefaultWebAuthnConfiguration

	party := schema.WebAuthnRelyingParty{WebAuthnBase: config.WebAuthnBase}

	for _, origin := range origins {
		party.Origins = append(party.Origins, &url.URL{Scheme: "https", Host: origin})
	}

	if modify != nil {
		modify(&party)
	}

	config.RelyingParties = map[string]schema.WebAuthnRelyingParty{tRelyingPartyID: party}

	return &config
}

func tSetupWebAuthnRelyingPartyUser(t *testing.T, mock *mocks.MockAutheliaCtx, data *session.WebAuthn) {
	t.Helper()

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.Username = testUsername
	us.AuthenticationMethodRefs.UsernameAndPassword = true
	us.WebAuthn = data

	require.NoError(t, mock.Ctx.SaveSession(&us))
}

func tWebAuthnRelyingPartyUser() *model.WebAuthnUser {
	return &model.WebAuthnUser{ID: 1, RPID: tRelyingPartyID, Username: testUsername, UserID: tRelyingPartyUserID}
}

func TestWebAuthnRelyingPartyRegistrationPUT(t *testing.T) {
	testCases := []struct {
		name           string
		host           string
		config         *schema.WebAuthn
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       *regexp.Regexp
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldUseRelyingPartyIDOnRelatedOrigin",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost),
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
						Return(tWebAuthnRelyingPartyUser(), nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return(nil, nil),
				)
			},
			regexp.MustCompile(`^\{"status":"OK","data":\{"publicKey":\{"rp":\{"name":"Authelia","id":"login\.example\.com"}`),
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldUseRelyingPartyDisplayName",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(func(party *schema.WebAuthnRelyingParty) {
				party.DisplayName = "Example"
			}, tRelyingPartyHost, tRelatedOriginHost),
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
						Return(tWebAuthnRelyingPartyUser(), nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return(nil, nil),
				)
			},
			regexp.MustCompile(`^\{"status":"OK","data":\{"publicKey":\{"rp":\{"name":"Example","id":"login\.example\.com"}`),
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldFailUnlistedOrigin",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost),
			nil,
			regexp.MustCompile(`^\{"status":"KO","code":"webauthn_register_failed","message":"Unable to register your security key."}$`),
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge for user 'john': error occurred provisioning the configuration", "error occurred finding the relying party: no related origin found for origin 'https://auth.example2.com'")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.WebAuthn = *tc.config

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)
			mock.Ctx.Request.SetBodyString(`{"description":"test"}`)

			tSetupWebAuthnRelyingPartyUser(t, mock, nil)

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnRegistrationPUT(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestWebAuthnRelyingPartyRegistrationPOST(t *testing.T) {
	authenticator := newTWebAuthnAuthenticator(t)

	testCases := []struct {
		name           string
		host           string
		config         *schema.WebAuthn
		have           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldRegisterOnRelatedOrigin",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost),
			authenticator.Attestation(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge),
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
						Return(tWebAuthnRelyingPartyUser(), nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return(nil, nil),
					mock.StorageMock.
						EXPECT().
						SaveWebAuthnCredential(mock.Ctx, gomock.Any()).
						DoAndReturn(func(_ any, credential model.WebAuthnCredential) error {
							assert.Equal(t, tRelyingPartyID, credential.RPID)
							assert.Equal(t, authenticator.kid, credential.KID.Bytes())

							return nil
						}),
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.NotifierMock.
						EXPECT().
						Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: testEmail}, "Second Factor Method Added", gomock.Any(), gomock.Any()).
						Return(nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusCreated,
			nil,
		},
		{
			"ShouldFailOriginNotInRelyingParty",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost),
			authenticator.Attestation(t, tRelyingPartyID, "https://evil.example.net", tRelyingPartyChallenge),
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
						Return(tWebAuthnRelyingPartyUser(), nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return(nil, nil),
				)
			},
			`{"status":"KO","code":"webauthn_register_failed","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, "Error occurred validating a WebAuthn registration challenge for user 'john': error comparing the response to the WebAuthn session data", mock.Hook.LastEntry().Message)
				assert.Contains(t, fmt.Sprint(mock.Hook.LastEntry().Data["error"]), "https://evil.example.net")
			},
		},
		{
			"ShouldFailRelyingPartyIDOfAnotherRelyingParty",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost),
			authenticator.Attestation(t, tRelatedOriginHost, tRelatedOrigin, tRelyingPartyChallenge),
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
						Return(tWebAuthnRelyingPartyUser(), nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return(nil, nil),
				)
			},
			`{"status":"KO","code":"webauthn_register_failed","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, "Error occurred validating a WebAuthn registration challenge for user 'john': error comparing the response to the WebAuthn session data", mock.Hook.LastEntry().Message)
			},
		},
		{
			"ShouldApplyRelyingPartyFiltering",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(func(party *schema.WebAuthnRelyingParty) {
				party.Filtering.ProhibitedAAGUIDs = []uuid.UUID{tRelyingPartyAAGUID}
			}, tRelyingPartyHost, tRelatedOriginHost),
			authenticator.Attestation(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge),
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
						Return(tWebAuthnRelyingPartyUser(), nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return(nil, nil),
				)
			},
			`{"status":"KO","code":"webauthn_register_failed","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error occurred processing the credential filtering", "error checking webauthn credential: filters have been configured which prohibit the AAGUID '01020304-0506-0708-0102-030405060708' from registration")
			},
		},
		{
			"ShouldNotApplyGlobalFilteringToRelyingParty",
			tRelatedOriginHost,
			func() *schema.WebAuthn {
				config := tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost)
				config.Filtering.ProhibitedAAGUIDs = []uuid.UUID{tRelyingPartyAAGUID}

				return config
			}(),
			authenticator.Attestation(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge),
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
						Return(tWebAuthnRelyingPartyUser(), nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return(nil, nil),
					mock.StorageMock.
						EXPECT().
						SaveWebAuthnCredential(mock.Ctx, gomock.Any()).
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetailsCached(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.NotifierMock.
						EXPECT().
						Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: testEmail}, "Second Factor Method Added", gomock.Any(), gomock.Any()).
						Return(nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusCreated,
			nil,
		},
		{
			"ShouldFailUnlistedOrigin",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost),
			authenticator.Attestation(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge),
			nil,
			`{"status":"KO","code":"webauthn_register_failed","message":"Unable to register your security key."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error occurred provisioning the configuration", "error occurred finding the relying party: no related origin found for origin 'https://auth.example2.com'")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.WebAuthn = *tc.config

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)
			mock.Ctx.Request.SetBodyString(tc.have)

			tSetupWebAuthnRelyingPartyUser(t, mock, &session.WebAuthn{
				Description: "test",
				SessionData: &webauthn.SessionData{
					Challenge:        tRelyingPartyChallenge,
					RelyingPartyID:   tRelyingPartyID,
					UserID:           []byte(tRelyingPartyUserID),
					Expires:          time.Now().Add(time.Minute),
					UserVerification: protocol.VerificationPreferred,
					CredParams:       []protocol.CredentialParameter{{Type: protocol.PublicKeyCredentialType, Algorithm: webauthncose.AlgES256}},
					Extensions:       protocol.SessionExtensions{Requested: []string{protocol.ExtensionCredProps}},
				},
			})

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnRegistrationPOST(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestWebAuthnRelyingPartyAssertionGET(t *testing.T) {
	authenticator := newTWebAuthnAuthenticator(t)

	testCases := []struct {
		name            string
		host            string
		config          *schema.WebAuthn
		attestationType string
		expected        *regexp.Regexp
		expectedStatus  int
		expectedf       func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldUseRelyingPartyIDOnRelatedOrigin",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost),
			"packed",
			regexp.MustCompile(`"rpId":"login\.example\.com"`),
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.NotContains(t, string(mock.Ctx.Response.Body()), `"appid"`)
			},
		},
		{
			"ShouldUseRequestOriginForAppIDOnRelatedOrigin",
			tRelatedOriginHost,
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost),
			"fido-u2f",
			regexp.MustCompile(`"extensions":\{"appid":"https://auth\.example2\.com"}`),
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldUseRequestOriginForAppIDOnRelyingPartyOrigin",
			tRelyingPartyHost,
			tWebAuthnRelyingPartyConfig(nil, tRelatedOriginHost, tRelyingPartyHost),
			"fido-u2f",
			regexp.MustCompile(`"extensions":\{"appid":"https://login\.example\.com:8080"}`),
			fasthttp.StatusOK,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.WebAuthn = *tc.config

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)

			tSetupWebAuthnRelyingPartyUser(t, mock, nil)

			credential := authenticator.Credential(t, tRelyingPartyID, tc.attestationType)
			credential.AttestationFormat = tc.attestationType

			gomock.InOrder(
				mock.UserProviderMock.
					EXPECT().
					GetDetails(testUsername).
					Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
				mock.StorageMock.
					EXPECT().
					LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
					Return(tWebAuthnRelyingPartyUser(), nil),
				mock.StorageMock.
					EXPECT().
					LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
					Return([]model.WebAuthnCredential{credential}, nil),
			)

			WebAuthnAssertionGET(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestWebAuthnRelyingPartyAssertionGETShouldFailUnlistedOrigin(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.WebAuthn = *tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tRelatedOriginHost)

	tSetupWebAuthnRelyingPartyUser(t, mock, nil)

	WebAuthnAssertionGET(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"KO","code":"mfa_validation_failed","message":"Authentication failed, please retry later."}`, string(mock.Ctx.Response.Body()))

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn authentication challenge for user 'john': error occurred provisioning the configuration", "error occurred finding the relying party: no related origin found for origin 'https://auth.example2.com'")
}

func TestWebAuthnRelyingPartyAssertionPOST(t *testing.T) {
	authenticator := newTWebAuthnAuthenticator(t)

	testCases := []struct {
		name            string
		host            string
		have            string
		credentialRPID  string
		attestationType string
		expectSuccess   bool
		expectedRPID    string
	}{
		{
			"ShouldAssertOnRelatedOriginWithCredentialRegisteredOnRelyingPartyOrigin",
			tRelatedOriginHost,
			authenticator.Assertion(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge, nil),
			tRelyingPartyID,
			"none",
			true,
			tRelyingPartyID,
		},
		{
			"ShouldAssertOnRelyingPartyOrigin",
			tRelyingPartyHost,
			authenticator.Assertion(t, tRelyingPartyID, tRelyingPartyOrigin, tRelyingPartyChallenge, nil),
			tRelyingPartyID,
			"none",
			true,
			tRelyingPartyID,
		},
		{
			"ShouldSetLegacyCredentialRPIDToRelyingPartyID",
			tRelatedOriginHost,
			authenticator.Assertion(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge, nil),
			"",
			"none",
			true,
			tRelyingPartyID,
		},
		{
			"ShouldSetLegacyFIDOU2FCredentialRPIDToRequestOrigin",
			tRelatedOriginHost,
			authenticator.Assertion(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge, nil),
			"",
			"fido-u2f",
			true,
			tRelatedOrigin,
		},
		{
			"ShouldFailClientDataOriginNotInRelyingParty",
			tRelatedOriginHost,
			authenticator.Assertion(t, tRelyingPartyID, "https://evil.example.net", tRelyingPartyChallenge, nil),
			tRelyingPartyID,
			"none",
			false,
			"",
		},
		{
			"ShouldFailAuthenticatorDataForAnotherRelyingPartyID",
			tRelatedOriginHost,
			authenticator.Assertion(t, tRelatedOriginHost, tRelatedOrigin, tRelyingPartyChallenge, nil),
			tRelyingPartyID,
			"none",
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.WebAuthn = *tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)
			mock.Ctx.Request.SetBodyString(tc.have)

			tSetupWebAuthnRelyingPartyUser(t, mock, &session.WebAuthn{
				SessionData: &webauthn.SessionData{
					Challenge:        tRelyingPartyChallenge,
					RelyingPartyID:   tRelyingPartyID,
					UserID:           []byte(tRelyingPartyUserID),
					Expires:          time.Now().Add(time.Minute),
					UserVerification: protocol.VerificationPreferred,
				},
			})

			credential := authenticator.Credential(t, tc.credentialRPID, tc.attestationType)

			calls := []any{
				mock.UserProviderMock.
					EXPECT().
					GetDetails(testUsername).
					Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
				mock.StorageMock.
					EXPECT().
					LoadWebAuthnUser(mock.Ctx, tRelyingPartyID, testUsername).
					Return(tWebAuthnRelyingPartyUser(), nil),
				mock.StorageMock.
					EXPECT().
					LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
					Return([]model.WebAuthnCredential{credential}, nil),
			}

			if tc.expectSuccess {
				calls = append(calls,
					mock.StorageMock.
						EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, gomock.Any()).
						DoAndReturn(func(_ any, updated model.WebAuthnCredential) error {
							assert.Equal(t, tc.expectedRPID, updated.RPID)
							assert.Equal(t, uint32(1), updated.SignCount)

							return nil
						}),
				)
			}

			calls = append(calls,
				mock.StorageMock.
					EXPECT().
					AppendAuthenticationLog(mock.Ctx, gomock.Any()).
					DoAndReturn(func(_ any, attempt model.AuthenticationAttempt) error {
						assert.Equal(t, tc.expectSuccess, attempt.Successful)

						return nil
					}),
			)

			gomock.InOrder(calls...)

			WebAuthnAssertionPOST(mock.Ctx)

			if tc.expectSuccess {
				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
				assert.Regexp(t, regexp.MustCompile(`^\{"status":"OK"`), string(mock.Ctx.Response.Body()))
			} else {
				assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
				assert.Equal(t, `{"status":"KO","code":"mfa_validation_failed","message":"Authentication failed, please retry later."}`, string(mock.Ctx.Response.Body()))
			}
		})
	}
}

func TestWebAuthnRelyingPartyCredentialPUT(t *testing.T) {
	testCases := []struct {
		name           string
		host           string
		credentialRPID string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldModifyCredentialOfRelyingPartyOnRelatedOrigin",
			tRelatedOriginHost,
			tRelyingPartyID,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
						Return([]model.WebAuthnCredential{{ID: 1, RPID: tRelyingPartyID, Username: testUsername}}, nil),
					mock.StorageMock.
						EXPECT().
						UpdateWebAuthnCredentialDescription(mock.Ctx, testUsername, 1, "abc").
						Return(nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldFailCredentialOfHostnameRPIDOnRelatedOrigin",
			tRelatedOriginHost,
			tRelatedOriginHost,
			nil,
			`{"status":"KO","code":"operation_failed","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred modifying WebAuthn credential for user 'john'", "the credential with id '1' belongs to the relying party 'auth.example2.com' but the request is for the relying party 'login.example.com'")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.WebAuthn = *tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)
			mock.Ctx.SetUserValue("credentialID", "1")
			mock.Ctx.Request.SetBodyString(`{"description":"abc"}`)

			tSetupWebAuthnRelyingPartyUser(t, mock, nil)

			mock.StorageMock.
				EXPECT().
				LoadWebAuthnCredentialByID(mock.Ctx, 1).
				Return(&model.WebAuthnCredential{ID: 1, RPID: tc.credentialRPID, Username: testUsername}, nil)

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnCredentialPUT(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestWebAuthnRelyingPartyFirstFactorPasskeyGET(t *testing.T) {
	testCases := []struct {
		name           string
		config         *schema.WebAuthn
		expected       *regexp.Regexp
		expectedStatus int
	}{
		{
			"ShouldUseRelyingPartyIDOnRelatedOrigin",
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost),
			regexp.MustCompile(`"rpId":"login\.example\.com"`),
			fasthttp.StatusOK,
		},
		{
			"ShouldFailUnlistedOrigin",
			tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost),
			regexp.MustCompile(`^\{"status":"KO","code":"mfa_validation_failed","message":"Authentication failed, please retry later."}$`),
			fasthttp.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.WebAuthn = *tc.config
			mock.Ctx.Configuration.WebAuthn.EnablePasskeyLogin = true

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tRelatedOriginHost)

			FirstFactorPasskeyGET(false)(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))
		})
	}
}

func TestWebAuthnRelyingPartyFirstFactorPasskeyPOST(t *testing.T) {
	authenticator := newTWebAuthnAuthenticator(t)

	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.WebAuthn = *tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost, tRelatedOriginHost)
	mock.Ctx.Configuration.WebAuthn.EnablePasskeyLogin = true

	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tRelatedOriginHost)
	mock.Ctx.Request.SetBodyString(authenticator.Assertion(t, tRelyingPartyID, tRelatedOrigin, tRelyingPartyChallenge, []byte(tRelyingPartyUserID)))

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.WebAuthn = &session.WebAuthn{
		SessionData: &webauthn.SessionData{
			Challenge:        tRelyingPartyChallenge,
			RelyingPartyID:   tRelyingPartyID,
			Expires:          time.Now().Add(time.Minute),
			UserVerification: protocol.VerificationPreferred,
		},
	}

	require.NoError(t, mock.Ctx.SaveSession(&us))

	credential := authenticator.Credential(t, tRelyingPartyID, "none")
	credential.Discoverable = true

	gomock.InOrder(
		mock.StorageMock.EXPECT().
			LoadWebAuthnUserByUserID(mock.Ctx, tRelyingPartyID, tRelyingPartyUserID).
			Return(&model.WebAuthnUser{RPID: tRelyingPartyID, UserID: tRelyingPartyUserID, Username: testUsername}, nil),
		mock.StorageMock.EXPECT().
			LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, tRelyingPartyID, testUsername).
			Return([]model.WebAuthnCredential{credential}, nil),
		mock.StorageMock.EXPECT().
			UpdateWebAuthnCredentialSignIn(mock.Ctx, gomock.Any()).
			DoAndReturn(func(_ any, updated model.WebAuthnCredential) error {
				assert.Equal(t, tRelyingPartyID, updated.RPID)

				return nil
			}),
		mock.UserProviderMock.EXPECT().
			GetDetails(testUsername).
			Return(&authentication.UserDetails{Username: testUsername}, nil),
		mock.StorageMock.EXPECT().
			LoadBannedIP(mock.Ctx, model.NewIP(mock.Ctx.RemoteIP())).
			Return(nil, nil),
		mock.StorageMock.EXPECT().
			LoadBannedUser(mock.Ctx, testUsername).
			Return(nil, nil),
		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(mock.Ctx, gomock.Any()).
			Return(nil),
		mock.StorageMock.EXPECT().
			UpdateUserSignInDateByUsername(mock.Ctx, testUsername).
			Return(nil),
	)

	FirstFactorPasskeyPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))
}

func TestWebAuthnRelyingPartyUserInfoPOSTShouldNotChangePreferenceOnUnlistedOrigin(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.WebAuthn = *tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tRelatedOriginHost)

	tSetupWebAuthnRelyingPartyUser(t, mock, nil)

	mock.StorageMock.
		EXPECT().
		CountUnusedRecoveryCodesByUsername(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(0, nil)

	gomock.InOrder(
		mock.StorageMock.
			EXPECT().
			LoadPreferred2FAMethod(mock.Ctx, testUsername).
			Return(model.SecondFactorMethodWebAuthn, nil),
		mock.StorageMock.
			EXPECT().
			LoadUserInfo(mock.Ctx, testUsername).
			Return(model.UserInfo{Method: model.SecondFactorMethodWebAuthn, HasTOTP: true, HasWebAuthn: true}, nil),
		mock.UserProviderMock.
			EXPECT().
			GetDetailsCached(testUsername).
			Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName}, nil),
	)

	UserInfoPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	actual := model.UserInfo{}

	mock.GetResponseData(t, &actual)

	assert.Equal(t, model.SecondFactorMethodWebAuthn, actual.Method)
}

func TestWebAuthnRelyingPartyMethodPreferencePOSTShouldAcceptWebAuthnOnUnlistedOrigin(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.WebAuthn = *tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tRelatedOriginHost)
	mock.Ctx.Request.SetBodyString(`{"method":"webauthn"}`)

	tSetupWebAuthnRelyingPartyUser(t, mock, nil)

	mock.StorageMock.
		EXPECT().
		SavePreferred2FAMethod(mock.Ctx, testUsername, model.SecondFactorMethodWebAuthn).
		Return(nil)

	MethodPreferencePOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))
}

func TestWebAuthnRelyingPartyConfigurationGET(t *testing.T) {
	testCases := []struct {
		name     string
		host     string
		expected []string
	}{
		{"ShouldIncludeWebAuthnOnRelyingPartyOrigin", tRelyingPartyHost, []string{model.SecondFactorMethodTOTP, model.SecondFactorMethodWebAuthn, model.SecondFactorMethodDuo}},
		{"ShouldExcludeWebAuthnOnUnlistedOrigin", tRelatedOriginHost, []string{model.SecondFactorMethodTOTP, model.SecondFactorMethodDuo}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.WebAuthn = *tWebAuthnRelyingPartyConfig(nil, tRelyingPartyHost)
			mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&mock.Ctx.Configuration)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)

			ConfigurationGET(mock.Ctx)

			mock.Assert200OK(t, configurationBody{AvailableMethods: tc.expected})
		})
	}
}
