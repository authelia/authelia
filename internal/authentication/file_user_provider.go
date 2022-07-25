package authentication

import (
	_ "embed" // Embed users_database.template.yml.
	"fmt"
	"os"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/go-crypt/crypt"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// FileUserProvider is a provider reading details from a file.
type FileUserProvider struct {
	config   *schema.FileAuthenticationBackendConfig
	hash     crypt.Hash
	database *DatabaseModel
	lock     *sync.Mutex
}

// NewFileUserProvider creates a new instance of FileUserProvider.
func NewFileUserProvider(config *schema.FileAuthenticationBackendConfig) *FileUserProvider {
	logger := logging.Logger()

	errs := checkDatabase(config.Path)
	if errs != nil {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(1)
	}

	databaseYAML, err := readDatabase(config.Path)
	if err != nil {
		// Panic since the file does not exist when Authelia is starting.
		panic(err)
	}

	var database *DatabaseModel

	if database, err = databaseYAML.ToDatabaseModel(); err != nil {
		panic(err)
	}

	var hash crypt.Hash

	fmt.Printf("%+v\n", config.Password)

	switch config.Password.Algorithm {
	case "sha2crypt", "sha512":
		hash = crypt.NewSHA2CryptHash().
			WithRounds(config.Password.SHA2Crypt.Rounds).
			WithSaltLength(config.Password.SHA2Crypt.SaltLength)
	case "argon2", "argon2id":
		hash = crypt.NewArgon2Hash().
			WithVariant(crypt.NewArgon2Variant(config.Password.Argon2.Variant)).
			WithT(config.Password.Argon2.Iterations).
			WithM(config.Password.Argon2.Memory).
			WithP(config.Password.Argon2.Parallelism).
			WithK(config.Password.Argon2.KeyLength).
			WithS(config.Password.Argon2.SaltLength)
	case "pbkdf2":
		hash = crypt.NewPBKDF2Hash().
			WithVariant(crypt.NewPBKDF2Variant(config.Password.PBKDF2.Variant)).
			WithIterations(config.Password.PBKDF2.Iterations).
			WithKeyLength(config.Password.PBKDF2.KeyLength).
			WithSaltLength(config.Password.PBKDF2.SaltLength)
	case "scrypt":
		hash = crypt.NewScryptHash().
			WithLN(config.Password.SCrypt.Rounds).
			WithP(config.Password.SCrypt.Parallelism).
			WithR(config.Password.SCrypt.BlockSize)
	case "bcrypt":
		hash = crypt.NewBcryptHash().
			WithVariant(crypt.NewBcryptVariant(config.Password.BCrypt.Variant)).
			WithCost(config.Password.BCrypt.Rounds)
	}

	return &FileUserProvider{
		config:   config,
		hash:     hash,
		database: database,
		lock:     &sync.Mutex{},
	}
}

func checkDatabase(path string) []error {
	_, err := os.Stat(path)
	if err != nil {
		errs := []error{
			fmt.Errorf("Unable to find database file: %v", path),
			fmt.Errorf("Generating database file: %v", path),
		}

		err := generateDatabaseFromTemplate(path)
		if err != nil {
			errs = append(errs, err)
		} else {
			errs = append(errs, fmt.Errorf("Generated database at: %v", path))
		}

		return errs
	}

	return nil
}

//go:embed users_database.template.yml
var cfg []byte

func generateDatabaseFromTemplate(path string) error {
	err := os.WriteFile(path, cfg, 0600)
	if err != nil {
		return fmt.Errorf("Unable to generate %v: %v", path, err)
	}

	return nil
}

func readDatabase(path string) (*YAMLDatabaseModel, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read database from file %s: %s", path, err)
	}

	db := YAMLDatabaseModel{}

	err = yaml.Unmarshal(content, &db)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse database: %s", err)
	}

	ok, err := govalidator.ValidateStruct(db)
	if err != nil {
		return nil, fmt.Errorf("Invalid schema of database: %s", err)
	}

	if !ok {
		return nil, fmt.Errorf("The database format is invalid: %s", err)
	}

	return &db, nil
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *FileUserProvider) CheckUserPassword(username string, password string) (match bool, err error) {
	if details, ok := p.database.Users[username]; ok {
		return details.Digest.MatchAdvanced(password)
	}

	return false, ErrUserNotFound
}

// GetDetails retrieve the groups a user belongs to.
func (p *FileUserProvider) GetDetails(username string) (*UserDetails, error) {
	if details, ok := p.database.Users[username]; ok {
		return &UserDetails{
			Username:    username,
			DisplayName: details.DisplayName,
			Groups:      details.Groups,
			Emails:      []string{details.Email},
		}, nil
	}

	return nil, fmt.Errorf("User '%s' does not exist in database", username)
}

// UpdatePassword update the password of the given user.
func (p *FileUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	details, ok := p.database.Users[username]
	if !ok {
		return ErrUserNotFound
	}

	if details.Digest, err = p.hash.Hash(newPassword); err != nil {
		return err
	}

	p.lock.Lock()
	p.database.Users[username] = details

	b, err := yaml.Marshal(p.database.ToYAMLDatabaseModel())
	if err != nil {
		p.lock.Unlock()
		return err
	}

	err = os.WriteFile(p.config.Path, b, fileAuthenticationMode)
	p.lock.Unlock()

	return err
}

// StartupCheck implements the startup check provider interface.
func (p *FileUserProvider) StartupCheck() (err error) {
	return nil
}
