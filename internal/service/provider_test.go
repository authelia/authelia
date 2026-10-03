// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProvisioners(t *testing.T) {
	provisioners := GetProvisioners()

	assert.Len(t, provisioners, 6)
}
