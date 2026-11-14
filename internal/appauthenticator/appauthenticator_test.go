// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"math/big"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exBase       = "https://auth.example.com/authelia"
	exVerifier   = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"
	exChallenge  = "ICEiIyQlJicoKSorLC0uLzAxMjM0NTY3ODk6Ozw9Pj8"
	exApproval   = "b-ZGgysHmyFD-nVfpSL6T_ujhcTC48BiTmjNunybxAo"
	exTransport  = "YCYm6f6_ROxusCuhobH8lv5vX30zdDsRc4mOU1sKWIo"
	exDeviceID   = "QEFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaW1xdXl8"
	exPullID     = "YGFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6e3x9fn8"
	exCheckVer   = "gIGCg4SFhoeIiYqLjI2Oj5CRkpOUlZaXmJmam5ydnp8"
	exCheckChal  = "rN4zJ28AG73PbsJ2nR8lHWH0eIqNuKSVrG4BwHaRHGE"
	exApproveChl = "oKGio6SlpqeoqaqrrK2ur7CxsrO0tba3uLm6u7y9vr8"
)

func TestShouldComputeCodeChallenges(t *testing.T) {
	assert.Equal(t, "6oZqdX5MOLq_qBJ8vppAnT4fk6AP8UiP9zX8-Rev_9A", CodeChallengeS256(exVerifier))
	assert.Equal(t, exCheckChal, CodeChallengeS256(exCheckVer))
}

func TestShouldBuildSignedStrings(t *testing.T) {
	reg := RegistrationString(exBase, exChallenge, "42", exVerifier, exApproval, exTransport)

	assert.Equal(t, "authelia-app-v1:register::aHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tL2F1dGhlbGlh::ICEiIyQlJicoKSorLC0uLzAxMjM0NTY3ODk6Ozw9Pj8::42::AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8::b-ZGgysHmyFD-nVfpSL6T_ujhcTC48BiTmjNunybxAo::YCYm6f6_ROxusCuhobH8lv5vX30zdDsRc4mOU1sKWIo", reg)

	h := sha256.Sum256([]byte(reg))

	assert.Equal(t, "4TCF54IWzcc8U4Auql_lix1PaIG8BFYy4ec5ozcx-C4", base64.RawURLEncoding.EncodeToString(h[:]))

	body := []byte(`{"code_challenge":"rN4zJ28AG73PbsJ2nR8lHWH0eIqNuKSVrG4BwHaRHGE","code_challenge_method":"S256"}`)

	assert.Equal(t, "hPv5ytNt4r5MEYTeJjhAPdlNU8tOhOQDOXXdi4QpoAs", BodyHash(body))

	assert.Equal(t,
		"authelia-app-v1:request::aHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tL2F1dGhlbGlh::QEFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaW1xdXl8::POST::aHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tL2F1dGhlbGlhL2FwaS9zZWNvbmRmYWN0b3IvYXBwL2NoYWxsZW5nZS9ZR0ZpWTJSbFptZG9hV3ByYkcxdWIzQnhjbk4wZFhaM2VIbDZlM3g5Zm44::1794125087::wMHCw8TFxsfIycrLzM3Ozw::hPv5ytNt4r5MEYTeJjhAPdlNU8tOhOQDOXXdi4QpoAs",
		RequestString(exBase, exDeviceID, "POST", exBase+"/api/secondfactor/app/challenge/"+exPullID, "1794125087", "wMHCw8TFxsfIycrLzM3Ozw", BodyHash(body)))

	assert.Equal(t,
		"authelia-app-v1:approve::aHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tL2F1dGhlbGlh::QEFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaW1xdXl8::oKGio6SlpqeoqaqrrK2ur7CxsrO0tba3uLm6u7y9vr8::07::gIGCg4SFhoeIiYqLjI2Oj5CRkpOUlZaXmJmam5ydnp8",
		ApprovalString(exBase, exDeviceID, exApproveChl, "07", exCheckVer))

	assert.Equal(t,
		"authelia-app-v1:request::aHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tL2F1dGhlbGlh::QEFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaW1xdXl8::DELETE::aHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tL2F1dGhlbGlhL2FwaS9zZWNvbmRmYWN0b3IvYXBwL2RldmljZQ::1794125090::0NHS09TV1tfY2drb3N3e3w::47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
		RequestString(exBase, exDeviceID, "DELETE", exBase+"/api/secondfactor/app/device", "1794125090", "0NHS09TV1tfY2drb3N3e3w", BodyHash(nil)))
}

func TestShouldHashEmptyBody(t *testing.T) {
	assert.Equal(t, EmptyBodyHash, BodyHash(nil))
	assert.Equal(t, EmptyBodyHash, BodyHash([]byte{}))
}

func TestShouldVerifySignatures(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	spki, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)

	pub, err := ParsePublicKey(spki)
	require.NoError(t, err)

	msg := "hello"
	h := sha256.Sum256([]byte(msg))

	der, err := ecdsa.SignASN1(rand.Reader, priv, h[:])
	require.NoError(t, err)

	assert.True(t, VerifySignature(pub, msg, base64.RawURLEncoding.EncodeToString(der)))
	assert.False(t, VerifySignature(pub, "other", base64.RawURLEncoding.EncodeToString(der)))
	assert.False(t, VerifySignature(pub, msg, "!!!"))
	assert.False(t, VerifySignature(nil, msg, base64.RawURLEncoding.EncodeToString(der)))

	r, s := new(big.Int), new(big.Int)

	parseDER(t, der, r, s)

	n := priv.Curve.Params().N

	other := new(big.Int).Sub(n, s)

	for _, v := range []*big.Int{s, other} {
		assert.True(t, VerifySignature(pub, msg, base64.RawURLEncoding.EncodeToString(encodeDER(t, r, v))))
	}

	sig2 := append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...)

	assert.False(t, VerifySignature(pub, msg, base64.RawURLEncoding.EncodeToString(sig2)))
}

func TestShouldRejectP384(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	spki, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)

	_, err = ParsePublicKey(spki)
	assert.Error(t, err)

	_, err = ParsePublicKey([]byte("garbage"))
	assert.Error(t, err)
}

func TestShouldComputeThumbprint(t *testing.T) {
	h := sha256.Sum256([]byte("abc"))

	assert.Equal(t, base64.RawURLEncoding.EncodeToString(h[:]), Thumbprint([]byte("abc")))
}

func TestShouldParseAuthorization(t *testing.T) {
	good := `AutheliaApp device_id="QEFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaW1xdXl8", ts="1794125087", nonce="wMHCw8TFxsfIycrLzM3Ozw", sig="abc_-DEF"`

	a, err := ParseAuthorization([]byte(good))
	require.NoError(t, err)

	assert.Equal(t, exDeviceID, a.DeviceID)
	assert.Equal(t, "1794125087", a.Timestamp)
	assert.Equal(t, "wMHCw8TFxsfIycrLzM3Ozw", a.Nonce)
	assert.Equal(t, "abc_-DEF", a.Signature)
	assert.Equal(t, int64(1794125087), a.Time.Unix())

	bad := []string{
		``,
		`AutheliaApp ts="1794125087", device_id="a", nonce="wMHCw8TFxsfIycrLzM3Ozw", sig="abc"`,
		`AutheliaApp device_id="a",ts="1794125087", nonce="wMHCw8TFxsfIycrLzM3Ozw", sig="abc"`,
		`Bearer device_id="a", ts="1794125087", nonce="wMHCw8TFxsfIycrLzM3Ozw", sig="abc"`,
		`AutheliaApp device_id="a", ts="1794125087", nonce="wMHCw8TFxsfIycrLzM3Oz", sig="abc"`,
		`AutheliaApp device_id="a", ts="1794125087123", nonce="wMHCw8TFxsfIycrLzM3Ozw", sig="abc"`,
		`AutheliaApp device_id="a", ts="1794125087", nonce="wMHCw8TFxsfIycrLzM3Ozw", sig=""`,
	}

	for _, v := range bad {
		_, err = ParseAuthorization([]byte(v))
		assert.Error(t, err, v)
	}
}

func TestShouldGenerateNumbers(t *testing.T) {
	leading := map[byte]bool{}

	for i := 0; i < 20000; i++ {
		n, err := NewNumber(4)
		require.NoError(t, err)
		require.Len(t, n, 4)
		require.Regexp(t, `^\d{4}$`, n)

		leading[n[0]] = true
	}

	assert.Len(t, leading, 10)

	_, err := NewNumber(1)
	assert.Error(t, err)

	_, err = NewNumber(5)
	assert.Error(t, err)
}

func TestShouldGenerateIDs(t *testing.T) {
	id, err := NewID()
	require.NoError(t, err)

	assert.Len(t, id, 43)
	assert.True(t, IsValidID(id))
	assert.True(t, IsValidVerifier(id))
	assert.True(t, IsValidCodeChallenge(id))
	assert.True(t, IsBase64URL(id))
	assert.Len(t, HashID(id), 32)

	assert.False(t, IsValidID(""))
	assert.False(t, IsValidID("a=b"))
	assert.False(t, IsValidVerifier("abc"))
	assert.False(t, IsBase64URL(""))
	assert.False(t, IsBase64URL("ab=="))
}

func TestShouldNormalizeBaseURL(t *testing.T) {
	testCases := []struct {
		in, out string
		err     bool
	}{
		{"https://Auth.Example.com/", "https://auth.example.com", false},
		{"https://auth.example.com:443/authelia/", "https://auth.example.com/authelia", false},
		{"https://auth.example.com:8443/authelia", "https://auth.example.com:8443/authelia", false},
		{"https://auth.example.com", "https://auth.example.com", false},
		{"https://user:pass@auth.example.com/a", "https://auth.example.com/a", false},
		{"http://127.0.0.1:9091", "", true},
		{"https://127.0.0.1:9091", "https://127.0.0.1:9091", false},
		{"https://auth.example.com./a", "", true},
		{"https://[fe80::1%25eth0]/a", "", true},
		{"https://b\u00fccher.example/a", "", true},
		{"https://my_host.example/a", "", true},
		{"https://-bad-.example/", "", true},
		{"https://auth.example.com:99999/a", "", true},
		{"https://auth.example.com:0443/a", "", true},
		{"https://auth.example.com:0/a", "", true},
		{"https://auth.example.com:65535/a", "https://auth.example.com:65535/a", false},
		{"https://[::1]:8443/a", "https://[::1]:8443/a", false},
		{"https://auth.example.com/a%20b", "", true},
		{"https://auth.example.com/a/../b", "", true},
		{"https://auth.example.com/a//b", "", true},
		{"ftp://auth.example.com", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			u, err := url.Parse(tc.in)
			require.NoError(t, err)

			out, err := NewBaseURL(u)

			if tc.err {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}

	_, err := NewBaseURL(nil)
	assert.Error(t, err)
}

func TestShouldBuildQRText(t *testing.T) {
	assert.Equal(t, "authelia-app://auth.example.com/authelia/api/register/app/abc", QRText("https://auth.example.com/authelia", "abc"))
}

func parseDER(t *testing.T, der []byte, r, s *big.Int) {
	t.Helper()

	var sig struct{ R, S *big.Int }

	_, err := asn1Unmarshal(der, &sig)
	require.NoError(t, err)

	r.Set(sig.R)
	s.Set(sig.S)
}

func encodeDER(t *testing.T, r, s *big.Int) []byte {
	t.Helper()

	out, err := asn1Marshal(struct{ R, S *big.Int }{r, s})
	require.NoError(t, err)

	return out
}

var (
	asn1Unmarshal = asn1.Unmarshal
	asn1Marshal   = asn1.Marshal
)
