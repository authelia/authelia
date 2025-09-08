package commands

import (
	"bytes"
	"runtime/debug"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuildInfoCmd(t *testing.T) {
	cmd := newBuildInfoCmd(&CmdCtx{})

	assert.NotNil(t, cmd)
}

func TestRunBuildInfo(t *testing.T) {
	newFlags := func(verbose bool) *pflag.FlagSet {
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

		flags.Bool("verbose", false, "")

		if verbose {
			err := flags.Set("verbose", "true")

			require.NoError(t, err)
		}

		return flags
	}

	testCases := []struct {
		name     string
		flags    *pflag.FlagSet
		expected string
		err      string
	}{
		{
			"Successful",
			newFlags(false),
			"Last Tag: unknown\nState: untagged dirty\nBranch: master\nCommit: unknown\nBuild Number: 0\nBuild OS: linux\nBuild Arch: amd64\nBuild Compiler: gc\nBuild Date: \nDevelopment: false\nExtra: \n\nGo:\n    Version: go1.25.0\n    Module Path: github.com/authelia/authelia/v4\n    Executable Path: github.com/authelia/authelia/v4/internal/commands.test\n",
			"",
		},
		{
			"SuccessfulVerbose",
			newFlags(true),
			"Last Tag: unknown\nState: untagged dirty\nBranch: master\nCommit: unknown\nBuild Number: 0\nBuild OS: linux\nBuild Arch: amd64\nBuild Compiler: gc\nBuild Date: \nDevelopment: false\nExtra: \n\nGo:\n    Version: go1.25.0\n    Module Path: github.com/authelia/authelia/v4\n    Executable Path: github.com/authelia/authelia/v4/internal/commands.test\n",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			err := runBuildInfo(buf, tc.flags)

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

func TestRunBuildInfoOutput(t *testing.T) {
	testCases := []struct {
		name     string
		verbose  bool
		info     *debug.BuildInfo
		expected []string
		err      string
	}{
		{
			"ShouldHandleNormal",
			false,
			&debug.BuildInfo{
				Main: debug.Module{
					Path: "github.com/authelia/authelia/v4",
				},
			},
			[]string{
				"Last Tag: unknown\nState: untagged dirty\nBranch: master\nCommit: unknown\nBuild Number: 0\nBuild OS: linux\nBuild Arch: amd64\nBuild Compiler: gc\nBuild Date: \nDevelopment: false\nExtra: \n\nGo:\n    Version: \n    Module Path: github.com/authelia/authelia/v4",
			},
			"",
		},
		{
			"ShouldHandleVerbose",
			true,
			&debug.BuildInfo{
				Main: debug.Module{
					Path: "github.com/authelia/authelia/v4",
				},
				Deps: []*debug.Module{
					{
						Path:    "github.com/a/fake/pkg",
						Version: "v1.0.0",
					},
				},
			},
			[]string{
				"Last Tag: unknown\nState: untagged dirty\nBranch: master\nCommit: unknown\nBuild Number: 0\nBuild OS: linux\nBuild Arch: amd64\nBuild Compiler: gc\nBuild Date: \nDevelopment: false\nExtra: \n\nGo:\n    Version: \n    Module Path: github.com/authelia/authelia/v4",
				"Dependencies:\n        github.com/a/fake/pkg@v1.0.0 ()\n",
			},
			"",
		},
		{
			"ShouldHandleError",
			false,
			nil,
			nil,
			"failed to read build info",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)

			err := runBuildInfoOutput(buf, tc.verbose, tc.info)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}

			for _, s := range tc.expected {
				assert.Contains(t, buf.String(), s)
			}
		})
	}
}
