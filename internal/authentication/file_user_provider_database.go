package authentication

import (
	"fmt"
	"os"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/go-crypt/crypt"
	"gopkg.in/yaml.v3"
)

// NewFileUserDatabase creates a new FileUserDatabase.
func NewFileUserDatabase(filePath string) (database *FileUserDatabase) {
	return &FileUserDatabase{
		RWMutex: &sync.RWMutex{},
		Path:    filePath,
		Users:   map[string]DatabaseUserDetails{},
	}
}

// FileUserDatabase is a user details database that is concurrency safe database and can be reloaded.
type FileUserDatabase struct {
	*sync.RWMutex

	Path  string
	Users map[string]DatabaseUserDetails
}

// Save the database to disk.
func (m *FileUserDatabase) Save() (err error) {
	m.RLock()

	defer m.RUnlock()

	if err = m.ToDatabaseModel().Write(m.Path); err != nil {
		return err
	}

	return nil
}

// Load the database from disk.
func (m *FileUserDatabase) Load() (err error) {
	yml := &DatabaseModel{Users: map[string]UserDetailsModel{}}

	if err = yml.Read(m.Path); err != nil {
		return fmt.Errorf("error reading the authentication database: %w", err)
	}

	m.Lock()

	defer m.Unlock()

	if err = yml.ReadToFileUserDatabase(m); err != nil {
		return fmt.Errorf("error decoding the authentication database: %w", err)
	}

	return nil
}

// GetUserDetails get a DatabaseUserDetails given a username as a value type where the username must be the users actual
// username.
func (m *FileUserDatabase) GetUserDetails(username string) (user DatabaseUserDetails, err error) {
	m.RLock()

	defer m.RUnlock()

	if details, ok := m.Users[username]; ok {
		return details, nil
	}

	return user, ErrUserNotFound
}

// SetUserDetails sets the DatabaseUserDetails for a given user.
func (m *FileUserDatabase) SetUserDetails(username string, details *DatabaseUserDetails) {
	if details == nil {
		return
	}

	m.Lock()

	m.Users[username] = *details

	m.Unlock()
}

// ToDatabaseModel converts the FileUserDatabase into the DatabaseModel for saving.
func (m *FileUserDatabase) ToDatabaseModel() (model *DatabaseModel) {
	model = &DatabaseModel{
		Users: map[string]UserDetailsModel{},
	}

	m.RLock()

	for user, details := range m.Users {
		model.Users[user] = details.ToUserDetailsModel()
	}

	m.RUnlock()

	return model
}

// DatabaseUserDetails is the model of user details in the file database.
type DatabaseUserDetails struct {
	Username    string
	Digest      crypt.Digest
	Disabled    bool
	DisplayName string
	Email       string
	Groups      []string
}

// ToUserDetails converts DatabaseUserDetails into a *UserDetails given a username.
func (m DatabaseUserDetails) ToUserDetails() (details *UserDetails) {
	return &UserDetails{
		Username:    m.Username,
		DisplayName: m.DisplayName,
		Emails:      []string{m.Email},
		Groups:      m.Groups,
	}
}

// ToUserDetailsModel converts DatabaseUserDetails into a UserDetailsModel.
func (m DatabaseUserDetails) ToUserDetailsModel() (model UserDetailsModel) {
	return UserDetailsModel{
		HashedPassword: m.Digest.Encode(),
		DisplayName:    m.DisplayName,
		Email:          m.Email,
		Groups:         m.Groups,
	}
}

// DatabaseModel is the model of users file database.
type DatabaseModel struct {
	Users map[string]UserDetailsModel `yaml:"users" valid:"required"`
}

// ReadToFileUserDatabase reads the DatabaseModel into a FileUserDatabase.
func (m *DatabaseModel) ReadToFileUserDatabase(db *FileUserDatabase) (err error) {
	users := map[string]DatabaseUserDetails{}

	var udm *DatabaseUserDetails

	for user, details := range m.Users {
		if udm, err = details.ToDatabaseUserDetailsModel(user); err != nil {
			return fmt.Errorf("failed to parse hash for user '%s': %w", user, err)
		}

		users[user] = *udm
	}

	db.Users = users

	return nil
}

// Read a DatabaseModel from disk.
func (m *DatabaseModel) Read(filePath string) (err error) {
	var (
		content []byte
		ok      bool
	)

	if content, err = os.ReadFile(filePath); err != nil {
		return fmt.Errorf("failed to read the '%s' file: %w", filePath, err)
	}

	if len(content) == 0 {
		return ErrNoContent
	}

	if err = yaml.Unmarshal(content, m); err != nil {
		return fmt.Errorf("could not parse the YAML database: %w", err)
	}

	if ok, err = govalidator.ValidateStruct(m); err != nil {
		return fmt.Errorf("could not validate the schema: %w", err)
	}

	if !ok {
		return fmt.Errorf("the schema is invalid")
	}

	return nil
}

// Write a DatabaseModel to disk.
func (m *DatabaseModel) Write(fileName string) (err error) {
	var (
		data []byte
	)

	if data, err = yaml.Marshal(m); err != nil {
		return err
	}

	return os.WriteFile(fileName, data, fileAuthenticationMode)
}

// UserDetailsModel is the model of user details in the file database.
type UserDetailsModel struct {
	HashedPassword string   `yaml:"password" valid:"required"`
	DisplayName    string   `yaml:"displayname" valid:"required"`
	Email          string   `yaml:"email"`
	Groups         []string `yaml:"groups"`
}

// ToDatabaseUserDetailsModel converts a UserDetailsModel into a *DatabaseUserDetails.
func (m UserDetailsModel) ToDatabaseUserDetailsModel(username string) (model *DatabaseUserDetails, err error) {
	var d crypt.Digest

	if d, err = crypt.Decode(m.HashedPassword); err != nil {
		return nil, err
	}

	return &DatabaseUserDetails{
		Username:    username,
		Digest:      d,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Groups:      m.Groups,
	}, nil
}
