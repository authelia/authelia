// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/events"
)

func TestWebhookSchemaIdentifiersMatchDescriptors(t *testing.T) {
	names := events.Registered()

	require.NotEmpty(t, names)

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			descriptor, ok := events.Lookup(name)

			require.True(t, ok)
			assert.Equal(t, descriptor.DataSchema, fmt.Sprintf(urlFormatJSONSchemaWebhooks, events.ContractVersion, name))
		})
	}
}

func TestWebhookSchemasArePublished(t *testing.T) {
	t.Chdir(dirRepositoryRoot)

	entries, err := os.ReadDir(webhookJSONSchemaTestDir())

	require.NoError(t, err)

	actual := make([]string, 0, len(entries))

	for _, entry := range entries {
		actual = append(actual, entry.Name())
	}

	documents := webhookJSONSchemaDocuments()

	expected := make([]string, 0, len(documents))

	for _, document := range documents {
		expected = append(expected, document.Name+extJSON)
	}

	assert.ElementsMatch(t, expected, actual, "The published webhook documents do not match the event registry. Regenerate them with `%s`.", cmdWebhookJSONSchemaRegenerate)
}

func TestWebhookSchemasAreCurrent(t *testing.T) {
	t.Chdir(dirRepositoryRoot)

	r, err := newWebhookJSONSchemaReflector(dirEvents)

	require.NoError(t, err)

	dir := webhookJSONSchemaTestDir()

	for _, document := range webhookJSONSchemaDocuments() {
		t.Run(document.Name, func(t *testing.T) {
			path := filepath.Join(dir, document.Name+extJSON)

			published, err := os.ReadFile(path)

			require.NoError(t, err)

			buf := &bytes.Buffer{}

			require.NoError(t, encodeJSONSchema(buf, reflectWebhookJSONSchema(r, document.Value, document.Name)))

			assert.Equal(t, buf.String(), string(published), "The published document '%s' is stale: it no longer matches the Go type it is generated from. Regenerate it with `%s`.", path, cmdWebhookJSONSchemaRegenerate)
		})
	}
}

func webhookJSONSchemaTestDir() string {
	return webhookJSONSchemaDir(filepath.Join(dirDocs, dirDocsStatic, dirDocsStaticJSONSchemas, dirDocsStaticJSONSchemasWebhooks))
}

const cmdWebhookJSONSchemaRegenerate = "go run ./cmd/authelia-gen docs json-schema webhooks"
