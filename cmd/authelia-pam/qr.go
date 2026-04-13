package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// renderQRCode encodes the given payload as a Unicode half-block QR code suitable
// for terminal display.
func renderQRCode(payload string) (string, error) {
	code, err := qr.Encode(payload, qr.M, qr.Auto)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR code: %w", err)
	}

	size := code.Bounds().Max.X

	const quiet = 1

	var b strings.Builder

	for y := -quiet; y < size+quiet; y += 2 {
		b.WriteByte(' ')

		for x := -quiet; x < size+quiet; x++ {
			top := moduleOn(code, x, y, size)
			bot := moduleOn(code, x, y+1, size)

			switch {
			case top && bot:
				b.WriteString("\u2588") // Full block.
			case top && !bot:
				b.WriteString("\u2580") // Upper half.
			case !top && bot:
				b.WriteString("\u2584") // Lower half.
			default:
				b.WriteString(" ")
			}
		}

		b.WriteByte('\n')
	}

	return b.String(), nil
}

// moduleOn reports whether the QR module at (x, y) is set, treating quiet-zone
// coordinates as unset.
func moduleOn(code barcode.Barcode, x, y, size int) bool {
	if x < 0 || y < 0 || x >= size || y >= size {
		return false
	}

	return code.At(x, y) == color.Black
}
