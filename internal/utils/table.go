package utils

import (
	"fmt"
	"io"
	"strings"
)

// RenderTable renders headers and rows as a bordered, fixed-width text table.
func RenderTable(w io.Writer, headers []string, rows [][]string) error {
	widths := make([]int, len(headers))

	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	sep := func() string {
		var sb strings.Builder

		sb.WriteString("+")

		for _, w := range widths {
			sb.WriteString(strings.Repeat("-", w+2))
			sb.WriteString("+")
		}

		return sb.String()
	}

	row := func(cells []string) string {
		var sb strings.Builder

		sb.WriteString("|")

		for i, cell := range cells {
			sb.WriteString(" ")
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", widths[i]-len(cell)+1))
			sb.WriteString("|")
		}

		return sb.String()
	}

	line := sep()

	for _, s := range []string{line, row(headers), line} {
		if _, err := fmt.Fprintln(w, s); err != nil {
			return err
		}
	}

	for _, r := range rows {
		if _, err := fmt.Fprintln(w, row(r)); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(w, line)

	return err
}
