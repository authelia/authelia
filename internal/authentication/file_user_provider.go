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
	config   *schema.FileAuthenticationBackendConfig
	hash     crypt.Hash
	database *FileUserDatabase
}

// NewFileUserProvider creates a new instance of FileUserProvider.
func NewFileUserProvider(config *schema.FileAuthenticationBackendConfig) (provider *FileUserProvider) {
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
	var d *FileUserDatabaseUser

	if d, err = p.database.GetUserDetails(username); err != nil {
		return nil, err
	}

	return &UserDetails{
		Username:    username,
		DisplayName: d.DisplayName,
		Groups:      d.Groups,
		Emails:      []string{d.Email},
	}, nil
}

// UpdatePassword update the password of the given user.
func (p *FileUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	var details *FileUserDatabaseUser

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	if details.Digest, err = p.hash.Hash(newPassword); err != nil {
		return err
	}

	p.database.SetUserDetails(username, details)

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// StartupCheck implements the startup check provider interface.
func (p *FileUserProvider) StartupCheck() (err error) {
	switch p.config.Password.Algorithm {
	case "argon2", "argon2id", "":
		p.hash = crypt.NewArgon2Hash().
			WithVariant(crypt.NewArgon2Variant(p.config.Password.Argon2.Variant)).
			WithT(p.config.Password.Argon2.Iterations).
			WithM(p.config.Password.Argon2.Memory).
			WithP(p.config.Password.Argon2.Parallelism).
			WithK(p.config.Password.Argon2.KeyLength).
			WithS(p.config.Password.Argon2.SaltLength)
	case "sha2crypt", "sha512":
		p.hash = crypt.NewSHA2CryptHash().
			WithVariant(crypt.NewSHA2CryptVariant(p.config.Password.SHA2Crypt.Variant)).
			WithRounds(p.config.Password.SHA2Crypt.Rounds).
			WithSaltLength(p.config.Password.SHA2Crypt.SaltLength)
	case "pbkdf2":
		p.hash = crypt.NewPBKDF2Hash().
			WithVariant(crypt.NewPBKDF2Variant(p.config.Password.PBKDF2.Variant)).
			WithIterations(p.config.Password.PBKDF2.Iterations).
			WithKeyLength(p.config.Password.PBKDF2.KeyLength).
			WithSaltLength(p.config.Password.PBKDF2.SaltLength)
	case "scrypt":
		p.hash = crypt.NewScryptHash().
			WithLN(p.config.Password.SCrypt.Rounds).
			WithP(p.config.Password.SCrypt.Parallelism).
			WithR(p.config.Password.SCrypt.BlockSize)
	case "bcrypt":
		p.hash = crypt.NewBcryptHash().
			WithVariant(crypt.NewBcryptVariant(p.config.Password.BCrypt.Variant)).
			WithCost(p.config.Password.BCrypt.Cost)
	default:
		return fmt.Errorf("algorithm '%s' is unknown", p.config.Password.Algorithm)
	}

	logger := logging.Logger()

	if errs := checkDatabase(p.config.Path); errs != nil {
		for _, err = range errs {
			logger.Error(err)
		}

		return fmt.Errorf("one or more errors occurred checking the authentication database")
	}

	p.database = NewFileUserDatabase(p.config.Path)

	if err = p.database.Load(); err != nil {
		return err
	}

	return nil
}

func checkDatabase(path string) []error {
	var err error

	if _, err = os.Stat(path); err != nil {
		errs := []error{
			fmt.Errorf("Unable to find database file: %v", path),
			fmt.Errorf("Generating database file: %v", path),
		}

		if err = generateDatabaseFromTemplate(path); err != nil {
			errs = append(errs, err)
		} else {
			errs = append(errs, fmt.Errorf("Generated database at: %v", path))
		}

		return errs
	}

	return nil
}

//go:embed users_database.template.yml
var userYAMLTemplate []byte

func generateDatabaseFromTemplate(path string) error {
	err := os.WriteFile(path, userYAMLTemplate, 0600)
	if err != nil {
		return fmt.Errorf("Unable to generate %v: %v", path, err)
	}

	return nil
}
