package commands

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunBuildInfo(t *testing.T) {
	testCases := []struct {
		name    string
		verbose bool
		err     string
	}{
		{
			"Successful",
			false,
			"",
		},
		{
			"SuccessfulVerbose",
			true,
			"",
		},
	}

	r := regexp.MustCompile(`^Last Tag: (v\d+\.\d+\.\d+|unknown)\nState: (tagged|untagged) (clean|dirty)\nBranch: [^\s\n]+\nCommit: ([0-9a-f]{40}|unknown)\nBuild Number: \d+\nBuild OS: (linux|darwin|windows|freebsd)\nBuild Arch: (amd64|arm|arm64)\nBuild Compiler: gc\nBuild Date: \nDevelopment: (true|false)\nExtra: \n\nGo:\n\s+Version: go\d+\.\d+\.\d+\n\s+Module Path: github.com/authelia/authelia/v4\n\s+Executable Path: github.com/authelia/authelia/v4/internal/commands.test`)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := runBuildInfo(buf, tc.verbose)

			assert.Regexp(t, r, buf.String())

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}

			buf.Reset()
		})
	}
}
