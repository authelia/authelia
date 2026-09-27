// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"fmt"
	"time"
)

// ConformanceTrace records what the suite did for one module, for that module's subtest to log. The runner works on
// its own goroutine, which outlives any one subtest, so it records rather than logging directly. It records nothing
// unless the suite was asked to debug.
type ConformanceTrace struct {
	enabled bool
	lines   []string
}

// NewConformanceTrace returns a ConformanceTrace which records only when enabled.
func NewConformanceTrace(enabled bool) *ConformanceTrace {
	return &ConformanceTrace{enabled: enabled}
}

// Logf records a line, stamped with the time in UTC so it can be lined up with the container logs.
func (t *ConformanceTrace) Logf(format string, args ...any) {
	if t == nil || !t.enabled {
		return
	}

	t.lines = append(t.lines, time.Now().UTC().Format("15:04:05.000")+" "+fmt.Sprintf(format, args...))
}

// Lines returns the recorded lines.
func (t *ConformanceTrace) Lines() []string {
	if t == nil {
		return nil
	}

	return t.lines
}
