// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIWebhookSchemasAreCurrent(t *testing.T) {
	t.Chdir(dirRepositoryRoot)

	block, err := openAPIWebhookSchemas(dirEvents)

	require.NoError(t, err)

	data, err := os.ReadFile(fileAPIOpenAPI)

	require.NoError(t, err)

	assert.Equal(t, block, openAPIWebhookSchemasBlock(t, string(data)), "The webhook event schemas in '%s' are stale: they no longer match the Go types they are generated from. Regenerate them with `%s`.", fileAPIOpenAPI, cmdOpenAPIRegenerate)
}

func TestOpenAPIWebhookSchemaReferencesResolve(t *testing.T) {
	t.Chdir(dirRepositoryRoot)

	data, err := os.ReadFile(fileAPIOpenAPI)

	require.NoError(t, err)

	defined := make(map[string]bool)

	for _, document := range openAPIWebhookSchemaDocuments() {
		defined[openAPIWebhookSchemaComponent(document.Name)] = true
	}

	referenced := make(map[string]bool)

	for _, match := range reOpenAPIWebhookSchemaRef.FindAllStringSubmatch(string(data), -1) {
		referenced[match[1]] = true
	}

	require.NotEmpty(t, referenced)

	for name := range referenced {
		assert.True(t, defined[name], "The specification references the component '%s' which no longer exists, most likely because the Go type it is generated from was renamed.", name)
	}

	for name := range defined {
		assert.True(t, referenced[name], "The component '%s' is generated but never referenced, most likely because a Go type was renamed or an event was removed.", name)
	}
}

func TestOpenAPIReplaceBlockShouldErrorWithoutMarkers(t *testing.T) {
	testCases := []struct {
		name string
		have string
		err  string
	}{
		{
			"ShouldErrorOnMissingMarkers",
			"schemas:\n",
			"could not find the generated block markers '" + openAPIWebhookSchemaMarkerStart + "' and '" + openAPIWebhookSchemaMarkerEnd + "'",
		},
		{
			"ShouldErrorOnDuplicateStartMarker",
			"    " + openAPIWebhookSchemaMarkerStart + "\n    " + openAPIWebhookSchemaMarkerStart + "\n    " + openAPIWebhookSchemaMarkerEnd + "\n",
			"the generated block start marker occurs more than once",
		},
		{
			"ShouldErrorOnDuplicateEndMarker",
			"    " + openAPIWebhookSchemaMarkerStart + "\n    " + openAPIWebhookSchemaMarkerEnd + "\n    " + openAPIWebhookSchemaMarkerEnd + "\n",
			"the generated block end marker occurs more than once",
		},
		{
			"ShouldErrorOnReversedMarkers",
			"    " + openAPIWebhookSchemaMarkerEnd + "\n    " + openAPIWebhookSchemaMarkerStart + "\n",
			"the generated block end marker occurs before the start marker",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := openAPIReplaceBlock([]byte(tc.have), "")

			assert.EqualError(t, err, tc.err)
		})
	}
}

func TestOpenAPIReplaceBlockShouldPreserveSurroundingContent(t *testing.T) {
	have := "before\n    " + openAPIWebhookSchemaMarkerStart + "\n    stale\n    " + openAPIWebhookSchemaMarkerEnd + "\nafter\n"

	out, err := openAPIReplaceBlock([]byte(have), "    fresh")

	require.NoError(t, err)
	assert.Equal(t, "before\n    "+openAPIWebhookSchemaMarkerStart+"\n    fresh\n    "+openAPIWebhookSchemaMarkerEnd+"\nafter\n", string(out))
}

func TestOpenAPIWebhookSchemaDocumentsShouldCoverEveryPayload(t *testing.T) {
	documents := openAPIWebhookSchemaDocuments()

	names := make([]string, 0, len(documents))

	for _, document := range documents {
		names = append(names, document.Name)
	}

	assert.True(t, sort.StringsAreSorted(names), "The documents must be in a deterministic order so the generated block is stable.")

	for _, name := range []string{"Envelope", "EnvelopeBatch", "Notification", "NotificationValues", "Recipient"} {
		assert.Contains(t, names, name)
	}
}

func openAPIWebhookSchemasBlock(t *testing.T, data string) string {
	t.Helper()

	lines := strings.Split(data, "\n")

	start, end := -1, -1

	for i, line := range lines {
		switch strings.TrimSpace(line) {
		case openAPIWebhookSchemaMarkerStart:
			start = i
		case openAPIWebhookSchemaMarkerEnd:
			end = i
		}
	}

	require.NotEqual(t, -1, start)
	require.NotEqual(t, -1, end)

	return strings.Join(lines[start+1:end], "\n")
}
