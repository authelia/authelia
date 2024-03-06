package random

const (
	// DefaultN is the default value of n.
	DefaultN = 72
)

const (
	// CharSetAlphabeticLower are literally just valid alphabetic lowercase printable ASCII chars.
	CharSetAlphabeticLower = "abcdefghijklmnopqrstuvwxyz"

	// CharSetAlphabeticUpper are literally just valid alphabetic uppercase printable ASCII chars.
	CharSetAlphabeticUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// CharSetAlphabetic are literally just valid alphabetic printable ASCII chars.
	CharSetAlphabetic = CharSetAlphabeticLower + CharSetAlphabeticUpper

	// CharSetNumeric are literally just valid numeric chars.
	CharSetNumeric = "0123456789"

	// CharSetNumericHex are literally just valid hexadecimal printable ASCII chars.
	CharSetNumericHex = CharSetNumeric + "ABCDEF"

	// CharSetSymbolic are literally just valid symbolic printable ASCII chars.
	CharSetSymbolic = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

	// CharSetSymbolicRFC3986Unreserved are RFC3986 unreserved symbol characters.
	// See https://datatracker.ietf.org/doc/html/rfc3986#section-2.3.
	CharSetSymbolicRFC3986Unreserved = "-._~"

	// CharSetAlphaNumeric are literally just valid alphanumeric printable ASCII chars.
	CharSetAlphaNumeric = CharSetAlphabetic + CharSetNumeric

	// CharSetASCII are literally just valid printable ASCII chars.
	CharSetASCII = CharSetAlphabetic + CharSetNumeric + CharSetSymbolic

	// CharSetRFC3986Unreserved are RFC3986 unreserved characters.
	// See https://datatracker.ietf.org/doc/html/rfc3986#section-2.3.
	CharSetRFC3986Unreserved = CharSetAlphabetic + CharSetNumeric + CharSetSymbolicRFC3986Unreserved

	// CharSetUnambiguousUpper are a set of unambiguous uppercase characters.
	CharSetUnambiguousUpper = "ABCDEFGHJKLMNPQRTUVWYXZ2346789"
)
