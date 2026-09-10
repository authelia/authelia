// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestOutputWriterShouldReconstructConsoleOutput(t *testing.T) {
	testCases := []struct {
		name     string
		writes   []string
		expected string
	}{
		{
			"ShouldEmitOutputActionsOnly",
			[]string{
				`{"Action":"run","Test":"TestExample"}` + "\n" +
					`{"Action":"output","Test":"TestExample","Output":"=== RUN   TestExample\n"}` + "\n" +
					`{"Action":"output","Test":"TestExample","Output":"--- PASS: TestExample (0.00s)\n"}` + "\n" +
					`{"Action":"pass","Test":"TestExample"}` + "\n",
			},
			"=== RUN   TestExample\n--- PASS: TestExample (0.00s)\n",
		},
		{
			"ShouldPassThroughNonEventLines",
			[]string{"# github.com/authelia/authelia/v4/internal/suites\nsuite.go:1:1: undefined: x\n"},
			"# github.com/authelia/authelia/v4/internal/suites\nsuite.go:1:1: undefined: x\n",
		},
		{
			"ShouldBufferPartialLinesAcrossWrites",
			[]string{
				`{"Action":"output","Output":"first`,
				`\n"}` + "\n" + `{"Action":"output","Output":"second\n"}` + "\n",
			},
			"first\nsecond\n",
		},
		{
			"ShouldRetainTrailingPartialLine",
			[]string{`{"Action":"output","Output":"incomplete\n"}`},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			writer := &testOutputWriter{out: out}

			for _, write := range tc.writes {
				n, err := writer.Write([]byte(write))

				assert.NoError(t, err)
				assert.Equal(t, len(write), n)
			}

			assert.Equal(t, tc.expected, out.String())
		})
	}
}

func TestTestOutputWriterShouldPlaceFramingInsideBuildkiteGroups(t *testing.T) {
	event := func(test, output string) string {
		return fmt.Sprintf(`{"Action":"output","Test":%q,"Output":%q}`, test, output) + "\n"
	}

	testCases := []struct {
		name     string
		events   []string
		expected string
	}{
		{
			"ShouldPlaceTheFramingOfTheTestWhichStartedAGroupInsideIt",
			[]string{
				event("TestSuite", "=== RUN   TestSuite\n"),
				event("TestSuite/TestBasic", "=== RUN   TestSuite/TestBasic\n"),
				event("TestSuite/TestBasic", "--- OIDC Conformance Plan: basic\n"),
			},
			"--- OIDC Conformance Plan: basic\n=== RUN   TestSuite\n=== RUN   TestSuite/TestBasic\n",
		},
		{
			"ShouldLeaveTheFramingOfTheLastGroupBeforeTheNextOne",
			[]string{
				event("TestSuite/TestBasic/Codereuse", "=== RUN   TestSuite/TestBasic/Codereuse\n"),
				event("TestSuite/TestBasicFormPost", "=== RUN   TestSuite/TestBasicFormPost\n"),
				event("TestSuite/TestBasicFormPost", "--- OIDC Conformance Plan: basic-form-post\n"),
			},
			"=== RUN   TestSuite/TestBasic/Codereuse\n--- OIDC Conformance Plan: basic-form-post\n=== RUN   TestSuite/TestBasicFormPost\n",
		},
		{
			"ShouldPrintTheFramingOfEveryTestBeforeTheNextOutput",
			[]string{
				event("TestSuite/TestBasic/Server", "=== RUN   TestSuite/TestBasic/Server\n"),
				event("TestSuite/TestBasic/MaxAge1", "=== RUN   TestSuite/TestBasic/MaxAge1\n"),
				event("TestSuite/TestBasic/MaxAge1", "    suite_test.go:1: Module finished with the result 'FAILED'\n"),
				event("TestSuite/TestBasic/MaxAge1", "^^^ +++\n"),
			},
			"=== RUN   TestSuite/TestBasic/Server\n=== RUN   TestSuite/TestBasic/MaxAge1\n    suite_test.go:1: Module finished with the result 'FAILED'\n^^^ +++\n",
		},
		{
			"ShouldPrintHeldFramingBeforeTheSummary",
			[]string{
				event("TestSuite/TestBasic/Server", "=== RUN   TestSuite/TestBasic/Server\n"),
				event("TestSuite/TestBasic/Server", "--- PASS: TestSuite/TestBasic/Server (0.00s)\n"),
			},
			"=== RUN   TestSuite/TestBasic/Server\n--- PASS: TestSuite/TestBasic/Server (0.00s)\n",
		},
		{
			"ShouldPrintOutputWhichBelongsToNoTest",
			[]string{
				event("TestSuite", "=== RUN   TestSuite\n"),
				event("", "PASS\n"),
			},
			"=== RUN   TestSuite\nPASS\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			writer := &testOutputWriter{out: out, grouped: true}

			for _, write := range tc.events {
				_, err := writer.Write([]byte(write))

				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected, out.String())
		})
	}

	t.Run("ShouldPrintEverythingInPlaceWhenNotGrouped", func(t *testing.T) {
		out := &bytes.Buffer{}
		writer := &testOutputWriter{out: out}

		for _, write := range []string{
			event("TestSuite/TestBasic", "=== RUN   TestSuite/TestBasic\n"),
			event("TestSuite/TestBasic", "--- OIDC Conformance Plan: basic\n"),
		} {
			_, err := writer.Write([]byte(write))

			assert.NoError(t, err)
		}

		assert.Equal(t, "=== RUN   TestSuite/TestBasic\n--- OIDC Conformance Plan: basic\n", out.String())
	})
}

func TestTestOutputWriterShouldKeepTheSummaryFromStartingABuildkiteGroup(t *testing.T) {
	events := `{"Action":"output","Test":"TestSuite","Output":"--- FAIL: TestSuite (227.08s)\n"}` + "\n" +
		`{"Action":"output","Test":"TestSuite/TestBasic","Output":"    --- FAIL: TestSuite/TestBasic (116.14s)\n"}` + "\n" +
		`{"Action":"output","Output":"FAIL\n"}` + "\n"

	testCases := []struct {
		name      string
		buildkite bool
		grouped   bool
		expected  string
	}{
		{"ShouldIndentTheTopLevelSummaryInBuildkite", true, false, " --- FAIL: TestSuite (227.08s)\n    --- FAIL: TestSuite/TestBasic (116.14s)\nFAIL\n"},
		{"ShouldIndentTheTopLevelSummaryInBuildkiteWhenGrouped", true, true, " --- FAIL: TestSuite (227.08s)\n    --- FAIL: TestSuite/TestBasic (116.14s)\nFAIL\n"},
		{"ShouldLeaveTheSummaryAloneElsewhere", false, false, "--- FAIL: TestSuite (227.08s)\n    --- FAIL: TestSuite/TestBasic (116.14s)\nFAIL\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			writer := &testOutputWriter{out: out, buildkite: tc.buildkite, grouped: tc.grouped}

			_, err := writer.Write([]byte(events))

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, out.String())
		})
	}
}
