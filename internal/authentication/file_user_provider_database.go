package authentication

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type FileUserProviderDatabase interface {
	Save() (err error)
	Load() (err error)
	GetUserDetails(username string) (user FileUserDatabaseUserDetails, err error)
	SetUserDetails(username string, details *FileUserDatabaseUserDetails)
}

// NewFileUserDatabase creates a new FileUserDatabase.
func NewFileUserDatabase(filePath string, searchEmail, searchCI bool) (database *FileUserDatabase) {
	return &FileUserDatabase{
		RWMutex:     &sync.RWMutex{},
		Path:        filePath,
		Users:       map[string]FileUserDatabaseUserDetails{},
		Emails:      map[string]string{},
		Aliases:     map[string]string{},
		SearchEmail: searchEmail,
		SearchCI:    searchCI,
	}
}

// FileUserDatabase is a user details database that is concurrency safe database and can be reloaded.
type FileUserDatabase struct {
	*sync.RWMutex `json:"-"`

	Users map[string]FileUserDatabaseUserDetails `json:"users" jsonschema:"required,title=Users" jsonschema_description:"The dictionary of users."`

	Path    string            `json:"-"`
	Emails  map[string]string `json:"-"`
	Aliases map[string]string `json:"-"`

	SearchEmail bool `json:"-"`
	SearchCI    bool `json:"-"`
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
	yml := &FileDatabaseModel{Users: map[string]FileDatabaseUserDetailsModel{}}

	if err = yml.Read(m.Path); err != nil {
		return fmt.Errorf("error reading the authentication database: %w", err)
	}

	m.Lock()

	defer m.Unlock()

	if err = yml.ReadToFileUserDatabase(m); err != nil {
		return fmt.Errorf("error decoding the authentication database: %w", err)
	}

	return m.LoadAliases()
}

// LoadAliases performs the loading of alias information from the database.
func (m *FileUserDatabase) LoadAliases() (err error) {
	if m.SearchEmail || m.SearchCI {
		for k, user := range m.Users {
			if m.SearchEmail && user.Email != "" {
				if err = m.loadAliasEmail(k, user); err != nil {
					return err
				}
			}

			if m.SearchCI {
				if err = m.loadAlias(k); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (m *FileUserDatabase) loadAlias(k string) (err error) {
	u := strings.ToLower(k)

	if u != k {
		return fmt.Errorf("error loading authentication database: username '%s' is not lowercase but this is required when case-insensitive search is enabled", k)
	}

	for username, details := range m.Users {
		if k == username {
			continue
		}

		if strings.EqualFold(u, details.Email) {
			return fmt.Errorf("error loading authentication database: username '%s' is configured as an email for user with username '%s' which isn't allowed when case-insensitive search is enabled", u, username)
		}
	}

	m.Aliases[u] = k

	return nil
}

func (m *FileUserDatabase) loadAliasEmail(k string, user FileUserDatabaseUserDetails) (err error) {
	e := strings.ToLower(user.Email)

	var duplicates []string

	for username, details := range m.Users {
		if k == username {
			continue
		}

		if strings.EqualFold(e, details.Email) {
			duplicates = append(duplicates, username)
		}
	}

	if len(duplicates) != 0 {
		duplicates = append(duplicates, k)

		return fmt.Errorf("error loading authentication database: email '%s' is configured for for more than one user (users are '%s') which isn't allowed when email search is enabled", e, strings.Join(duplicates, "', '"))
	}

	if _, ok := m.Users[e]; ok && k != e {
		return fmt.Errorf("error loading authentication database: email '%s' is also a username which isn't allowed when email search is enabled", e)
	}

	m.Emails[e] = k

	return nil
}

// GetUserDetails get a FileUserDatabaseUserDetails given a username as a value type where the username must be the users actual
// username.
func (m *FileUserDatabase) GetUserDetails(username string) (user FileUserDatabaseUserDetails, err error) {
	m.RLock()

	defer m.RUnlock()

	u := strings.ToLower(username)

	if m.SearchEmail {
		if key, ok := m.Emails[u]; ok {
			return m.Users[key], nil
		}
	}

	if m.SearchCI {
		if key, ok := m.Aliases[u]; ok {
			return m.Users[key], nil
		}
	}

	if details, ok := m.Users[username]; ok {
		return details, nil
	}

	return user, ErrUserNotFound
}

// SetUserDetails sets the FileUserDatabaseUserDetails for a given user.
func (m *FileUserDatabase) SetUserDetails(username string, details *FileUserDatabaseUserDetails) {
	if details == nil {
		return
	}

	m.Lock()

	m.Users[username] = *details

	m.Unlock()
}

// ToDatabaseModel converts the FileUserDatabase into the FileDatabaseModel for saving.
func (m *FileUserDatabase) ToDatabaseModel() (model *FileDatabaseModel) {
	model = &FileDatabaseModel{
		Users: map[string]FileDatabaseUserDetailsModel{},
	}

	m.RLock()

	for user, details := range m.Users {
		model.Users[user] = details.ToUserDetailsModel()
	}

	m.RUnlock()

	return model
}

// FileUserDatabaseUserDetails is the model of user details in the file database.
type FileUserDatabaseUserDetails struct {
	Username    string                 `json:"-"`
	Password    *schema.PasswordDigest `json:"password" jsonschema:"required,title=Password" jsonschema_description:"The hashed password for the user."`
	DisplayName string                 `json:"displayname" jsonschema:"required,title=Display Name" jsonschema_description:"The display name for the user."`
	Email       string                 `json:"email" jsonschema:"title=Email" jsonschema_description:"The email for the user."`
	Groups      []string               `json:"groups" jsonschema:"title=Groups" jsonschema_description:"The groups list for the user."`
	Disabled    bool                   `json:"disabled" jsonschema:"default=false,title=Disabled" jsonschema_description:"The disabled status for the user."`
}

// ToUserDetails converts FileUserDatabaseUserDetails into a *UserDetails given a username.
func (m FileUserDatabaseUserDetails) ToUserDetails() (details *UserDetails) {
	return &UserDetails{
		Username:    m.Username,
		DisplayName: m.DisplayName,
		Emails:      []string{m.Email},
		Groups:      m.Groups,
	}
}

// ToUserDetailsModel converts FileUserDatabaseUserDetails into a FileDatabaseUserDetailsModel.
func (m FileUserDatabaseUserDetails) ToUserDetailsModel() (model FileDatabaseUserDetailsModel) {
	return FileDatabaseUserDetailsModel{
		Password:    m.Password.Encode(),
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Groups:      m.Groups,
	}
}

// FileDatabaseModel is the model of users file database.
type FileDatabaseModel struct {
	Users map[string]FileDatabaseUserDetailsModel `yaml:"users" json:"users" valid:"required" jsonschema:"required,title=Users" jsonschema_description:"The dictionary of users."`
}

// ReadToFileUserDatabase reads the FileDatabaseModel into a FileUserDatabase.
func (m *FileDatabaseModel) ReadToFileUserDatabase(db *FileUserDatabase) (err error) {
	users := map[string]FileUserDatabaseUserDetails{}

	var udm *FileUserDatabaseUserDetails

	for user, details := range m.Users {
		if udm, err = details.ToDatabaseUserDetailsModel(user); err != nil {
			return fmt.Errorf("failed to parse hash for user '%s': %w", user, err)
		}

		users[user] = *udm
	}

	db.Users = users

	return nil
}

// Read a FileDatabaseModel from disk.
func (m *FileDatabaseModel) Read(filePath string) (err error) {
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

// Write a FileDatabaseModel to disk.
func (m *FileDatabaseModel) Write(fileName string) (err error) {
	var (
		data []byte
	)

	if data, err = yaml.Marshal(m); err != nil {
		return err
	}

	return os.WriteFile(fileName, data, fileAuthenticationMode)
}

// FileDatabaseUserDetailsModel is the model of user details in the file database.
type FileDatabaseUserDetailsModel struct {
	Password    string   `yaml:"password" valid:"required"`
	DisplayName string   `yaml:"displayname" valid:"required"`
	Email       string   `yaml:"email"`
	Groups      []string `yaml:"groups"`
	Disabled    bool     `yaml:"disabled"`
}

// ToDatabaseUserDetailsModel converts a FileDatabaseUserDetailsModel into a *FileUserDatabaseUserDetails.
func (m FileDatabaseUserDetailsModel) ToDatabaseUserDetailsModel(username string) (model *FileUserDatabaseUserDetails, err error) {
	var d algorithm.Digest

	if d, err = crypt.Decode(m.Password); err != nil {
		return nil, err
	}

	return &FileUserDatabaseUserDetails{
		Username:    username,
		Password:    schema.NewPasswordDigest(d),
		Disabled:    m.Disabled,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Groups:      m.Groups,
	}, nil
}
