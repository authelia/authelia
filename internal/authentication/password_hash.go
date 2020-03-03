package authentication

import (
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/authelia/authelia/internal/utils"
	"github.com/simia-tech/crypt"
)

// PasswordHash represents all characteristics of a password hash.
// Authelia only supports salted SHA512 or salted argon2id method, i.e., $6$ mode or $argon2id$ mode.
type PasswordHash struct {
	Algorithm   string
	Iterations  int
	Salt        string
	Key         string
	Memory      int
	Parallelism int
}

// ParseHash extracts all characteristics of a hash given its string representation.
func ParseHash(hash string) (*PasswordHash, error) {
	parts := strings.Split(hash, "$")
	h := &PasswordHash{}

	if parts[1] == HashingAlgorithmSHA512 {
		if len(parts) != 5 {
			return nil, fmt.Errorf("Cannot parse the SHA512 hash %s", hash)
		}
		_, err := fmt.Sscanf(parts[2], "rounds=%d", &h.Iterations)
		if err != nil {
			return nil, fmt.Errorf("Cannot match pattern 'rounds=<int>' to find the number of rounds. Cause: %s", err)
		}
		h.Salt = parts[3]
		h.Key = parts[4]
		h.Algorithm = HashingAlgorithmSHA512
	} else if parts[1] == HashingAlgorithmArgon2id {
		if len(parts) != 6 {
			return nil, fmt.Errorf("Cannot parse the Argon2id hash %s", hash)
		}

		var version int
		_, err := fmt.Sscanf(parts[2], "v=%d", &version)
		if version < 19 {
			return nil, fmt.Errorf("Argon2id versions less than v19 are not supported (hash is version %d).", version)
		} else if version > 19 {
			return nil, fmt.Errorf("Argon2id versions greater than v19 are not supported (hash is version %d).", version)
		}

		_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &h.Memory, &h.Iterations, &h.Parallelism)
		if err != nil {
			return nil, fmt.Errorf("Cannot match pattern 'm=<int>,t=<int>,p=<int>' to find the argon2id params. Cause: %s", err)
		}
		h.Salt = parts[4]
		h.Key = parts[5]
		h.Algorithm = HashingAlgorithmArgon2id
		if !utils.IsStringBase64Valid(h.Key) {
			return nil, fmt.Errorf("Cannot parse hash argon2id key contains invalid base64 characters.")
		}
		if !utils.IsStringBase64Valid(h.Salt) {
			return nil, fmt.Errorf("Cannot parse hash argon2id salt contains invalid base64 characters.")
		}
	} else {
		return nil, fmt.Errorf("Authelia only supports salted SHA512 hashing ($6$) and salted argon2id ($argon2id$), not $%s$", parts[1])
	}
	return h, nil
}

// RandomString generate a random string of n characters.
func RandomString(n int) string {
	prime, err := cryptorand.Prime(cryptorand.Reader, 1024)
	if err != nil {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(prime.Int64())
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = HashingPossibleSaltCharacters[rand.Intn(len(HashingPossibleSaltCharacters))]
	}
	return string(b)
}

// HashPassword generate a salt and hash the password with the salt and a constant
// number of rounds.
func HashPassword(password, salt, algorithm string, iterations, memory, parallelism, saltLength int) (string, error) {
	var settings string

	if algorithm != HashingAlgorithmArgon2id && algorithm != HashingAlgorithmSHA512 {
		return "", fmt.Errorf("Hashing algorithm input of '%s' is invalid, only values of %s and %s are supported.", algorithm, HashingAlgorithmArgon2id, HashingAlgorithmSHA512)
	}

	if salt == "" {
		if saltLength < 1 {
			return "", fmt.Errorf("Salt length input of %d is invalid, it must be 1 or higher.", saltLength)
		} else if saltLength > 16 {
			return "", fmt.Errorf("Salt length input of %d is invalid, it must be 16 or lower.", saltLength)
		}
	} else if len(salt) > 16 {
		return "", fmt.Errorf("Salt input of %s is invalid (%d characters), it must be 16 or fewer characters.", salt, len(salt))
	} else if !utils.IsStringBase64Valid(salt) {
		return "", fmt.Errorf("Salt input of %s is invalid, only characters [a-zA-Z0-9+/] are valid for input.", salt)
	}
	if algorithm == HashingAlgorithmArgon2id {
		if memory < 8 {
			return "", fmt.Errorf("Memory (argon2id) input of %d is invalid, it must be 8 or higher.", memory)
		}
		if parallelism < 1 {
			return "", fmt.Errorf("Parallelism (argon2id) input of %d is invalid, it must be 1 or higher.", parallelism)
		}
		if memory < parallelism*8 {
			return "", fmt.Errorf("Memory (argon2id) input of %d is invalid with a paraellelism input of %d, it must be %d (parallelism * 8) or higher.", memory, parallelism, parallelism*8)
		}
	}

	if salt == "" {
		salt = RandomString(saltLength)
	}
	if algorithm == HashingAlgorithmArgon2id {
		settings, _ = crypt.Argon2idSettings(memory, iterations, parallelism, salt)
	} else if algorithm == HashingAlgorithmSHA512 {
		settings = fmt.Sprintf("$6$rounds=%d$%s", iterations, salt)
	}

	hash, err := crypt.Crypt(password, settings)
	if err != nil {
		log.Fatal(err)
	}
	return hash, nil
}

// CheckPassword check a password against a hash.
func CheckPassword(password, hash string) (bool, error) {
	passwordHash, err := ParseHash(hash)
	if err != nil {
		return false, err
	}
	expectedHash, err := HashPassword(password, passwordHash.Salt, passwordHash.Algorithm, passwordHash.Iterations, passwordHash.Memory, passwordHash.Parallelism, len(passwordHash.Salt))
	if err != nil {
		return false, err
	}
	return hash == expectedHash, nil
}
