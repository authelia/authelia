// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import "regexp"

const (
	cmdOpenAPIRegenerate           = "go run ./cmd/authelia-gen docs api openapi"
	cmdWebhookJSONSchemaRegenerate = "go run ./cmd/authelia-gen docs json-schema webhooks"
)

var reOpenAPIWebhookSchemaRef = regexp.MustCompile(`'#/components/schemas/(` + regexp.QuoteMeta(openAPIWebhookSchemaComponentPrefix) + `[^']+)'`)
