package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
)

func (rs *RodSession) verifyIsOIDC(t *testing.T, page *rod.Page, pattern, url string) {
	page.MustElementR("body", pattern)
	rs.verifyURLIs(t, page, url)
}

func (rs *RodSession) verifyIsOIDCErrorPage(t *testing.T, page *rod.Page, errorCode, errorDescription, errorURI, state string) {
	testCases := []struct {
		ElementID, ElementText string
	}{
		{"error", errorCode},
		{"error_description", errorDescription},
		{"error_uri", errorURI},
		{"state", state},
	}

	for _, tc := range testCases {
		t.Run(tc.ElementID, func(t *testing.T) {
			if tc.ElementText == "" {
				t.Skip("Test Skipped as the element is not expected.")
			}

			text, err := rs.WaitElementLocatedByID(t, page, tc.ElementID).Text()
			assert.NoError(t, err)
			assert.Equal(t, tc.ElementText, text)
		})
	}
}
