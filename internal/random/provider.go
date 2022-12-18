package random

// Provider of random functions and functionality.
type Provider interface {
	// Generate returns random data as bytes with the standard random.DefaultN length and can contain any byte values
	// (including unreadable byte values).
	Generate() []byte

	// GenerateCustom returns random data as bytes with n length and can contain only byte values from the provided
	// values. If n is less than 1 then DefaultN is used instead.
	GenerateCustom(n int, charset []byte) (data []byte)

	// GenerateString is an overload of GenerateCustom which takes a characters string and returns a string.
	GenerateString(n int, characters string) (data string)
}
