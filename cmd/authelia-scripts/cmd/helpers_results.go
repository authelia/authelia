package cmd

import (
	"bytes"
	"encoding/json"
	"io"
)

type testEvent struct {
	Action string `json:"Action"`
	Output string `json:"Output"`
}

// testOutputWriter reconstructs the console output of a `go test -json` event stream so the machine
// readable results can be captured without making the job log unreadable.
type testOutputWriter struct {
	out io.Writer
	buf bytes.Buffer
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
			if _, err = w.out.Write(line); err != nil {
				return len(p), err
			}

			continue
		}

		if event.Action == "output" {
			if _, err = io.WriteString(w.out, event.Output); err != nil {
				return len(p), err
			}
		}
	}
}
