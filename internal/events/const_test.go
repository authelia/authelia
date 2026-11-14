// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events_test

import "regexp"

const (
	pathOpenAPI       = "../../api/openapi.yml"
	webhookBatch      = "batch"
	webhookValidation = "validation"
)

var reTemplateAction = regexp.MustCompile(`\{\{-?.*?-?\}\}`)
