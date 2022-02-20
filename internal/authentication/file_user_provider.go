package authentication

import (
	_ "embed" // Embed users_database.template.yml.
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
)

// FileUserProvider is a provider reading details from a file.
type FileUserProvider struct {
	config   *schema.FileAuthenticationBackendConfiguration
	database *DatabaseModel

	lock *sync.Mutex

	log *logrus.Logger
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
func NewFileUserProvider(config *schema.FileAuthenticationBackendConfiguration) (provider *FileUserProvider) {
	provider = &FileUserProvider{
		config: config,
		lock:   &sync.Mutex{},
		log:    logging.Logger(),
	}

	return provider
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

	return false, ErrUserNotFound
}

// GetCurrentDetails returns GetDetails.
func (p *FileUserProvider) GetCurrentDetails(username string) (details *model.UserDetails, err error) {
	return p.GetDetails(username)
}

// GetDetails retrieve the groups a user belongs to.
func (p *FileUserProvider) GetDetails(username string) (*model.UserDetails, error) {
	if details, ok := p.database.Users[username]; ok {
		return &model.UserDetails{
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

	algorithm, err := ConfigAlgoToCryptoAlgo(p.config.Password.Algorithm)
	if err != nil {
		return err
	}

	hash, err := HashPassword(
		newPassword, "", algorithm, p.config.Password.Iterations,
		p.config.Password.Memory*1024, p.config.Password.Parallelism,
		p.config.Password.KeyLength, p.config.Password.SaltLength)

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

	err = os.WriteFile(p.config.Path, b, fileAuthenticationMode)
	p.lock.Unlock()

	return err
}

func (p *FileUserProvider) load() (err error) {
	database := &DatabaseModel{}

	if err = fileProviderReadPathToStruct(p.config.Path, database); err != nil {
		return err
	}

	if err = fileProviderValidateDatabaseSchema(p.config.Path, database); err != nil {
		return err
	}

	p.lock.Lock()

	p.database = database

	p.lock.Unlock()

	return nil
}

// StartupCheck implements the startup check provider interface.
func (p *FileUserProvider) StartupCheck() (err error) {
	if err = fileProviderEnsureDatabaseExists(p.config.Path); err != nil {
		return err
	}

	if err = p.load(); err != nil {
		return err
	}

	return nil
}
