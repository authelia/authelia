package authentication

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	"gopkg.in/yaml.v2"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// FileUserProvider is a provider reading details from a file.
type FileUserProvider struct {
	configuration *schema.FileAuthenticationBackendConfiguration
	database      *DatabaseModel
	lock          *sync.Mutex
}

// UserDetailsModel is the model of user details in the file database.
type UserDetailsModel struct {
	HashedPassword string   `yaml:"password" valid:"required"`
	Email          string   `yaml:"email"`
	Groups         []string `yaml:"groups"`
}

// DatabaseModel is the model of users file database.
type DatabaseModel struct {
	Users map[string]UserDetailsModel `yaml:"users" valid:"required"`
}

// NewFileUserProvider creates a new instance of FileUserProvider.
func NewFileUserProvider(configuration *schema.FileAuthenticationBackendConfiguration) *FileUserProvider {
	database, err := readDatabase(configuration.Path)
	if err != nil {
		// Panic since the file does not exist when Authelia is starting.
		panic(err.Error())
	}

	// Early check whether hashed passwords are correct for all users
	err = checkPasswordHashes(database)
	if err != nil {
		panic(err.Error())
	}

	return &FileUserProvider{
		configuration: configuration,
		database:      database,
		lock:          &sync.Mutex{},
	}
}

func checkPasswordHashes(database *DatabaseModel) error {
	for u, v := range database.Users {
		_, err := ParseHash(v.HashedPassword)
		if err != nil {
			return fmt.Errorf("Unable to parse hash of user %s: %s", u, err)
		}
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
		hashedPassword := strings.ReplaceAll(details.HashedPassword, "{CRYPT}", "")
		ok, err := CheckPassword(password, hashedPassword)
		if err != nil {
			return false, err
		}
		return ok, nil
	}
	return false, fmt.Errorf("User '%s' does not exist in database", username)
}

// GetDetails retrieve the groups a user belongs to.
func (p *FileUserProvider) GetDetails(username string) (*UserDetails, error) {
	if details, ok := p.database.Users[username]; ok {
		return &UserDetails{
			Username: username,
			Groups:   details.Groups,
			Emails:   []string{details.Email},
		}, nil
	}
	return nil, fmt.Errorf("User '%s' does not exist in database", username)
}

// UpdatePassword update the password of the given user.
func (p *FileUserProvider) UpdatePassword(username string, newPassword string) error {
	details, ok := p.database.Users[username]
	if !ok {
		return fmt.Errorf("User '%s' does not exist in database", username)
	}

	var algorithm string
	if p.configuration.PasswordHashing.Algorithm == "argon2id" {
		algorithm = HashingAlgorithmArgon2id
	} else if p.configuration.PasswordHashing.Algorithm == "sha512" {
		algorithm = HashingAlgorithmSHA512
	} else {
		return errors.New("Invalid algorithm in configuration. It should be `argon2id` or `sha512`")
	}

	hash, err := HashPassword(
		newPassword, "", algorithm, p.configuration.PasswordHashing.Iterations,
		p.configuration.PasswordHashing.Memory*1024, p.configuration.PasswordHashing.Parallelism,
		p.configuration.PasswordHashing.KeyLength, p.configuration.PasswordHashing.SaltLength)

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
	err = ioutil.WriteFile(p.configuration.Path, b, 0644)
	p.lock.Unlock()
	return err
}
