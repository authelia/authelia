// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		reauthentication, done := ConformanceConsentProgress(true, false, true)
		assert.True(t, reauthentication)
		assert.True(t, done)
	})
}

func TestConformanceScreenshotDataURI(t *testing.T) {
	small, large := []byte("small"), make([]byte, conformanceScreenshotLimit+1)

	t.Run("ShouldPreferAPNGWhichFits", func(t *testing.T) {
		uri, err := conformanceScreenshotDataURI(func(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error) {
			assert.Equal(t, proto.PageCaptureScreenshotFormatPng, format)

			return small, nil
		})

		require.NoError(t, err)
		assert.Equal(t, "data:image/png;base64,"+base64.StdEncoding.EncodeToString(small), uri)
	})

	t.Run("ShouldFallBackToAJPEGWhenThePNGIsTooLarge", func(t *testing.T) {
		var asked []string

		uri, err := conformanceScreenshotDataURI(func(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error) {
			asked = append(asked, fmt.Sprintf("%s:%d", format, quality))

			if format == proto.PageCaptureScreenshotFormatPng {
				return large, nil
			}

			return small, nil
		})

		require.NoError(t, err)
		assert.Equal(t, "data:image/jpeg;base64,"+base64.StdEncoding.EncodeToString(small), uri)
		assert.Equal(t, []string{"png:0", "jpeg:80"}, asked)
	})

	t.Run("ShouldFailWhenNothingFits", func(t *testing.T) {
		_, err := conformanceScreenshotDataURI(func(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error) {
			return large, nil
		})

		assert.ErrorContains(t, err, "500KB")
	})

	t.Run("ShouldFailWhenTheCaptureFails", func(t *testing.T) {
		_, err := conformanceScreenshotDataURI(func(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error) {
			return nil, errors.New("target closed")
		})

		assert.ErrorContains(t, err, "target closed")
	})
}

func TestConformanceIsCallbackRejectsAnUnparsableURL(t *testing.T) {
	assert.False(t, conformanceIsCallback("%zz/test/a/alias/callback"))
}
