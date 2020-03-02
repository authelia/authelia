package authentication

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/simia-tech/crypt"
)

// PasswordHash represents all characteristics of a password hash.
// Authelia only supports salted SHA512 or salted argon2id method, i.e., $6$ mode or $argon2id$ mode.
type PasswordHash struct {
	Type        string
	Rounds      int
	Salt        string
	Key         string
	Memory      int
	Parallelism int
}

var defaultPasswordType = Argon2id
var defaultPasswordArgon2idRounds = 3
var defaultPasswordArgon2idMemory = 64 * 1024
var defaultPasswordArgon2idParallelism = 2
var defaultPasswordSHA512Rounds = 5000
var defaultPasswordSaltLength = 16

// ParseHash extracts all characteristics of a hash given its string representation.
func ParseHash(hash string) (*PasswordHash, error) {
	parts := strings.Split(hash, "$")
	h := &PasswordHash{}

	if parts[1] == SHA512 {
		if len(parts) != 5 {
			return nil, fmt.Errorf("Cannot parse the SHA512 hash %s", hash)
		}
		_, err := fmt.Sscanf(parts[2], "rounds=%d", &h.Rounds)
		if err != nil {
			return nil, errors.New("Cannot match pattern 'rounds=<int>' to find the number of rounds")
		}
		h.Salt = parts[3]
		h.Key = parts[4]
		h.Type = SHA512
	} else if parts[1] == Argon2id {
		if len(parts) != 6 {
			return nil, fmt.Errorf("Cannot parse the Argon2id hash %s", hash)
		}

		_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &h.Memory, &h.Rounds, &h.Parallelism)
		if err != nil {
			return nil, errors.New("Cannot match pattern 'm=<int>,t=<int>,p=<int>' to find the argon2id params")
		}
		h.Salt = parts[4]
		h.Key = parts[5]
		h.Type = Argon2id
	} else {
		return nil, fmt.Errorf("Authelia only supports salted SHA512 hashing ($6$) and salted argon2id ($argon2id$), not $%s$", parts[1])
	}
	return h, nil
}

// The set of letters RandomString can pick in.
var possibleLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandomString generate a random string of n characters.
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = possibleLetters[rand.Intn(len(possibleLetters))]
	}
	return string(b)
}

// HashPassword generate a salt and hash the password with the salt and a constant
// number of rounds.
func HashPassword(password string, salt string) string {
	if salt == "" {
		if defaultPasswordType == Argon2id {
			salt, _ = crypt.Argon2idSettings(defaultPasswordArgon2idMemory, defaultPasswordArgon2idRounds, defaultPasswordArgon2idParallelism, RandomString(defaultPasswordSaltLength))
		} else if defaultPasswordType == SHA512 {
			salt = fmt.Sprintf("$6$rounds=%d$%s", defaultPasswordSHA512Rounds, RandomString(defaultPasswordSaltLength))
		}
	}
	hash, err := crypt.Crypt(password, salt)
	if err != nil {
		log.Fatal(err)
	}
	return hash
}

// CheckPassword check a password against a hash.
func CheckPassword(password string, hash string) (bool, error) {
	passwordHash, err := ParseHash(hash)
	if err != nil {
		return false, err
	}
	var salt string
	if passwordHash.Type == Argon2id {
		salt, err = crypt.Argon2idSettings(passwordHash.Memory, passwordHash.Rounds, passwordHash.Parallelism, passwordHash.Salt)
		if err != nil {
			return false, err
		}
	} else if passwordHash.Type == SHA512 {
		salt = fmt.Sprintf("$6$rounds=%d$%s$", passwordHash.Rounds, passwordHash.Salt)
	}
	pHash := HashPassword(password, salt)
	return pHash == hash, nil
}
