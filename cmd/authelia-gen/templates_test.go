// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldFailToLoadBadTemplate(t *testing.T) {
	assert.Panics(t, func() {
		mustLoadTmplFS("bad tmpl")
	})
}
