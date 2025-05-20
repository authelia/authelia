package session

import (
	"bytes"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/hashicorp/go-msgpack/v2/codec"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
)

func TestUserSession_SetFactors(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(session *UserSession)
		expect *UserSession
	}{
		{
			"ShouldSetOneFactorPassword",
			func(session *UserSession) {
				session.SetOneFactorPassword(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true)
			},
			&UserSession{
				Username:                  "john",
				Groups:                    []string{"abc", "123"},
				Emails:                    []string{"john@example.com"},
				KeepMeLoggedIn:            true,
				LastActivity:              10000,
				FirstFactorAuthnTimestamp: 10000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					UsernameAndPassword:          true,
				},
			},
		},
		{
			"ShouldSetOneFactorPasskey",
			func(session *UserSession) {
				session.SetOneFactorPasskey(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true, true, true, true)
			},
			&UserSession{
				Username:                  "john",
				Groups:                    []string{"abc", "123"},
				Emails:                    []string{"john@example.com"},
				KeepMeLoggedIn:            true,
				LastActivity:              10000,
				FirstFactorAuthnTimestamp: 10000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					WebAuthn:             true,
					WebAuthnHardware:     true,
					WebAuthnUserVerified: true,
					WebAuthnUserPresence: true,
				},
			},
		},
		{
			"ShouldSetTwoFactorPassword",
			func(session *UserSession) {
				session.SetOneFactorPasskey(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true, true, true, true)
				session.SetTwoFactorPassword(time.Unix(20000, 0))
			},
			&UserSession{
				Username:                   "john",
				Groups:                     []string{"abc", "123"},
				Emails:                     []string{"john@example.com"},
				KeepMeLoggedIn:             true,
				LastActivity:               20000,
				FirstFactorAuthnTimestamp:  10000,
				SecondFactorAuthnTimestamp: 20000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					UsernameAndPassword:          true,
					WebAuthn:                     true,
					WebAuthnHardware:             true,
					WebAuthnUserVerified:         true,
					WebAuthnUserPresence:         true,
				},
			},
		},
		{
			"ShouldSetOneFactorPasswordAndTwoFactorDuo",
			func(session *UserSession) {
				session.SetOneFactorPassword(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true)
				session.SetTwoFactorDuo(time.Unix(20000, 0))
			},
			&UserSession{
				Username:                   "john",
				Groups:                     []string{"abc", "123"},
				Emails:                     []string{"john@example.com"},
				KeepMeLoggedIn:             true,
				LastActivity:               20000,
				FirstFactorAuthnTimestamp:  10000,
				SecondFactorAuthnTimestamp: 20000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					UsernameAndPassword:          true,
					Duo:                          true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			session := &UserSession{}

			tc.setup(session)

			assert.Equal(t, tc.expect, session)
		})
	}
}

func TestUserSession_AuthenticationLevel(t *testing.T) {
	testCases := []struct {
		name     string
		have     *UserSession
		passkey  bool
		expected authentication.Level
	}{
		{
			"ShouldHandleAnonymous",
			&UserSession{},
			false,
			authentication.NotAuthenticated,
		},
		{
			"ShouldHandleTwoFactorTOTP",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					TOTP:                         true,
				},
			},
			false,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthn",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					WebAuthn:                     true,
				},
			},
			false,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorDuo",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					Duo:                          true,
				},
			},
			false,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthnPasskey",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					WebAuthn:                     true,
				},
			},
			true,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthnPasskeyWithoutKnowledge",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					WebAuthn: true,
				},
			},
			true,
			authentication.OneFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthnPasskeyWithoutKnowledgeWithUserVerification",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					WebAuthn:             true,
					WebAuthnUserVerified: true,
				},
			},
			true,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleNoAMR",
			&UserSession{
				Username:                 "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{},
			},
			true,
			authentication.NotAuthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.AuthenticationLevel(tc.passkey))
		})
	}
}

func TestUserSession_SetTwoFactorWebAuthn(t *testing.T) {
	testCases := []struct {
		name                         string
		at                           time.Time
		hardware, presence, verified bool
		expected                     authorization.AuthenticationMethodsReferences
	}{
		{
			"ShouldHandleHardware",
			time.Unix(1000, 0),
			true,
			true,
			true,
			authorization.AuthenticationMethodsReferences{
				WebAuthn:             true,
				WebAuthnHardware:     true,
				WebAuthnUserPresence: true,
				WebAuthnUserVerified: true,
			},
		},
		{
			"ShouldHandleSoftware",
			time.Unix(1000, 0),
			false,
			true,
			true,
			authorization.AuthenticationMethodsReferences{
				WebAuthn:             true,
				WebAuthnSoftware:     true,
				WebAuthnUserPresence: true,
				WebAuthnUserVerified: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := &UserSession{}

			actual.SetTwoFactorWebAuthn(tc.at, tc.hardware, tc.presence, tc.verified)

			assert.Equal(t, tc.expected, actual.AuthenticationMethodRefs)
		})
	}
}

func TestUserSession_Misc(t *testing.T) {
	session := &UserSession{}

	assert.Equal(t, Identity{}, session.Identity())
	assert.Equal(t, "", session.GetUsername())
	assert.Equal(t, "", session.GetDisplayName())
	assert.Nil(t, session.GetGroups())
	assert.Nil(t, session.GetEmails())
	assert.True(t, session.IsAnonymous())

	session.Username = "abc"

	assert.Equal(t, Identity{Username: "abc"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())

	session.DisplayName = "A B C"

	assert.Equal(t, Identity{Username: "abc", DisplayName: "A B C"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())

	session.Emails = []string{"abc@example.com", "xyz@example.com"}

	assert.Equal(t, Identity{Username: "abc", DisplayName: "A B C", Email: "abc@example.com"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())
	assert.Equal(t, []string{"abc@example.com", "xyz@example.com"}, session.GetEmails())

	session.Groups = []string{"agroup", "bgroup"}
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())
	assert.Equal(t, []string{"abc@example.com", "xyz@example.com"}, session.GetEmails())
	assert.Equal(t, []string{"agroup", "bgroup"}, session.GetGroups())
}

func TestUserSession_MessagePack(t *testing.T) {
	testCases := []struct {
		name     string
		have     *UserSession
		expected []byte
		err      string
	}{
		{
			"ShouldEncodeAnonymous",
			&UserSession{},
			[]byte{0x82, 0xa3, 0x74, 0x74, 0x6c, 0xd3, 0xff, 0x23, 0x40, 0x1, 0x0, 0xd4, 0x40, 0x0, 0xaa, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x80},
			"",
		},
		{
			"ShouldEncodeUserSession",
			&UserSession{CookieDomain: "example.com", Username: "john", KeepMeLoggedIn: true},
			[]byte{0x85, 0xa6, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0xab, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0xa8, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0xa4, 0x6a, 0x6f, 0x68, 0x6e, 0xb1, 0x6b, 0x65, 0x65, 0x70, 0x5f, 0x6d, 0x65, 0x5f, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x64, 0x5f, 0x69, 0x6e, 0xc3, 0xa3, 0x74, 0x74, 0x6c, 0xd3, 0xff, 0x23, 0x40, 0x1, 0x0, 0xd4, 0x40, 0x0, 0xaa, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x80},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.have.ToMessagePack()

			actual, err := a.MarshalMsg(nil)
			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, actual)
				assert.Equal(t, tc.expected, actual)

				result := &UserSessionMessagePack{}

				var o []byte

				o, err = result.UnmarshalMsg(actual)

				assert.Equal(t, a, result)
				assert.NoError(t, err)
				assert.Equal(t, []byte{}, o)

				assert.Equal(t, tc.have, result.ToUserSession())
			}
		})
	}
}

func TestUserSession_MessagePackAlt(t *testing.T) {
	testCases := []struct {
		name     string
		have     *UserSession
		expected []byte
		err      string
	}{
		{
			"ShouldEncodeAnonymous",
			&UserSession{},
			[]byte{0x81, 0xa3, 0x74, 0x74, 0x6c, 0xd3, 0xff, 0x23, 0x40, 0x1, 0x0, 0xd4, 0x40, 0x0},
			"",
		},
		{
			"ShouldEncodeUserSession",
			&UserSession{CookieDomain: "example.com", Username: "john", KeepMeLoggedIn: true},
			[]byte{0x84, 0xa6, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0xab, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0xa8, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0xa4, 0x6a, 0x6f, 0x68, 0x6e, 0xa8, 0x72, 0x65, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0xc3, 0xa3, 0x74, 0x74, 0x6c, 0xd3, 0xff, 0x23, 0x40, 0x1, 0x0, 0xd4, 0x40, 0x0},
			"",
		},
		{
			"ShouldEncodeUserSessionWithElevations",
			&UserSession{CookieDomain: "example.com", Username: "john", KeepMeLoggedIn: true, Elevations: Elevations{User: &Elevation{
				ID:       1,
				RemoteIP: net.ParseIP("127.0.0.1"),
				Expires:  time.UnixMicro(9000000000000).UTC(),
			}}},
			[]byte{0x85, 0xa6, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0xab, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0xa8, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0xa4, 0x6a, 0x6f, 0x68, 0x6e, 0xa8, 0x72, 0x65, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0xc3, 0xa3, 0x74, 0x74, 0x6c, 0xd3, 0xff, 0x23, 0x40, 0x1, 0x0, 0xd4, 0x40, 0x0, 0xaa, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x81, 0xa4, 0x75, 0x73, 0x65, 0x72, 0x83, 0xa2, 0x69, 0x64, 0x1, 0xa2, 0x69, 0x70, 0xb0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x1, 0xa3, 0x65, 0x78, 0x70, 0xd3, 0x0, 0x0, 0x8, 0x2f, 0x79, 0xcd, 0x90, 0x0},
			"",
		},
		{
			"ShouldEncodeUserSessionWithMFASessions",
			&UserSession{CookieDomain: "example.com", Username: "john", KeepMeLoggedIn: true,
				TOTP: &TOTP{
					Issuer:    "authelia.com",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    "abc123",
					Expires:   time.UnixMicro(90000000000000).UTC(),
				},
				WebAuthn: &WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:            "abc123",
						RelyingPartyID:       "authelia.com",
						UserID:               []byte("1k23nm12jkl3n1jk23n"),
						AllowedCredentialIDs: [][]byte{[]byte("12n3hjk1n3jk1h2n3n")},
						Expires:              time.UnixMicro(90000000000000).UTC(),
						UserVerification:     protocol.VerificationRequired,
						Extensions: map[string]any{
							"fido_u2f": true,
						},
						CredParams: []protocol.CredentialParameter{
							{
								Type:      protocol.PublicKeyCredentialType,
								Algorithm: webauthncose.AlgES256,
							},
							{
								Type:      protocol.PublicKeyCredentialType,
								Algorithm: webauthncose.AlgES384,
							},
							{
								Type:      protocol.PublicKeyCredentialType,
								Algorithm: webauthncose.AlgES512,
							},
						},
					},
					Description: "",
				},
			},
			[]byte{0x86, 0xa6, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0xab, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0xa8, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0xa4, 0x6a, 0x6f, 0x68, 0x6e, 0xa8, 0x72, 0x65, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0xc3, 0xa8, 0x77, 0x65, 0x62, 0x61, 0x75, 0x74, 0x68, 0x6e, 0x88, 0xb3, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x91, 0xb2, 0x31, 0x32, 0x6e, 0x33, 0x68, 0x6a, 0x6b, 0x31, 0x6e, 0x33, 0x6a, 0x6b, 0x31, 0x68, 0x32, 0x6e, 0x33, 0x6e, 0xa9, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x65, 0x6e, 0x67, 0x65, 0xa6, 0x61, 0x62, 0x63, 0x31, 0x32, 0x33, 0xaa, 0x63, 0x72, 0x65, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x93, 0x82, 0xa3, 0x61, 0x6c, 0x67, 0xf9, 0xa4, 0x74, 0x79, 0x70, 0x65, 0xaa, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2d, 0x6b, 0x65, 0x79, 0x82, 0xa3, 0x61, 0x6c, 0x67, 0xd0, 0xdd, 0xa4, 0x74, 0x79, 0x70, 0x65, 0xaa, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2d, 0x6b, 0x65, 0x79, 0x82, 0xa3, 0x61, 0x6c, 0x67, 0xd0, 0xdc, 0xa4, 0x74, 0x79, 0x70, 0x65, 0xaa, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2d, 0x6b, 0x65, 0x79, 0xa7, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0xa4, 0x5, 0x5d, 0x4a, 0x80, 0xaa, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x81, 0xa8, 0x66, 0x69, 0x64, 0x6f, 0x5f, 0x75, 0x32, 0x66, 0xc3, 0xa4, 0x72, 0x70, 0x49, 0x64, 0xac, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6c, 0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0xb0, 0x75, 0x73, 0x65, 0x72, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0xa8, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0xa7, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0xb3, 0x31, 0x6b, 0x32, 0x33, 0x6e, 0x6d, 0x31, 0x32, 0x6a, 0x6b, 0x6c, 0x33, 0x6e, 0x31, 0x6a, 0x6b, 0x32, 0x33, 0x6e, 0xa4, 0x74, 0x6f, 0x74, 0x70, 0x86, 0xa3, 0x69, 0x73, 0x73, 0xac, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6c, 0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0xa3, 0x61, 0x6c, 0x67, 0xa4, 0x53, 0x48, 0x41, 0x31, 0xa6, 0x64, 0x69, 0x67, 0x69, 0x74, 0x73, 0x6, 0xa6, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x1e, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0xa6, 0x61, 0x62, 0x63, 0x31, 0x32, 0x33, 0xa3, 0x65, 0x78, 0x70, 0xd3, 0x0, 0x0, 0x51, 0xda, 0xc2, 0x7, 0xa0, 0x0, 0xa3, 0x74, 0x74, 0x6c, 0xd3, 0xff, 0x23, 0x40, 0x1, 0x0, 0xd4, 0x40, 0x0},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.have.ToMessagePack()

			buf := bytes.NewBuffer(nil)
			handle := codec.MsgpackHandle{}

			encoder := codec.NewEncoder(buf, &handle)

			err := encoder.Encode(a)
			actual := buf.Bytes()

			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, actual)
				assert.Equal(t, tc.expected, actual)

				result := &UserSessionMessagePack{}

				decoder := codec.NewDecoderBytes(actual, &handle)

				err = decoder.Decode(result)

				assert.Equal(t, a, result)
				assert.NoError(t, err)

				assert.Equal(t, tc.have, result.ToUserSession())
			}
		})
	}
}
