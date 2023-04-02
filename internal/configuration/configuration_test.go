// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldHaveSameChecksumForBothTemplates(t *testing.T) {
	sumRoot, err := utils.HashSHA256FromPath("../../config.template.yml")
	assert.NoError(t, err)

	sumInternal, err := utils.HashSHA256FromPath("./config.template.yml")
	assert.NoError(t, err)

	assert.Equal(t, sumRoot, sumInternal, "Ensure both ./config.template.yml and ./internal/configuration/config.template.yml are exactly the same.")
}
