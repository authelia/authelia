package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserDetailerActivationWhiteBox(t *testing.T) {
	activation := &UserDetailerActivation{}

	assert.Nil(t, activation.Parent())
	assert.Nil(t, activation.address())
}
