package commands

import (
	"bytes"
	"net"
	"net/url"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func TestGetSubjectAndObjectFromFlags(t *testing.T) {
	testCases := []struct {
		name             string
		url              string
		method           string
		setmethod        bool
		username         string
		groups           []string
		ip               string
		expectederr      bool
		expectedmethod   string
		expectedusername string
		expectedgroups   []string
		expectedip       string
	}{
		{
			name:             "ShouldSetAllFields",
			url:              "https://example.com/admin?x=1",
			method:           "POST",
			setmethod:        true,
			username:         "alice",
			groups:           []string{"admin", "dev"},
			ip:               "203.0.113.5",
			expectederr:      false,
			expectedmethod:   "POST",
			expectedusername: "alice",
			expectedgroups:   []string{"admin", "dev"},
			expectedip:       "203.0.113.5",
		},
		{
			name:             "ShouldUseDefaultsWithOnlyURL",
			url:              "https://example.com/",
			method:           "",
			setmethod:        false,
			username:         "",
			groups:           nil,
			ip:               "",
			expectederr:      false,
			expectedmethod:   "GET",
			expectedusername: "",
			expectedgroups:   []string{},
			expectedip:       "",
		},
		{
			name:        "ShouldErrorOnInvalidURL",
			url:         "http://!@#*(!@&#*(!@$&!(*@",
			expectederr: true,
		},
		{
			name:             "ShouldIgnoreInvalidIP",
			url:              "http://example.com/a",
			method:           "DELETE",
			setmethod:        true,
			username:         "bob",
			groups:           []string{"users"},
			ip:               "not-an-ip",
			expectederr:      false,
			expectedmethod:   "DELETE",
			expectedusername: "bob",
			expectedgroups:   []string{"users"},
			expectedip:       "",
		},
		{
			name:             "ShouldAllowEmptyMethod",
			url:              "https://example.org/x",
			method:           "",
			setmethod:        true,
			username:         "",
			groups:           nil,
			ip:               "",
			expectederr:      false,
			expectedmethod:   "",
			expectedusername: "",
			expectedgroups:   []string{},
			expectedip:       "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			flags := cmd.Flags()
			flags.String("url", "", "")
			flags.String("method", "GET", "")
			flags.String("username", "", "")
			flags.StringSlice("groups", nil, "")
			flags.String("ip", "", "")

			require.NoError(t, flags.Set("url", tc.url))

			if tc.setmethod {
				require.NoError(t, flags.Set("method", tc.method))
			}

			require.NoError(t, flags.Set("username", tc.username))

			if tc.groups != nil {
				require.NoError(t, flags.Set("groups", strings.Join(tc.groups, ",")))
			}

			require.NoError(t, flags.Set("ip", tc.ip))

			subject, object, err := getSubjectAndObjectFromFlags(cmd)

			if tc.expectederr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedusername, subject.Username)
			assert.Equal(t, tc.expectedgroups, subject.Groups)

			if tc.expectedip == "" {
				assert.Nil(t, subject.IP)
			} else if assert.NotNil(t, subject.IP) {
				assert.Equal(t, tc.expectedip, subject.IP.String())
			}

			assert.Equal(t, tc.expectedmethod, object.Method)
		})
	}
}

func TestGetSubjectAndObjectFromFlagErrors(t *testing.T) {
	testCases := []struct {
		name     string
		url      bool
		method   bool
		username bool
		groups   bool
		ip       bool
		err      string
	}{
		{
			name:     "ShouldErrorOnMissingURLFlag",
			url:      false,
			method:   true,
			username: true,
			groups:   true,
			ip:       true,
			err:      "flag accessed but not defined: url",
		},
		{
			name:     "ShouldErrorOnMissingMethodFlag",
			url:      true,
			method:   false,
			username: true,
			groups:   true,
			ip:       true,
			err:      "flag accessed but not defined: method",
		},
		{
			name:     "ShouldErrorOnMissingUsernameFlag",
			url:      true,
			method:   true,
			username: false,
			groups:   true,
			ip:       true,
			err:      "flag accessed but not defined: username",
		},
		{
			name:     "ShouldErrorOnMissingGroupsFlag",
			url:      true,
			method:   true,
			username: true,
			groups:   false,
			ip:       true,
			err:      "flag accessed but not defined: groups",
		},
		{
			name:     "ShouldErrorOnMissingIPFlag",
			url:      true,
			method:   true,
			username: true,
			groups:   true,
			ip:       false,
			err:      "flag accessed but not defined: ip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			flags := cmd.Flags()

			if tc.url {
				flags.String("url", "", "")

				require.NoError(t, flags.Set("url", "https://example.com/"))
			}

			if tc.method {
				flags.String("method", "", "")
			}

			if tc.username {
				flags.String("username", "", "")
			}

			if tc.groups {
				flags.StringSlice("groups", nil, "")
			}

			if tc.ip {
				flags.String("ip", "", "")
			}

			subject, object, err := getSubjectAndObjectFromFlags(cmd)

			assert.EqualError(t, err, tc.err)
			assert.Equal(t, authorization.Subject{}, subject)
			assert.Equal(t, authorization.Object{}, object)
		})
	}
}

func TestHitMissMay(t *testing.T) {
	testCases := []struct {
		name     string
		input    []bool
		expected string
	}{
		{name: "ShouldReturnHitForAllTrue", input: []bool{true, true, true}, expected: "hit"},
		{name: "ShouldReturnMissForAllFalse", input: []bool{false, false}, expected: "miss"},
		{name: "ShouldReturnMayForMixed", input: []bool{true, false}, expected: "may"},
		{name: "ShouldReturnHitForSingleTrue", input: []bool{true}, expected: "hit"},
		{name: "ShouldReturnMissForSingleFalse", input: []bool{false}, expected: "miss"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := hitMissMay(tc.input...)
			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestAccessControlCheckWriteOutput(t *testing.T) {
	u, err := url.ParseRequestURI("https://example.com/path?query=1")
	require.NoError(t, err)

	object := authorization.NewObject(u, fasthttp.MethodGet)
	subject := authorization.Subject{
		Username: "alice",
		Groups:   []string{"grp1", "grp2"},
		IP:       net.ParseIP("127.0.0.1"),
	}

	testCases := []struct {
		name           string
		results        []authorization.RuleMatchResult
		defaultPolicy  string
		verbose        bool
		expectContains []string
	}{
		{
			name:          "ShouldApplyDefaultPolicyWhenNoRules",
			results:       nil,
			defaultPolicy: "default",
			verbose:       true,
			expectContains: []string{
				"Performing policy check for request to",
				"The policy 'default' from the default policy will be applied to this request as no rules matched the request.",
			},
		},
		{
			name: "ShouldApplyPolicyWhenMatchedRule",
			results: []authorization.RuleMatchResult{
				{
					Rule:               &authorization.AccessControlRule{Policy: authorization.Bypass},
					MatchDomain:        true,
					MatchResources:     true,
					MatchQuery:         true,
					MatchMethods:       true,
					MatchNetworks:      true,
					MatchSubjects:      true,
					MatchSubjectsExact: true,
					Skipped:            false,
				},
			},
			defaultPolicy: "default",
			verbose:       true,
			expectContains: []string{
				"The policy 'bypass' from rule #1 will be applied to this request.",
			},
		},
		{
			name: "ShouldPreferPotentialWhenBeforeApplied",
			results: []authorization.RuleMatchResult{
				{
					Rule:               &authorization.AccessControlRule{Policy: authorization.Bypass},
					MatchDomain:        true,
					MatchResources:     true,
					MatchQuery:         true,
					MatchMethods:       true,
					MatchNetworks:      true,
					MatchSubjects:      true,
					MatchSubjectsExact: false,
					Skipped:            false,
				},
				{
					Rule:               &authorization.AccessControlRule{Policy: authorization.OneFactor},
					MatchDomain:        true,
					MatchResources:     true,
					MatchQuery:         true,
					MatchMethods:       true,
					MatchNetworks:      true,
					MatchSubjects:      true,
					MatchSubjectsExact: true,
					Skipped:            false,
				},
			},
			defaultPolicy: "default",
			verbose:       true,
			expectContains: []string{
				"will potentially be applied to this request",
				"rule #1",
				"rule #2",
				"bypass",
				"one_factor",
			},
		},
		{
			name: "ShouldBreakOnSkippedWhenNotVerbose",
			results: []authorization.RuleMatchResult{
				{
					Rule:               &authorization.AccessControlRule{Policy: authorization.OneFactor},
					MatchDomain:        true,
					MatchResources:     true,
					MatchQuery:         true,
					MatchMethods:       true,
					MatchNetworks:      true,
					MatchSubjects:      true,
					MatchSubjectsExact: true,
					Skipped:            true,
				},
				{
					Rule:               &authorization.AccessControlRule{Policy: authorization.Bypass},
					MatchDomain:        true,
					MatchResources:     true,
					MatchQuery:         true,
					MatchMethods:       true,
					MatchNetworks:      true,
					MatchSubjects:      true,
					MatchSubjectsExact: true,
					Skipped:            false,
				},
			},
			defaultPolicy: "default",
			verbose:       false,
			expectContains: []string{
				"The policy 'default' from the default policy will be applied to this request as no rules matched the request.",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			accessControlCheckWriteOutput(&buf, object, subject, tc.results, tc.defaultPolicy, tc.verbose)
			out := buf.String()

			for _, s := range tc.expectContains {
				assert.Contains(t, out, s)
			}
		})
	}
}
