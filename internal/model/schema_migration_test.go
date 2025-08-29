package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaMigration(t *testing.T) {
	m := &SchemaMigration{}

	assert.False(t, m.NotEmpty())
	assert.Equal(t, 0, m.Before())
	assert.Equal(t, -1, m.After())

	m.Up = true

	assert.Equal(t, -1, m.Before())
	assert.Equal(t, 0, m.After())

	m.Query = "abc"

	assert.True(t, m.NotEmpty())

	m.Version = 5

	assert.Equal(t, 4, m.Before())
	assert.Equal(t, 5, m.After())
}
