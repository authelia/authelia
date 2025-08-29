package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunBuildInfo(t *testing.T) {
	testCases := []struct {
		name     string
		verbose  bool
		expected string
		err      string
	}{
		{
			"Successful",
			false,
			"Last Tag: unknown\nState: untagged dirty\nBranch: master\nCommit: unknown\nBuild Number: 0\nBuild OS: linux\nBuild Arch: amd64\nBuild Compiler: gc\nBuild Date: \nDevelopment: false\nExtra: \n\nGo:\n    Version: go1.25.0\n    Module Path: github.com/authelia/authelia/v4\n    Executable Path: github.com/authelia/authelia/v4/internal/commands.test\n",
			"",
		},
		{
			"SuccessfulVerbose",
			true,
			"Last Tag: unknown\nState: untagged dirty\nBranch: master\nCommit: unknown\nBuild Number: 0\nBuild OS: linux\nBuild Arch: amd64\nBuild Compiler: gc\nBuild Date: \nDevelopment: false\nExtra: \n\nGo:\n    Version: go1.25.0\n    Module Path: github.com/authelia/authelia/v4\n    Executable Path: github.com/authelia/authelia/v4/internal/commands.test\n",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			err := runBuildInfo(buf, tc.verbose)

			assert.Contains(t, buf.String(), tc.expected)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}

			buf.Reset()
		})
	}
}
