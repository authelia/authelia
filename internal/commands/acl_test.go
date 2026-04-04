package commands

import (
	"bytes"
	"net"
	"net/url"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewACLCommand(t *testing.T) {
	var cmd *cobra.Command

	cmd = newAccessControlCommand(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newAccessControlCheckCommand(&CmdCtx{})
	assert.NotNil(t, cmd)
}

func TestGetSubjectAndObjectFromFlagErrors(t *testing.T) {
	testCases := []struct {
		name     string
		url      bool
		urlValue string
		method   bool
		username bool
		groups   bool
		ip       bool
		subject  authorization.Subject
		object   authorization.Object
		err      string
	}{
		{
			name:     "ShouldErrorOnMissingURLFlag",
			method:   true,
			username: true,
			groups:   true,
			ip:       true,
			err:      "flag accessed but not defined: url",
		},
		{
			name:     "ShouldErrorOnInvalidURLFlag",
			method:   true,
			url:      true,
			urlValue: "http://%@#(*$@()#*&$invalid",
			username: true,
			groups:   true,
			ip:       true,
			err:      "parse \"http://%@#(*$@()#*&$invalid\": invalid character \"#\" in host name",
		},
		{
			name:     "ShouldErrorOnMissingMethodFlag",
			url:      true,
			username: true,
			groups:   true,
			ip:       true,
			err:      "flag accessed but not defined: method",
		},
		{
			name:   "ShouldErrorOnMissingUsernameFlag",
			url:    true,
			method: true,
			groups: true,
			ip:     true,
			err:    "flag accessed but not defined: username",
		},
		{
			name:     "ShouldErrorOnMissingGroupsFlag",
			url:      true,
			method:   true,
			username: true,
			ip:       true,
			err:      "flag accessed but not defined: groups",
		},
		{
			name:     "ShouldErrorOnMissingIPFlag",
			url:      true,
			method:   true,
			username: true,
			groups:   true,
			err:      "flag accessed but not defined: ip",
		},
		{
			name:     "ShouldNotErrorWithAllFlagsSet",
			url:      true,
			method:   true,
			username: true,
			groups:   true,
			ip:       true,
			subject:  authorization.Subject{Username: "john", Groups: []string{"example"}, IP: net.ParseIP("127.0.0.1")},
			object:   authorization.Object{URL: &url.URL{Scheme: "https", Host: "example.com", Path: "/"}, Domain: "example.com", Method: fasthttp.MethodGet, Path: "/"},
			err:      "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cobra.Command{Use: "test"}
			flags := cmd.Flags()

			if tc.url {
				flags.String("url", "", "")

				if tc.urlValue != "" {
					require.NoError(t, flags.Set("url", tc.urlValue))
				} else {
					require.NoError(t, flags.Set("url", "https://example.com/"))
				}
			}

			if tc.method {
				flags.String("method", "", "")

				require.NoError(t, flags.Set("method", fasthttp.MethodGet))
			}

			if tc.username {
				flags.String("username", "", "")

				require.NoError(t, flags.Set("username", "john"))
			}

			if tc.groups {
				flags.StringSlice("groups", nil, "")

				require.NoError(t, flags.Set("groups", "example"))
			}

			if tc.ip {
				flags.String("ip", "", "")

				require.NoError(t, flags.Set("ip", "127.0.0.1"))
			}

			subject, object, err := getSubjectAndObjectFromFlags(cmd)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			assert.Equal(t, tc.subject, subject)
			assert.Equal(t, tc.object, object)
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

func TestAccessControlCheckRunE(t *testing.T) {
	testCases := []struct {
		name           string
		config         func() *schema.Configuration
		flags          map[string]string
		err            string
		expectContains []string
	}{
		{
			"ShouldSucceedBypassRule",
			func() *schema.Configuration {
				return &schema.Configuration{
					AccessControl: schema.AccessControl{
						DefaultPolicy: "deny",
						Rules: []schema.AccessControlRule{
							{
								Domains: schema.AccessControlRuleDomains{"example.com"},
								Policy:  "bypass",
							},
						},
					},
				}
			},
			map[string]string{"url": "https://example.com/", "method": fasthttp.MethodGet, "username": "john", "ip": "10.0.0.1"},
			"",
			[]string{"bypass", "rule #1"},
		},
		{
			"ShouldSucceedDefaultPolicyDenyNoRules",
			func() *schema.Configuration {
				return &schema.Configuration{
					AccessControl: schema.AccessControl{
						DefaultPolicy: "deny",
					},
				}
			},
			map[string]string{"url": "https://example.com/", "method": fasthttp.MethodGet},
			"",
			[]string{"The default policy 'deny' will be applied to ALL requests as no rules are configured."},
		},
		{
			"ShouldSucceedDefaultPolicyNoMatch",
			func() *schema.Configuration {
				return &schema.Configuration{
					AccessControl: schema.AccessControl{
						DefaultPolicy: "two_factor",
						Rules: []schema.AccessControlRule{
							{
								Domains: schema.AccessControlRuleDomains{"other.com"},
								Policy:  "bypass",
							},
						},
					},
				}
			},
			map[string]string{"url": "https://example.com/", "method": fasthttp.MethodGet},
			"",
			[]string{"two_factor", "default policy"},
		},
		{
			"ShouldSucceedWithVerboseOutput",
			func() *schema.Configuration {
				return &schema.Configuration{
					AccessControl: schema.AccessControl{
						DefaultPolicy: "deny",
						Rules: []schema.AccessControlRule{
							{
								Domains: schema.AccessControlRuleDomains{"example.com"},
								Policy:  "one_factor",
							},
						},
					},
				}
			},
			map[string]string{"url": "https://example.com/", "method": fasthttp.MethodPost, "username": "alice", "groups": "admins", "ip": "192.168.1.1", "verbose": "true"},
			"",
			[]string{"one_factor", "rule #1", "alice", "admins", "192.168.1.1"},
		},
		{
			"ShouldSucceedWithMultipleRules",
			func() *schema.Configuration {
				return &schema.Configuration{
					AccessControl: schema.AccessControl{
						DefaultPolicy: "deny",
						Rules: []schema.AccessControlRule{
							{
								Domains: schema.AccessControlRuleDomains{"other.com"},
								Policy:  "bypass",
							},
							{
								Domains: schema.AccessControlRuleDomains{"example.com"},
								Policy:  "two_factor",
							},
						},
					},
				}
			},
			map[string]string{"url": "https://example.com/", "method": fasthttp.MethodGet},
			"",
			[]string{"two_factor", "rule #2"},
		},
		{
			"ShouldErrInvalidURL",
			func() *schema.Configuration {
				return &schema.Configuration{
					AccessControl: schema.AccessControl{
						DefaultPolicy: "deny",
					},
				}
			},
			map[string]string{"url": "://invalid", "method": fasthttp.MethodGet},
			"missing protocol scheme",
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := NewCmdCtx()
			cmdCtx.config = tc.config()
			cmdCtx.cconfig = NewCmdCtxConfig()

			cmd := &cobra.Command{Use: "check-policy"}

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			cmd.Flags().String("url", "", "")
			cmd.Flags().String("method", fasthttp.MethodGet, "")
			cmd.Flags().String("username", "", "")
			cmd.Flags().StringSlice("groups", nil, "")
			cmd.Flags().String("ip", "", "")
			cmd.Flags().Bool("verbose", false, "")

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmdCtx.AccessControlCheckRunE(cmd, nil)

			if tc.err == "" {
				assert.NoError(t, err)

				for _, s := range tc.expectContains {
					assert.Contains(t, buf.String(), s)
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
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
				"The default policy 'default' will be applied to ALL requests as no rules are configured.",
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
		{
			name: "ShouldHandleMaybeMatch",
			results: []authorization.RuleMatchResult{
				{
					Rule:               &authorization.AccessControlRule{Policy: authorization.OneFactor},
					MatchDomain:        true,
					MatchResources:     true,
					MatchQuery:         true,
					MatchMethods:       true,
					MatchNetworks:      true,
					MatchSubjects:      false,
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
					MatchSubjectsExact: false,
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
					MatchSubjectsExact: false,
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

			err := runAccessControlCheck(&buf, object, subject, tc.results, tc.defaultPolicy, tc.verbose)
			assert.NoError(t, err)

			out := buf.String()

			for _, s := range tc.expectContains {
				assert.Contains(t, out, s)
			}
		})
	}
}
