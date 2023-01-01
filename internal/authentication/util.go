package authentication

import "golang.org/x/crypto/md4" //nolint:staticcheck

// String returns a string representation of an authentication.Level.
func (l Level) String() string {
	switch l {
	case NotAuthenticated:
		return "not_authenticated"
	case OneFactor:
		return "one_factor"
	case TwoFactor:
		return "two_factor"
	default:
		return "invalid"
	}
}

// NTHash calculates the NTLM hash of an ASCII string (in).
func NTHash(in string) []byte {
	/* Prepare a byte array to return. */
	u16 := []byte{}

	/* Add all bytes, as well as the 0x00 of UTF-16. */
	for _, b := range []byte(in) {
		u16 = append(u16, b)
		u16 = append(u16, 0x00)
	}

	/* Hash the byte array with MD4. */
	mdfour := md4.New()
	mdfour.Write(u16)

	/* Return the output. */
	return mdfour.Sum(nil)
}
