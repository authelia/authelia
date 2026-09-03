// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsURLPath(t *testing.T) {
	callback := fmt.Sprintf(externalIdentityCallbackPathFmt, "conformance-rp-basic")

	assert.True(t, isURLPath("https://login.example.com:8080"+callback+"?code=abc&state=def", callback))
	assert.False(t, isURLPath("https://login.example.com:8080"+callback+"-other", callback))
	assert.False(t, isURLPath("https://conformance.example.com:8443/test/a/alias/authorize", callback))
	assert.False(t, isURLPath("%zz", callback))
}
