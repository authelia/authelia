// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageCleanBefore(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		have     string
		expected time.Time
		err      string
	}{
		{
			"ShouldUseNowWhenTheFlagIsAbsent",
			"",
			now,
			"",
		},
		{
			"ShouldSubtractADurationInTheCommonSyntax",
			"30 days",
			now.Add(-time.Hour * 24 * 30),
			"",
		},
		{
			"ShouldSubtractADurationInTheGoSyntax",
			"1h",
			now.Add(-time.Hour),
			"",
		},
		{
			"ShouldRaiseErrorOnAnUnparsableDuration",
			"not a duration",
			time.Time{},
			"error parsing the value of the --before flag: could not parse 'not a duration' as a duration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String(cmdFlagNameBefore, "", "")

			if tc.have != "" {
				require.NoError(t, cmd.Flags().Set(cmdFlagNameBefore, tc.have))
			}

			before, err := storageCleanBefore(cmd, now)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, before)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
