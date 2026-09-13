// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
)

func TestConformanceRelyingPartyAuthenticated(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		level    int
		expected bool
	}{
		{"ShouldAcceptTheLinkedAccountSignedIn", "harry", 1, true},
		{"ShouldAcceptTheLinkedAccountWithASecondFactor", "harry", 2, true},
		{"ShouldRejectNoSession", "", 0, false},
		{"ShouldRejectAnotherAccount", "john", 1, false},
		{"ShouldRejectTheLinkedAccountWithoutAnAuthenticationLevel", "harry", 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConformanceRelyingPartyAuthenticated(tc.username, tc.level, "harry"))
		})
	}

	assert.False(t, ConformanceRelyingPartyAuthenticated("", 1, ""), "an account must be expected for a session to count as signed in")
}

func TestConformanceAssertRelyingPartyLeg(t *testing.T) {
	testCases := []struct {
		name     string
		module   string
		accepted bool
		err      string
	}{
		{"ShouldPassAnAcceptedHappyPath", "oidcc-client-test", true, ""},
		{"ShouldFailARejectedHappyPath", "oidcc-client-test", false, "expected Authelia to accept the identity asserted by the provider but it rejected it, redirecting to 'https://login.example.com:8080/'"},
		{"ShouldPassARejectedInvalidIssuer", "oidcc-client-test-invalid-iss", false, ""},
		{"ShouldFailAnAcceptedInvalidUserInfoSubject", "oidcc-client-test-userinfo-invalid-sub", true, "expected Authelia to reject the identity asserted by the provider but it accepted it, redirecting to 'https://login.example.com:8080/'"},
		{"ShouldPassEitherOutcomeOfMultipleMatchingKeys", "oidcc-client-test-kid-absent-multiple-jwks", true, ""},
		{"ShouldPassEitherOutcomeOfMultipleMatchingKeysRejected", "oidcc-client-test-kid-absent-multiple-jwks", false, ""},
		{"ShouldPassAModuleWithoutAnExpectation", "oidcc-client-test-some-future-module", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ConformanceAssertRelyingPartyLeg(tc.module, ConformanceRelyingPartyLeg{Location: "https://login.example.com:8080/", Accepted: tc.accepted, Notified: !tc.accepted})

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestConformanceRelyingPartyExpectationsNameClientModules(t *testing.T) {
	for module := range conformanceRelyingPartyExpectations {
		assert.Truef(t, strings.HasPrefix(module, "oidcc-client-test"), "'%s' is not a Relying Party module", module)
	}

	assert.True(t, conformanceRelyingPartyExpectations["oidcc-client-test"], "the first module of each plan is the one the identity is linked in, so it must be one Authelia accepts")
}

func TestConformanceRelyingPartyUsers(t *testing.T) {
	suiteURL, err := url.ParseRequestURI(oidcConformanceBaseURL)
	require.NoError(t, err)

	autheliaURL, err := url.ParseRequestURI(oidcConformanceAutheliaURL)
	require.NoError(t, err)

	usernames, emails := map[string]string{}, map[string]string{}

	for _, builder := range conformance.RelyingPartyBuilders("k3j2x9ab", "authelia", suiteURL, autheliaURL) {
		user, ok := conformanceRelyingPartyUsers[builder.ProviderID()]

		require.Truef(t, ok, "the '%s' provider has no account to link its identity to", builder.ProviderID())
		require.NotEmpty(t, user.username)
		require.NotEmpty(t, user.email)

		assert.NotContainsf(t, usernames, user.username, "the '%s' and '%s' providers link the same account, so their plans would contend for its session", usernames[user.username], builder.ProviderID())
		assert.NotContainsf(t, emails, user.email, "the '%s' and '%s' providers link accounts with the same email, so their plans could read each other's one-time codes", emails[user.email], builder.ProviderID())

		usernames[user.username], emails[user.email] = builder.ProviderID(), builder.ProviderID()

		assert.Containsf(t, readTestFile(t, "OIDCConformance/users.yml"), "  "+user.username+":\n", "the account '%s' is not in the suite's user database", user.username)
		assert.Containsf(t, readTestFile(t, "OIDCConformance/users.yml"), "email: "+user.email+"\n", "the email '%s' is not in the suite's user database", user.email)
	}
}

func TestConformanceRunner_RelyingPartyDrivesWaitingModulesWithTheBrowser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/runner":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"m1"}`))
		case "/api/runner/m1/wait-state":
			_, _ = w.Write([]byte(`{"state":"WAITING"}`))
		case "/api/info/m1":
			_, _ = w.Write([]byte(`{"_id":"m1","status":"WAITING","result":"UNKNOWN"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	runner := NewConformanceRelyingPartyRunner(client, nil, "plan1", []ConformancePlanModule{{TestModule: "oidcc-client-test"}}, "conformance-rp-basic")

	require.NotNil(t, runner.relyingParty)
	assert.Equal(t, "harry", runner.relyingParty.user.username)
	assert.False(t, runner.relyingParty.linked)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range runner.Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 1)

	assert.Equal(t, "ClientTest", outcomes[0].Name)
	assert.Equal(t, "m1", outcomes[0].ID)
	assert.EqualError(t, outcomes[0].Err, "the module needs a browser but the runner has none",
		"a waiting Relying Party module is driven by signing in with the provider, not by visiting the suite's urls")
}

func readTestFile(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(name)
	require.NoError(t, err)

	return string(data)
}

func TestConformanceAssertRelyingPartyLegRequiresANotificationForARejection(t *testing.T) {
	for _, module := range []string{"oidcc-client-test-invalid-iss", "oidcc-client-test-kid-absent-multiple-jwks", "oidcc-client-test-some-future-module"} {
		t.Run(module, func(t *testing.T) {
			err := ConformanceAssertRelyingPartyLeg(module, ConformanceRelyingPartyLeg{Location: "https://login.example.com:8080/?external_identity_error=true"})

			assert.EqualError(t, err, "expected Authelia to show an error notification when it rejected the identity asserted by the provider but it did not, redirecting to 'https://login.example.com:8080/?external_identity_error=true'")
		})
	}

	assert.NoError(t, ConformanceAssertRelyingPartyLeg("oidcc-client-test", ConformanceRelyingPartyLeg{Accepted: true}), "an accepted identity is not notified")
}
