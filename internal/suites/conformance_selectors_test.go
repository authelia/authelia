// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceSelectorsExistInThePortal(t *testing.T) {
	portal := conformancePortalSource(t)

	for _, tc := range []struct {
		name     string
		selector string
	}{
		{"FirstFactorStage", conformanceSelectorFirstFactor},
		{"ConsentStage", conformanceSelectorConsent},
		{"ConsentAccept", conformanceSelectorConsentAccept},
		{"ConsentAuthenticate", conformanceSelectorConsentAuthenticate},
		{"ConsentReauthenticationField", conformanceSelectorConsentReauthentication},
		{"ConsentReauthenticationPassword", conformanceSelectorConsentPassword},
		{"Username", conformanceSelectorUsername},
		{"Password", conformanceSelectorPassword},
		{"SignIn", conformanceSelectorSignIn},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, id := range conformanceSelectorIdentifiers(tc.selector) {
				assert.Containsf(t, portal, id, "the portal no longer renders anything with the id %q, which %q depends on", id, tc.selector)
			}
		})
	}

	t.Run("AutheliaErrorToast", func(t *testing.T) {
		assert.Contains(t, portal, "notification", "the toast no longer carries the notification class")
		assert.Contains(t, portal, "data-[type=error]", "the toast no longer distinguishes an error by data-type")
	})

	t.Run("AutheliaCompletionError", func(t *testing.T) {
		assert.Contains(t, portal, `data-testid={"openid-completion-outcome"}`, "the completion view no longer carries its outcome test id")
		assert.Contains(t, portal, "data-outcome={outcome}", "the completion view no longer exposes its outcome")
		assert.Contains(t, portal, `error ? "error"`, "the completion view no longer names the error outcome 'error'")
	})
}

func conformanceSelectorIdentifiers(selector string) (ids []string) {
	for _, part := range strings.Fields(selector) {
		if !strings.HasPrefix(part, "#") {
			continue
		}

		ids = append(ids, strings.TrimPrefix(part, "#"))
	}

	return ids
}

func conformancePortalSource(t *testing.T) string {
	t.Helper()

	builder := &strings.Builder{}

	root, err := os.OpenRoot("../../web/src")
	require.NoError(t, err)

	defer root.Close()

	require.NoError(t, fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".tsx") {
			return nil
		}

		data, err := root.ReadFile(path)
		if err != nil {
			return err
		}

		builder.Write(data)

		return nil
	}))

	require.NotEmpty(t, builder.String(), "no portal sources were read, so this test proves nothing")

	return builder.String()
}
