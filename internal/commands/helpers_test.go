package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestGetStorageProvider(t *testing.T) {
	config = &schema.Configuration{}

	assert.Nil(t, getStorageProvider(nil))
}
