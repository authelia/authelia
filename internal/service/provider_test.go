package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProvisioners(t *testing.T) {
	provisioners := GetProvisioners()

	assert.Len(t, provisioners, 5)
}
