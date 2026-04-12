package suites

import (
	"time"
)

const externalSuiteNameDocs = "docs"

func init() {
	ExternalGlobalRegistry.Register(externalSuiteNameDocs, ExternalSuite{
		Description: "Hugo documentation site rendering tests",
		TestTimeout: 1 * time.Minute,
	})
}
