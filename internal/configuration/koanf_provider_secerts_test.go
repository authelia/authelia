package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnErrOnReadBytes(t *testing.T) {
	p := NewSecretsProvider(".", NewProvider())

	_, err := p.ReadBytes()

	assert.EqualError(t, err, "provider does not support this method")
}
