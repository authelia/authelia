// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestGitHubProviderAuthorizationRequest(t *testing.T) {
	provider := newGitHubProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
		Scopes: []string{"read:user", "user:email"}, TokenEndpointAuthMethod: "client_secret_basic",
	}, newProviderClient(nil))

	request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/api/identity/github/callback"})
	require.NoError(t, err)

	assert.Len(t, request.State, 43)
	assert.Empty(t, request.Nonce, "GitHub issues no ID Token, so there is no nonce to bind it to")
	assert.Len(t, request.CodeVerifier, 43)

	uri, err := url.Parse(request.URL)
	require.NoError(t, err)

	assert.Equal(t, "https", uri.Scheme)
	assert.Equal(t, "github.com", uri.Host)
	assert.Equal(t, "/login/oauth/authorize", uri.Path)

	query := uri.Query()

	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "Iv1.abc", query.Get("client_id"))
	assert.Equal(t, "read:user user:email", query.Get("scope"))
	assert.Equal(t, "https://auth.example.com/api/identity/github/callback", query.Get("redirect_uri"))
	assert.Equal(t, request.State, query.Get("state"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
	assert.NotEmpty(t, query.Get("code_challenge"))
	assert.False(t, query.Has("nonce"))
}

func TestGitHubProviderComplete(t *testing.T) {
	testCases := []struct {
		Name         string
		Method       string
		UserStatus   int
		User         any
		EmailsStatus int
		Emails       any
		Expected     *IdentityClaims
		Error        string
	}{
		{
			Name:         "ShouldReturnTheUserWithTheVerifiedPrimaryEmail",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat", "name": "The Octocat"},
			EmailsStatus: http.StatusOK,
			Emails: []map[string]any{
				{"email": "octocat@users.noreply.github.com", "primary": false, "verified": true},
				{"email": "octocat@github.com", "primary": true, "verified": true},
			},
			Expected: &IdentityClaims{Issuer: "https://github.com/login/oauth", Subject: "583231", PreferredUsername: "octocat", Name: "The Octocat", Email: "octocat@github.com"},
		},
		{
			Name:         "ShouldAuthenticateWithThePostMethod",
			Method:       "client_secret_post",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat", "name": "The Octocat"},
			EmailsStatus: http.StatusOK,
			Emails:       []map[string]any{},
			Expected:     &IdentityClaims{Issuer: "https://github.com/login/oauth", Subject: "583231", PreferredUsername: "octocat", Name: "The Octocat"},
		},
		{
			Name:         "ShouldUseTheLoginWithoutAName",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat", "name": nil},
			EmailsStatus: http.StatusOK,
			Emails:       []map[string]any{},
			Expected:     &IdentityClaims{Issuer: "https://github.com/login/oauth", Subject: "583231", PreferredUsername: "octocat", Name: "octocat"},
		},
		{
			Name:         "ShouldOmitAnUnverifiedPrimaryEmail",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat"},
			EmailsStatus: http.StatusOK,
			Emails:       []map[string]any{{"email": "octocat@github.com", "primary": true, "verified": false}},
			Expected:     &IdentityClaims{Issuer: "https://github.com/login/oauth", Subject: "583231", PreferredUsername: "octocat", Name: "octocat"},
		},
		{
			Name:         "ShouldRaiseErrorWhenTheConfiguredEmailScopeIsNotGranted",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat"},
			EmailsStatus: http.StatusNotFound,
			Emails:       map[string]any{"message": "Not Found"},
			Error:        "error requesting the github emails: the github response is invalid: the endpoint returned status code 404",
		},
		{
			Name:       "ShouldRaiseErrorOnAbsentID",
			Method:     "client_secret_basic",
			UserStatus: http.StatusOK,
			User:       map[string]any{"login": "octocat"},
			Error:      "error requesting the github user: the github response is invalid: the 'id' is required but it is absent",
		},
		{
			Name:       "ShouldRaiseErrorOnErrorStatus",
			Method:     "client_secret_basic",
			UserStatus: http.StatusUnauthorized,
			User:       map[string]any{"message": "Bad credentials"},
			Error:      "error requesting the github user: the github response is invalid: the endpoint returned status code 401",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := http.NewServeMux()

			mux.HandleFunc("/login/oauth/access_token", func(rw http.ResponseWriter, r *http.Request) {
				require.NoError(t, r.ParseForm())

				assert.Equal(t, "authorization_code", r.PostForm.Get("grant_type"))
				assert.Equal(t, "the-code", r.PostForm.Get("code"))
				assert.Equal(t, "the-verifier", r.PostForm.Get("code_verifier"))
				assert.Equal(t, "https://auth.example.com/cb", r.PostForm.Get("redirect_uri"))

				switch tc.Method {
				case "client_secret_basic":
					username, password, ok := r.BasicAuth()

					assert.True(t, ok)
					assert.Equal(t, "Iv1.abc", username)
					assert.Equal(t, "secret", password)
				case "client_secret_post":
					assert.Equal(t, "Iv1.abc", r.PostForm.Get("client_id"))
					assert.Equal(t, "secret", r.PostForm.Get("client_secret"))
				}

				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "gho_token", "token_type": "bearer", "scope": "read:user,user:email"})
			})

			mux.HandleFunc("/user", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer gho_token", r.Header.Get(headerAuthorization))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get(headerAccept))
				assert.Equal(t, "2026-03-10", r.Header.Get(headerGitHubAPIVersion))

				rw.Header().Set(headerContentType, "application/json; charset=utf-8")
				rw.WriteHeader(tc.UserStatus)

				_ = json.NewEncoder(rw).Encode(tc.User)
			})

			mux.HandleFunc("/user/emails", func(rw http.ResponseWriter, r *http.Request) {
				if tc.Emails == nil {
					t.Errorf("the emails must not be requested")
				}

				assert.Equal(t, "Bearer gho_token", r.Header.Get(headerAuthorization))
				assert.Equal(t, "100", r.URL.Query().Get("per_page"))

				rw.Header().Set(headerContentType, "application/json; charset=utf-8")
				rw.WriteHeader(tc.EmailsStatus)

				_ = json.NewEncoder(rw).Encode(tc.Emails)
			})

			server := httptest.NewServer(mux)

			defer server.Close()

			claims, err := newTestGitHubProvider(server.URL, tc.Method).Complete(context.Background(), CompletionRequest{Code: "the-code", CodeVerifier: "the-verifier", RedirectURI: "https://auth.example.com/cb", Now: time.Now()})

			if tc.Error != "" {
				assert.Nil(t, claims)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)

			claims.RawProviders = nil

			assert.Equal(t, tc.Expected, claims)
		})
	}
}

func TestGitHubProviderMetadata(t *testing.T) {
	provider := newGitHubProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
		AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}},
	}, newProviderClient(nil))

	assert.Equal(t, "github", provider.ID())
	assert.Equal(t, "GitHub", provider.Name())
	assert.Equal(t, "", provider.LogoURI())
	assert.Equal(t, ProviderTypeGitHub, provider.Type())
	assert.Equal(t, "https://github.com/login/oauth", provider.Issuer())
	assert.Equal(t, ResponseModeQuery, provider.ResponseMode())
	assert.Equal(t, []string{"pwd"}, provider.AuthenticationMethodsReference(nil))
	assert.Equal(t, []string{"pwd", "kba"}, newTestGitHubProvider("https://github.com", "client_secret_basic").AuthenticationMethodsReference(nil))
}

func newTestGitHubProvider(base, method string) *GitHubProvider {
	provider := newGitHubProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
		Scopes: []string{"read:user", "user:email"}, TokenEndpointAuthMethod: method,
	}, newProviderClient(nil))

	provider.client.RetryMax = 0
	provider.tokenEndpoint = base + "/login/oauth/access_token"
	provider.userEndpoint = base + "/user"
	provider.emailsEndpoint = base + "/user/emails"
	provider.organizationsEndpoint = base + "/user/orgs"
	provider.teamsEndpoint = base + "/user/teams"

	return provider
}

func TestGitHubProviderCompleteContentType(t *testing.T) {
	testCases := []struct {
		Name        string
		ContentType string
		Error       string
	}{
		{"ShouldAcceptJSON", "application/json; charset=utf-8", ""},
		{"ShouldAcceptGitHubJSON", mimeApplicationGitHubJSON, ""},
		{"ShouldAcceptGitHubJSONWithParameters", "application/vnd.github+json; charset=utf-8", ""},
		{"ShouldRejectHTML", "text/html", "error requesting the github user: the github response is invalid: the content type 'text/html' is not 'application/json' or 'application/vnd.github+json'"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := http.NewServeMux()

			mux.HandleFunc("/login/oauth/access_token", func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "gho_token", "token_type": "bearer"})
			})

			mux.HandleFunc("/user", func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, tc.ContentType)

				_ = json.NewEncoder(rw).Encode(map[string]any{"id": 1, "login": "john"})
			})

			mux.HandleFunc("/user/emails", func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, tc.ContentType)

				_ = json.NewEncoder(rw).Encode([]map[string]any{{"email": "john@example.com", "primary": true, "verified": true}})
			})

			server := httptest.NewServer(mux)

			defer server.Close()

			claims, err := newTestGitHubProvider(server.URL, "client_secret_basic").Complete(context.Background(), CompletionRequest{Code: "the-code", CodeVerifier: "the-verifier", RedirectURI: "https://auth.example.com/cb", Now: time.Now()})

			if tc.Error != "" {
				assert.Nil(t, claims)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)

			claims.RawProviders = nil

			assert.Equal(t, &IdentityClaims{Issuer: githubIssuer, Subject: "1", PreferredUsername: "john", Name: "john", Email: "john@example.com"}, claims)
		})
	}
}

func TestGitHubProviderCompleteShouldRetrieveEveryPageOfEmails(t *testing.T) {
	const filler = `{"email":"filler@example.com","primary":false,"verified":true}`

	fullPage := "[" + strings.TrimSuffix(strings.Repeat(filler+",", 100), ",") + "]"

	testCases := []struct {
		name     string
		lastPage int
		expected string
		count    int
		err      string
	}{
		{"ShouldFindThePrimaryEmailOnALaterPage", 2, "primary@example.com", 101, ""},
		{"ShouldFailWhenThePageLimitIsExceeded", githubListPagesLimit + 1, "", 0, "error requesting the github emails: the github response is invalid: the list exceeds 10 pages"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, mimeApplicationJSON)

				switch r.URL.Path {
				case "/login/oauth/access_token":
					_, _ = rw.Write([]byte(`{"access_token":"token","token_type":"Bearer"}`))
				case "/user":
					_, _ = rw.Write([]byte(`{"id":1,"login":"octocat"}`))
				case "/user/emails":
					assert.Equal(t, "100", r.URL.Query().Get("per_page"))

					if r.URL.Query().Get("page") == strconv.Itoa(tc.lastPage) {
						_, _ = rw.Write([]byte(`[{"email":"primary@example.com","primary":true,"verified":true}]`))

						return
					}

					_, _ = rw.Write([]byte(fullPage))
				}
			}))

			defer server.Close()

			claims, err := testGitHubComplete(newTestGitHubProvider(server.URL, "client_secret_basic"))

			if tc.err != "" {
				assert.Nil(t, claims)
				require.EqualError(t, err, tc.err)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expected, claims.Email)
			assert.Len(t, claims.RawProviders["github"]["emails"], tc.count)
		})
	}
}

func TestGitHubProviderCompleteShouldExposeRawClaimsToTheClaimsMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/login/oauth/access_token":
			_, _ = rw.Write([]byte(`{"access_token":"token","token_type":"Bearer"}`))
		case "/user":
			_ = json.NewEncoder(rw).Encode(map[string]any{"id": 1, "login": "octocat", "company": "GitHub"})
		default:
			_, _ = rw.Write([]byte(`[]`))
		}
	}))

	defer server.Close()

	provider := newClaimsMappedProvider(newTestGitHubProvider(server.URL, "client_secret_basic"), map[string]string{"name": "claims.github.company"})

	claims, err := provider.Complete(context.Background(), CompletionRequest{Code: "the-code", CodeVerifier: "the-verifier", RedirectURI: "https://auth.example.com/cb", Now: time.Now()})
	require.NoError(t, err)

	assert.Equal(t, "GitHub", claims.Name)
	assert.Equal(t, "octocat", claims.PreferredUsername)
}

func TestGitHubProviderCompleteShouldExposeEmailsToTheClaimsMapping(t *testing.T) {
	const mappingPresence = `has(claims.github.emails) ? "present" : "absent"`

	testCases := []struct {
		name        string
		scopes      []string
		status      int
		contentType string
		body        string
		mapping     string
		expected    string
		present     bool
		failure     string
		unrequested bool
	}{
		{
			name:     "ShouldExposeEveryPropertyOfEveryEmail",
			scopes:   []string{"user:email"},
			status:   http.StatusOK,
			body:     `[{"email":"primary@example.com","primary":true,"verified":true,"visibility":"public"},{"email":"other@example.com","primary":false,"verified":false,"visibility":null}]`,
			mapping:  `claims.github.emails.filter(e, !e.verified)[0].email + "," + claims.github.emails[0].visibility`,
			expected: "other@example.com,public",
			present:  true,
		},
		{
			name:     "ShouldExposeAnEmptyList",
			scopes:   []string{"user:email"},
			status:   http.StatusOK,
			body:     `[]`,
			mapping:  `has(claims.github.emails) && size(claims.github.emails) == 0 ? "empty" : "other"`,
			expected: "empty",
			present:  true,
		},
		{
			name:     "ShouldMatchTheDomainOfAnEmailRegardlessOfCase",
			scopes:   []string{"user:email"},
			status:   http.StatusOK,
			body:     `[{"email":"Jane@Example.COM","primary":true,"verified":true}]`,
			mapping:  `claims.github.emails.exists(e, e.verified && e.email.lowerAscii().endsWith("@example.com")) ? "match" : "other"`,
			expected: "match",
			present:  true,
		},
		{
			name:        "ShouldNotRequestTheEmailsWithoutAnEmailScope",
			scopes:      []string{"read:user"},
			status:      http.StatusOK,
			body:        `[{"email":"primary@example.com","primary":true,"verified":true}]`,
			mapping:     mappingPresence,
			expected:    "absent",
			unrequested: true,
		},
		{
			name:     "ShouldRequestTheEmailsWithTheUserScope",
			scopes:   []string{"user"},
			status:   http.StatusOK,
			body:     `[]`,
			mapping:  mappingPresence,
			expected: "present",
			present:  true,
		},
		{
			name:    "ShouldFailOnANullEntry",
			scopes:  []string{"user:email"},
			status:  http.StatusOK,
			body:    `[null]`,
			failure: "the github response is invalid: the email address at index 0 is not an object",
		},
		{
			name:    "ShouldFailOnAnEntryWithoutAProperty",
			scopes:  []string{"user:email"},
			status:  http.StatusOK,
			body:    `[{"email":"primary@example.com","primary":true,"verified":true},{"email":"other@example.com","primary":false}]`,
			failure: "the github response is invalid: the email address at index 1 requires the 'email', 'primary', and 'verified' properties",
		},
		{
			name:    "ShouldFailWhenTheEmailScopeIsNotGranted",
			scopes:  []string{"user:email"},
			status:  http.StatusForbidden,
			body:    `{"message":"Forbidden"}`,
			failure: "the github response is invalid: the endpoint returned status code 403",
		},
		{
			name:    "ShouldFailWhenTheUserScopeIsNotGranted",
			scopes:  []string{"user"},
			status:  http.StatusNotFound,
			body:    `{"message":"Not Found"}`,
			failure: "the github response is invalid: the endpoint returned status code 404",
		},
		{
			name:    "ShouldFailOnAServerError",
			scopes:  []string{"user:email"},
			status:  http.StatusInternalServerError,
			body:    `{"message":"Server Error"}`,
			failure: "/user/emails?page=1&per_page=100 giving up after",
		},
		{
			name:        "ShouldFailOnAnInvalidContentType",
			scopes:      []string{"user:email"},
			status:      http.StatusOK,
			contentType: "text/html",
			body:        `[]`,
			failure:     "the github response is invalid: the content type 'text/html' is not 'application/json' or 'application/vnd.github+json'",
		},
		{
			name:    "ShouldFailOnMalformedJSON",
			scopes:  []string{"user:email"},
			status:  http.StatusOK,
			body:    `[{`,
			failure: "the github response is invalid: unexpected end of JSON input",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, mimeApplicationJSON)

				switch r.URL.Path {
				case "/login/oauth/access_token":
					_, _ = rw.Write([]byte(`{"access_token":"token","token_type":"Bearer"}`))
				case "/user":
					_, _ = rw.Write([]byte(`{"id":1,"login":"octocat"}`))
				case "/user/emails":
					if tc.unrequested {
						t.Errorf("the emails must not be requested")
					}

					if tc.contentType != "" {
						rw.Header().Set(headerContentType, tc.contentType)
					}

					rw.WriteHeader(tc.status)

					_, _ = rw.Write([]byte(tc.body))
				}
			}))

			defer server.Close()

			inner := newTestGitHubProvider(server.URL, "client_secret_basic")
			inner.scopes = tc.scopes

			claims, err := testGitHubComplete(inner)

			if tc.failure != "" {
				assert.Nil(t, claims)
				require.ErrorContains(t, err, "error requesting the github emails: ")
				assert.ErrorContains(t, err, tc.failure)

				return
			}

			require.NoError(t, err)

			if tc.present {
				assert.Contains(t, claims.RawProviders["github"], "emails")
			} else {
				assert.NotContains(t, claims.RawProviders["github"], "emails")
			}

			mapped, err := testGitHubComplete(newClaimsMappedProvider(inner, map[string]string{"name": tc.mapping}))
			require.NoError(t, err)

			assert.Equal(t, tc.expected, mapped.Name)
		})
	}
}

// This is the expression from the Email Addresses section of the GitHub integration documentation.
func TestGitHubProviderCompleteShouldApplyTheDocumentedEmailsClaimsMapping(t *testing.T) {
	const mapping = `!has(claims.github.emails) || !claims.github.emails.exists(e, e.verified && e.email.lowerAscii().endsWith("@example.com"))`

	testCases := []struct {
		name     string
		scopes   []string
		body     string
		disabled bool
	}{
		{"ShouldAllowAVerifiedAddress", []string{"user:email"}, `[{"email":"jane@example.com","primary":true,"verified":true}]`, false},
		{"ShouldAllowAVerifiedAddressRegardlessOfCase", []string{"user:email"}, `[{"email":"Jane@Example.COM","primary":true,"verified":true}]`, false},
		{"ShouldAllowAVerifiedSecondaryAddress", []string{"user"}, `[{"email":"jane@other.example","primary":true,"verified":true},{"email":"jane@example.com","primary":false,"verified":true}]`, false},
		{"ShouldDisableAnUnverifiedAddress", []string{"user:email"}, `[{"email":"jane@example.com","primary":true,"verified":false}]`, true},
		{"ShouldDisableAnAddressAtAnotherDomain", []string{"user:email"}, `[{"email":"jane@other.example","primary":true,"verified":true}]`, true},
		{"ShouldDisableASimilarDomain", []string{"user:email"}, `[{"email":"jane@notexample.com","primary":true,"verified":true}]`, true},
		{"ShouldDisableNoAddresses", []string{"user:email"}, `[]`, true},
		{"ShouldDisableWithoutAnEmailScope", []string{"read:user"}, `[{"email":"jane@example.com","primary":true,"verified":true}]`, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, mimeApplicationJSON)

				switch r.URL.Path {
				case "/login/oauth/access_token":
					_, _ = rw.Write([]byte(`{"access_token":"token","token_type":"Bearer"}`))
				case "/user":
					_, _ = rw.Write([]byte(`{"id":1,"login":"octocat"}`))
				case "/user/emails":
					_, _ = rw.Write([]byte(tc.body))
				}
			}))

			defer server.Close()

			inner := newTestGitHubProvider(server.URL, "client_secret_basic")
			inner.scopes = tc.scopes

			_, err := testGitHubComplete(newClaimsMappedProvider(inner, map[string]string{"disabled": mapping}))

			if tc.disabled {
				assert.ErrorIs(t, err, ErrIdentityDisabled)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newTestGitHubOrganizationsServer(t *testing.T, requested map[string]int, handler func(rw http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requested[r.URL.Path]++

		rw.Header().Set(headerContentType, mimeApplicationJSON)

		switch r.URL.Path {
		case "/login/oauth/access_token":
			_, _ = rw.Write([]byte(`{"access_token":"token","token_type":"Bearer"}`))
		case "/user":
			_ = json.NewEncoder(rw).Encode(map[string]any{"id": 1, "login": "octocat", "two_factor_authentication": true})
		case "/user/emails":
			_, _ = rw.Write([]byte(`[]`))
		default:
			handler(rw, r)
		}
	}))
}

func testGitHubComplete(provider Provider) (*IdentityClaims, error) {
	return provider.Complete(context.Background(), CompletionRequest{Code: "the-code", CodeVerifier: "the-verifier", RedirectURI: "https://auth.example.com/cb", Now: time.Now()})
}

func TestGitHubProviderCompleteShouldRetrieveOrganizationsAndTeamsWithReadOrg(t *testing.T) {
	requested := map[string]int{}

	server := newTestGitHubOrganizationsServer(t, requested, func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, githubAPIVersion, r.Header.Get(headerGitHubAPIVersion))
		assert.Equal(t, "100", r.URL.Query().Get("per_page"))

		if r.URL.Path == "/user/teams" {
			_ = json.NewEncoder(rw).Encode([]map[string]any{{"slug": "core", "organization": map[string]any{"login": "authelia"}}})

			return
		}

		if r.URL.Query().Get("page") == "1" {
			page := make([]map[string]any, 100)

			for i := range page {
				page[i] = map[string]any{"login": "filler"}
			}

			_ = json.NewEncoder(rw).Encode(page)

			return
		}

		_ = json.NewEncoder(rw).Encode([]map[string]any{{"login": "authelia"}})
	})

	defer server.Close()

	inner := newTestGitHubProvider(server.URL, "client_secret_basic")
	inner.scopes = []string{"read:user", "read:org"}

	testCases := []struct {
		name     string
		mapping  string
		expected bool
	}{
		{"ShouldDisableWithoutTheOrganization", `!claims.github.organizations.exists(o, o.login == "other")`, true},
		{"ShouldNotDisableWithTheOrganizationFromALaterPage", `!claims.github.organizations.exists(o, o.login == "authelia")`, false},
		{"ShouldNotDisableWithTheTeam", `!claims.github.teams.exists(t, t.organization.login == "authelia" && t.slug == "core")`, false},
		{"ShouldDisableWithoutTheTeam", `!claims.github.teams.exists(t, t.organization.login == "authelia" && t.slug == "other")`, true},
		{"ShouldExposeTwoFactorAuthentication", `claims.github.two_factor_authentication != true`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := testGitHubComplete(newClaimsMappedProvider(inner, map[string]string{"disabled": tc.mapping}))

			if tc.expected {
				assert.ErrorIs(t, err, ErrIdentityDisabled)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	assert.Equal(t, 2*len(testCases), requested["/user/orgs"])
	assert.Equal(t, len(testCases), requested["/user/teams"])
}

func TestGitHubProviderCompleteShouldNotRetrieveOrganizationsWithoutReadOrg(t *testing.T) {
	requested := map[string]int{}

	server := newTestGitHubOrganizationsServer(t, requested, func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[]`))
	})

	defer server.Close()

	claims, err := testGitHubComplete(newTestGitHubProvider(server.URL, "client_secret_basic"))
	require.NoError(t, err)

	assert.Equal(t, "octocat", claims.RawProviders["github"]["login"])
	assert.NotContains(t, claims.RawProviders["github"], "organizations")
	assert.NotContains(t, claims.RawProviders["github"], "teams")
	assert.Empty(t, claims.RawUserInfo)
	assert.Zero(t, requested["/user/orgs"])
	assert.Zero(t, requested["/user/teams"])
}

func TestGitHubProviderCompleteShouldFailClosedRetrievingOrganizations(t *testing.T) {
	full := make([]map[string]any, 100)

	for i := range full {
		full[i] = map[string]any{"login": "filler"}
	}

	testCases := []struct {
		name     string
		handler  func(rw http.ResponseWriter, r *http.Request)
		expected string
	}{
		{
			"ShouldFailWhenOrganizationsAreForbidden",
			func(rw http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/user/orgs" {
					rw.WriteHeader(http.StatusForbidden)

					return
				}

				_, _ = rw.Write([]byte(`[]`))
			},
			"error requesting the github organizations: the github response is invalid: the endpoint returned status code 403",
		},
		{
			"ShouldFailWhenTeamsAreForbidden",
			func(rw http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/user/teams" {
					rw.WriteHeader(http.StatusForbidden)

					return
				}

				_, _ = rw.Write([]byte(`[]`))
			},
			"error requesting the github teams: the github response is invalid: the endpoint returned status code 403",
		},
		{
			"ShouldFailWhenThereAreTooManyPages",
			func(rw http.ResponseWriter, _ *http.Request) {
				_ = json.NewEncoder(rw).Encode(full)
			},
			"error requesting the github organizations: the github response is invalid: the list exceeds 10 pages",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestGitHubOrganizationsServer(t, map[string]int{}, tc.handler)

			defer server.Close()

			provider := newTestGitHubProvider(server.URL, "client_secret_basic")
			provider.scopes = []string{"read:user", "read:org"}

			claims, err := testGitHubComplete(provider)

			assert.Nil(t, claims)
			assert.EqualError(t, err, tc.expected)
		})
	}
}
