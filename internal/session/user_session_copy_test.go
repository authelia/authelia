package session

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSessionDeepCopyShouldNotShareReferences(t *testing.T) {
	original := newPopulatedUserSession()

	assertNoSharedReferences(t, "UserSession", reflect.ValueOf(original), reflect.ValueOf(original.deepCopy()))
}

func TestUserSessionDeepCopyShouldPreserveNilReferences(t *testing.T) {
	session := UserSession{}.deepCopy()

	assert.Nil(t, session.AuthenticationMethodRefs.Extra)
	assert.Nil(t, session.PasswordResetUsername)
	assert.Nil(t, session.TOTP)
	assert.Nil(t, session.WebAuthn)
	assert.Nil(t, session.Elevations.User)

	session = UserSession{WebAuthn: &WebAuthn{}, Elevations: Elevations{User: &Elevation{}}}.deepCopy()

	require.NotNil(t, session.WebAuthn)
	assert.Nil(t, session.WebAuthn.SessionData)
	require.NotNil(t, session.Elevations.User)
	assert.Nil(t, session.Elevations.User.RemoteIP)

	session = UserSession{WebAuthn: &WebAuthn{SessionData: &webauthn.SessionData{}}}.deepCopy()

	require.NotNil(t, session.WebAuthn.SessionData)
	assert.Nil(t, session.WebAuthn.UserID)
	assert.Nil(t, session.WebAuthn.AllowedCredentialIDs)
	assert.Nil(t, session.WebAuthn.Extensions)
	assert.Nil(t, session.WebAuthn.CredParams)
}

func TestUserSessionDeepCopyShouldPreserveValues(t *testing.T) {
	original := newPopulatedUserSession()

	assert.Equal(t, original, original.deepCopy())
}

func TestStrategyCacheShouldIsolateMutationsBetweenConsumers(t *testing.T) {
	strategy := newTestStrategyWithRepository(t, newCountingRepository(), nil)
	ctx := newTestCachingContext()

	userSession := newPopulatedUserSession()
	userSession.Username = testUsername
	userSession.CookieDomain = testDomain

	require.NoError(t, strategy.Save(ctx, &userSession))

	first, err := strategy.Get(ctx)
	require.NoError(t, err)

	first.AuthenticationMethodRefs.Extra[0] = "mutated"
	*first.PasswordResetUsername = "mutated"
	first.TOTP.Issuer = "mutated"
	first.WebAuthn.Description = "mutated"
	first.WebAuthn.UserID[0] = 'z'
	first.WebAuthn.AllowedCredentialIDs[0][0] = 'z'
	first.WebAuthn.Extensions["extension"] = "mutated"
	first.Elevations.User.ID = 99
	first.Elevations.User.RemoteIP[len(first.Elevations.User.RemoteIP)-1] = 9

	second, err := strategy.Get(ctx)
	require.NoError(t, err)

	assert.Equal(t, "extra", second.AuthenticationMethodRefs.Extra[0])
	assert.Equal(t, "reset", *second.PasswordResetUsername)
	assert.Equal(t, "issuer", second.TOTP.Issuer)
	assert.Equal(t, "description", second.WebAuthn.Description)
	assert.Equal(t, []byte("userid"), second.WebAuthn.UserID)
	assert.Equal(t, []byte("credential"), second.WebAuthn.AllowedCredentialIDs[0])
	assert.Equal(t, "value", second.WebAuthn.Extensions["extension"])
	assert.Equal(t, 1, second.Elevations.User.ID)
	assert.Equal(t, "192.0.2.1", second.Elevations.User.RemoteIP.String())
}

func newPopulatedUserSession() (session UserSession) {
	reset := "reset"

	session = NewUserSession(testUsername)

	session.CookieDomain = testDomain
	session.PublicID = "pid"
	session.AuthenticationMethodRefs.Extra = []string{"extra"}
	session.PasswordResetUsername = &reset
	session.TOTP = &TOTP{Issuer: "issuer", Algorithm: "SHA1", Digits: 6, Period: 30, Secret: "secret"}
	session.Elevations.User = &Elevation{ID: 1, RemoteIP: net.ParseIP("192.0.2.1"), Expires: time.Unix(1700000000, 0).UTC()}
	session.WebAuthn = &WebAuthn{
		Description: "description",
		SessionData: &webauthn.SessionData{
			Challenge:            "challenge",
			UserID:               []byte("userid"),
			AllowedCredentialIDs: [][]byte{[]byte("credential")},
			Extensions:           protocol.AuthenticationExtensions{"extension": "value"},
			CredParams:           []protocol.CredentialParameter{{Type: protocol.PublicKeyCredentialType}},
		},
	}

	return session
}

// assertNoSharedReferences walks the original and its copy in lockstep and fails for every pointer, slice, or map which
// both of them refer to, as such a value is reachable for mutation through either. A time.Time is skipped as it holds a
// pointer to a shared location which is never mutated.
func assertNoSharedReferences(t *testing.T, path string, original, copied reflect.Value) {
	if original.Type() == reflect.TypeOf(time.Time{}) {
		return
	}

	switch original.Kind() {
	case reflect.Pointer:
		assertNoSharedPointer(t, path, original, copied)
	case reflect.Interface:
		if !original.IsNil() && !copied.IsNil() {
			assertNoSharedReferences(t, path, original.Elem(), copied.Elem())
		}
	case reflect.Slice:
		assertNoSharedSlice(t, path, original, copied)
	case reflect.Map:
		if !original.IsNil() && !copied.IsNil() {
			assert.NotEqual(t, original.Pointer(), copied.Pointer(), "%s is the same map in the copy", path)
		}
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			assertNoSharedReferences(t, path+"."+original.Type().Field(i).Name, original.Field(i), copied.Field(i))
		}
	}
}

func assertNoSharedPointer(t *testing.T, path string, original, copied reflect.Value) {
	if original.IsNil() || copied.IsNil() {
		return
	}

	assert.NotEqual(t, original.Pointer(), copied.Pointer(), "%s is the same pointer in the copy", path)

	assertNoSharedReferences(t, path, original.Elem(), copied.Elem())
}

func assertNoSharedSlice(t *testing.T, path string, original, copied reflect.Value) {
	if original.IsNil() || copied.IsNil() || original.Len() == 0 {
		return
	}

	assert.NotEqual(t, original.Pointer(), copied.Pointer(), "%s shares its backing array with the copy", path)

	for i := 0; i < original.Len() && i < copied.Len(); i++ {
		assertNoSharedReferences(t, fmt.Sprintf("%s[%d]", path, i), original.Index(i), copied.Index(i))
	}
}
