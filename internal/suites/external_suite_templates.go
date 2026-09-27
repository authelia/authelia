// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"time"
)

const externalSuiteNameTemplates = "templates"

func init() {
	ExternalGlobalRegistry.Register(externalSuiteNameTemplates, ExternalSuite{
		Description: "React-email template rendering tests",
		TestTimeout: 1 * time.Minute,
	})
}
