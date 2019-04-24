package authentication

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/asaskevich/govalidator"

	"gopkg.in/yaml.v2"
)

// FileUserProvider is a provider reading details from a file.
type FileUserProvider struct {
	path     *string
	database *DatabaseModel
	lock     *sync.Mutex
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
func NewFileUserProvider(filepath string) *FileUserProvider {
	database, err := readDatabase(filepath)
	if err != nil {
		// Panic since the file does not exist when Authelia is starting.
		panic(err)
	}
	return &FileUserProvider{
		path:     &filepath,
		database: database,
		lock:     &sync.Mutex{},
	}
}

func readDatabase(path string) (*DatabaseModel, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	db := DatabaseModel{}
	err = yaml.Unmarshal(content, &db)
	if err != nil {
		return nil, err
	}

	ok, err := govalidator.ValidateStruct(db)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("The database format is invalid: %s", err.Error())
	}
	return &db, nil
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *FileUserProvider) CheckUserPassword(username string, password string) (bool, error) {
	if details, ok := p.database.Users[username]; ok {
		hashedPassword := details.HashedPassword[7:] // Remove {CRYPT}
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
			Groups: details.Groups,
			Emails: []string{details.Email},
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

	hash := HashPassword(newPassword, nil)
	details.HashedPassword = fmt.Sprintf("{CRYPT}%s", hash)

	p.lock.Lock()
	p.database.Users[username] = details

	b, err := yaml.Marshal(p.database)
	if err != nil {
		p.lock.Unlock()
		return err
	}
	err = ioutil.WriteFile(*p.path, b, 0644)
	p.lock.Unlock()
	return err
}
