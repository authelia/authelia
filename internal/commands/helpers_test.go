package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStorageProvider(t *testing.T) {
	assert.Nil(t, getStorageProvider(NewCmdCtx()))
}
