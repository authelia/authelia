package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserOpaqueIdentifier(t *testing.T) {
	testCases := []struct {
		name                        string
		service, sectorID, username string
	}{
		{
			"ShouldHandleEmptyStrings",
			"",
			"",
			"",
		},
		{
			"ShouldHandleNonEmptyStrings",
			"abc",
			"123",
			"ggg",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := NewUserOpaqueIdentifier(tc.service, tc.sectorID, tc.username)
			require.NoError(t, err)

			assert.Equal(t, tc.service, id.Service)
			assert.Equal(t, tc.username, id.Username)
			assert.Equal(t, tc.sectorID, id.SectorID)

			assert.NotNil(t, id.Identifier)
		})
	}
}
