package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOneTimeCode(t *testing.T) {
	testCases := []struct {
		name       string
		username   string
		characters int
		duration   time.Duration
		expected   *OneTimeCode
		err        string
	}{
		{
			"Success",
			"username",
			1,
			time.Hour,
			&OneTimeCode{},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewOneTimeCode(nil, tc.username, tc.characters, tc.duration)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
