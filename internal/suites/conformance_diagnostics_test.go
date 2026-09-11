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

func TestConformanceAcceptedWarningReport(t *testing.T) {
	entries := []ConformanceLogEntry{
		{Msg: "Checked the id_token", Result: "SUCCESS", Src: "ValidateIdToken"},
		{Msg: "acr value in id_token is not (one of the) requested values", Result: "WARNING", Src: "ValidateIdTokenACRClaimAgainstAcrValuesRequest",
			Fields: map[string]any{"requirements": []any{"OIDCC-3.1.2.1", "OIDCC-15.1"}, "id_token_acr": "0"}},
	}

	report := ConformanceAcceptedWarningReport("oidcc-ensure-request-with-acr-values-succeeds", "FINISHED", "the acr is the fallback", entries)

	assert.Contains(t, report, "Module 'oidcc-ensure-request-with-acr-values-succeeds' finished with the accepted result 'WARNING' and the status 'FINISHED': the acr is the fallback.")
	assert.Contains(t, report, "[WARNING] acr value in id_token is not (one of the) requested values")
	assert.Contains(t, report, "condition: ValidateIdTokenACRClaimAgainstAcrValuesRequest")
	assert.Contains(t, report, "id_token_acr: 0")
	assert.NotContains(t, report, "Checked the id_token", "only the graded entries are reported")

	assert.Equal(t, "Module 'oidcc-server' finished with the accepted result 'WARNING' and the status 'FINISHED': reason. The module logged no warning entries.",
		ConformanceAcceptedWarningReport("oidcc-server", "FINISHED", "reason", nil))
}

func TestConformanceDiagnosticsOmitsTheConformanceUIsBookkeeping(t *testing.T) {
	detail := ConformanceDiagnostics([]ConformanceLogEntry{
		{Msg: "acr value in id_token is not (one of the) requested values", Result: "WARNING",
			Fields: map[string]any{"blockId": "a1c507", "updatedAt": float64(1789035070157), "id_token_acr": "0"}},
	})

	assert.Contains(t, detail, "id_token_acr: 0")
	assert.NotContains(t, detail, "blockId")
	assert.NotContains(t, detail, "updatedAt")
}
