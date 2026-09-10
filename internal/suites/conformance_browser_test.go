// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConformanceClassifyPage(t *testing.T) {
	testCases := []struct {
		name                                string
		url                                 string
		firstFactor, consent, autheliaError bool
		expected                            ConformancePageState
	}{
		{
			"ShouldDetectTheCallbackByPath",
			"https://conformance.example.com:8443/test/a/conformance-basic-authelia1a2b3c4d/callback?code=abc",
			false, false, false,
			ConformancePageCallback,
		},
		{
			"ShouldPreferTheCallbackOverAnyStageStillInTheDOM",
			"https://conformance.example.com:8443/test/a/alias/callback",
			true, false, false,
			ConformancePageCallback,
		},
		{
			"ShouldDetectTheFirstFactorPage",
			"https://login.example.com:8080/",
			true, false, false,
			ConformancePageFirstFactor,
		},
		{
			"ShouldDetectTheConsentPage",
			"https://login.example.com:8080/consent/openid/decision",
			false, true, false,
			ConformancePageConsent,
		},
		{
			"ShouldDetectAnAutheliaError",
			"https://login.example.com:8080/",
			false, false, true,
			ConformancePageAutheliaError,
		},
		{
			// The error toast is a portalled overlay which co-exists with whatever stage is on screen, so it must not
			// end the leg on a page the driver could have worked.
			"ShouldPreferTheFirstFactorPageOverACoexistingErrorToast",
			"https://login.example.com:8080/",
			true, false, true,
			ConformancePageFirstFactor,
		},
		{
			"ShouldPreferTheConsentPageOverACoexistingErrorToast",
			"https://login.example.com:8080/consent/openid/decision",
			false, true, true,
			ConformancePageConsent,
		},
		{
			"ShouldReportUnknownWhenNothingMatches",
			"https://login.example.com:8080/",
			false, false, false,
			ConformancePageUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConformanceClassifyPage(tc.url, tc.firstFactor, tc.consent, tc.autheliaError))
		})
	}
}

func TestConformanceConsentSettleSelector(t *testing.T) {
	t.Run("ShouldWatchThePasswordFieldAfterSubmittingReauthentication", func(t *testing.T) {
		// The stage and the URL are identical either side of this submission, so watching the stage would wait out
		// the whole budget on a form that had already moved on. This is what oidcc-prompt-login turns on.
		assert.Equal(t, conformanceSelectorConsentReauthentication, conformanceConsentSettleSelector(true))
	})

	t.Run("ShouldWatchTheStageAfterSubmittingTheDecision", func(t *testing.T) {
		// Granting consent leaves the form altogether, so the stage clearing is the signal.
		assert.Equal(t, conformanceSelectorConsent, conformanceConsentSettleSelector(false))
	})

	t.Run("ShouldWatchDifferentThingsForTheTwoSteps", func(t *testing.T) {
		assert.NotEqual(t, conformanceConsentSettleSelector(true), conformanceConsentSettleSelector(false))
	})
}
