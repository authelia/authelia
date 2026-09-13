// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserSessionOpenIDConnectStateDefaults(t *testing.T) {
	testCases := []struct {
		Name string
		Have UserSession
	}{
		{
			Name: "ShouldNotBeSetOnNewSession",
			Have: NewDefaultUserSession(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Nil(t, tc.Have.ExternalIdentity)
			assert.Nil(t, tc.Have.ExternalIdentityPending)
		})
	}
}
