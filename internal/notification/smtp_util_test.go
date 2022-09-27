package notification

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestNewMIMECharacteristics(t *testing.T) {
	alphanumericPlus8Bit := utils.AlphaNumericCharacters + string([]byte{163})

	testCases := []struct {
		name     string
		expected MIMECharacteristics
		have     []byte
	}{
		{
			"ShouldDetectMessageCharacteristics7Bit",
			MIMECharacteristics{},
			createMIMEBytes(utils.AlphaNumericCharacters, true, 5, 150),
		},
		{
			"ShouldDetectMessageCharacteristicsLongLine",
			MIMECharacteristics{LongLines: true},
			createMIMEBytes(utils.AlphaNumericCharacters, true, 3, 1200),
		},
		{
			"ShouldDetectMessageCharacteristicsCRLF",
			MIMECharacteristics{LineFeeds: true},
			createMIMEBytes(utils.AlphaNumericCharacters, false, 5, 150),
		},
		{
			"ShouldDetectMessageCharacteristics8Bit",
			MIMECharacteristics{Characters8BIT: true},
			createMIMEBytes(alphanumericPlus8Bit, true, 3, 150),
		},
		{
			"ShouldDetectMessageCharacteristicsAll",
			MIMECharacteristics{true, true, true},
			createMIMEBytes(alphanumericPlus8Bit, false, 3, 1200),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMIMECharacteristics(tc.have)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func createMIMEBytes(charset string, crlf bool, lines, length int) []byte {
	buf := &bytes.Buffer{}

	for i := 0; i < lines; i++ {
		for j := 0; j < length/100; j++ {
			buf.WriteString(utils.RandomString(100, charset, false))
		}

		if n := length % 100; n != 0 {
			buf.WriteString(utils.RandomString(n, charset, false))
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
