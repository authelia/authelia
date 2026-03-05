package authentication

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"go.yaml.in/yaml/v4"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
)

type FileUserProviderDatabase interface {
	Save() (err error)
	Load() (err error)
	GetUserDetails(username string) (user FileUserDatabaseUserDetails, err error)
	SetUserDetails(username string, details *FileUserDatabaseUserDetails)
}

// NewFileUserDatabase creates a new FileUserDatabase.
func NewFileUserDatabase(filePath string, searchEmail, searchCI bool, extra map[string]expression.ExtraAttribute) (database *FileUserDatabase) {
	return &FileUserDatabase{
		RWMutex:     &sync.RWMutex{},
		Path:        filePath,
		Users:       map[string]FileUserDatabaseUserDetails{},
		Emails:      map[string]string{},
		Aliases:     map[string]string{},
		SearchEmail: searchEmail,
		SearchCI:    searchCI,
		Extra:       extra,
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

	Extra map[string]expression.ExtraAttribute
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

	if err = yml.ReadToFileUserDatabase(m, m.Extra); err != nil {
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
	Username       string                 `json:"-"`
	Password       *schema.PasswordDigest `json:"password" jsonschema:"required,title=Password" jsonschema_description:"The hashed password for the user."`
	DisplayName    string                 `json:"displayname" jsonschema:"required,title=Display Name" jsonschema_description:"The display name for the user."`
	GivenName      string                 `json:"given_name,omitempty" jsonschema:"title=Given Name" jsonschema_description:"The given name for the user."`
	MiddleName     string                 `json:"middle_name,omitempty" jsonschema:"title=Middle Name" jsonschema_description:"The middle name for the user."`
	FamilyName     string                 `json:"family_name,omitempty" jsonschema:"title=Family Name" jsonschema_description:"The family name for the user."`
	Nickname       string                 `json:"nickname,omitempty" jsonschema:"title=Nickname" jsonschema_description:"The nickname for the user."`
	Gender         string                 `json:"gender,omitempty" jsonschema:"title=Gender" jsonschema_description:"The gender for the user."`
	Birthdate      string                 `json:"birthdate,omitempty" jsonschema:"title=Birthdate" jsonschema_description:"The birthdate for the user."`
	Website        *url.URL               `json:"website,omitempty" jsonschema:"title=Website" jsonschema_description:"The website URL for the user."`
	Profile        *url.URL               `json:"profile,omitempty" jsonschema:"title=Profile" jsonschema_description:"The profile URL for the user."`
	Picture        *url.URL               `json:"picture,omitempty" jsonschema:"title=Picture" jsonschema_description:"The picture URL for the user."`
	ZoneInfo       string                 `json:"zoneinfo,omitempty" jsonschema:"title=Zone Information" jsonschema_description:"The time zone for the user."`
	Locale         *language.Tag          `json:"locale,omitempty" jsonschema:"title=Locale" jsonschema_description:"The BCP47 locale for the user."`
	PhoneNumber    string                 `json:"phone_number,omitempty" jsonschema:"title=Phone Number" jsonschema_description:"The phone number for the user."`
	PhoneExtension string                 `json:"phone_extension,omitempty" jsonschema:"title=Phone Extension" jsonschema_description:"The phone extension for the user."`
	Email          string                 `json:"email" jsonschema:"title=Email" jsonschema_description:"The email for the user."`
	Groups         []string               `json:"groups" jsonschema:"title=Groups" jsonschema_description:"The groups list for the user."`
	Disabled       bool                   `json:"disabled" jsonschema:"default=false,title=Disabled" jsonschema_description:"The disabled status for the user."`

	Address *FileUserDatabaseUserDetailsAddressModel `json:"address,omitempty" jsonschema:"title=Address" jsonschema_description:"The address for the user."`

	Extra map[string]any `json:"extra" jsonschema:"title=Extra" jsonschema_description:"The extra attributes for the user."`
}

type FileUserDatabaseUserDetailsAddressModel struct {
	StreetAddress string `yaml:"street_address" json:"street_address,omitempty" jsonschema:"title=Street Address" jsonschema_description:"The street address for the user."`
	Locality      string `yaml:"locality" json:"locality,omitempty" jsonschema:"title=Locality" jsonschema_description:"The locality for the user."`
	Region        string `yaml:"region" json:"region,omitempty" jsonschema:"title=Region" jsonschema_description:"The region for the user."`
	PostalCode    string `yaml:"postal_code" json:"postal_code,omitempty" jsonschema:"title=Postal Code" jsonschema_description:"The postal code or postcode for the user."`
	Country       string `yaml:"country" json:"country,omitempty" jsonschema:"title=Country" jsonschema_description:"The country for the user."`
}

// ToUserDetails converts FileUserDatabaseUserDetails into a *UserDetails.
func (m FileUserDatabaseUserDetails) ToUserDetails() (details *UserDetails) {
	var emails []string

	if m.Email != "" {
		emails = append(emails, m.Email)
	}

	return &UserDetails{
		Username:    m.Username,
		DisplayName: m.DisplayName,
		Emails:      emails,
		Groups:      m.Groups,
	}
}

// ToExtendedUserDetails converts FileUserDatabaseUserDetails into a *UserDetailsExtended.
func (m FileUserDatabaseUserDetails) ToExtendedUserDetails() (details *UserDetailsExtended) {
	details = &UserDetailsExtended{
		GivenName:      m.GivenName,
		FamilyName:     m.FamilyName,
		MiddleName:     m.MiddleName,
		Nickname:       m.Nickname,
		Profile:        m.Profile,
		Picture:        m.Picture,
		Website:        m.Website,
		Gender:         m.Gender,
		Birthdate:      m.Birthdate,
		ZoneInfo:       m.ZoneInfo,
		Locale:         m.Locale,
		PhoneNumber:    m.PhoneNumber,
		PhoneExtension: m.PhoneExtension,
		UserDetails:    m.ToUserDetails(),
		Extra:          m.Extra,
	}

	if m.Address != nil {
		details.Address = &UserDetailsAddress{
			StreetAddress: m.Address.StreetAddress,
			Locality:      m.Address.Locality,
			Region:        m.Address.Region,
			PostalCode:    m.Address.PostalCode,
			Country:       m.Address.Country,
		}
	}

	return details
}

// ToUserDetailsModel converts FileUserDatabaseUserDetails into a FileDatabaseUserDetailsModel.
func (m FileUserDatabaseUserDetails) ToUserDetailsModel() (model FileDatabaseUserDetailsModel) {
	model = FileDatabaseUserDetailsModel{
		Password:       m.Password.Encode(),
		DisplayName:    m.DisplayName,
		GivenName:      m.GivenName,
		MiddleName:     m.MiddleName,
		FamilyName:     m.FamilyName,
		Nickname:       m.Nickname,
		Gender:         m.Gender,
		Birthdate:      m.Birthdate,
		ZoneInfo:       m.ZoneInfo,
		PhoneNumber:    m.PhoneNumber,
		PhoneExtension: m.PhoneExtension,
		Email:          m.Email,
		Groups:         m.Groups,
		Address:        m.Address,
		Extra:          m.Extra,
	}

	if m.Website != nil {
		model.Website = m.Website.String()
	}

	if m.Profile != nil {
		model.Profile = m.Profile.String()
	}

	if m.Picture != nil {
		model.Picture = m.Picture.String()
	}

	if m.Locale != nil {
		model.Locale = m.Locale.String()
	}

	return model
}

// FileDatabaseModel is the model of users file database.
type FileDatabaseModel struct {
	Users map[string]FileDatabaseUserDetailsModel `yaml:"users" json:"users" valid:"required" jsonschema:"required,title=Users" jsonschema_description:"The dictionary of users."`
}

// ReadToFileUserDatabase reads the FileDatabaseModel into a FileUserDatabase.
func (m *FileDatabaseModel) ReadToFileUserDatabase(db *FileUserDatabase, extra map[string]expression.ExtraAttribute) (err error) {
	users := map[string]FileUserDatabaseUserDetails{}

	var udm *FileUserDatabaseUserDetails

	for username, details := range m.Users {
		if err = details.ValidateExtra(username, extra); err != nil {
			return err
		}

		if udm, err = details.ToDatabaseUserDetailsModel(username); err != nil {
			return err
		}

		users[username] = *udm
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
	Password       string   `yaml:"password" valid:"required"` //nolint:gosec // This is a hash, not a raw password.
	DisplayName    string   `yaml:"displayname" valid:"required"`
	Email          string   `yaml:"email"`
	Groups         []string `yaml:"groups"`
	GivenName      string   `yaml:"given_name"`
	MiddleName     string   `yaml:"middle_name"`
	FamilyName     string   `yaml:"family_name"`
	Nickname       string   `yaml:"nickname"`
	Gender         string   `yaml:"gender"`
	Birthdate      string   `yaml:"birthdate"`
	Website        string   `yaml:"website"`
	Profile        string   `yaml:"profile"`
	Picture        string   `yaml:"picture"`
	ZoneInfo       string   `yaml:"zoneinfo"`
	Locale         string   `yaml:"locale"`
	PhoneNumber    string   `yaml:"phone_number"`
	PhoneExtension string   `yaml:"phone_extension"`
	Disabled       bool     `yaml:"disabled"`

	Address *FileUserDatabaseUserDetailsAddressModel `yaml:"address"`

	Extra map[string]any `yaml:"extra"`
}

//nolint:gocyclo
func (m FileDatabaseUserDetailsModel) ValidateExtra(username string, extra map[string]expression.ExtraAttribute) (err error) {
	for name, value := range m.Extra {
		attribute, ok := extra[name]
		if !ok {
			return fmt.Errorf("error occurred validating extra attributes for user '%s': attribute '%s' is unknown", username, name)
		}

		mv := attribute.IsMultiValued()
		vt := attribute.GetValueType()

		if !mv {
			switch value.(type) {
			case string:
				if vt == ValueTypeString {
					continue
				}
			case int, int64, int32, float64, float32:
				if vt == ValueTypeInteger {
					continue
				}
			case bool:
				if vt == ValueTypeBoolean {
					continue
				}
			default:
				return fmt.Errorf("error occurred validating extra attributes for user '%s': attribute '%s' has the unknown type '%T'", username, name, value)
			}

			return fmt.Errorf("error occurred validating extra attributes for user '%s': attribute '%s' has the known type '%T' but '%s' is the expected type", username, name, value, vt)
		}

		values, ok := value.([]any)
		if !ok {
			return fmt.Errorf("error occurred validating extra attributes for user '%s': attribute '%s' has the type '%T' but '[]%s' is the expected type", username, name, value, vt)
		}

		for _, v := range values {
			switch v.(type) {
			case string:
				if vt == ValueTypeString {
					continue
				}
			case int, int64, int32, float64, float32:
				if vt == ValueTypeInteger {
					continue
				}
			case bool:
				if vt == ValueTypeBoolean {
					continue
				}
			default:
				return fmt.Errorf("error occurred validating extra attributes for user '%s': attribute '%s' has the unknown item type '%T'", username, name, v)
			}
		}
	}

	return nil
}

// ToDatabaseUserDetailsModel converts a FileDatabaseUserDetailsModel into a *FileUserDatabaseUserDetails.
func (m FileDatabaseUserDetailsModel) ToDatabaseUserDetailsModel(username string) (model *FileUserDatabaseUserDetails, err error) {
	var d algorithm.Digest

	if d, err = crypt.Decode(m.Password); err != nil {
		return nil, fmt.Errorf("error occurred decoding the password hash for '%s': %w", username, err)
	}

	model = &FileUserDatabaseUserDetails{
		Username:       username,
		Password:       schema.NewPasswordDigest(d),
		Disabled:       m.Disabled,
		DisplayName:    m.DisplayName,
		Email:          m.Email,
		GivenName:      m.GivenName,
		MiddleName:     m.MiddleName,
		FamilyName:     m.FamilyName,
		Nickname:       m.Nickname,
		Gender:         m.Gender,
		Birthdate:      m.Birthdate,
		ZoneInfo:       m.ZoneInfo,
		PhoneNumber:    m.PhoneNumber,
		PhoneExtension: m.PhoneExtension,
		Groups:         m.Groups,
		Address:        m.Address,
		Extra:          m.Extra,
	}

	if m.Website != "" {
		if model.Website, err = parseAttributeURI(username, "", "website", m.Website); err != nil {
			return nil, err
		}
	}

	if m.Profile != "" {
		if model.Profile, err = parseAttributeURI(username, "", "profile", m.Profile); err != nil {
			return nil, err
		}
	}

	if m.Picture != "" {
		if model.Picture, err = parseAttributeURI(username, "", "picture", m.Picture); err != nil {
			return nil, err
		}
	}

	if m.Locale != "" {
		var tag language.Tag

		if tag, err = language.Parse(m.Locale); err != nil {
			return nil, fmt.Errorf("error occurred parsing user details for '%s': failed to parse the locale attribute with value '%s': %w", username, m.Locale, err)
		}

		model.Locale = &tag
	}

	return model, nil
}
