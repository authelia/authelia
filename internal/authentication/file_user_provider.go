package authentication

import (
	_ "embed" // Embed users_database.template.yml.
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/argon2"
	"github.com/go-crypt/crypt/algorithm/bcrypt"
	"github.com/go-crypt/crypt/algorithm/pbkdf2"
	"github.com/go-crypt/crypt/algorithm/scrypt"
	"github.com/go-crypt/crypt/algorithm/shacrypt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/logging"
)

// FileUserProvider is a provider reading details from a file.
type FileUserProvider struct {
	config        *schema.AuthenticationBackendFile
	hash          algorithm.Hash
	database      FileUserProviderDatabase
	mutex         sync.Mutex
	timeoutReload time.Time
}

// NewFileUserProvider creates a new instance of FileUserProvider.
func NewFileUserProvider(config *schema.AuthenticationBackendFile) (provider *FileUserProvider) {
	return &FileUserProvider{
		config:        config,
		timeoutReload: time.Now().Add(-1 * time.Second),
		database:      NewFileUserDatabase(config.Path, config.Search.Email, config.Search.CaseInsensitive, getExtra(config)),
	}
}

func getExtra(config *schema.AuthenticationBackendFile) (extra map[string]expression.ExtraAttribute) {
	extra = make(map[string]expression.ExtraAttribute, len(config.ExtraAttributes))

	if len(config.ExtraAttributes) != 0 {
		for name, attribute := range config.ExtraAttributes {
			extra[name] = attribute
		}
	}

	return extra
}

// Reload the database.
func (p *FileUserProvider) Reload() (reloaded bool, err error) {
	now := time.Now()

	p.mutex.Lock()

	defer p.mutex.Unlock()

	if now.Before(p.timeoutReload) {
		return false, nil
	}

	switch err = p.database.Load(); {
	case err == nil:
		p.setTimeoutReload(now)
	case errors.Is(err, ErrNoContent):
		return false, nil
	default:
		return false, fmt.Errorf("failed to reload: %w", err)
	}

	p.setTimeoutReload(now)

	return true, nil
}

func (p *FileUserProvider) Close() (err error) {
	return nil
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *FileUserProvider) CheckUserPassword(username string, password string) (match bool, err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return false, err
	}

	if details.Disabled {
		return false, ErrUserNotFound
	}

	return details.Password.MatchAdvanced(password)
}

// GetDetails retrieves a user's information, excluding disabled users.
func (p *FileUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var d FileUserDatabaseUserDetails

	if d, err = p.database.GetUserDetails(username); err != nil {
		return nil, err
	}

	if d.Disabled {
		return nil, ErrUserNotFound
	}

	return d.ToUserDetails(), nil
}

// GetUser retrieves a user's information, including disabled users.
func (p *FileUserProvider) GetUser(username string) (details *UserDetails, err error) {
	var d FileUserDatabaseUserDetails

	if d, err = p.database.GetUserDetails(username); err != nil {
		return nil, err
	}

	return d.ToUserDetails(), nil
}

// ListUsers returns a list of all users and their attributes.
func (p *FileUserProvider) ListUsers() (userList []UserDetails, err error) {
	if _, err := p.Reload(); err != nil {
		return nil, fmt.Errorf("failed to reload user database: %w", err)
	}

	allUsers := p.database.GetAllUsers()
	userList = make([]UserDetails, 0, len(allUsers))

	for username, details := range allUsers {
		user := UserDetails{
			Username:    username,
			DisplayName: details.DisplayName,
			Emails:      []string{details.Email},
			Groups:      details.Groups,
		}
		userList = append(userList, user)
	}

	return userList, nil
}

func (p *FileUserProvider) GetDetailsExtended(username string) (details *UserDetailsExtended, err error) {
	var d FileUserDatabaseUserDetails

	if d, err = p.database.GetUserDetails(username); err != nil {
		return nil, err
	}

	if d.Disabled {
		return nil, ErrUserNotFound
	}

	return d.ToExtendedUserDetails(), nil
}

// UpdatePassword updates the password of the given user.
func (p *FileUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	if details.Disabled {
		return ErrUserDisabled
	}

	var digest algorithm.Digest

	if digest, err = p.hash.Hash(newPassword); err != nil {
		return err
	}

	details.Password = schema.NewPasswordDigest(digest)

	p.database.SetUserDetails(details.Username, &details)

	p.mutex.Lock()
	p.setTimeoutReload(time.Now())
	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// ChangePassword validates the old password then changes the password of the given user.
func (p *FileUserProvider) ChangePassword(username string, oldPassword string, newPassword string) (err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return fmt.Errorf("%w : %v", ErrUserNotFound, err)
	}

	if details.Disabled {
		return ErrUserDisabled
	}

	if strings.TrimSpace(newPassword) == "" {
		return ErrPasswordWeak
	}

	if oldPassword == newPassword {
		return ErrPasswordWeak
	}

	oldPasswordCorrect, err := p.CheckUserPassword(username, oldPassword)

	if err != nil {
		return ErrAuthenticationFailed
	}

	if !oldPasswordCorrect {
		return ErrIncorrectPassword
	}

	var digest algorithm.Digest

	if digest, err = p.hash.Hash(newPassword); err != nil {
		return fmt.Errorf("%w : %v", ErrOperationFailed, err)
	}

	details.Password = schema.NewPasswordDigest(digest)

	p.database.SetUserDetails(details.Username, &details)

	p.mutex.Lock()
	p.setTimeoutReload(time.Now())
	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return fmt.Errorf("%w : %v", ErrOperationFailed, err)
	}

	return nil
}

// ChangeDisplayName changes the display name for a specific user.
func (p *FileUserProvider) ChangeDisplayName(username string, newDisplayName string) (err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	if details.Disabled {
		return ErrUserDisabled
	}

	if newDisplayName == "" {
		return ErrEmptyInput
	}

	details.DisplayName = newDisplayName

	p.database.SetUserDetails(details.Username, &details)

	p.mutex.Lock()
	p.setTimeoutReload(time.Now())
	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// ChangeEmail changes the groups for a specific user.
func (p *FileUserProvider) ChangeEmail(username string, newEmail string) (err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	if details.Disabled {
		return ErrUserDisabled
	}

	if newEmail == "" {
		return ErrEmptyInput
	}

	details.Email = newEmail

	p.database.SetUserDetails(details.Username, &details)

	p.mutex.Lock()
	p.setTimeoutReload(time.Now())
	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// ChangeGroups changes the groups for a specific user.
func (p *FileUserProvider) ChangeGroups(username string, newGroups []string) (err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	if details.Disabled {
		return ErrUserDisabled
	}

	details.Groups = newGroups

	p.database.SetUserDetails(details.Username, &details)

	p.mutex.Lock()
	p.setTimeoutReload(time.Now())
	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// SetDisabled enables or disables a user.
func (p *FileUserProvider) SetDisabled(username string, disabled bool) (err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	details.Disabled = disabled

	p.database.SetUserDetails(details.Username, &details)

	p.mutex.Lock()

	p.setTimeoutReload(time.Now())

	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// AddUser creates a new user in the file database. Takes additional, optional values via opts.
func (p *FileUserProvider) AddUser(username, displayName, password string, opts ...func(options *NewUserOptionalDetailsOpts)) (err error) {
	var digest algorithm.Digest

	if digest, err = p.hash.Hash(password); err != nil {
		return err
	}

	options := &NewUserOptionalDetailsOpts{}

	for _, opt := range opts {
		opt(options)
	}

	details := FileUserDatabaseUserDetails{
		Username:    username,
		DisplayName: displayName,
		Password:    schema.NewPasswordDigest(digest),
		Email:       options.Email,
		Groups:      options.Groups,
		Disabled:    false,
	}

	p.database.SetUserDetails(details.Username, &details)

	p.mutex.Lock()

	p.setTimeoutReload(time.Now())

	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// UpdateUser modifies an existing user in the file database. Takes new values via opts.
func (p *FileUserProvider) UpdateUser(username string, opts ...func(options *ModifyUserDetailsOpts)) (err error) {
	var details FileUserDatabaseUserDetails

	if details, err = p.database.GetUserDetails(username); err != nil {
		return err
	}

	options := &ModifyUserDetailsOpts{}

	for _, opt := range opts {
		opt(options)
	}

	var digest algorithm.Digest = details.Password

	if *options.Password != "" {
		if digest, err = p.hash.Hash(*options.Password); err != nil {
			return err
		}
	}

	details = FileUserDatabaseUserDetails{
		Username:    username,
		DisplayName: *options.DisplayName,
		Password:    schema.NewPasswordDigest(digest),
		Email:       *options.Email,
		Groups:      options.Groups,
		Disabled:    *options.Disabled,
	}

	p.database.SetUserDetails(username, &details)

	p.mutex.Lock()
	p.setTimeoutReload(time.Now())
	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user from the file database.
func (p *FileUserProvider) DeleteUser(username string) (err error) {
	p.database.DeleteUserDetails(username)

	p.mutex.Lock()

	p.setTimeoutReload(time.Now())

	p.mutex.Unlock()

	if err = p.database.Save(); err != nil {
		return err
	}

	return nil
}

// StartupCheck implements the startup check provider interface.
func (p *FileUserProvider) StartupCheck() (err error) {
	if err = checkDatabase(p.config.Path); err != nil {
		logging.Logger().WithError(err).Errorf("Error checking user authentication YAML database")

		return fmt.Errorf("one or more errors occurred checking the authentication database")
	}

	if p.hash, err = NewFileCryptoHashFromConfig(p.config.Password); err != nil {
		return err
	}

	if p.database == nil {
		p.database = NewFileUserDatabase(p.config.Path, p.config.Search.Email, p.config.Search.CaseInsensitive, getExtra(p.config))
	}

	if err = p.database.Load(); err != nil {
		return err
	}

	return nil
}

func (p *FileUserProvider) setTimeoutReload(now time.Time) {
	p.timeoutReload = now.Add(time.Second / 2)
}

// NewFileCryptoHashFromConfig returns a crypt.Hash given a valid configuration.
func NewFileCryptoHashFromConfig(config schema.AuthenticationBackendFilePassword) (hash algorithm.Hash, err error) {
	switch config.Algorithm {
	case hashArgon2, "":
		hash, err = argon2.New(
			argon2.WithVariantName(config.Argon2.Variant),
			argon2.WithT(config.Argon2.Iterations),
			argon2.WithM(uint32(config.Argon2.Memory)), //nolint:gosec // Validated at runtime.
			argon2.WithP(config.Argon2.Parallelism),
			argon2.WithK(config.Argon2.KeyLength),
			argon2.WithS(config.Argon2.SaltLength),
		)
	case hashSHA2Crypt:
		hash, err = shacrypt.New(
			shacrypt.WithVariantName(config.SHA2Crypt.Variant),
			shacrypt.WithIterations(config.SHA2Crypt.Iterations),
			shacrypt.WithSaltLength(config.SHA2Crypt.SaltLength),
		)
	case hashPBKDF2:
		hash, err = pbkdf2.New(
			pbkdf2.WithVariantName(config.PBKDF2.Variant),
			pbkdf2.WithIterations(config.PBKDF2.Iterations),
			pbkdf2.WithSaltLength(config.PBKDF2.SaltLength),
		)
	case hashSCrypt:
		hash, err = scrypt.New(
			scrypt.WithLN(config.SCrypt.Iterations),
			scrypt.WithP(config.SCrypt.Parallelism),
			scrypt.WithR(config.SCrypt.BlockSize),
			scrypt.WithKeyLength(config.SCrypt.KeyLength),
			scrypt.WithSaltLength(config.SCrypt.SaltLength),
		)
	case hashBCrypt:
		hash, err = bcrypt.New(
			bcrypt.WithVariantName(config.BCrypt.Variant),
			bcrypt.WithIterations(config.BCrypt.Cost),
		)
	default:
		return nil, fmt.Errorf("algorithm '%s' is unknown", config.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize hash settings: %w", err)
	}

	return hash, nil
}

func checkDatabase(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.WriteFile(path, userYAMLTemplate, 0600); err != nil {
			return fmt.Errorf("user authentication database file doesn't exist at path '%s' and could not be generated: %w", path, err)
		}

		return fmt.Errorf("user authentication database file doesn't exist at path '%s' and has been generated", path)
	} else if err != nil {
		return fmt.Errorf("error checking user authentication database file: %w", err)
	}

	return nil
}

//go:embed users_database.template.yml
var userYAMLTemplate []byte
