package model

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRandomNullUUID(t *testing.T) {
	have, err := NewRandomNullUUID()
	require.NoError(t, err)
	assert.NotEmpty(t, have)
}

func TestMustNullUUID(t *testing.T) {
	assert.NotPanics(t, func() {
		have := MustNullUUID(NewRandomNullUUID())

		require.NotNil(t, have)
	})

	assert.Panics(t, func() {
		_ = MustNullUUID(uuid.NullUUID{}, fmt.Errorf("bad"))
	})
}

func TestParseNullUUID(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected uuid.NullUUID
		err      string
	}{
		{
			"ShouldParseEmptyString",
			"",
			uuid.NullUUID{},
			"",
		},
		{
			"ShouldNotParseX",
			"x",
			uuid.NullUUID{},
			"invalid UUID length: 1",
		},
		{
			"ShouldParseRealUUIDv4",
			"32668311-6935-4c5e-99d1-507f0040de06",
			uuid.NullUUID{Valid: true, UUID: uuid.Must(uuid.Parse("32668311-6935-4c5e-99d1-507f0040de06"))},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseNullUUID(tc.have)

			assert.Equal(t, tc.expected, actual)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
