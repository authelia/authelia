package authentication

// Level is the type representing a level of authentication
type Level int

const (
	// NotAuthenticated if the user is not authenticated yet.
	NotAuthenticated Level = iota
	// OneFactor if the user has passed first factor only.
	OneFactor Level = iota
	// TwoFactor if the user has passed two factors.
	TwoFactor Level = iota
)

const (
	// TOTP Method using Time-Based One-Time Password applications like Google Authenticator
	TOTP = "totp"
	// U2F Method using U2F devices like Yubikeys
	U2F = "u2f"
	// Push Method using Duo application to receive push notifications.
	Push = "mobile_push"
)

// PossibleMethods is the set of all possible 2FA methods.
var PossibleMethods = []string{TOTP, U2F, Push}

const (
	//Argon2id Hash Identifier
	HashingAlgorithmArgon2id = "argon2id"
	//SHA512 Hash Identifier
	HashingAlgorithmSHA512 = "6"
)

// These are the default values from the upstream crypt module, we use them to for GetInt, and they need to be checked when updating  github.com/simia-tech/crypt
const (
	HashingDefaultArgon2idTime        = 1
	HashingDefaultArgon2idMemory      = 32 * 1024
	HashingDefaultArgon2idParallelism = 4
	HashingDefaultArgon2idKeyLength   = 32
	HashingDefaultSHA512Iterations    = 5000
)

// Valid Hashing runes
var HashingPossibleSaltCharacters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/")
