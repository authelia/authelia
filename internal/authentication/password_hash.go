package authentication

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/simia-tech/crypt"
)

// PasswordHash represents all characteristics of a password hash.
// Authelia only supports salted SHA512 or salted argon2id method, i.e., $6$ mode or $argon2id$ mode.

type PasswordHash struct {
	Algorithm   string
	Rounds      int
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
		_, err := fmt.Sscanf(parts[2], "rounds=%d", &h.Rounds)
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
			return nil, fmt.Errorf("Argon2 versions less than v19 are not supported (hash is version %d)", version)
		} else if version > 19 {
			return nil, fmt.Errorf("Argon2 versions greater than v19 are not supported (hash is version %d)", version)
		}

		_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &h.Memory, &h.Rounds, &h.Parallelism)
		if err != nil {
			return nil, fmt.Errorf("Cannot match pattern 'm=<int>,t=<int>,p=<int>' to find the argon2id params. Cause: %s", err)
		}
		h.Salt = parts[4]
		h.Key = parts[5]
		h.Algorithm = HashingAlgorithmArgon2id
	} else {
		return nil, fmt.Errorf("Authelia only supports salted SHA512 hashing ($6$) and salted argon2id ($argon2id$), not $%s$", parts[1])
	}
	return h, nil
}

// RandomString generate a random string of n characters.
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = HashingPossibleSaltCharacters[rand.Intn(len(HashingPossibleSaltCharacters))]
	}
	return string(b)
}

// HashPassword generate a salt and hash the password with the salt and a constant
// number of rounds.
func HashPassword(password, salt, algorithm string, rounds, memory, parallelism, saltLength int) (string, error) {
	var settings string

	if algorithm != HashingAlgorithmArgon2id && algorithm != HashingAlgorithmSHA512 {
		return "", fmt.Errorf("Hashing Algorithm '%s' is Invalid (only support values of %s and %s).", algorithm, HashingAlgorithmArgon2id, HashingAlgorithmSHA512)
	}

	if algorithm == HashingAlgorithmArgon2id {
		if memory < 8 {
			return "", fmt.Errorf("Memory for argon2id must be above 8, you set it to %d.", memory)
		}
		if parallelism < 1 {
			return "", fmt.Errorf("Parallelism for argon2id must be above 0, you set it to %d.", parallelism)
		}
		if salt == "" && saltLength < 1 {
			return "", fmt.Errorf("Salt length is  %d.", parallelism)
		}
		if memory < parallelism*8 {
			return "", fmt.Errorf("Memory for argon2id must be above %d (parallelism * 8), you set memory to %d and parallelism to %d.", parallelism*8, memory, parallelism)
		}
	}

	if algorithm == HashingAlgorithmArgon2id {
		if salt != "" {
			settings, _ = crypt.Argon2idSettings(memory, rounds, parallelism, salt)
		} else {
			settings, _ = crypt.Argon2idSettings(memory, rounds, parallelism)
		}
	} else if algorithm == HashingAlgorithmSHA512 {
		if salt != "" {
			settings = fmt.Sprintf("$6$rounds=%d$%s", rounds, salt)
		} else {
			settings = fmt.Sprintf("$6$rounds=%d$%s", rounds, RandomString(saltLength))
		}
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
	expectedHash, err := HashPassword(password, passwordHash.Salt, passwordHash.Algorithm, passwordHash.Rounds, passwordHash.Memory, passwordHash.Parallelism, len(passwordHash.Salt))
	if err != nil {
		return false, err
	}
	return hash == expectedHash, nil
}
