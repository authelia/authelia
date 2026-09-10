// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceDiagnostics(t *testing.T) {
	// The shape a real scope failure takes: the message names the problem, the condition's own arguments name the
	// claim. Taken from AbstractVerifyScopesReturnedInClaims, which logs exactly these three arguments.
	scopeFailure := `[
	  {"result":"SUCCESS","src":"CallUserInfoEndpoint","msg":"Userinfo endpoint response was 200"},
	  {"result":"INFO","src":"ExtractUserInfo","msg":"Extracted userinfo"},
	  {"result":"WARNING","src":"VerifyScopesReturnedInUserInfoClaims","msg":"'claims' in userinfo doesn't contain all scope items of scope in authorization request","requirements":["OIDCC-5.4"],
	   "expected_scope_items":["name","given_name","birthdate"],"actual_scope_items":["name"],"missing_items":["given_name","birthdate"],
	   "_id":"abc","testId":"t1","testOwner":"anon","time":1789021616378}
	]`

	var entries []ConformanceLogEntry

	require.NoError(t, json.Unmarshal([]byte(scopeFailure), &entries))

	out := ConformanceDiagnostics(entries)

	t.Run("ShouldReportTheGradedEntry", func(t *testing.T) {
		assert.Contains(t, out, "[WARNING]")
		assert.Contains(t, out, "doesn't contain all scope items")
		assert.Contains(t, out, "condition: VerifyScopesReturnedInUserInfoClaims")
		assert.Contains(t, out, `requirements: ["OIDCC-5.4"]`)
	})

	t.Run("ShouldReportTheConditionArgumentsThatExplainIt", func(t *testing.T) {
		assert.Contains(t, out, `missing_items: ["given_name","birthdate"]`)
		assert.Contains(t, out, `expected_scope_items: ["name","given_name","birthdate"]`)
		assert.Contains(t, out, `actual_scope_items: ["name"]`)
	})

	t.Run("ShouldOmitUngradedNarrativeAndBookkeeping", func(t *testing.T) {
		assert.NotContains(t, out, "Userinfo endpoint response was 200")
		assert.NotContains(t, out, "Extracted userinfo")
		assert.NotContains(t, out, "testOwner")
		assert.NotContains(t, out, "1789021616378")
	})

	t.Run("ShouldOrderFieldsStablyAcrossRuns", func(t *testing.T) {
		for i := 0; i < 8; i++ {
			var repeat []ConformanceLogEntry

			require.NoError(t, json.Unmarshal([]byte(scopeFailure), &repeat))
			assert.Equal(t, out, ConformanceDiagnostics(repeat))
		}
	})
}

func TestConformanceDiagnosticsReportsNothingWhenEveryEntryPassed(t *testing.T) {
	entries := []ConformanceLogEntry{
		{Result: "SUCCESS", Msg: "fine"},
		{Result: "INFO", Msg: "also fine"},
	}

	assert.Empty(t, ConformanceDiagnostics(entries))
}

func TestConformanceDiagnosticsTruncatesAValueLargeEnoughToBuryTheRest(t *testing.T) {
	entries := []ConformanceLogEntry{{
		Result: "FAILURE",
		Msg:    "token endpoint returned an error",
		Fields: map[string]any{"result": "FAILURE", "msg": "token endpoint returned an error", "http_response_body": strings.Repeat("x", conformanceDiagnosticValue+500)},
	}}

	out := ConformanceDiagnostics(entries)

	assert.Contains(t, out, "500 bytes truncated")
	assert.Less(t, len(out), conformanceDiagnosticValue+400)
}

func TestConformanceDiagnosticsCapsTheNumberOfEntries(t *testing.T) {
	entries := make([]ConformanceLogEntry, 0, conformanceDiagnosticEntries+5)

	for i := 0; i < conformanceDiagnosticEntries+5; i++ {
		entries = append(entries, ConformanceLogEntry{Result: "FAILURE", Msg: "failed"})
	}

	out := ConformanceDiagnostics(entries)

	assert.Equal(t, conformanceDiagnosticEntries, strings.Count(out, "[FAILURE]"))
	assert.Contains(t, out, "further entries omitted")
}
