package authentication

import (
	"context"
	"errors"
	"time"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

// DBUserProvider is a provider reading details from a sql database.
type DBUserProvider struct {
	config   *schema.AuthenticationBackendDB
	database storage.AuthenticationStorageProvider
	hash     algorithm.Hash
}

// NewDBUserProvider creates a new instance of DBUserProvider.
func NewDBUserProvider(config *schema.AuthenticationBackendDB, database storage.AuthenticationStorageProvider) (provider *DBUserProvider) {
	return &DBUserProvider{
		config:   config,
		database: database,
	}
}

// StartupCheck implements authentication.UserProvider.StartupCheck().
func (p *DBUserProvider) StartupCheck() (err error) {
	if p.config == nil {
		return errors.New("nil configuration provided")
	}

	if p.hash, err = NewCryptoHashFromConfig(p.config.Password); err != nil {
		return err
	}

	return nil
}

// CheckUserPassword implements authentication.UserProvider.CheckUserPassword().
func (p *DBUserProvider) CheckUserPassword(username string, password string) (valid bool, err error) {
	var user model.User

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)

	defer cancel()

	if user, err = p.database.LoadUser(ctx, username, p.config.Search.Email); err != nil {
		logging.Logger().WithError(err).Warn("error loading user info")
		return false, ErrUserNotFound
	}

	if user.Disabled {
		logging.Logger().WithError(err).Warn("user is disabled")
		return false, ErrUserNotFound
	}

	var d algorithm.Digest

	if d, err = crypt.Decode(string(user.Password)); err != nil {
		logging.Logger().WithError(err).Warn("error decoding user's password hash")
		return false, ErrInvalidPassword
	}

	return d.MatchAdvanced(password)
}

// UpdatePassword implements authentication.UserProvider.UpdatePassword().
func (p *DBUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	var user model.User

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)

	defer cancel()

	if user, err = p.database.LoadUser(ctx, username, p.config.Search.Email); err != nil {
		logging.Logger().WithError(err).Warn("error loading user info")
		return ErrUserNotFound
	}

	if user.Disabled {
		logging.Logger().WithError(err).Warn("trying to update password to disabled user")
		return ErrUserNotFound
	}

	// the user real username.
	username = user.Username

	var passwordDigest string

	if passwordDigest, err = p.hashPassword(newPassword); err != nil {
		logging.Logger().WithError(err).Warn("error hashing user's password")
		return ErrUpdatingUserPassword
	}

	if err = p.database.UpdateUserPassword(ctx, username, passwordDigest); err != nil {
		logging.Logger().WithError(err).Warn("error updating user's passord")
		return ErrUpdatingUserPassword
	}

	return nil
}

func (p *DBUserProvider) hashPassword(password string) (passwordDigest string, err error) {
	var digest algorithm.Digest

	if digest, err = p.hash.Hash(password); err != nil {
		return "", err
	}

	passwordDigest = schema.NewPasswordDigest(digest).String()

	return
}

// GetDetails implements authentication.UserProvider.GetDetails().
func (p *DBUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var user *UserDetailsExtended

	if user, err = p.GetDetailsExtended(username); err != nil {
		return nil, err
	}

	if user.Disabled {
		logging.Logger().WithError(err).Warn("user is disabled")
		return nil, ErrUserNotFound
	}

	return &user.UserDetails, nil
}

// GetDetailsExtended load user extended info, returns error if user not exists.
func (p *DBUserProvider) GetDetailsExtended(username string) (*UserDetailsExtended, error) {
	var user model.User

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)

	defer cancel()

	if user, err = p.database.LoadUser(ctx, username, p.config.Search.Email); err != nil {
		logging.Logger().WithError(err).Warn("error loading user info")
		return nil, ErrUserNotFound
	}

	return &UserDetailsExtended{
		UserDetails: UserDetails{
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Emails:      []string{user.Email},
			Groups:      user.Groups,
		},
		Disabled: user.Disabled,
	}, nil
}

// AddUser adds a user given the new user's information.
func (p *DBUserProvider) AddUser(username, displayname, password string, opts ...func(options *NewUserDetailsOpts)) (err error) {
	if username == "" {
		return ErrInvalidUsername
	}

	if password == "" {
		return ErrInvalidPassword
	}

	options := &NewUserDetailsOpts{}

	for _, opt := range opts {
		opt(options)
	}

	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout*time.Second)

	defer cancel()

	if options.Email, err = parseEmail(options.Email); err != nil {
		return ErrInvalidEmail
	}

	var passwordDigest string

	if passwordDigest, err = p.hashPassword(password); err != nil {
		logging.Logger().WithError(err).Warn("error generating password hash for user")
		return ErrCreatingUser
	}

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		return ErrCreatingUser
	} else if exists {
		return ErrUserExists
	}

	err = p.database.CreateUser(ctx, model.User{
		Username:    username,
		Password:    []byte(passwordDigest),
		DisplayName: displayname,
		Email:       options.Email,
		Groups:      options.Groups,
		Disabled:    options.Disabled,
	})

	if err != nil {
		logging.Logger().WithError(err).Warn("error creating user")
		return ErrCreatingUser
	}

	return nil
}

// DeleteUser deletes user given the username.
func (p *DBUserProvider) DeleteUser(username string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout*time.Second)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		logging.Logger().WithError(err).Warn("error deleting user")
		return ErrDeletingUser
	} else if !exists {
		return ErrUserNotFound
	}

	if err = p.database.DeleteUser(ctx, username); err != nil {
		logging.Logger().WithError(err).Warn("error deleting user")
		return ErrDeletingUser
	}

	return nil
}

// ChangePassword validates the old password then changes the password of the given user.
func (p *DBUserProvider) ChangePassword(username, oldPassword, newPassword string) (err error) {
	return errors.New("not implemented")
}

// ChangeDisplayName changes the display name for a specific user.
func (p *DBUserProvider) ChangeDisplayName(username, newDisplayName string) (err error) {
	return errors.New("not implemented")
}

// ChangeEmail changes the email for a specific user.
func (p *DBUserProvider) ChangeEmail(username, newEmail string) (err error) {
	return errors.New("not implemented")
}

// ChangeGroups changes the groups for a specific user.
func (p *DBUserProvider) ChangeGroups(username string, newGroups []string) (err error) {
	return errors.New("not implemented")
}

// ListUsers returns a list of all users and their attributes.
func (p *DBUserProvider) ListUsers() (userList []UserDetails, err error) {
	return userList, errors.New("not implemented")
}
