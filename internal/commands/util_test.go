package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadXNormalizedPaths(t *testing.T) {
	testCases := []struct {
		name, expectErr                string
		haveConfigs, expectConfigs     []string
		haveDirectory, expectDirectory string
	}{
		{
			name:        "ShouldNormalizeDirectory",
			expectErr:   "",
			haveConfigs: []string{"/abc/123.yml"}, expectConfigs: []string{"/abc/123.yml"},
			haveDirectory: "/etc/authelia/", expectDirectory: "/etc/authelia",
		},
		{
			name:        "ShouldErrOnConfigInDirectory",
			expectErr:   "failed to load config directory '/etc/authelia': the file '/etc/authelia/123.yml' is in that directory which is not supported",
			haveConfigs: []string{"/etc/authelia/123.yml"}, expectConfigs: nil,
			haveDirectory: "/etc/authelia/", expectDirectory: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualConfigs, actualDirectory, actualErr := loadXNormalizedPaths(tc.haveConfigs, tc.haveDirectory)

			assert.Equal(t, tc.expectConfigs, actualConfigs)
			assert.Equal(t, tc.expectDirectory, actualDirectory)

			if tc.expectErr != "" {
				assert.EqualError(t, actualErr, tc.expectErr)
			} else {
				assert.Nil(t, actualErr)
			}
		})
	}
}
