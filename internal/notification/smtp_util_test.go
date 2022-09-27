package notification

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestNewMIMECharacteristics(t *testing.T) {
	testCases := []struct {
		name     string
		expected MIMECharacteristics
		have     []byte
	}{
		{
			"ShouldDetectMessageCharacteristics7Bit",
			MIMECharacteristics{},
			createMIMEBytes(false, true, 5, 150),
		},
		{
			"ShouldDetectMessageCharacteristicsLongLine",
			MIMECharacteristics{LongLines: true},
			createMIMEBytes(false, true, 3, 1200),
		},
		{
			"ShouldDetectMessageCharacteristicsLF",
			MIMECharacteristics{LineFeeds: true},
			createMIMEBytes(false, false, 5, 150),
		},
		{
			"ShouldDetectMessageCharacteristics8Bit",
			MIMECharacteristics{Characters8BIT: true},
			createMIMEBytes(true, true, 3, 150),
		},
		{
			"ShouldDetectMessageCharacteristicsLongLineAndLF",
			MIMECharacteristics{true, true, false},
			createMIMEBytes(false, false, 3, 1200),
		},
		{
			"ShouldDetectMessageCharacteristicsLongLineAnd8Bit",
			MIMECharacteristics{true, false, true},
			createMIMEBytes(true, true, 3, 1200),
		},
		{
			"ShouldDetectMessageCharacteristics8BitAndLF",
			MIMECharacteristics{false, true, true},
			createMIMEBytes(true, false, 3, 150),
		},
		{
			"ShouldDetectMessageCharacteristicsAll",
			MIMECharacteristics{true, true, true},
			createMIMEBytes(true, false, 3, 1200),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMIMECharacteristics(tc.have)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func createMIMEBytes(include8bit, crlf bool, lines, length int) []byte {
	buf := &bytes.Buffer{}

	for i := 0; i < lines; i++ {
		for j := 0; j < length/100; j++ {
			switch {
			case include8bit:
				buf.Write(utils.RandomBytes(50, utils.AlphaNumericCharacters, false))
				buf.Write([]byte{163})
				buf.Write(utils.RandomBytes(49, utils.AlphaNumericCharacters, false))
			default:
				buf.Write(utils.RandomBytes(100, utils.AlphaNumericCharacters, false))
			}
		}

		if n := length % 100; n != 0 {
			buf.Write(utils.RandomBytes(n, utils.AlphaNumericCharacters, false))
		}

		switch {
		case crlf:
			buf.Write([]byte{13, 10})
		default:
			buf.Write([]byte{10})
		}
	}

	return buf.Bytes()
}
