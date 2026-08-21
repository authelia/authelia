package schema

// PBKDF2VariantDefaultIterations returns the default number of iterations for the given PBKDF2 variant.
func PBKDF2VariantDefaultIterations(variant string) int {
	switch variant {
	case SHA512Lower, "":
		return defaultIterationsPBKDF2SHA512
	case SHA384Lower:
		return defaultIterationsPBKDF2SHA384
	case SHA256Lower:
		return defaultIterationsPBKDF2SHA256
	case SHA224Lower:
		return defaultIterationsPBKDF2SHA224
	default:
		return defaultIterationsPBKDF2SHA1
	}
}
