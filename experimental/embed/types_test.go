package embed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeBasics(t *testing.T) {
	c := &Configuration{}
	assert.NotNil(t, c.ToInternal())

	p := &Providers{}
	assert.NotNil(t, p.ToInternal())
}
