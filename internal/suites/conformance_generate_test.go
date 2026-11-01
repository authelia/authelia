// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCConformanceGenerateReportsAFileItCannotWrite(t *testing.T) {
	t.Setenv("SUITE_TMP_PATH", filepath.Join(t.TempDir(), "missing"))

	assert.Error(t, oidcConformanceGenerate())
	assert.Error(t, oidcConformanceWriteJSON(SuiteTmpPath(oidcConformancePlansFile), []OIDCConformancePlanFile{}))
}

func TestOIDCConformanceReadPlans(t *testing.T) {
	t.Run("ShouldReportPlansWhichWereNeverGenerated", func(t *testing.T) {
		t.Setenv("SUITE_TMP_PATH", t.TempDir())

		_, err := oidcConformanceReadPlans()

		assert.ErrorContains(t, err, "error reading the generated conformance plans, was the suite set up?")
	})

	t.Run("ShouldReportPlansWhichCannotBeDecoded", func(t *testing.T) {
		dir := t.TempDir()

		t.Setenv("SUITE_TMP_PATH", dir)

		require.NoError(t, os.WriteFile(filepath.Join(dir, oidcConformancePlansFile), []byte("{"), 0600))

		_, err := oidcConformanceReadPlans()

		assert.Error(t, err)
	})
}
