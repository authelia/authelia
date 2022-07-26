package authentication

import (
	"fmt"
	"os"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/go-crypt/crypt"
	"gopkg.in/yaml.v3"
)

func NewFileUserDatabase(p string) (database *FileUserDatabase) {
	return &FileUserDatabase{
		Mutex: &sync.Mutex{},
		Path:  p,
		Users: map[string]FileUserDatabaseUser{},
	}
}

type FileUserDatabase struct {
	*sync.Mutex

	Path  string
	Users map[string]FileUserDatabaseUser
}

func (m *FileUserDatabase) Save() (err error) {
	m.Lock()

	if err = m.ToDatabaseModel().Write(m.Path); err != nil {
		m.Unlock()

		return err
	}

	m.Unlock()

	return nil
}

func (m *FileUserDatabase) Load() (err error) {
	yml := &DatabaseModel{Users: map[string]UserDetailsModel{}}

	if err = yml.Read(m.Path); err != nil {
		return fmt.Errorf("error reading the authentication database: %w", err)
	}

	var db *FileUserDatabase

	if db, err = yml.ToDatabaseModel(); err != nil {
		return fmt.Errorf("error decoding the authentication database: %w", err)
	}

	m.Lock()

	m.Users = db.Users

	m.Unlock()

	return nil
}

func (m FileUserDatabase) GetUserDetails(username string) (user *FileUserDatabaseUser, err error) {
	if details, ok := m.Users[username]; ok {
		return &details, nil
	}

	return nil, ErrUserNotFound
}

func (m *FileUserDatabase) SetUserDetails(username string, details *FileUserDatabaseUser) {
	m.Lock()

	m.Users[username] = *details

	m.Unlock()
}

func (m FileUserDatabase) ToDatabaseModel() (model *DatabaseModel) {
	model = &DatabaseModel{
		Users: map[string]UserDetailsModel{},
	}

	for user, details := range m.Users {
		model.Users[user] = details.ToUserDetailsModel()
	}

	return model
}

// FileUserDatabaseUser is the model of user details in the file database.
type FileUserDatabaseUser struct {
	Digest      crypt.Digest
	DisplayName string
	Email       string
	Groups      []string
}

func (m FileUserDatabaseUser) ToUserDetailsModel() (model UserDetailsModel) {
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

func (m DatabaseModel) ToDatabaseModel() (model *FileUserDatabase, err error) {
	model = &FileUserDatabase{
		Users: map[string]FileUserDatabaseUser{},
	}

	var udm *FileUserDatabaseUser

	for user, details := range m.Users {
		if udm, err = details.ToDatabaseUserDetailsModel(); err != nil {
			return nil, fmt.Errorf("failed to parse hash for user '%s': %w", user, err)
		}

		model.Users[user] = *udm
	}

	return model, nil
}

func (m *DatabaseModel) Read(p string) (err error) {
	var (
		content []byte
		ok      bool
	)

	if content, err = os.ReadFile(p); err != nil {
		return fmt.Errorf("failed to read the '%s' file: %w", p, err)
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

func (m *DatabaseModel) Write(p string) (err error) {
	var (
		data []byte
	)

	if data, err = yaml.Marshal(m); err != nil {
		return err
	}

	return os.WriteFile(p, data, fileAuthenticationMode)
}

// UserDetailsModel is the model of user details in the file database.
type UserDetailsModel struct {
	HashedPassword string   `yaml:"password" valid:"required"`
	DisplayName    string   `yaml:"displayname" valid:"required"`
	Email          string   `yaml:"email"`
	Groups         []string `yaml:"groups"`
}

func (m UserDetailsModel) ToDatabaseUserDetailsModel() (model *FileUserDatabaseUser, err error) {
	var d crypt.Digest

	if d, err = crypt.Decode(m.HashedPassword); err != nil {
		return nil, err
	}

	return &FileUserDatabaseUser{
		Digest:      d,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Groups:      m.Groups,
	}, nil
}
