// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCmdName(t *testing.T) {
	testCases := []struct {
		name     string
		have     *cobra.Command
		expected string
	}{
		{
			"ShouldResolveRootCmd",
			newRootCmd(),
			"authelia-gen",
		},
		{
			"ShouldResolveDocsCmd",
			newDocsCmd(),
			"docs",
		},
		{
			"ShouldResolveDocsSubCmd",
			newRootCmd().Commands()[0].Commands()[0],
			"code.keys",
		},
		{
			"ShouldResolveMiscLocaleMoveCmdWithoutArgumentHint",
			newMiscCmd().Commands()[1],
			"misc.locale-move",
		},
		{
			"ShouldResolveMiscReleaseCmdWithoutArgumentSpec",
			newMiscCmd().Commands()[3],
			"misc.release",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, resolveCmdName(tc.have))
		})
	}
}

func TestRootCmdGetArgs(t *testing.T) {
	testCases := []struct {
		name     string
		have     func() *cobra.Command
		args     []string
		expected []string
	}{
		{
			"ShouldReturnRootCmdArgs",
			func() *cobra.Command {
				cmd := newRootCmd()

				cmd.SetArgs([]string{"a", "b"})

				return cmd.Commands()[0]
			},
			[]string{"c", "d"},
			[]string{"authelia-gen", "code", "c", "d"},
		},
		{
			"ShouldReturnRootCmdWithoutArgs",
			func() *cobra.Command {
				cmd := newRootCmd()

				return cmd.Commands()[0]
			},
			nil,
			[]string{"authelia-gen", "code"},
		},
		{
			"ShouldReturnRootCmdArgsWithoutArgumentSpec",
			func() *cobra.Command {
				for _, subCmd := range newRootCmd().Commands() {
					if subCmd.Name() == cmdUseMisc {
						return subCmd.Commands()[3]
					}
				}

				return nil
			},
			nil,
			[]string{"authelia-gen", "misc", "release"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, rootCmdGetArgs(tc.have(), tc.args))
		})
	}
}

func TestSortCmds(t *testing.T) {
	testCases := []struct {
		name     string
		have     *cobra.Command
		expected []string
	}{
		{
			"ShouldSortRootCmd",
			newRootCmd(),
			[]string{"code", "commit-lint", "github", "locales", "misc", "docs"},
		},
		{
			"ShouldSortDocsCmd",
			newDocsCmd(),
			[]string{"cli", "data", pathJSONSchema, cmdUseManage, "seo", "date"},
		},
		{
			"ShouldSortDocsSEOCmd",
			newDocsSEOCmd(),
			[]string{"openid-connect"},
		},
		{
			"ShouldSortGitHubCmd",
			newGitHubCmd(),
			[]string{"issue-templates"},
		},
		{
			"ShouldSortLocalesCmd",
			newLocalesCmd(),
			nil,
		},
		{
			"ShouldSortDocsDataCmd",
			newDocsDataCmd(),
			[]string{"keys", "misc"},
		},
		{
			"ShouldSortMiscCmd",
			newMiscCmd(),
			[]string{"contributors", "locale-move", "oidc", "release"},
		},
		{
			"ShouldSortMiscOIDCCmd",
			newMiscOIDCCmd(),
			[]string{"conformance"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := sortCmds(tc.have)

			n := len(tc.expected)

			require.Len(t, actual, n)

			for i := 0; i < n; i++ {
				assert.Equal(t, tc.expected[i], actual[i].Name())
			}
		})
	}
}
