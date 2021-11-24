package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldHaveSameChecksumForBothTemplates(t *testing.T) {
	sumRoot, err := utils.HashSHA256FromPath("../../configuration.template.yml")
	assert.NoError(t, err)

	sumInternal, err := utils.HashSHA256FromPath("./configuration.template.yml")
	assert.NoError(t, err)

	assert.Equal(t, sumRoot, sumInternal, "Ensure both ./configuration.template.yml and ./internal/configuration/configuration.template.yml are exactly the same.")
}
