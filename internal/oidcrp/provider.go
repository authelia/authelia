package oidcrp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// Provider represents a single configured external OpenID Connect 1.0 Provider.
type Provider struct {
	ID                                  string
	Name                                string
	Issuer                              string
	ClientID                            string
	Scopes                              []string
	TrustAuthenticationMethodsReference bool

	clientSecret            string
	tokenEndpointAuthMethod string
	alg                     string

	authorizationEndpoint string
	tokenEndpoint         string
	jwksURI               string

	discover bool
	resolved bool
	mutex    sync.Mutex

	keys   KeySet
	client *retryablehttp.Client
}

// AuthorizationEndpoint returns the resolved authorization endpoint.
func (p *Provider) AuthorizationEndpoint() string {
	return p.authorizationEndpoint
}

// TokenEndpoint returns the resolved token endpoint.
func (p *Provider) TokenEndpoint() string {
	return p.tokenEndpoint
}

// ValidateIDToken validates a raw ID Token against this provider.
func (p *Provider) ValidateIDToken(ctx context.Context, raw, nonce string, now time.Time) (claims *IdentityClaims, err error) {
	if err = p.Resolve(ctx); err != nil {
		return nil, err
	}

	return ValidateIDToken(ctx, p.keys, raw, ValidateOptions{
		Issuer:   p.Issuer,
		ClientID: p.ClientID,
		Nonce:    nonce,
		Alg:      p.alg,
		JWKSURI:  p.jwksURI,
		Now:      now,
		Leeway:   time.Minute,
	})
}

// Resolve performs OpenID Connect Discovery 1.0 for this provider if it has not already been performed successfully.
// Discovery is deliberately not performed while the providers are constructed: the availability of a third party must
// never determine whether Authelia is able to start, as that would take down every other authentication method with
// it. A failure here is retried the next time the provider is used.
func (p *Provider) Resolve(ctx context.Context) (err error) {
	if !p.discover {
		return nil
	}

	p.mutex.Lock()

	defer p.mutex.Unlock()

	if p.resolved {
		return nil
	}

	var discovery *Discovery

	if discovery, err = Discover(ctx, p.client, p.Issuer); err != nil {
		return fmt.Errorf("error resolving provider '%s': %w", p.ID, err)
	}

	if p.authorizationEndpoint == "" {
		p.authorizationEndpoint = discovery.AuthorizationEndpoint
	}

	if p.tokenEndpoint == "" {
		p.tokenEndpoint = discovery.TokenEndpoint
	}

	if p.jwksURI == "" {
		p.jwksURI = discovery.JWKSURI
	}

	p.resolved = true

	return nil
}

func newProvider(config *schema.AuthenticationBackendOpenIDConnectProvider, caCertPool *x509.CertPool) (provider *Provider) {
	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = 2
	client.HTTPClient = &http.Client{Timeout: time.Second * 10, Transport: newProviderTransport(caCertPool)}

	provider = &Provider{
		ID:                                  config.ID,
		Name:                                config.Name,
		Issuer:                              config.Issuer,
		ClientID:                            config.ClientID,
		Scopes:                              config.Scopes,
		TrustAuthenticationMethodsReference: config.AuthenticationMethodsReference.Trust,
		clientSecret:                        config.ClientSecret,
		tokenEndpointAuthMethod:             config.TokenEndpointAuthMethod,
		alg:                                 config.IDTokenSignedResponseAlg,
		authorizationEndpoint:               config.Endpoints.Authorization,
		tokenEndpoint:                       config.Endpoints.Token,
		jwksURI:                             config.Endpoints.JSONWebKeys,
		discover:                            !config.Discovery.Disable,
		client:                              client,
	}

	if len(config.JSONWebKeys) != 0 {
		provider.keys = &staticKeySet{jwks: oidc.NewJSONWebKeySetPublic(config.JSONWebKeys)}
	} else {
		provider.keys = oauthelia2.NewDefaultJWKSFetcherStrategy(oauthelia2.JWKSFetcherWithHTTPClient(client))
	}

	return provider
}

func newProviderTransport(caCertPool *x509.CertPool) (transport *http.Transport) {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = t.Clone()
	} else {
		transport = &http.Transport{}
	}

	transport.TLSClientConfig = &tls.Config{RootCAs: caCertPool, MinVersion: tls.VersionTLS12}

	return transport
}

type staticKeySet struct {
	jwks *jose.JSONWebKeySet
}

func (s *staticKeySet) Resolve(_ context.Context, _ string, _ bool) (jwks *jose.JSONWebKeySet, err error) {
	return s.jwks, nil
}
