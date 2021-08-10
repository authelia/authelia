package authentication

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/simia-tech/crypt"

	"github.com/authelia/authelia/v4/internal/utils"
)

// PasswordHash represents all characteristics of a password hash.
// Authelia only supports salted SHA512 or salted argon2id method, i.e., $6$ mode or $argon2id$ mode.
type PasswordHash struct {
	Algorithm   CryptAlgo
	Iterations  int
	Salt        string
	Key         string
	KeyLength   int
	Memory      int
	Parallelism int
}

// ConfigAlgoToCryptoAlgo returns a CryptAlgo and nil error if valid, otherwise it returns argon2id and an error.
func ConfigAlgoToCryptoAlgo(fromConfig string) (CryptAlgo, error) {
	switch fromConfig {
	case argon2id:
		return HashingAlgorithmArgon2id, nil
	case sha512:
		return HashingAlgorithmSHA512, nil
	default:
		return HashingAlgorithmArgon2id, errors.New("Invalid algorithm in configuration. It should be `argon2id` or `sha512`")
	}
}

// ParseHash extracts all characteristics of a hash given its string representation.
func ParseHash(hash string) (passwordHash *PasswordHash, err error) {
	parts := strings.Split(hash, "$")

	// This error can be ignored as it's always nil.
	c, parameters, salt, key, _ := crypt.DecodeSettings(hash)
	code := CryptAlgo(c)
	h := &PasswordHash{}

	h.Salt = salt
	h.Key = key

	if h.Key != parts[len(parts)-1] {
		return nil, fmt.Errorf("Hash key is not the last parameter, the hash is likely malformed (%s)", hash)
	}

	if h.Key == "" {
		return nil, fmt.Errorf("Hash key contains no characters or the field length is invalid (%s)", hash)
	}

	_, err = crypt.Base64Encoding.DecodeString(h.Salt)
	if err != nil {
		return nil, errors.New("Salt contains invalid base64 characters")
	}

	switch code {
	case HashingAlgorithmSHA512:
		h.Iterations = parameters.GetInt("rounds", HashingDefaultSHA512Iterations)
		h.Algorithm = HashingAlgorithmSHA512

		if parameters["rounds"] != "" && parameters["rounds"] != strconv.Itoa(h.Iterations) {
			return nil, fmt.Errorf("SHA512 iterations is not numeric (%s)", parameters["rounds"])
		}
	case HashingAlgorithmArgon2id:
		version := parameters.GetInt("v", 0)
		if version < 19 {
			if version == 0 {
				return nil, fmt.Errorf("Argon2id version parameter not found (%s)", hash)
			}

			return nil, fmt.Errorf("Argon2id versions less than v19 are not supported (hash is version %d)", version)
		} else if version > 19 {
			return nil, fmt.Errorf("Argon2id versions greater than v19 are not supported (hash is version %d)", version)
		}

		h.Algorithm = HashingAlgorithmArgon2id
		h.Memory = parameters.GetInt("m", HashingDefaultArgon2idMemory)
		h.Iterations = parameters.GetInt("t", HashingDefaultArgon2idTime)
		h.Parallelism = parameters.GetInt("p", HashingDefaultArgon2idParallelism)
		h.KeyLength = parameters.GetInt("k", HashingDefaultArgon2idKeyLength)

		decodedKey, err := crypt.Base64Encoding.DecodeString(h.Key)

		if err != nil {
			return nil, errors.New("Hash key contains invalid base64 characters")
		}

		if len(decodedKey) != h.KeyLength {
			return nil, fmt.Errorf("Argon2id key length parameter (%d) does not match the actual key length (%d)", h.KeyLength, len(decodedKey))
		}
	default:
		return nil, fmt.Errorf("Authelia only supports salted SHA512 hashing ($6$) and salted argon2id ($argon2id$), not $%s$", code)
	}

	return h, nil
}

// HashPassword generate a salt and hash the password with the salt and a constant number of rounds.
func HashPassword(password, salt string, algorithm CryptAlgo, iterations, memory, parallelism, keyLength, saltLength int) (hash string, err error) {
	var settings string

	if algorithm != HashingAlgorithmArgon2id && algorithm != HashingAlgorithmSHA512 {
		return "", fmt.Errorf("Hashing algorithm input of '%s' is invalid, only values of %s and %s are supported", algorithm, HashingAlgorithmArgon2id, HashingAlgorithmSHA512)
	}

	if algorithm == HashingAlgorithmArgon2id {
		err := validateArgon2idSettings(memory, parallelism, iterations, keyLength)
		if err != nil {
			return "", err
		}
	}

	err = validateSalt(salt, saltLength)
	if err != nil {
		return "", err
	}

	if salt == "" {
		salt = crypt.Base64Encoding.EncodeToString([]byte(utils.RandomString(saltLength, HashingPossibleSaltCharacters)))
	}

	settings = getCryptSettings(salt, algorithm, iterations, memory, parallelism, keyLength)

	// This error can be ignored because we check for it before a user gets here.
	hash, _ = crypt.Crypt(password, settings)

	return hash, nil
}

// CheckPassword check a password against a hash.
func CheckPassword(password, hash string) (ok bool, err error) {
	expectedHash, err := ParseHash(hash)
	if err != nil {
		return false, err
	}

	passwordHashString, err := HashPassword(password, expectedHash.Salt, expectedHash.Algorithm, expectedHash.Iterations, expectedHash.Memory, expectedHash.Parallelism, expectedHash.KeyLength, len(expectedHash.Salt))
	if err != nil {
		return false, err
	}

	passwordHash, err := ParseHash(passwordHashString)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare([]byte(passwordHash.Key), []byte(expectedHash.Key)) == 1, nil
}

func getCryptSettings(salt string, algorithm CryptAlgo, iterations, memory, parallelism, keyLength int) (settings string) {
	switch algorithm {
	case HashingAlgorithmArgon2id:
		settings, _ = crypt.Argon2idSettings(memory, iterations, parallelism, keyLength, salt)
	case HashingAlgorithmSHA512:
		settings = fmt.Sprintf("$6$rounds=%d$%s", iterations, salt)
	default:
		panic("invalid password hashing algorithm provided")
	}

	return settings
}

// validateSalt checks the salt input and settings are valid and returns it and a nil error if they are, otherwise returns an error.
func validateSalt(salt string, saltLength int) error {
	if salt == "" {
		if saltLength < 8 {
			return fmt.Errorf("Salt length input of %d is invalid, it must be 8 or higher", saltLength)
		}

		return nil
	}

	decodedSalt, err := crypt.Base64Encoding.DecodeString(salt)
	if err != nil {
		return fmt.Errorf("Salt input of %s is invalid, only base64 strings are valid for input", salt)
	}

	if len(decodedSalt) < 8 {
		return fmt.Errorf("Salt input of %s is invalid (%d characters), it must be 8 or more characters", decodedSalt, len(decodedSalt))
	}

	return nil
}

// validateArgon2idSettings checks the argon2id settings are valid.
func validateArgon2idSettings(memory, parallelism, iterations, keyLength int) error {
	// Caution: Increasing any of the values in the below block has a high chance in old passwords that cannot be verified.
	if memory < 8 {
		return fmt.Errorf("Memory (argon2id) input of %d is invalid, it must be 8 or higher", memory)
	}

	if parallelism < 1 {
		return fmt.Errorf("Parallelism (argon2id) input of %d is invalid, it must be 1 or higher", parallelism)
	}

	if memory < parallelism*8 {
		return fmt.Errorf("Memory (argon2id) input of %d is invalid with a parallelism input of %d, it must be %d (parallelism * 8) or higher", memory, parallelism, parallelism*8)
	}

	if keyLength < 16 {
		return fmt.Errorf("Key length (argon2id) input of %d is invalid, it must be 16 or higher", keyLength)
	}

	if iterations < 1 {
		return fmt.Errorf("Iterations (argon2id) input of %d is invalid, it must be 1 or more", iterations)
	}

	// Caution: Increasing any of the values in the above block has a high chance in old passwords that cannot be verified.
	return nil
}
