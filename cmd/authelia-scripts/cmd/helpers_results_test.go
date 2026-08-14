package cmd

import (
	"bytes"
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
