package storage

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsentPreConfigRows_NilRows(t *testing.T) {
	testCases := []struct {
		name string
		op   string
	}{
		{
			name: "ShouldReturnFalseOnNextWhenRowsNil",
			op:   "next",
		},
		{
			name: "ShouldReturnNilOnCloseWhenRowsNil",
			op:   "close",
		},
		{
			name: "ShouldReturnErrNoRowsOnGetWhenRowsNil",
			op:   "get",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &ConsentPreConfigRows{rows: nil}

			switch tc.op {
			case "next":
				assert.False(t, r.Next())
			case "close":
				assert.NoError(t, r.Close())
			case "get":
				cfg, err := r.Get()
				require.Error(t, err)
				assert.ErrorIs(t, err, sql.ErrNoRows)
				assert.Nil(t, cfg)
			default:
				t.Fatalf("unknown op: %s", tc.op)
			}
		})
	}
}
