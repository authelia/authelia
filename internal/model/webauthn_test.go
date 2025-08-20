package model_test

import (
	"crypto/rand"
	"database/sql"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestWebAuthnMisc(t *testing.T) {
	u := uuid.Must(uuid.Parse("cb69481e-8ff7-4039-93ec-0a2729a154a8"))

	uuidBinary, err := u.MarshalBinary()
	require.NoError(t, err)

	testCases := []struct {
		name                          string
		have                          *model.WebAuthnUser
		expectedID                    []byte
		expectedName                  string
		expectedDisplayName           string
		expectedCredentials           []webauthn.Credential
		expectedCredentialDescriptors []protocol.CredentialDescriptor
	}{
		{
			"ShouldHandleBasicElements",
			&model.WebAuthnUser{
				ID:          1,
				RPID:        "https://example.com",
				Username:    "john",
				UserID:      "abc",
				DisplayName: "John Smith",
			},
			[]byte("abc"),
			"john",
			"John Smith",
			[]webauthn.Credential{},
			[]protocol.CredentialDescriptor{},
		},
		{
			"ShouldHandleCredentials",
			&model.WebAuthnUser{
				ID: 1,
				Credentials: []model.WebAuthnCredential{
					{
						ID:              1,
						KID:             model.NewBase64([]byte("abc")),
						PublicKey:       []byte("notapubkey"),
						AttestationType: "packed",
						Present:         true,
						Verified:        true,
						BackupEligible:  false,
						BackupState:     false,
						AAGUID:          model.MustNullUUID(model.ParseNullUUID("cb69481e-8ff7-4039-93ec-0a2729a154a8")),
						SignCount:       10,
						CloneWarning:    false,
						Attachment:      "cross-platform",
						Transport:       "usb",
						Attestation:     []byte(`{"clientDataJSON":"ZXhhbXBsZQ==","clientDataHash":"ZXhhbXBsZQ==","authenticatorData":"ZXhhbXBsZQ==","publicKeyAlgorithm":1,"object":"ZXhhbXBsZQ=="}`),
					},
				},
			},
			[]byte{},
			"",
			"",
			[]webauthn.Credential{
				{
					ID:              []byte("abc"),
					PublicKey:       []byte("notapubkey"),
					AttestationType: "packed",
					Transport:       []protocol.AuthenticatorTransport{protocol.USB},
					Flags: webauthn.CredentialFlags{
						UserPresent:    true,
						UserVerified:   true,
						BackupState:    false,
						BackupEligible: false,
					},
					Authenticator: webauthn.Authenticator{
						AAGUID:       uuidBinary,
						Attachment:   protocol.CrossPlatform,
						SignCount:    10,
						CloneWarning: false,
					},
					Attestation: webauthn.CredentialAttestation{
						ClientDataJSON:     []byte("example"),
						ClientDataHash:     []byte("example"),
						AuthenticatorData:  []byte("example"),
						PublicKeyAlgorithm: 1,
						Object:             []byte("example"),
					},
				},
			},
			[]protocol.CredentialDescriptor{
				{
					Type:            "public-key",
					CredentialID:    []byte("abc"),
					Transport:       []protocol.AuthenticatorTransport{protocol.USB},
					AttestationType: "packed",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedID, tc.have.WebAuthnID())
			assert.Equal(t, tc.expectedName, tc.have.WebAuthnName())
			assert.Equal(t, tc.expectedDisplayName, tc.have.WebAuthnDisplayName())
			assert.Equal(t, tc.expectedCredentials, tc.have.WebAuthnCredentials())
			assert.Equal(t, tc.expectedCredentialDescriptors, tc.have.WebAuthnCredentialDescriptors())
		})
	}
}

func TestWebAuthnUser(t *testing.T) {
	testCases := []struct {
		name            string
		have            model.WebAuthnUser
		expectedFIDOU2F bool
	}{
		{
			"ShouldNotHaveFIDOU2FByDefault",
			model.WebAuthnUser{},
			false,
		},
		{
			"ShouldNotHaveFIDOU2FWithIncorrectAttestationType",
			model.WebAuthnUser{
				Credentials: []model.WebAuthnCredential{
					{
						AttestationType: "random",
					},
				},
			},
			false,
		},
		{
			"ShouldHaveFIDOU2FWithCorrectAttestationType",
			model.WebAuthnUser{
				Credentials: []model.WebAuthnCredential{
					{
						AttestationType: "fido-u2f",
					},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedFIDOU2F, tc.have.HasFIDOU2F())
		})
	}
}

func TestWebAuthnCredential(t *testing.T) {
	testCases := []struct {
		name          string
		have          *model.WebAuthnCredential
		config        *webauthn.Config
		now           time.Time
		authenticator webauthn.Authenticator
		expected      *model.WebAuthnCredential
	}{
		{
			name: "ShouldUpdate",
			have: &model.WebAuthnCredential{
				SignCount:  1,
				RPID:       "",
				LastUsedAt: sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			},
			config:        &webauthn.Config{RPID: "https://example.com", RPOrigins: []string{"org.example.com"}},
			now:           time.Unix(10, 0),
			authenticator: webauthn.Authenticator{SignCount: 2, CloneWarning: false},
			expected: &model.WebAuthnCredential{
				SignCount:  2,
				RPID:       "https://example.com",
				LastUsedAt: sql.NullTime{Time: time.Unix(10, 0), Valid: true},
			},
		},
		{
			name: "ShouldUpdateFIDOU2F",
			have: &model.WebAuthnCredential{
				SignCount:       1,
				RPID:            "",
				LastUsedAt:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
				AttestationType: "fido-u2f",
			},
			config:        &webauthn.Config{RPID: "https://example.com", RPOrigins: []string{"org.example.com"}},
			now:           time.Unix(10, 0),
			authenticator: webauthn.Authenticator{SignCount: 2, CloneWarning: false},
			expected: &model.WebAuthnCredential{
				SignCount:       2,
				RPID:            "org.example.com",
				LastUsedAt:      sql.NullTime{Time: time.Unix(10, 0), Valid: true},
				AttestationType: "fido-u2f",
			},
		},
		{
			name: "ShouldNotUpdateExistingRPID",
			have: &model.WebAuthnCredential{
				SignCount:       1,
				LastUsedAt:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
				AttestationType: "fido-u2f",
				RPID:            "another.example.com",
			},
			config:        &webauthn.Config{RPID: "https://example.com", RPOrigins: []string{"org.example.com"}},
			now:           time.Unix(10, 0),
			authenticator: webauthn.Authenticator{SignCount: 2, CloneWarning: false},
			expected: &model.WebAuthnCredential{
				SignCount:       2,
				RPID:            "another.example.com",
				LastUsedAt:      sql.NullTime{Time: time.Unix(10, 0), Valid: true},
				AttestationType: "fido-u2f",
			},
		},
		{
			name: "ShouldUpdateCloneWarning",
			have: &model.WebAuthnCredential{
				SignCount:       1,
				LastUsedAt:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
				AttestationType: "fido-u2f",
				RPID:            "another.example.com",
			},
			config:        &webauthn.Config{RPID: "https://example.com", RPOrigins: []string{"org.example.com"}},
			now:           time.Unix(10, 0),
			authenticator: webauthn.Authenticator{SignCount: 2, CloneWarning: true},
			expected: &model.WebAuthnCredential{
				SignCount:       2,
				RPID:            "another.example.com",
				LastUsedAt:      sql.NullTime{Time: time.Unix(10, 0), Valid: true},
				AttestationType: "fido-u2f",
				CloneWarning:    true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.have.UpdateSignInInfo(tc.config, tc.now, tc.authenticator)

			assert.Equal(t, tc.expected, tc.have)
		})
	}
}

func TestWebAuthnCredential_ToData(t *testing.T) {
	toTimePtr := func(in time.Time) *time.Time {
		return &in
	}

	toStrPtr := func(in string) *string {
		return &in
	}

	testCases := []struct {
		name     string
		have     model.WebAuthnCredential
		expected model.WebAuthnCredentialData
	}{
		{
			"ShouldParseToData",
			model.WebAuthnCredential{
				SignCount:       2,
				RPID:            "org.example.com",
				LastUsedAt:      sql.NullTime{Time: time.Unix(10, 0), Valid: true},
				AttestationType: "fido-u2f",
				Transport:       "nfc,usb",
			},
			model.WebAuthnCredentialData{
				SignCount:       2,
				RPID:            "org.example.com",
				Transports:      []string{"nfc", "usb"},
				LastUsedAt:      toTimePtr(time.Unix(10, 0)),
				AttestationType: "fido-u2f",
			},
		},
		{
			"ShouldParseToData",
			model.WebAuthnCredential{
				KID:             model.NewBase64([]byte("abc")),
				SignCount:       2,
				RPID:            "org.example.com",
				LastUsedAt:      sql.NullTime{Time: time.Unix(10, 0), Valid: true},
				AttestationType: "fido-u2f",
				Transport:       "nfc,usb",
				PublicKey:       []byte("abc"),
				AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("b4e159da-a52b-4690-81dd-08972950db5f")), Valid: true},
			},
			model.WebAuthnCredentialData{
				KID:             "YWJj",
				SignCount:       2,
				RPID:            "org.example.com",
				Transports:      []string{"nfc", "usb"},
				LastUsedAt:      toTimePtr(time.Unix(10, 0)),
				AttestationType: "fido-u2f",
				PublicKey:       "YWJj",
				AAGUID:          toStrPtr("b4e159da-a52b-4690-81dd-08972950db5f"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.have.ToData()

			assert.Equal(t, tc.expected, actual)

			actual2, err := actual.ToCredential()

			require.NoError(t, err)

			assert.Equal(t, tc.have, *actual2)
		})
	}
}

func TestWebAuthnCredentialData_ToCredential(t *testing.T) {
	toTimePtr := func(in time.Time) *time.Time {
		return &in
	}

	toStrPtr := func(in string) *string {
		return &in
	}

	testCases := []struct {
		name     string
		have     model.WebAuthnCredentialData
		expected *model.WebAuthnCredential
		err      string
	}{
		{
			"ShouldParseToData",
			model.WebAuthnCredentialData{
				SignCount:       2,
				RPID:            "org.example.com",
				Transports:      []string{"nfc", "usb"},
				LastUsedAt:      toTimePtr(time.Unix(10, 0)),
				AttestationType: "fido-u2f",
			},
			&model.WebAuthnCredential{
				SignCount:       2,
				RPID:            "org.example.com",
				LastUsedAt:      sql.NullTime{Time: time.Unix(10, 0), Valid: true},
				AttestationType: "fido-u2f",
				Transport:       "nfc,usb",
			},
			"",
		},
		{
			"ShouldErrBadAAGUID",
			model.WebAuthnCredentialData{
				SignCount:       2,
				RPID:            "org.example.com",
				Transports:      []string{"nfc", "usb"},
				LastUsedAt:      toTimePtr(time.Unix(10, 0)),
				AttestationType: "fido-u2f",
				AAGUID:          toStrPtr("not-a-aaguid"),
			},
			nil,
			"error occurred parsing aaguid: invalid UUID length: 12",
		},
		{
			"ShouldErrBadKID",
			model.WebAuthnCredentialData{
				SignCount:       2,
				RPID:            "org.example.com",
				Transports:      []string{"nfc", "usb"},
				LastUsedAt:      toTimePtr(time.Unix(10, 0)),
				AttestationType: "fido-u2f",
				KID:             "---123===123",
			},
			nil,
			"error occurred deocding kid: illegal base64 data at input byte 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.have.ToCredential()

			assert.Equal(t, tc.expected, actual)

			if len(tc.err) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestNewWebAuthnCredential(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Clock = &mock.Clock

	testCases := []struct {
		name                        string
		rpid, username, description string
		credential                  *webauthn.Credential
		expected                    model.WebAuthnCredential
	}{
		{
			"ShouldGenerateStandard",
			"abc.example.com",
			"john",
			"example",
			&webauthn.Credential{
				Authenticator: webauthn.Authenticator{
					AAGUID: []byte{180, 225, 89, 218, 165, 43, 70, 144, 129, 221, 8, 151, 41, 80, 219, 95},
				},
				Transport: []protocol.AuthenticatorTransport{
					protocol.NFC,
					protocol.USB,
				},
			},
			model.WebAuthnCredential{
				Username:    "john",
				RPID:        "abc.example.com",
				Description: "example",
				Transport:   "nfc,usb",
				CreatedAt:   mock.Clock.Now(),
				AAGUID:      uuid.NullUUID{UUID: uuid.Must(uuid.Parse("b4e159da-a52b-4690-81dd-08972950db5f")), Valid: true},
				Attestation: []byte(`{"clientDataJSON":null,"clientDataHash":null,"authenticatorData":null,"publicKeyAlgorithm":0,"object":null}`),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := model.NewWebAuthnCredential(mock.Ctx, tc.rpid, tc.username, tc.description, tc.credential)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestWebAuthnCredentialImportExport(t *testing.T) {
	have := model.WebAuthnCredentialExport{
		WebAuthnCredentials: []model.WebAuthnCredential{
			{
				ID:              0,
				CreatedAt:       time.Now(),
				LastUsedAt:      sql.NullTime{Time: time.Now(), Valid: true},
				RPID:            "example",
				Username:        "john",
				Description:     "akey",
				KID:             model.NewBase64(MustRead(20)),
				PublicKey:       MustRead(128),
				AttestationType: "fido-u2f",
				Transport:       "",
				AAGUID:          model.MustNullUUID(model.NewRandomNullUUID()),
				SignCount:       20,
				CloneWarning:    false,
			},
			{
				ID:              0,
				CreatedAt:       time.Now(),
				LastUsedAt:      sql.NullTime{Valid: false},
				RPID:            "example2",
				Username:        "john2",
				Description:     "bkey",
				KID:             model.NewBase64(MustRead(60)),
				PublicKey:       MustRead(64),
				AttestationType: "packed",
				Transport:       "",
				AAGUID:          uuid.NullUUID{Valid: false},
				SignCount:       30,
				CloneWarning:    true,
			},
		},
	}

	out, err := yaml.Marshal(&have)
	require.NoError(t, err)

	imported := model.WebAuthnCredentialExport{}

	require.NoError(t, yaml.Unmarshal(out, &imported))
	require.Equal(t, len(have.WebAuthnCredentials), len(imported.WebAuthnCredentials))

	for i, actual := range imported.WebAuthnCredentials {
		t.Run(actual.Description, func(t *testing.T) {
			expected := have.WebAuthnCredentials[i]

			assert.Equal(t, expected.KID, actual.KID)
			assert.Equal(t, expected.PublicKey, actual.PublicKey)
			assert.Equal(t, expected.SignCount, actual.SignCount)
			assert.Equal(t, expected.AttestationType, actual.AttestationType)
			assert.Equal(t, expected.RPID, actual.RPID)
			assert.Equal(t, expected.AAGUID.Valid, actual.AAGUID.Valid)
			assert.Equal(t, expected.AAGUID.UUID, actual.AAGUID.UUID)
			assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
			assert.WithinDuration(t, expected.LastUsedAt.Time, actual.LastUsedAt.Time, time.Second)
			assert.Equal(t, expected.LastUsedAt.Valid, actual.LastUsedAt.Valid)
			assert.Equal(t, expected.CloneWarning, actual.CloneWarning)
			assert.Equal(t, expected.Description, actual.Description)
			assert.Equal(t, expected.Username, actual.Username)
		})
	}
}

func MustRead(n int) []byte {
	data := make([]byte, n)

	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	return data
}
