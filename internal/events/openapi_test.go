// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/authelia/authelia/v4/internal/events"
)

func TestOpenAPIShouldDocumentEveryRegisteredEventType(t *testing.T) {
	data, err := os.ReadFile(pathOpenAPI)
	require.NoError(t, err)

	spec := struct {
		Webhooks map[string]any `yaml:"webhooks"`
	}{}

	require.NoError(t, yaml.Unmarshal(untemplate(data), &spec))

	documented := make([]string, 0, len(spec.Webhooks))

	for name := range spec.Webhooks {
		documented = append(documented, name)
	}

	sort.Strings(documented)

	expected := append([]string{webhookBatch, webhookValidation}, events.Registered()...)

	sort.Strings(expected)

	assert.Equal(t, expected, documented,
		"the webhooks section of %s must document exactly the registered event types plus %q and %q",
		pathOpenAPI, webhookBatch, webhookValidation)
}

const (
	pathOpenAPI       = "../../api/openapi.yml"
	webhookBatch      = "batch"
	webhookValidation = "validation"
)

var reTemplateAction = regexp.MustCompile(`\{\{-?.*?-?\}\}`)

func untemplate(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))

	n := 0

	for _, line := range lines {
		if strings.TrimSpace(line) != "" && strings.TrimSpace(reTemplateAction.ReplaceAllString(line, "")) == "" {
			continue
		}

		out = append(out, reTemplateAction.ReplaceAllStringFunc(line, func(string) string {
			n++

			return "x" + strconv.Itoa(n)
		}))
	}

	return []byte(strings.Join(out, "\n"))
}
