package authentication

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/simia-tech/crypt"
	"gopkg.in/yaml.v2"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// FileUserProvider is a provider reading details from a file.
type FileUserProvider struct {
	configuration *schema.FileAuthenticationBackendConfiguration
	database      *DatabaseModel
	lock          *sync.Mutex

	// TODO: Remove this. This is only here to temporarily fix the username enumeration security flaw in #949.
	fakeHash string
}

// UserDetailsModel is the model of user details in the file database.
type UserDetailsModel struct {
	HashedPassword string   `yaml:"password" valid:"required"`
	DisplayName    string   `yaml:"displayname" valid:"required"`
	Email          string   `yaml:"email"`
	Groups         []string `yaml:"groups"`
}

// DatabaseModel is the model of users file database.
type DatabaseModel struct {
	Users map[string]UserDetailsModel `yaml:"users" valid:"required"`
}

// NewFileUserProvider creates a new instance of FileUserProvider.
func NewFileUserProvider(configuration *schema.FileAuthenticationBackendConfiguration) *FileUserProvider {
	errs := checkDatabase(configuration.Path)
	if errs != nil {
		for _, err := range errs {
			logging.Logger().Error(err)
		}

		os.Exit(1)
	}

	database, err := readDatabase(configuration.Path)
	if err != nil {
		// Panic since the file does not exist when Authelia is starting.
		panic(err)
	}

	// Early check whether hashed passwords are correct for all users
	err = checkPasswordHashes(database)
	if err != nil {
		panic(err)
	}

	var cryptAlgo CryptAlgo = HashingAlgorithmArgon2id
	// TODO: Remove this. This is only here to temporarily fix the username enumeration security flaw in #949.
	// This generates a hash that should be usable to do a fake CheckUserPassword
	if configuration.Password.Algorithm == sha512 {
		cryptAlgo = HashingAlgorithmSHA512
	}

	settings := getCryptSettings(utils.RandomString(configuration.Password.SaltLength, HashingPossibleSaltCharacters),
		cryptAlgo, configuration.Password.Iterations, configuration.Password.Memory*1024, configuration.Password.Parallelism,
		configuration.Password.KeyLength)
	data := crypt.Base64Encoding.EncodeToString([]byte(utils.RandomString(configuration.Password.KeyLength, HashingPossibleSaltCharacters)))
	fakeHash := fmt.Sprintf("%s$%s", settings, data)

	return &FileUserProvider{
		configuration: configuration,
		database:      database,
		lock:          &sync.Mutex{},
		fakeHash:      fakeHash,
	}
}

func checkPasswordHashes(database *DatabaseModel) error {
	for u, v := range database.Users {
		v.HashedPassword = strings.ReplaceAll(v.HashedPassword, "{CRYPT}", "")
		_, err := ParseHash(v.HashedPassword)

		if err != nil {
			return fmt.Errorf("Unable to parse hash of user %s: %s", u, err)
		}

		database.Users[u] = v
	}

	return nil
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

func generateDatabaseFromTemplate(path string) error {
	f, err := cfg.Open("users_database.template.yml")
	if err != nil {
		return fmt.Errorf("Unable to open users_database.template.yml: %v", err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("Unable to read users_database.template.yml: %v", err)
	}

	err = ioutil.WriteFile(path, b, 0600)
	if err != nil {
		return fmt.Errorf("Unable to generate %v: %v", path, err)
	}

	return nil
}

func readDatabase(path string) (*DatabaseModel, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read database from file %s: %s", path, err)
	}

	db := DatabaseModel{}

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
func (p *FileUserProvider) CheckUserPassword(username string, password string) (bool, error) {
	if details, ok := p.database.Users[username]; ok {
		ok, err := CheckPassword(password, details.HashedPassword)
		if err != nil {
			return false, err
		}

		return ok, nil
	}

	// TODO: Remove this. This is only here to temporarily fix the username enumeration security flaw in #949.
	_, _ = CheckPassword(password, p.fakeHash)

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
func (p *FileUserProvider) UpdatePassword(username string, newPassword string) error {
	details, ok := p.database.Users[username]
	if !ok {
		return ErrUserNotFound
	}

	algorithm, err := ConfigAlgoToCryptoAlgo(p.configuration.Password.Algorithm)
	if err != nil {
		return err
	}

	hash, err := HashPassword(
		newPassword, "", algorithm, p.configuration.Password.Iterations,
		p.configuration.Password.Memory*1024, p.configuration.Password.Parallelism,
		p.configuration.Password.KeyLength, p.configuration.Password.SaltLength)

	if err != nil {
		return err
	}

	details.HashedPassword = hash

	p.lock.Lock()
	p.database.Users[username] = details

	b, err := yaml.Marshal(p.database)
	if err != nil {
		p.lock.Unlock()
		return err
	}

	err = ioutil.WriteFile(p.configuration.Path, b, fileAuthenticationMode)
	p.lock.Unlock()

	return err
}
