// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteCoverage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".nyc_output")

	require.NoError(t, writeCoverage(dir, `{"/node/src/app/src/index.tsx":{"path":"/node/src/app/src/index.tsx"}}`))
	require.NoError(t, writeCoverage(dir, `{}`))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 2, "each document's coverage is written to a file of its own, so the report merges them rather than one replacing another")

	wd, err := os.Getwd()
	require.NoError(t, err)

	web := strings.TrimSuffix(wd, "internal/suites") + "web/"

	var found bool

	for _, entry := range entries {
		assert.True(t, strings.HasPrefix(entry.Name(), "coverage-") && strings.HasSuffix(entry.Name(), ".json"), entry.Name())

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		require.NoError(t, err)

		if strings.Contains(string(data), "index.tsx") {
			found = true

			assert.NotContains(t, string(data), "/node/src/app/", "the paths the frontend was built at must be rewritten to this checkout")
			assert.Contains(t, string(data), web+"src/index.tsx")
		}
	}

	assert.True(t, found)
}

func TestConformanceCoverageScript(t *testing.T) {
	assert.Contains(t, conformanceCoverageScript, "'beforeunload'")
	assert.Contains(t, conformanceCoverageScript, "window."+conformanceCoverageBinding+"(")
	assert.Contains(t, conformanceCoverageScript, "window.__coverage__")
}
