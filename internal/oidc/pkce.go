package oidc

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewPKCEVerifier returns a new PKCEVerifier.
func NewPKCEVerifier(n int) *PKCEVerifier {
	return &PKCEVerifier{
		verifier: utils.RandomString(n, utils.CharSetRFC3986Unreserved, true),
	}
}

// PKCEVerifier is a struct that holds the random verifier value and provides easy methods to get the
// 'code_challenge' parameter from the CodeChallenge func, the 'code_challenge_method' parameter from the
// CodeChallengeMethod func and the 'code_verifier' parameter from the CodeVerifier func.
type PKCEVerifier struct {
	plain    bool
	verifier string
}

// CodeChallenge returns the base64 URL encoded SHA256 sum of the verifier when using S256 and the verifier when not to be
// used with the code_challenge parameter on the authorization endpoint.
func (v *PKCEVerifier) CodeChallenge() string {
	if v.plain {
		return v.verifier
	}

	challenge := sha256.New()

	challenge.Write([]byte(v.verifier))

	return base64.RawURLEncoding.EncodeToString(challenge.Sum([]byte{}))
}

// CodeChallengeMethod returns the value to be used with the code_challenge_method parameter on the authorization endpoint.
func (v *PKCEVerifier) CodeChallengeMethod() string {
	if v.plain {
		return PKCEChallengeMethodPlain
	}

	return PKCEChallengeMethodSHA256
}

// CodeVerifier returns the verifier in the clear to be used with the code_verifier parameter on the token endpoint.
func (v *PKCEVerifier) CodeVerifier() string {
	return v.verifier
}

// SetAuthorizationEndpointParameters sets parameters to a url.Values form related to the OAuth 2.0 Authorization
// Endpoint.
func (v *PKCEVerifier) SetAuthorizationEndpointParameters(form *url.Values) {
	form.Set(FormCodeChallenge, v.CodeChallenge())
	form.Set(FormCodeChallengeMethod, v.CodeChallengeMethod())
}

// SetTokenEndpointParameters sets parameters to a url.Values form related to the OAuth 2.0 Token Endpoint.
func (v *PKCEVerifier) SetTokenEndpointParameters(form *url.Values) {
	form.Set(FormCodeVerifier, v.CodeVerifier())
}
