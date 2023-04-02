// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStorageProvider(t *testing.T) {
	assert.Nil(t, getStorageProvider(NewCmdCtx()))
}
