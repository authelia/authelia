package authentication

import (
	_ "embed" // Embed users_database.template.yml.
	"fmt"
	"os"

	"github.com/go-crypt/crypt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// FileUserProvider is a provider reading details from a file.
type FileUserProvider struct {
	config   *schema.FileAuthenticationBackend
	hash     crypt.Hash
	database *FileUserDatabase
}

// NewFileUserProvider creates a new instance of FileUserProvider.
func NewFileUserProvider(config *schema.FileAuthenticationBackend) (provider *FileUserProvider) {
	return &FileUserProvider{
		config:   config,
		database: NewFileUserDatabase(config.Path),
	}
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *FileUserProvider) CheckUserPassword(username string, password string) (match bool, err error) {
	if details, ok := p.database.Users[username]; ok {
		return details.Digest.MatchAdvanced(password)
	}

	return false, ErrUserNotFound
}

// GetDetails retrieve the groups a user belongs to.
func (p *FileUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var d DatabaseUserDetails

	if d, err = p.database.GetUserDetails(username); err != nil {
		return nil, err
	}

	return d.ToUserDetails(username), nil
}

// UpdatePassword update the password of the given user.
func (p *FileUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	var details DatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	if details.Digest, err = p.hash.Hash(newPassword); err != nil {
		return err
	}

	p.database.SetUserDetails(username, &details)

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// StartupCheck implements the startup check provider interface.
func (p *FileUserProvider) StartupCheck() (err error) {
	if err = checkDatabase(p.config.Path); err != nil {
		logging.Logger().WithError(err).Errorf("Error checking user authentication YAML database")

		return fmt.Errorf("one or more errors occurred checking the authentication database")
	}

	if p.hash, err = NewFileCryptoHashFromConfig(p.config); err != nil {
		return err
	}

	p.database = NewFileUserDatabase(p.config.Path)

	if err = p.database.Load(); err != nil {
		return err
	}

	return nil
}

// NewFileCryptoHashFromConfig returns a crypt.Hash given a valid configuration.
func NewFileCryptoHashFromConfig(config *schema.FileAuthenticationBackend) (hash crypt.Hash, err error) {
	switch config.Password.Algorithm {
	case hashArgon2, "":
		hash = crypt.NewArgon2Hash().
			WithVariant(crypt.NewArgon2Variant(config.Password.Argon2.Variant)).
			WithT(config.Password.Argon2.Iterations).
			WithM(config.Password.Argon2.Memory).
			WithP(config.Password.Argon2.Parallelism).
			WithK(config.Password.Argon2.KeyLength).
			WithS(config.Password.Argon2.SaltLength)
	case hashSHA2Crypt:
		hash = crypt.NewSHA2CryptHash().
			WithVariant(crypt.NewSHA2CryptVariant(config.Password.SHA2Crypt.Variant)).
			WithRounds(config.Password.SHA2Crypt.Iterations).
			WithSaltLength(config.Password.SHA2Crypt.SaltLength)
	case hashPBKDF2:
		hash = crypt.NewPBKDF2Hash().
			WithVariant(crypt.NewPBKDF2Variant(config.Password.PBKDF2.Variant)).
			WithIterations(config.Password.PBKDF2.Iterations).
			WithKeyLength(config.Password.PBKDF2.KeyLength).
			WithSaltLength(config.Password.PBKDF2.SaltLength)
	case hashSCrypt:
		hash = crypt.NewScryptHash().
			WithLN(config.Password.SCrypt.Iterations).
			WithP(config.Password.SCrypt.Parallelism).
			WithR(config.Password.SCrypt.BlockSize)
	case hashBCrypt:
		hash = crypt.NewBcryptHash().
			WithVariant(crypt.NewBcryptVariant(config.Password.BCrypt.Variant)).
			WithCost(config.Password.BCrypt.Cost)
	default:
		return nil, fmt.Errorf("algorithm '%s' is unknown", config.Password.Algorithm)
	}

	if err = hash.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate hash settings: %w", err)
	}

	return hash, nil
}

func checkDatabase(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.WriteFile(path, userYAMLTemplate, 0600); err != nil {
			return fmt.Errorf("user authentication database file doesn't exist at path '%s' and could not be generated: %w", path, err)
		}

		return fmt.Errorf("user authentication database file doesn't exist at path '%s' and has been generated", path)
	} else if err != nil {
		return fmt.Errorf("error checking user authentication database file: %w", err)
	}

	return nil
}

//go:embed users_database.template.yml
var userYAMLTemplate []byte
