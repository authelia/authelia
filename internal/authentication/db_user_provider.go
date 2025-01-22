package authentication

import (
	"context"
	"errors"
	"fmt"

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

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if user, err = p.database.LoadUser(ctx, username, p.config.Search.Email); err != nil {
		return false, fmt.Errorf("%w: %w", ErrUserNotFound, err)
	}

	if user.Disabled {
		return false, fmt.Errorf("%w: %w", ErrUserNotFound, ErrUserDisabled)
	}

	var d algorithm.Digest

	if d, err = crypt.Decode(string(user.Password)); err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidPassword, err)
	}

	return d.MatchAdvanced(password)
}

// UpdatePassword implements authentication.UserProvider.UpdatePassword().
func (p *DBUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	var user model.User

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if user, err = p.database.LoadUser(ctx, username, p.config.Search.Email); err != nil {
		return fmt.Errorf("%w: %w", ErrUserNotFound, err)
	}

	if user.Disabled {
		return fmt.Errorf("%w: %w", ErrUserNotFound, ErrUserDisabled)
	}

	// the user real username.
	username = user.Username

	var passwordDigest string

	if passwordDigest, err = p.hashPassword(newPassword); err != nil {
		return fmt.Errorf("%w: %w", ErrUpdatingUserPassword, err)
	}

	if err = p.database.UpdateUserPassword(ctx, username, passwordDigest); err != nil {
		return fmt.Errorf("%w: %w", ErrUpdatingUserPassword, err)
	}

	return nil
}

func (p *DBUserProvider) hashPassword(password string) (passwordDigest string, err error) {
	var digest algorithm.Digest

	if p.hash == nil {
		return "", fmt.Errorf("%w: hash algorithm not defined", ErrHashError)
	}

	if digest, err = p.hash.Hash(password); err != nil {
		return "", fmt.Errorf("%w: %w", ErrHashError, err)
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
		return nil, fmt.Errorf("%w: %w", ErrUserNotFound, ErrUserDisabled)
	}

	return &user.UserDetails, nil
}

// GetDetailsExtended load user extended info, returns error if user not exists.
func (p *DBUserProvider) GetDetailsExtended(username string) (*UserDetailsExtended, error) {
	var user model.User

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if user, err = p.database.LoadUser(ctx, username, p.config.Search.Email); err != nil {
		logging.Logger().WithError(err).Warn("error loading user info")
		return nil, ErrUserNotFound
	}

	var details = userModelToUserDetailsExtended(user)

	return &details, nil
}

// AddUser adds a user given the new user's information.
func (p *DBUserProvider) AddUser(username, displayname, password string, opts ...func(options *NewUserDetailsOpts)) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

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

	if options.Email, err = parseEmail(options.Email); err != nil {
		return ErrInvalidEmail
	}

	var passwordDigest string

	if passwordDigest, err = p.hashPassword(password); err != nil {
		return fmt.Errorf("%w: %w", ErrCreatingUser, err)
	}

	if ctx, err = p.database.BeginTX(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrCreatingUser, err)
	}

	if err = p.addUserTx(ctx, username, displayname, passwordDigest, options); err != nil {
		if rbErr := p.database.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("%w: %w: %w", ErrCreatingUser, err, rbErr)
		}

		return fmt.Errorf("%w: %w", ErrCreatingUser, err)
	}

	if err = p.database.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrCreatingUser, err)
	}

	return nil
}

func (p *DBUserProvider) addUserTx(ctx context.Context, username, displayname, password string, options *NewUserDetailsOpts) (err error) {
	if exists, err := p.database.UserExists(ctx, username); err != nil {
		return ErrCreatingUser
	} else if exists {
		return ErrUserExists
	}

	if err = p.database.CreateUser(ctx, username, options.Email, password); err != nil {
		return err
	}

	if err = p.database.UpdateUserGroups(ctx, username, options.Groups...); err != nil {
		return err
	}

	if displayname != "" {
		if err = p.database.UpdateUserDisplayName(ctx, username, displayname); err != nil {
			return err
		}
	}

	return nil
}

/*
// UpdateUser updates a user.
func (p *DBUserProvider) UpdateUser(username string, before, after *UserDetails) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		return fmt.Errorf("%w: %w", ErrUpdatingUser, err)
	} else if !exists {
		return ErrUserNotFound
	}

	data := model.User{}

	if err = p.database.UpdateUser(ctx, username, data); err != nil {
		return fmt.Errorf("%w: %w", ErrUpdatingUser, err)
	}

	return nil
}.
*/

// DeleteUser deletes user given the username.
func (p *DBUserProvider) DeleteUser(username string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		return fmt.Errorf("%w: error checking user: %w", ErrDeletingUser, err)
	} else if !exists {
		return ErrUserNotFound
	}

	if err = p.database.DeleteUser(ctx, username); err != nil {
		return fmt.Errorf("%w: %w", ErrDeletingUser, err)
	}

	return nil
}

// ChangeDisplayName changes the display name for a specific user.
func (p *DBUserProvider) ChangeDisplayName(username, newDisplayName string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		logging.Logger().WithError(err).Warn("error changing user's display name")
		return ErrUpdatingUser
	} else if !exists {
		return ErrUserNotFound
	}

	if err = p.database.UpdateUserDisplayName(ctx, username, newDisplayName); err != nil {
		logging.Logger().WithError(err).Warn("error changing user's display name")
		return ErrUpdatingUser
	}

	return nil
}

// ChangeEmail changes the email for a specific user.
func (p *DBUserProvider) ChangeEmail(username, newEmail string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		logging.Logger().WithError(err).Warn("error changing user's email")
		return ErrUpdatingUser
	} else if !exists {
		return ErrUserNotFound
	}

	if err = p.database.UpdateUserEmail(ctx, username, newEmail); err != nil {
		logging.Logger().WithError(err).Warn("error changing user's email")
		return ErrUpdatingUser
	}

	return nil
}

// ChangeGroups changes the groups for a specific user.
func (p *DBUserProvider) ChangeGroups(username string, newGroups []string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		logging.Logger().WithError(err).Warn("error checking if user exists")
		return ErrUpdatingUser
	} else if !exists {
		return ErrUserNotFound
	}

	if ctx, err = p.database.BeginTX(ctx); err != nil {
		logging.Logger().WithError(err).Warn("error creating transaction")
		return ErrUpdatingUser
	}

	if err = p.database.UpdateUserGroups(ctx, username, newGroups...); err != nil {
		if rbErr := p.database.Rollback(ctx); rbErr != nil {
			logging.Logger().WithError(err).Warn("failed to rollback user changes")
		}

		return ErrUpdatingUser
	}

	if err = p.database.Commit(ctx); err != nil {
		logging.Logger().WithError(err).Warn("failed to commit user creation")
		return ErrUpdatingUser
	}

	return nil
}

// ListUsers returns a list of all users and their attributes.
func (p *DBUserProvider) ListUsers() (userList []UserDetailsExtended, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	var models []model.User

	if models, err = p.database.ListUsers(ctx); err != nil {
		return userList, fmt.Errorf("%w: %w", ErrListingUser, err)
	}

	for _, u := range models {
		userList = append(userList, userModelToUserDetailsExtended(u))
	}

	return userList, nil
}

// DisableUser disables a user.
func (p *DBUserProvider) DisableUser(username string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		logging.Logger().WithError(err).Warn("error disabling user")
		return ErrUpdatingUser
	} else if !exists {
		return ErrUserNotFound
	}

	if err = p.database.UpdateUserStatus(ctx, username, true); err != nil {
		logging.Logger().WithError(err).Warn("error disabling user")
		return ErrUpdatingUser
	}

	return nil
}

// EnableUser enables a user.
func (p *DBUserProvider) EnableUser(username string) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	if exists, err := p.database.UserExists(ctx, username); err != nil {
		logging.Logger().WithError(err).Warn("error enabling user")
		return ErrUpdatingUser
	} else if !exists {
		return ErrUserNotFound
	}

	if err = p.database.UpdateUserStatus(ctx, username, false); err != nil {
		logging.Logger().WithError(err).Warn("error enabling user")
		return ErrUpdatingUser
	}

	return nil
}
