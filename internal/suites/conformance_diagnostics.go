// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	// conformanceDiagnosticEntries caps how many failing entries one module reports. A module logs an entry per
	// condition and only the graded ones are shown, so this is generous in practice; the exported plan archive holds
	// the complete log either way.
	conformanceDiagnosticEntries = 20

	// conformanceDiagnosticValue caps one field's rendered length. Conformance logs carry whole HTTP exchanges, JWKS
	// documents and tokens, and seven plans of those would bury the failure that matters in the CI output.
	conformanceDiagnosticValue = 2000
)

// conformanceDiagnosticResults are the log entry results worth reporting when a module fails. SUCCESS and INFO
// entries are the narrative of a run that worked and would drown the ones that did not.
var conformanceDiagnosticResults = map[string]bool{"FAILURE": true, "WARNING": true, "REVIEW": true}

// conformanceDiagnosticSkip are entry fields rendered separately or not worth rendering: bookkeeping the reader
// already has, and img, which is a base64 image.
var conformanceDiagnosticSkip = map[string]bool{
	"msg": true, "result": true, "src": true, "requirements": true,
	"_id": true, "testId": true, "testOwner": true, "time": true,
	"img": true, "upload": true,
}

// ConformanceDiagnostics renders the graded entries of a module's log for a failure message. Each entry keeps its
// condition's own arguments, which is where a conformance failure actually explains itself: the message says a claim
// was missing, the arguments say which one.
func ConformanceDiagnostics(entries []ConformanceLogEntry) string {
	builder := &strings.Builder{}

	reported := 0

	for _, entry := range entries {
		if !conformanceDiagnosticResults[entry.Result] {
			continue
		}

		if reported == conformanceDiagnosticEntries {
			fmt.Fprintf(builder, "\n  ... further entries omitted, see the exported plan archive for the complete log")

			break
		}

		reported++

		fmt.Fprintf(builder, "\n  [%s] %s", entry.Result, entry.Msg)

		if entry.Src != "" {
			fmt.Fprintf(builder, "\n    condition: %s", entry.Src)
		}

		if requirements := conformanceDiagnosticField(entry.Fields["requirements"]); requirements != "" {
			fmt.Fprintf(builder, "\n    requirements: %s", requirements)
		}

		for _, name := range conformanceDiagnosticNames(entry.Fields) {
			fmt.Fprintf(builder, "\n    %s: %s", name, conformanceDiagnosticField(entry.Fields[name]))
		}
	}

	if reported == 0 {
		return ""
	}

	return builder.String()
}

// conformanceDiagnosticNames returns the entry's condition-specific field names in a stable order, so that the same
// failure reads the same way twice.
func conformanceDiagnosticNames(fields map[string]any) (names []string) {
	for name := range fields {
		if conformanceDiagnosticSkip[name] {
			continue
		}

		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// conformanceDiagnosticField renders one value compactly, truncating anything long enough to bury the rest.
func conformanceDiagnosticField(value any) string {
	if value == nil {
		return ""
	}

	rendered, ok := value.(string)

	if !ok {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}

		rendered = string(data)
	}

	if len(rendered) > conformanceDiagnosticValue {
		return rendered[:conformanceDiagnosticValue] + fmt.Sprintf("... (%d bytes truncated)", len(rendered)-conformanceDiagnosticValue)
	}

	return rendered
}
