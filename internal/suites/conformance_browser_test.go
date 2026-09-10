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
			// The authorization request carries the callback in its redirect_uri parameter, so a check against the
			// whole URL would end the leg before Authelia had presented anything.
			"ShouldNotMistakeAnAuthorizationRequestCarryingTheCallbackForTheCallback",
			"https://login.example.com:8080/api/oidc/authorization?client_id=abc&redirect_uri=https://conformance.example.com:8443/test/a/alias/callback",
			false, false, false,
			ConformancePageUnknown,
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

func TestConformanceConsentProgress(t *testing.T) {
	t.Run("ShouldTreatTheRevealedPasswordFieldAsTheReauthenticationStep", func(t *testing.T) {
		// Accepting switches the step in place: the stage and the URL are unchanged, and only the password field's
		// appearance says so. Watching for the form to go away instead is what made oidcc-prompt-login wait out its
		// whole budget on a form that had done exactly what was asked.
		reauthentication, done := ConformanceConsentProgress(true, true, false)
		assert.True(t, reauthentication)
		assert.True(t, done)
	})

	t.Run("ShouldTreatTheStageClearingAsTheFlowLeaving", func(t *testing.T) {
		reauthentication, done := ConformanceConsentProgress(false, false, false)
		assert.False(t, reauthentication)
		assert.True(t, done)
	})

	t.Run("ShouldTreatAChangedURLAsTheFlowLeaving", func(t *testing.T) {
		reauthentication, done := ConformanceConsentProgress(false, true, true)
		assert.False(t, reauthentication)
		assert.True(t, done)
	})

	t.Run("ShouldKeepWaitingWhileTheFormIsStillOnItsDecisionStep", func(t *testing.T) {
		reauthentication, done := ConformanceConsentProgress(false, true, false)
		assert.False(t, reauthentication)
		assert.False(t, done)
	})

	t.Run("ShouldPreferTheReauthenticationStepOverTheFormLeaving", func(t *testing.T) {
		// A password field seen alongside a changed URL still means the step was revealed; answering it is what makes
		// the flow leave, so reporting the departure here would skip the answer.
		reauthentication, done := ConformanceConsentProgress(true, false, true)
		assert.True(t, reauthentication)
		assert.True(t, done)
	})
}
