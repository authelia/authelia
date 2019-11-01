package authentication

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/simia-tech/crypt"
)

// PasswordHash represents all characteristics of a password hash.
// Authelia only supports salted SHA512 method, i.e., $6$ mode.
type PasswordHash struct {
	// The number of rounds.
	Rounds int
	// The salt with a max size of 16 characters for SHA512.
	Salt string
	// The password hash.
	Hash string
}

// passwordHashFromString extracts all characteristics of a hash given its string representation.
func passwordHashFromString(hash string) (*PasswordHash, error) {
	// Only supports salted sha 512.
	if hash[:3] != "$6$" {
		return nil, errors.New("Authelia only supports salted SHA512 hashing")
	}
	parts := strings.Split(hash, "$")

	if len(parts) != 5 {
		return nil, errors.New("Cannot parse the hash")
	}

	roundsKV := strings.Split(parts[2], "=")
	if len(roundsKV) != 2 {
		return nil, errors.New("Cannot find the number of rounds")
	}

	rounds, err := strconv.ParseInt(roundsKV[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("Cannot find the number of rounds in the hash: %s", err.Error())
	}

	return &PasswordHash{
		Rounds: int(rounds),
		Salt:   parts[3],
		Hash:   parts[4],
	}, nil
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
		salt = fmt.Sprintf("$6$rounds=50000$%s", RandomString(16))
	}
	hash, err := crypt.Crypt(password, salt)
	if err != nil {
		log.Fatal(err)
	}
	return hash
}

// CheckPassword check a password against a hash.
func CheckPassword(password string, hash string) (bool, error) {
	passwordHash, err := passwordHashFromString(hash)
	if err != nil {
		return false, err
	}
	salt := fmt.Sprintf("$6$rounds=%d$%s$", passwordHash.Rounds, passwordHash.Salt)
	pHash := HashPassword(password, salt)
	return pHash == hash, nil
}
