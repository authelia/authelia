package oidc_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewDPoPResourceRequest(t *testing.T) {
	testCases := []struct {
		Name     string
		Method   string
		Target   string
		Token    string
		Proof    string
		Expected string
	}{
		{
			"ShouldReconstructHTTPSTarget",
			http.MethodGet,
			"https://app.example.com/resource?query=value",
			"token",
			"proof",
			"https://app.example.com/resource",
		},
		{
			"ShouldReconstructHTTPTarget",
			http.MethodPost,
			"http://app.example.com/api/v1",
			"token",
			"proof",
			"http://app.example.com/api/v1",
		},
		{
			"ShouldPreserveEscapedPath",
			http.MethodGet,
			"https://app.example.com/resource%3Fx=1",
			"token",
			"proof",
			"https://app.example.com/resource%3Fx=1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			target, err := url.Parse(tc.Target)
			require.NoError(t, err)

			r := oidc.NewDPoPResourceRequest(tc.Method, target, tc.Token, tc.Proof)

			assert.Equal(t, tc.Method, r.Method)
			assert.Equal(t, target.Host, r.Host)
			assert.Equal(t, tc.Proof, r.Header.Get(oidc.HeaderDPoP))
			assert.Equal(t, oidc.SchemeDPoP+" "+tc.Token, r.Header.Get("Authorization"))
			assert.Equal(t, tc.Expected, oauthelia2.RequestURL(r))
		})
	}
}

func TestValidateDPoPResourceAccess(t *testing.T) {
	testCases := []struct {
		Name          string
		Strategy      oauthelia2.DPoPStrategy
		JKT           string
		ExpectedNonce string
		Expected      string
	}{
		{
			"ShouldRejectUnboundToken",
			&testDPoPStrategy{},
			"",
			"",
			"The access token is not bound to a DPoP proof-of-possession key.",
		},
		{
			"ShouldRejectBoundTokenWithoutStrategy",
			nil,
			"abc",
			"",
			"The access token is bound to a DPoP proof-of-possession key but this issuer is not configured to validate DPoP proofs.",
		},
		{
			"ShouldReturnValidationError",
			&testDPoPStrategy{err: oauthelia2.ErrInvalidDPoPProof.WithHint("The DPoP proof has already been used.")},
			"abc",
			"",
			"The DPoP proof has already been used.",
		},
		{
			"ShouldMintNonceOnChallenge",
			&testDPoPStrategy{err: oauthelia2.ErrUseDPoPNonce.WithHint("The DPoP proof is missing the required 'nonce' claim."), nonce: "fresh-nonce"},
			"abc",
			"fresh-nonce",
			"The DPoP proof is missing the required 'nonce' claim.",
		},
		{
			"ShouldSucceed",
			&testDPoPStrategy{},
			"abc",
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			target, err := url.Parse("https://app.example.com/resource")
			require.NoError(t, err)

			r := oidc.NewDPoPResourceRequest(http.MethodGet, target, "token", "proof")

			nonce, err := oidc.ValidateDPoPResourceAccess(context.TODO(), tc.Strategy, r, "token", tc.JKT, false)

			assert.Equal(t, tc.ExpectedNonce, nonce)

			if tc.Expected == "" {
				assert.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Equal(t, tc.Expected, oauthelia2.ErrorToRFC6749Error(err).HintField)
		})
	}
}

type testDPoPStrategy struct {
	err   error
	nonce string
}

func (s *testDPoPStrategy) ValidateDPoPProof(ctx context.Context, method, url, proof string, requireNonce bool) (parsed *oauthelia2.DPoPProof, err error) {
	return nil, s.err
}

func (s *testDPoPStrategy) NewDPoPNonce(ctx context.Context) (nonce string, err error) {
	return s.nonce, nil
}

func (s *testDPoPStrategy) ValidateDPoPNonce(ctx context.Context, nonce string) (err error) {
	return nil
}

func (s *testDPoPStrategy) ValidateResourceAccess(ctx context.Context, r *http.Request, accessToken, boundJKT string, requireNonce bool) (parsed *oauthelia2.DPoPProof, err error) {
	if s.err != nil {
		return nil, s.err
	}

	return &oauthelia2.DPoPProof{Thumbprint: boundJKT}, nil
}
