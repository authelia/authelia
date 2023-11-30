package oidc_test

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/hmac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestRefreshTokenGrantHandler_CanSkipClientAuth(t *testing.T) {
	factory := func() (oauth2.RefreshTokenStrategy, *storage.MemoryStore, *fosite.AccessRequest) {
		return &oauth2.HMACSHAStrategy{
			Enigma: &hmac.HMACStrategy{Config: &fosite.Config{GlobalSecret: []byte("foobarfoobarfoobarfoobarfoobarfoobarfoobarfoobar")}},
			Config: &fosite.Config{
				AccessTokenLifespan:   time.Hour * 24,
				AuthorizeCodeLifespan: time.Hour * 24,
			},
		}, storage.NewMemoryStore(), fosite.NewAccessRequest(&fosite.DefaultSession{})
	}

	strategy, store, requester := factory()

	config := &fosite.Config{
		AccessTokenLifespan:      time.Hour,
		RefreshTokenLifespan:     time.Hour,
		ScopeStrategy:            fosite.HierarchicScopeStrategy,
		AudienceMatchingStrategy: fosite.DefaultAudienceMatchingStrategy,
		RefreshTokenScopes:       []string{oidc.ScopeOffline},
	}

	handler := &oidc.RefreshTokenGrantHandler{
		TokenRevocationStorage: store,
		RefreshTokenStrategy:   strategy,
		Config:                 config,
	}

	assert.False(t, handler.CanSkipClientAuth(context.TODO(), requester))
}

func TestRefreshTokenGrantHandler_HandleTokenEndpointRequest(t *testing.T) {
	sess := &fosite.DefaultSession{Subject: "othersub"}

	expiredSess := &fosite.DefaultSession{
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.RefreshToken: time.Now().UTC().Add(-time.Hour),
		},
	}

	type testCase struct {
		name      string
		setup     func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest)
		err       error
		expected  string
		rexpected *regexp.Regexp
		texpected func(t *testing.T, requester *fosite.AccessRequest)
	}

	testCases := []testCase{
		{
			name: "ShouldFailNotResponsible",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{"123"}
			},
			err:      fosite.ErrUnknownRequest,
			expected: "The handler is not responsible for this request.",
		},
		{
			name: "ShouldFailInvalidGrant",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken}}

				requester.Form.Add(oidc.FormParameterRefreshToken, "some.refreshtokensig")
			},
			err:      fosite.ErrInvalidGrant,
			expected: "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The refresh token has not been found: Could not find the requested resource(s).",
		},
		{
			name: "ShouldFailTokenValidButDoesNotExist",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken}}

				token, _, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)
				requester.Form.Add(oidc.FormParameterRefreshToken, token)
			},
			err:      fosite.ErrInvalidGrant,
			expected: "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The refresh token has not been found: Could not find the requested resource(s).",
		},
		{
			name: "ShouldFailBecauseClientDoesNotMatch",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:       &fosite.DefaultClient{ID: ""},
					GrantedScope: []string{oidc.ScopeOffline},
					Session:      sess,
				})
				require.NoError(t, err)
			},
			err:      fosite.ErrInvalidGrant,
			expected: "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The OAuth 2.0 Client ID from this request does not match the ID during the initial token issuance.",
		},
		{
			name: "ShouldFailExpiredToken",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					Session:        expiredSess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour * 2).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			err:       fosite.ErrInvalidGrant,
			rexpected: regexp.MustCompile(`^The provided authorization grant \(e.g., authorization code, resource owner credentials\) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client\. Token expired\. Refresh token expired at '\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(.\d+)? \+0000 UTC'\.$`),
		},
		{
			name: "ShouldFailOfflineScopeRequestedButClientNotPermitted",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			err:      fosite.ErrInvalidScope,
			expected: "The requested scope is invalid, unknown, or malformed. The OAuth 2.0 Client is not allowed to request scope 'foo'.",
		},
		{
			name: "ShouldPass",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.NotEqual(t, sess, requester.Session)
				assert.NotEqual(t, time.Now().UTC().Add(-time.Hour).Round(time.Hour), requester.RequestedAt)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.RequestedScope)
				assert.NotEqual(t, url.Values{"foo": []string{"bar"}}, requester.Form)
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.AccessToken))
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.RefreshToken))
			},
		},
		{
			name: "ShouldPassKeepOriginalID",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					ID:             "foo",
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.Equal(t, "foo", requester.GetID())
				assert.NotEqual(t, sess, requester.Session)
				assert.NotEqual(t, time.Now().UTC().Add(-time.Hour).Round(time.Hour), requester.RequestedAt)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.RequestedScope)
				assert.NotEqual(t, url.Values{"foo": []string{"bar"}}, requester.Form)
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.AccessToken))
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.RefreshToken))
			},
		},
		{
			name: "ShouldPassWithScope",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", "baz", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				requester.Form.Add(oidc.FormParameterScope, "foo bar baz offline")
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", "bar", "baz", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "bar", "baz", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.Equal(t, fosite.Arguments{"foo", "bar", "baz", oidc.ScopeOffline}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo", "bar", "baz", oidc.ScopeOffline}, requester.RequestedScope)
			},
		},
		{
			name: "ShouldPassWithScopeNarrowing",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", "baz", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				requester.Form.Add(oidc.FormParameterScope, "foo bar offline")
				requester.SetRequestedScopes(fosite.Arguments{"foo", "bar", oidc.ScopeOffline})

				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", "bar", "baz", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "bar", "baz", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.Equal(t, fosite.Arguments{"foo", "bar", oidc.ScopeOffline}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo", "bar", oidc.ScopeOffline}, requester.RequestedScope)
			},
		},
		{
			name: "ShouldFailWithScopeBroadening",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", "baz", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				requester.Form.Add(oidc.FormParameterScope, "foo bar offline")
				requester.SetRequestedScopes(fosite.Arguments{"foo", "bar", oidc.ScopeOffline})

				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", "baz", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "baz", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			err:      fosite.ErrInvalidScope,
			expected: "The requested scope is invalid, unknown, or malformed. The requested scope 'bar' was not originally granted by the resource owner.",
		},
		{
			name: "ShouldPassWithScopeBroadeningOnRefreshFlowScopeClientTrue",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &oidc.BaseClient{
					ID:                                     "foo",
					GrantTypes:                             fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:                                 []string{"foo", "bar", "baz", oidc.ScopeOffline},
					RefreshFlowIgnoreOriginalGrantedScopes: true,
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				requester.Form.Add(oidc.FormParameterScope, "foo bar offline")
				requester.SetRequestedScopes(fosite.Arguments{"foo", "bar", oidc.ScopeOffline})

				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", "baz", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "baz", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo", "bar", oidc.ScopeOffline}, requester.RequestedScope)
			},
		},
		{
			name: "ShouldFailWithScopeBroadeningOnRefreshFlowScopeClientFalse",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &oidc.BaseClient{
					ID:                                     "foo",
					GrantTypes:                             fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:                                 []string{"foo", "bar", "baz", oidc.ScopeOffline},
					RefreshFlowIgnoreOriginalGrantedScopes: false,
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				requester.Form.Add(oidc.FormParameterScope, "foo bar offline")
				requester.SetRequestedScopes(fosite.Arguments{"foo", "bar", oidc.ScopeOffline})

				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", "baz", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "baz", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			err:      fosite.ErrInvalidScope,
			expected: "The requested scope is invalid, unknown, or malformed. The requested scope 'bar' was not originally granted by the resource owner.",
		},
		{
			name: "ShouldPassWithCustomLifespans",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClientWithCustomTokenLifespans{
					DefaultClient: &fosite.DefaultClient{
						ID:         "foo",
						GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
						Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
					},
				}

				requester.Client.(*fosite.DefaultClientWithCustomTokenLifespans).SetTokenLifespans(&TestLifespans)

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.NotEqual(t, sess, requester.Session)
				assert.NotEqual(t, time.Now().UTC().Add(-time.Hour).Round(time.Hour), requester.RequestedAt)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.RequestedScope)
				assert.NotEqual(t, url.Values{"foo": []string{"bar"}}, requester.Form)

				require.WithinDuration(t, time.Now().Add(*TestLifespans.RefreshTokenGrantAccessTokenLifespan).UTC(), requester.GetSession().GetExpiresAt(fosite.AccessToken).UTC(), time.Minute)
				require.WithinDuration(t, time.Now().Add(*TestLifespans.RefreshTokenGrantRefreshTokenLifespan).UTC(), requester.GetSession().GetExpiresAt(fosite.RefreshToken).UTC(), time.Minute)
			},
		},
		{
			name: "ShouldFailWithoutOfflineScope",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar"},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo"},
					RequestedScope: fosite.Arguments{"foo", "bar"},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			err:      fosite.ErrScopeNotGranted,
			expected: "The token was not granted the requested scope. The OAuth 2.0 Client was not granted scope offline and may thus not perform the 'refresh_token' authorization grant.",
		},
		{
			name: "ShouldPassWithoutOfflineScopeWhenConfigured",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				config.RefreshTokenScopes = []string{}
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar"},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo"},
					RequestedScope: fosite.Arguments{"foo", "bar"},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.NotEqual(t, sess, requester.Session)
				assert.NotEqual(t, time.Now().UTC().Add(-time.Hour).Round(time.Hour), requester.RequestedAt)
				assert.Equal(t, fosite.Arguments{"foo"}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo"}, requester.RequestedScope)
				assert.NotEqual(t, url.Values{"foo": []string{"bar"}}, requester.Form)
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.AccessToken))
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.RefreshToken))
			},
		},
		{
			name: "ShouldFailOnReuse",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				req := &fosite.Request{
					Client:         requester.Client,
					GrantedScope:   fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope: fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					Session:        sess,
					Form:           url.Values{"foo": []string{"bar"}},
					RequestedAt:    time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				}
				err = store.CreateRefreshTokenSession(context.TODO(), sig, req)
				require.NoError(t, err)

				err = store.RevokeRefreshToken(context.TODO(), req.ID)
				require.NoError(t, err)
			},
			err:      fosite.ErrInactiveToken,
			expected: "Token is inactive because it is malformed, expired or otherwise invalid. Token validation failed. Token is inactive because it is malformed, expired or otherwise invalid. Token validation failed.",
		},
		{
			name: "ShouldFailOnMalformedAudience",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:          requester.Client,
					GrantedScope:    fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope:  fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					GrantedAudience: fosite.Arguments{string([]byte{0x00, 0x01, 0x02})},
					Session:         sess,
					Form:            url.Values{"foo": []string{"bar"}},
					RequestedAt:     time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Unable to parse requested audience '\x00\x01\x02'. parse '\\x00\\x01\\x02': net/url: invalid control character in URL",
		},
		{
			name: "ShouldFailOnUnauthorizedAudience",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:          requester.Client,
					GrantedScope:    fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope:  fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					GrantedAudience: fosite.Arguments{"https://foo.com"},
					Session:         sess,
					Form:            url.Values{"foo": []string{"bar"}},
					RequestedAt:     time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Requested audience 'https://foo.com' has not been whitelisted by the OAuth 2.0 Client.",
		},
		{
			name: "ShouldPassOnPermittedAudienceAndGrantPreviousAudiences",
			setup: func(config *fosite.Config, strategy oauth2.RefreshTokenStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
					Audience:   []string{"https://foo.com"},
					Scopes:     []string{"foo", "bar", oidc.ScopeOffline},
				}

				token, sig, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)

				requester.Form.Add(oidc.FormParameterRefreshToken, token)
				err = store.CreateRefreshTokenSession(context.TODO(), sig, &fosite.Request{
					Client:          requester.Client,
					GrantedScope:    fosite.Arguments{"foo", oidc.ScopeOffline},
					RequestedScope:  fosite.Arguments{"foo", "bar", oidc.ScopeOffline},
					GrantedAudience: fosite.Arguments{"https://foo.com"},
					Session:         sess,
					Form:            url.Values{"foo": []string{"bar"}},
					RequestedAt:     time.Now().UTC().Add(-time.Hour).Round(time.Hour),
				})
				require.NoError(t, err)
			},
			texpected: func(t *testing.T, requester *fosite.AccessRequest) {
				assert.NotEqual(t, sess, requester.Session)
				assert.NotEqual(t, time.Now().UTC().Add(-time.Hour).Round(time.Hour), requester.RequestedAt)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.GrantedScope)
				assert.Equal(t, fosite.Arguments{"foo", oidc.ScopeOffline}, requester.RequestedScope)
				assert.Equal(t, fosite.Arguments{"https://foo.com"}, requester.GrantedAudience)
				assert.Equal(t, fosite.Arguments(nil), requester.RequestedAudience)
				assert.NotEqual(t, url.Values{"foo": []string{"bar"}}, requester.Form)
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.AccessToken))
				assert.Equal(t, time.Now().Add(time.Hour).UTC().Round(time.Second), requester.GetSession().GetExpiresAt(fosite.RefreshToken))
			},
		},
	}

	factory := func() (oauth2.RefreshTokenStrategy, *storage.MemoryStore, *fosite.AccessRequest) {
		return &oauth2.HMACSHAStrategy{
			Enigma: &hmac.HMACStrategy{Config: &fosite.Config{GlobalSecret: []byte("foobarfoobarfoobarfoobarfoobarfoobarfoobarfoobar")}},
			Config: &fosite.Config{
				AccessTokenLifespan:   time.Hour * 24,
				AuthorizeCodeLifespan: time.Hour * 24,
			},
		}, storage.NewMemoryStore(), fosite.NewAccessRequest(&fosite.DefaultSession{})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy, store, requester := factory()

			config := &fosite.Config{
				AccessTokenLifespan:      time.Hour,
				RefreshTokenLifespan:     time.Hour,
				ScopeStrategy:            fosite.HierarchicScopeStrategy,
				AudienceMatchingStrategy: fosite.DefaultAudienceMatchingStrategy,
				RefreshTokenScopes:       []string{oidc.ScopeOffline},
			}

			handler := &oidc.RefreshTokenGrantHandler{
				TokenRevocationStorage: store,
				RefreshTokenStrategy:   strategy,
				Config:                 config,
			}

			requester.Form = url.Values{}

			tc.setup(config, strategy, store, requester)

			err := handler.HandleTokenEndpointRequest(context.TODO(), requester)
			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), err.Error())
				if tc.rexpected != nil {
					assert.Regexp(t, tc.rexpected, oidc.ErrorToDebugRFC6749Error(err))
				} else {
					assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.expected)
				}
			} else {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
			}

			if tc.texpected != nil {
				tc.texpected(t, requester)
			}
		})
	}
}

func TestRefreshTokenGrantHandler_PopulateTokenEndpointResponse(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(config *fosite.Config, strategy oauth2.CoreStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest)
		check    func(t *testing.T, strategy oauth2.CoreStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest, responder *fosite.AccessResponse)
		err      error
		expected string
	}{
		{
			name: "ShouldPass",
			setup: func(config *fosite.Config, strategy oauth2.CoreStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.ID = "req-id"
				requester.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				requester.RequestedScope = fosite.Arguments{"foo", "bar"}
				requester.GrantedScope = fosite.Arguments{"foo", "bar"}

				token, signature, err := strategy.GenerateRefreshToken(context.TODO(), nil)
				require.NoError(t, err)
				require.NoError(t, store.CreateRefreshTokenSession(context.TODO(), signature, requester))
				requester.Form.Add(oidc.FormParameterRefreshToken, token)
			},
			check: func(t *testing.T, strategy oauth2.CoreStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest, responder *fosite.AccessResponse) {
				signature := strategy.RefreshTokenSignature(context.Background(), requester.Form.Get(oidc.FormParameterRefreshToken))

				// The old refresh token should be deleted.
				_, err := store.GetRefreshTokenSession(context.TODO(), signature, nil)
				require.Error(t, err)

				assert.Equal(t, "req-id", requester.ID)
				require.NoError(t, strategy.ValidateAccessToken(context.TODO(), requester, responder.GetAccessToken()))
				require.NoError(t, strategy.ValidateRefreshToken(context.TODO(), requester, responder.ToMap()[oidc.FormParameterRefreshToken].(string)))
				assert.Equal(t, fosite.BearerAccessToken, responder.GetTokenType())
				assert.NotEmpty(t, responder.ToMap()["expires_in"])
				assert.Equal(t, "foo bar", responder.ToMap()[oidc.FormParameterScope])
			},
		},
		{
			name: "ShouldFailNotResponsible",
			setup: func(config *fosite.Config, strategy oauth2.CoreStrategy, store *storage.MemoryStore, requester *fosite.AccessRequest) {
				requester.GrantTypes = fosite.Arguments{"313"}
			},
			err:      fosite.ErrUnknownRequest,
			expected: "The handler is not responsible for this request.",
		},
	}

	factory := func() (oauth2.CoreStrategy, *storage.MemoryStore, *fosite.AccessRequest, *fosite.AccessResponse) {
		return &oauth2.HMACSHAStrategy{
			Enigma: &hmac.HMACStrategy{Config: &fosite.Config{GlobalSecret: []byte("foobarfoobarfoobarfoobarfoobarfoobarfoobarfoobar")}},
			Config: &fosite.Config{
				AccessTokenLifespan:   time.Hour * 24,
				AuthorizeCodeLifespan: time.Hour * 24,
			},
		}, storage.NewMemoryStore(), fosite.NewAccessRequest(&fosite.DefaultSession{}), fosite.NewAccessResponse()
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy, store, requester, responder := factory()

			config := &fosite.Config{
				AccessTokenLifespan:      time.Hour,
				ScopeStrategy:            fosite.HierarchicScopeStrategy,
				AudienceMatchingStrategy: fosite.DefaultAudienceMatchingStrategy,
			}

			h := oidc.RefreshTokenGrantHandler{
				TokenRevocationStorage: store,
				RefreshTokenStrategy:   strategy,
				AccessTokenStrategy:    strategy,
				Config:                 config,
			}

			requester.Client = &fosite.DefaultClient{}
			requester.Form = url.Values{}

			tc.setup(config, strategy, store, requester)

			err := h.PopulateTokenEndpointResponse(context.TODO(), requester, responder)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.expected)
			} else {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
			}

			if tc.check != nil {
				tc.check(t, strategy, store, requester, responder)
			}
		})
	}
}

func TestRefreshFlowSanitizeRestoreOriginalRequest(t *testing.T) {
	testCases := []struct {
		name      string
		requester fosite.Requester
		original  fosite.Requester
		expected  fosite.Arguments
	}{
		{
			"ShouldRestoreIDAndScopeWhenAccessRequest",
			&fosite.AccessRequest{
				Request: fosite.Request{
					ID: "test",
				},
			},
			&fosite.AccessRequest{
				Request: fosite.Request{
					ID:           "test2",
					GrantedScope: fosite.Arguments{abc, "123"},
				},
			},
			fosite.Arguments{abc, "123"},
		},
		{
			"ShouldRestoreIDAndNotRestoreScopeWhenRequest",
			&fosite.Request{
				ID: "test",
			},
			&fosite.AccessRequest{
				Request: fosite.Request{
					ID:           "test2",
					GrantedScope: fosite.Arguments{abc, "123"},
				},
			},
			fosite.Arguments(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := oidc.RefreshFlowSanitizeRestoreOriginalRequest(tc.requester, tc.original)

			assert.Equal(t, tc.original.GetID(), actual.GetID())
			assert.Equal(t, tc.expected, actual.GetGrantedScopes())
		})
	}
}

func TestRefreshTokenGrantHandler_HandleTokenEndpointRequest_Tx(t *testing.T) {
	type store struct {
		storage.Transactional
		oauth2.TokenRevocationStorage
	}

	testCases := []struct {
		name      string
		setup     func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage)
		err       error
		expected  string
		texpected func(t *testing.T, request *fosite.AccessRequest)
	}{
		{
			name: "ShouldRevokeSessionOnTokenReuse",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				request.Client = &fosite.DefaultClient{
					ID:         "foo",
					GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
				}
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(request, fosite.ErrInactiveToken).
					Times(1)
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					DeleteRefreshTokenSession(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				transactional.
					EXPECT().
					Commit(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInactiveToken,
			expected: "Token is inactive because it is malformed, expired or otherwise invalid. Token validation failed. Token is inactive because it is malformed, expired or otherwise invalid. Token validation failed.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			request := fosite.NewAccessRequest(&fosite.DefaultSession{})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			transactional := mocks.NewMockTransactional(ctrl)
			revocation := mocks.NewMockTokenRevocationStorage(ctrl)
			tc.setup(ctx, request, transactional, revocation)

			strategy := &oauth2.HMACSHAStrategy{
				Enigma: &hmac.HMACStrategy{Config: &fosite.Config{GlobalSecret: []byte("foobarfoobarfoobarfoobarfoobarfoobarfoobarfoobar")}},
				Config: &fosite.Config{
					AccessTokenLifespan:   time.Hour * 24,
					AuthorizeCodeLifespan: time.Hour * 24,
				},
			}

			handler := oidc.RefreshTokenGrantHandler{
				TokenRevocationStorage: store{
					transactional,
					revocation,
				},
				AccessTokenStrategy:  strategy,
				RefreshTokenStrategy: strategy,
				Config: &fosite.Config{
					AccessTokenLifespan:      time.Hour,
					ScopeStrategy:            fosite.HierarchicScopeStrategy,
					AudienceMatchingStrategy: fosite.DefaultAudienceMatchingStrategy,
				},
			}

			err := handler.HandleTokenEndpointRequest(ctx, request)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.expected)
			} else {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
			}

			if tc.texpected != nil {
				tc.texpected(t, request)
			}
		})
	}
}

func TestRefreshTokenGrantHandler_PopulateTokenEndpointResponse_Tx(t *testing.T) {
	type store struct {
		storage.Transactional
		oauth2.TokenRevocationStorage
	}

	testCases := []struct {
		name      string
		setup     func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage)
		err       error
		expected  string
		texpected func(t *testing.T, request *fosite.AccessRequest, response *fosite.AccessResponse)
	}{
		{
			name: "ShouldSuccessfullyCommitTxWhenNoErrors",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateAccessTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateRefreshTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				transactional.
					EXPECT().
					Commit(ctx).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "ShouldSuccessfullyRollbackTxWhenErrorsFromGetRefreshTokenSession",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(nil, errors.New("Whoops, a nasty database error occurred!")).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Whoops, a nasty database error occurred!",
		},
		{
			name: "ShouldErrorErrInvalidRequestWhenGetRefreshTokenSessionErrNotFound",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(nil, fosite.ErrNotFound).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Failed to refresh token because of multiple concurrent requests using the same token which is not allowed. Could not find the requested resource(s).",
		},
		{
			name: "ShouldSuccessfullyRollbackTxWhenErrorsFromRevokeAccessToken",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(errors.New("Whoops, a nasty database error occurred!")).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Whoops, a nasty database error occurred!",
		},
		{
			name: "ShouldErrorErrInvalidRequestWhenRevokeAccessTokenErrSerializationFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(fosite.ErrSerializationFailure).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Failed to refresh token because of multiple concurrent requests using the same token which is not allowed. The request could not be completed due to concurrent access",
		},
		{
			name: "ShouldErrorErrInactiveTokenWhenRevokeAccessTokenErrInvalidRequest",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(nil, fosite.ErrInactiveToken).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Failed to refresh token because of multiple concurrent requests using the same token which is not allowed. Token is inactive because it is malformed, expired or otherwise invalid. Token validation failed.",
		},
		{
			name: "ShouldSuccessfullyRollbackTxWhenErrorsFromRevokeRefreshTokenMaybeGracePeriod",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(errors.New("Whoops, a nasty database error occurred!")).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Whoops, a nasty database error occurred!",
		},
		{
			name: "ShouldErrorErrInvalidRequestWhenRevokeRefreshTokenMaybeGracePeriodErrSerializationFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(fosite.ErrSerializationFailure).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Failed to refresh token because of multiple concurrent requests using the same token which is not allowed. The request could not be completed due to concurrent access",
		},
		{
			name: "ShouldErrorErrInvalidRequestWhenCreateAccessTokenSessionErrSerializationFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateAccessTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(fosite.ErrSerializationFailure).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Failed to refresh token because of multiple concurrent requests using the same token which is not allowed. The request could not be completed due to concurrent access",
		},
		{
			name: "ShouldSuccessfullyRollbackTxWhenErrorsFromCreateAccessTokenSession",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateAccessTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(errors.New("Whoops, a nasty database error occurred!")).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Whoops, a nasty database error occurred!",
		},
		{
			name: "ShouldSuccessfullyRollbackTxWhenErrorsFromCreateRefreshTokenSession",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateAccessTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateRefreshTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(errors.New("Whoops, a nasty database error occurred!")).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Whoops, a nasty database error occurred!",
		},
		{
			name: "ShouldErrorErrInvalidRequestWhenCreateRefreshTokenSessionErrSerializationFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateAccessTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateRefreshTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(fosite.ErrSerializationFailure).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Failed to refresh token because of multiple concurrent requests using the same token which is not allowed. The request could not be completed due to concurrent access",
		},
		{
			name: "ShouldErrorErrServerErrorWhenBeginTxFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(nil, errors.New("Could not create transaction!")).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Could not create transaction!",
		},
		{
			name: "ShouldErrorErrServerErrorWhenRollbackTxFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(nil, fosite.ErrNotFound).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(errors.New("Could not rollback transaction!")).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. error: invalid_request; rollback error: Could not rollback transaction!",
		},
		{
			name: "ShouldErrorErrServerErrorWhenCommitTxFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateAccessTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateRefreshTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				transactional.
					EXPECT().
					Commit(ctx).
					Return(errors.New("Could not commit transaction!")).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrServerError,
			expected: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Could not commit transaction!",
		},
		{
			name: "ShouldErrorErrInvalidRequestWhenCommitTxErrSerializationFailure",
			setup: func(ctx context.Context, request *fosite.AccessRequest, transactional *mocks.MockTransactional, revocation *mocks.MockTokenRevocationStorage) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeRefreshToken}
				transactional.
					EXPECT().
					BeginTX(ctx).
					Return(ctx, nil).
					Times(1)
				revocation.
					EXPECT().
					GetRefreshTokenSession(ctx, gomock.Any(), nil).
					Return(request, nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeAccessToken(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					RevokeRefreshTokenMaybeGracePeriod(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateAccessTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				revocation.
					EXPECT().
					CreateRefreshTokenSession(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				transactional.
					EXPECT().
					Commit(ctx).
					Return(fosite.ErrSerializationFailure).
					Times(1)
				transactional.
					EXPECT().
					Rollback(ctx).
					Return(nil).
					Times(1)
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Failed to refresh token because of multiple concurrent requests using the same token which is not allowed. The request could not be completed due to concurrent access",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			transactional := mocks.NewMockTransactional(ctrl)
			revocation := mocks.NewMockTokenRevocationStorage(ctrl)

			ctx := context.Background()

			request := fosite.NewAccessRequest(&fosite.DefaultSession{})

			tc.setup(ctx, request, transactional, revocation)

			strategy := &oauth2.HMACSHAStrategy{
				Enigma: &hmac.HMACStrategy{Config: &fosite.Config{GlobalSecret: []byte("foobarfoobarfoobarfoobarfoobarfoobarfoobarfoobar")}},
				Config: &fosite.Config{
					AccessTokenLifespan:   time.Hour * 24,
					AuthorizeCodeLifespan: time.Hour * 24,
				},
			}

			handler := oidc.RefreshTokenGrantHandler{
				// Notice how we are passing in a store that has support for transactions!
				TokenRevocationStorage: store{
					transactional,
					revocation,
				},
				AccessTokenStrategy:  strategy,
				RefreshTokenStrategy: strategy,
				Config: &fosite.Config{
					AccessTokenLifespan:      time.Hour,
					ScopeStrategy:            fosite.HierarchicScopeStrategy,
					AudienceMatchingStrategy: fosite.DefaultAudienceMatchingStrategy,
				},
			}

			response := fosite.NewAccessResponse()

			err := handler.PopulateTokenEndpointResponse(ctx, request, response)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.expected)
			} else {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
			}

			if tc.texpected != nil {
				tc.texpected(t, request, response)
			}
		})
	}
}

var TestLifespans = fosite.ClientLifespanConfig{
	AuthorizationCodeGrantAccessTokenLifespan:  ptr(31 * time.Hour),
	AuthorizationCodeGrantIDTokenLifespan:      ptr(32 * time.Hour),
	AuthorizationCodeGrantRefreshTokenLifespan: ptr(33 * time.Hour),
	ClientCredentialsGrantAccessTokenLifespan:  ptr(34 * time.Hour),
	ImplicitGrantAccessTokenLifespan:           ptr(35 * time.Hour),
	ImplicitGrantIDTokenLifespan:               ptr(36 * time.Hour),
	JwtBearerGrantAccessTokenLifespan:          ptr(37 * time.Hour),
	PasswordGrantAccessTokenLifespan:           ptr(38 * time.Hour),
	PasswordGrantRefreshTokenLifespan:          ptr(39 * time.Hour),
	RefreshTokenGrantIDTokenLifespan:           ptr(40 * time.Hour),
	RefreshTokenGrantAccessTokenLifespan:       ptr(41 * time.Hour),
	RefreshTokenGrantRefreshTokenLifespan:      ptr(42 * time.Hour),
}

func ptr(d time.Duration) *time.Duration {
	return &d
}
