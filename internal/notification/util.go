package notification

import (
	"bytes"
)

func getSanitizedPassword(in string) (out string) {
	buf := bytes.NewBuffer(nil)

	switch {
	case len(in) > 20:
		buf.Write([]byte(in[0:5]))
		buf.Write([]byte(reAnyCharacter.ReplaceAllString(in[5:], "*")))
	case len(in) > 4:
		buf.Write([]byte(in[0:2]))
		buf.Write([]byte(reAnyCharacter.ReplaceAllString(in[2:], "*")))
	default:
		buf.Write([]byte(reAnyCharacter.ReplaceAllString(in, "*")))
	}

	return buf.String()
}
