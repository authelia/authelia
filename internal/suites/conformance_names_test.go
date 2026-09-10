// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConformanceSubtestNames(t *testing.T) {
	testCases := []struct {
		name     string
		have     []ConformancePlanModule
		expected []string
	}{
		{
			"ShouldStripPrefixAndPascalCase",
			[]ConformancePlanModule{{TestModule: "oidcc-server"}},
			[]string{"Server"},
		},
		{
			"ShouldPascalCaseEverySegment",
			[]ConformancePlanModule{{TestModule: "oidcc-ensure-request-without-nonce-succeeds-for-code-flow"}},
			[]string{"EnsureRequestWithoutNonceSucceedsForCodeFlow"},
		},
		{
			"ShouldKeepDigits",
			[]ConformancePlanModule{{TestModule: "oidcc-max-age-10000"}},
			[]string{"MaxAge10000"},
		},
		{
			"ShouldHandleModulesWithoutThePrefix",
			[]ConformancePlanModule{{TestModule: "oidcc-config-certification"}, {TestModule: "discovery-issuer-not-matching-config"}},
			[]string{"ConfigCertification", "DiscoveryIssuerNotMatchingConfig"},
		},
		{
			"ShouldNotSuffixUniqueModules",
			[]ConformancePlanModule{
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code"}},
				{TestModule: "oidcc-scope-address", Variant: map[string]string{"response_type": "code"}},
			},
			[]string{"Server", "ScopeAddress"},
		},
		{
			"ShouldSuffixRepeatedModulesWithTheirVariant",
			[]ConformancePlanModule{
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code id_token"}},
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code id_token token"}},
			},
			[]string{"ServerCodeIdToken", "ServerCodeIdTokenToken"},
		},
		{
			"ShouldOrderVariantKeysDeterministically",
			[]ConformancePlanModule{
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code", "response_mode": "form_post"}},
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code", "response_mode": "query"}},
			},
			[]string{"ServerFormPostCode", "ServerQueryCode"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConformanceSubtestNames(tc.have))
		})
	}
}

func TestConformanceResultAccepted(t *testing.T) {
	testCases := []struct {
		name     string
		module   string
		result   string
		expected bool
	}{
		{"ShouldAcceptPassedFromAnyModule", "oidcc-server", "PASSED", true},
		{"ShouldAcceptReviewFromAnyModule", "oidcc-server", "REVIEW", true},
		{"ShouldAcceptSkippedFromAnyModule", "oidcc-server", "SKIPPED", true},
		{"ShouldAcceptAWarningFromAListedModule", "oidcc-ensure-request-with-acr-values-succeeds", "WARNING", true},
		{"ShouldRejectAWarningFromAnUnlistedModule", "oidcc-scope-profile", "WARNING", false},
		{"ShouldRejectAFailureFromAListedModule", "oidcc-ensure-request-with-acr-values-succeeds", "FAILED", false},
		{"ShouldRejectAFailure", "oidcc-server", "FAILED", false},
		{"ShouldRejectAnUnknownResult", "oidcc-server", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, conformanceResultAccepted(tc.module, tc.result))
		})
	}

	for module, reason := range conformanceAcceptedWarnings {
		assert.NotEmptyf(t, reason, "the WARNING accepted from '%s' must say why it is expected", module)
	}
}

func TestConformanceArchivePath(t *testing.T) {
	t.Run("ShouldWriteWhereTheCIArtifactPathCollectsIt", func(t *testing.T) {
		t.Setenv("CI", "true")

		assert.Equal(t, filepath.Join("..", "..", "oidc-conformance-plans", "conformance-basic.zip"), conformanceArchivePath("basic"))
	})

	t.Run("ShouldWriteToTheTemporaryDirectoryOutsideCI", func(t *testing.T) {
		t.Setenv("CI", "")

		assert.Equal(t, filepath.Join(os.TempDir(), "authelia-suites-oidc-conformance-plans", composeProjectName(), "conformance-basic.zip"), conformanceArchivePath("basic"))
	})
}

func TestBuildkiteLogMarkers(t *testing.T) {
	t.Run("ShouldMarkGroupsInBuildkiteWhenDebugging", func(t *testing.T) {
		t.Setenv("BUILDKITE", "true")
		t.Setenv("SUITE_DEBUG", "true")

		assert.Equal(t, "--- OIDC Conformance Plan: basic", buildkiteGroup("OIDC Conformance Plan: basic"))
		assert.Equal(t, "^^^ +++", buildkiteExpandGroup())
	})

	t.Run("ShouldMarkNothingInBuildkiteWhenNotDebugging", func(t *testing.T) {
		t.Setenv("BUILDKITE", "true")
		t.Setenv("SUITE_DEBUG", "")

		assert.Empty(t, buildkiteGroup("OIDC Conformance Plan: basic"))
		assert.Empty(t, buildkiteExpandGroup())
	})

	t.Run("ShouldMarkNothingOutsideBuildkite", func(t *testing.T) {
		t.Setenv("BUILDKITE", "")
		t.Setenv("SUITE_DEBUG", "true")

		assert.Empty(t, buildkiteGroup("OIDC Conformance Plan: basic"))
		assert.Empty(t, buildkiteExpandGroup())
	})
}

func TestConformancePlanSummary(t *testing.T) {
	assert.Equal(t, "Plan 'basic': 45 accepted, 2 failed, 1 skipped", conformancePlanSummary("basic", 45, 2, 1))
}

func TestConformanceArchiveReference(t *testing.T) {
	t.Run("ShouldLinkTheJobsArtifactsInBuildkite", func(t *testing.T) {
		t.Setenv("CI", "true")
		t.Setenv("BUILDKITE_BUILD_URL", "https://buildkite.com/authelia/authelia/builds/58611")
		t.Setenv("BUILDKITE_JOB_ID", "0192")

		assert.Equal(t, "oidc-conformance-plans/conformance-basic.zip in the job's artifacts: https://buildkite.com/authelia/authelia/builds/58611/waterfall?jid=0192&tab=artifacts",
			conformanceArchiveReference("basic", "https://conformance.example.com:8443/log-detail.html?log=m1"))
	})

	t.Run("ShouldPointAtTheLiveLogElsewhere", func(t *testing.T) {
		t.Setenv("CI", "")
		t.Setenv("BUILDKITE_BUILD_URL", "")
		t.Setenv("BUILDKITE_JOB_ID", "")

		assert.Equal(t, conformanceArchivePath("basic")+", or the live log while the suite is up: https://conformance.example.com:8443/log-detail.html?log=m1",
			conformanceArchiveReference("basic", "https://conformance.example.com:8443/log-detail.html?log=m1"))
	})
}

func TestOIDCConformanceRelease(t *testing.T) {
	assert.Equal(t, "v4.39.24 (commit 7498f635a5f3)", oidcConformanceRelease("7498f635a5f3", "v4.39.24"))
	assert.Equal(t, "commit 7498f635a5f3", oidcConformanceRelease("7498f635a5f3", ""))
	assert.Empty(t, oidcConformanceRelease("", ""), "without a commit the builder's own version is left in place")
}
