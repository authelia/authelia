// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/authelia/jsonschema"

	"github.com/authelia/authelia/v4/internal/events"
)

func newDocsAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseDocsAPI,
		Short: "Generate the API specifications",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newDocsAPIOpenAPICmd())

	return cmd
}

func newDocsAPIOpenAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseDocsAPIOpenAPI,
		Short: "Generate the webhook event schema components of the OpenAPI specification",
		RunE:  docsAPIOpenAPIRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func docsAPIOpenAPIRunE(cmd *cobra.Command, _ []string) (err error) {
	var eventsDir, path string

	if eventsDir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDirEvents); err != nil {
		return err
	}

	if path, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagFileAPIOpenAPI); err != nil {
		return err
	}

	var block string

	if block, err = openAPIWebhookSchemas(eventsDir); err != nil {
		return err
	}

	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return err
	}

	var out []byte

	if out, err = openAPIReplaceBlock(data, block); err != nil {
		return fmt.Errorf("failed to update '%s': %w", path, err)
	}

	var f *os.File

	if f, err = os.Create(path); err != nil {
		return err
	}

	if _, err = f.Write(out); err != nil {
		_ = f.Close()

		return err
	}

	return f.Close()
}

func openAPIWebhookSchemas(dir string) (block string, err error) {
	var r *jsonschema.Reflector

	if r, err = newWebhookJSONSchemaReflector(dir); err != nil {
		return "", err
	}

	documents := openAPIWebhookSchemaDocuments()

	names := make(map[string]bool, len(documents))

	for _, document := range documents {
		names[document.Name] = true
	}

	doc := &yaml.Node{Kind: yaml.MappingNode}

	for _, document := range documents {
		var node *yaml.Node

		if node, err = openAPIWebhookSchemaNode(r, document, names); err != nil {
			return "", err
		}

		doc.Content = append(doc.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: openAPIWebhookSchemaComponent(document.Name)}, node)
	}

	buf := &bytes.Buffer{}

	encoder := yaml.NewEncoder(buf)

	encoder.SetIndent(2)

	if err = encoder.Encode(doc); err != nil {
		return "", err
	}

	if err = encoder.Close(); err != nil {
		return "", err
	}

	return openAPIIndent(buf.String()), nil
}

func openAPIWebhookSchemaNode(r *jsonschema.Reflector, document openAPIWebhookSchemaDocument, names map[string]bool) (node *yaml.Node, err error) {
	s := r.Reflect(document.Value)

	for name := range s.Definitions {
		if !names[name] {
			return nil, fmt.Errorf("the '%s' document defines the type '%s' which is not hoisted to a component of its own, add it to the document list", document.Name, name)
		}
	}

	s.Version = ""
	s.ID = ""
	s.Definitions = nil

	var data []byte

	if data, err = json.Marshal(s); err != nil {
		return nil, err
	}

	data = reJSONSchemaDefinitionRef.ReplaceAll(data, []byte(`"#/components/schemas/`+openAPIWebhookSchemaComponentPrefix+`${1}"`))

	node = &yaml.Node{}

	if err = yaml.Unmarshal(data, node); err != nil {
		return nil, err
	}

	node = node.Content[0]

	openAPIBlockStyle(node)

	return node, nil
}

func openAPIBlockStyle(node *yaml.Node) {
	node.Style = 0

	for _, child := range node.Content {
		openAPIBlockStyle(child)
	}
}

func openAPIWebhookSchemaDocuments() (documents []openAPIWebhookSchemaDocument) {
	values := map[string]any{
		"Envelope":           &events.Envelope{},
		"EnvelopeBatch":      &events.EnvelopeBatch{},
		"Notification":       &events.Notification{},
		"NotificationValues": &events.NotificationValues{},
		"Recipient":          &events.Recipient{},
	}

	for _, name := range events.Registered() {
		descriptor, ok := events.Lookup(name)
		if !ok {
			continue
		}

		value := descriptor.New()

		values[reflect.Indirect(reflect.ValueOf(value)).Type().Name()] = value
	}

	names := make([]string, 0, len(values))

	for name := range values {
		names = append(names, name)
	}

	sort.Strings(names)

	documents = make([]openAPIWebhookSchemaDocument, 0, len(names))

	for _, name := range names {
		documents = append(documents, openAPIWebhookSchemaDocument{Name: name, Value: values[name]})
	}

	return documents
}

func openAPIWebhookSchemaComponent(name string) string {
	return openAPIWebhookSchemaComponentPrefix + name
}

func openAPIIndent(block string) string {
	lines := strings.Split(strings.TrimRight(block, "\n"), "\n")

	for i, line := range lines {
		if line == "" {
			continue
		}

		lines[i] = openAPIWebhookSchemaIndent + line
	}

	return strings.Join(lines, "\n")
}

func openAPIReplaceBlock(data []byte, block string) (out []byte, err error) {
	lines := strings.Split(string(data), "\n")

	start, end := -1, -1

	for i, line := range lines {
		switch strings.TrimSpace(line) {
		case openAPIWebhookSchemaMarkerStart:
			if start != -1 {
				return nil, fmt.Errorf("the generated block start marker occurs more than once")
			}

			start = i
		case openAPIWebhookSchemaMarkerEnd:
			if end != -1 {
				return nil, fmt.Errorf("the generated block end marker occurs more than once")
			}

			end = i
		}
	}

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("could not find the generated block markers '%s' and '%s'", openAPIWebhookSchemaMarkerStart, openAPIWebhookSchemaMarkerEnd)
	}

	if end < start {
		return nil, fmt.Errorf("the generated block end marker occurs before the start marker")
	}

	updated := make([]string, 0, len(lines))

	updated = append(updated, lines[:start+1]...)
	updated = append(updated, strings.Split(block, "\n")...)
	updated = append(updated, lines[end:]...)

	return []byte(strings.Join(updated, "\n")), nil
}
