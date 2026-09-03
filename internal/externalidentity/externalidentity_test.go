// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeProviderErrorValue(t *testing.T) {
	testCases := []struct {
		Name     string
		Value    string
		Expected string
	}{
		{Name: "ShouldReturnEmptyValue", Value: "", Expected: ""},
		{Name: "ShouldReturnPlainValue", Value: "The resource owner denied the request.", Expected: "The resource owner denied the request."},
		{Name: "ShouldRemoveControlCharacters", Value: "access_denied\n2026-09-13 level=info msg=\"forged\"\r\x00\x1b[31m", Expected: "access_denied2026-09-13 level=info msg=\"forged\"[31m"},
		{Name: "ShouldPreserveMultibyteCharacters", Value: "l'accès est refusé", Expected: "l'accès est refusé"},
		{Name: "ShouldNotTruncateValueAtTheLimit", Value: strings.Repeat("é", providerErrorValueLimit), Expected: strings.Repeat("é", providerErrorValueLimit)},
		{Name: "ShouldTruncateValueOverTheLimit", Value: strings.Repeat("é", providerErrorValueLimit+1), Expected: strings.Repeat("é", providerErrorValueLimit) + "..."},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, SanitizeProviderErrorValue(tc.Value))
		})
	}
}
