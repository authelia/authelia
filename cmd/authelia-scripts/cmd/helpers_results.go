// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

type testEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
	Output string `json:"Output"`
}

type testOutputWriter struct {
	out     io.Writer
	buf     bytes.Buffer
	grouped bool
	pending []testFraming
	buildkite bool
	deferring bool
	deferred  bytes.Buffer
}

type testFraming struct {
	test string
	line string
}

func (w *testOutputWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)

	for {
		line, err := w.buf.ReadBytes('\n')
		if err != nil {
			w.buf.Write(line)

			return len(p), nil
		}

		var event testEvent

		if json.Unmarshal(line, &event) != nil {
			if err = w.write(string(line)); err != nil {
				return len(p), err
			}

			continue
		}

		if err = w.handle(event); err != nil {
			return len(p), err
		}
	}
}

func (w *testOutputWriter) handle(event testEvent) (err error) {
	if event.Action != "output" {
		return nil
	}

	if !w.grouped {
		return w.write(event.Output)
	}

	switch output := event.Output; {
	case isTestFramingLine(output):
		w.pending = append(w.pending, testFraming{test: event.Test, line: output})

		return nil
	case isBuildkiteMarkerLine(output) && !isTestSummaryLine(output):
		if err = w.release(func(framing testFraming) bool { return !isTestOrParent(framing.test, event.Test) }); err != nil {
			return err
		}

		if err = w.write(output); err != nil {
			return err
		}

		return w.release(func(testFraming) bool { return true })
	default:
		if err = w.release(func(testFraming) bool { return true }); err != nil {
			return err
		}

		return w.write(output)
	}
}

func (w *testOutputWriter) write(output string) (err error) {
	if w.buildkite && !w.deferring && isTopLevelTestSummaryLine(output) {
		w.deferring = true
	}

	if w.deferring {
		w.deferred.WriteString(output)

		return nil
	}

	_, err = io.WriteString(w.out, output)

	return err
}

func (w *testOutputWriter) Flush() (err error) {
	if err = w.release(func(testFraming) bool { return true }); err != nil {
		return err
	}

	_, err = w.out.Write(w.deferred.Bytes())

	w.deferred.Reset()

	return err
}

func (w *testOutputWriter) release(match func(framing testFraming) bool) (err error) {
	kept := w.pending[:0]

	for _, framing := range w.pending {
		if !match(framing) {
			kept = append(kept, framing)

			continue
		}

		if err = w.write(framing.line); err != nil {
			return err
		}
	}

	w.pending = kept

	return nil
}

func isTestOrParent(test, child string) bool {
	return test == child || strings.HasPrefix(child, test+"/")
}

func isTestFramingLine(output string) bool {
	for _, prefix := range []string{"=== RUN", "=== PAUSE", "=== CONT", "=== NAME"} {
		if strings.HasPrefix(output, prefix) {
			return true
		}
	}

	return false
}

func isTestSummaryLine(output string) bool {
	trimmed := strings.TrimLeft(output, " ")

	for _, prefix := range []string{"--- PASS: ", "--- FAIL: ", "--- SKIP: "} {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	return false
}

func isTopLevelTestSummaryLine(output string) bool {
	return isTestSummaryLine(output) && !strings.HasPrefix(output, " ")
}

func isBuildkiteMarkerLine(output string) bool {
	for _, prefix := range []string{"--- ", "+++ ", "~~~ ", "^^^ +++"} {
		if strings.HasPrefix(output, prefix) {
			return true
		}
	}

	return false
}
