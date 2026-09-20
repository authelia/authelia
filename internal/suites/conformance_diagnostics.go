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
	conformanceDiagnosticEntries = 20
	conformanceDiagnosticValue   = 2000
)

var (
	conformanceDiagnosticResults = map[string]bool{"FAILURE": true, "WARNING": true, "REVIEW": true}
	conformanceDiagnosticSkip    = map[string]bool{
		"msg": true, "result": true, "src": true, "requirements": true,
		"_id": true, "testId": true, "testOwner": true, "time": true,
		"blockId": true, "updatedAt": true, "img": true, "upload": true,
	}
)

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

// ConformanceAcceptedWarningReport renders a module's accepted WARNING for its subtest's log: why it is accepted,
// followed by the graded entries of the module's log, so the warning which was accepted is visible rather than only
// the fact that one was.
func ConformanceAcceptedWarningReport(module, status, reason string, entries []ConformanceLogEntry) string {
	report := fmt.Sprintf("Module '%s' finished with the accepted result 'WARNING' and the status '%s': %s.", module, status, reason)

	if detail := ConformanceDiagnostics(entries); detail != "" {
		return report + " Conformance log entries:" + detail
	}

	return report + " The module logged no warning entries."
}
