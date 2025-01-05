package authentication

import (
	"context"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
	// TODO: verify that table exists.
	if p.hash, err = NewCryptoHashFromConfig(p.config.Password); err != nil {
		return err
	}

	return nil
}

// CheckUserPassword implements authentication.UserProvider.CheckUserPassword().
func (p *DBUserProvider) CheckUserPassword(username string, password string) (valid bool, err error) {
	var user *model.User

	if user, err = p.loadUser(username); err != nil {
		return false, err
	}

	var d algorithm.Digest

	if d, err = crypt.Decode(string(user.Password)); err != nil {
		return false, err
	}

	return d.MatchAdvanced(password)
}

// GetDetails implements authentication.UserProvider.GetDetails().
func (p *DBUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	var user *model.User

	if user, err = p.loadUser(username); err != nil {
		return nil, err
	}

	return &UserDetails{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Emails:      []string{user.Email},
		Groups:      user.Groups,
	}, nil
}

// UpdatePassword implements authentication.UserProvider.UpdatePassword().
func (p *DBUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	var user *model.User

	if user, err = p.loadUser(username); err != nil {
		return err
	}

	var digest algorithm.Digest

	if digest, err = p.hash.Hash(newPassword); err != nil {
		return err
	}

	// the user real username.
	username = user.Username

	var passwordDigest = schema.NewPasswordDigest(digest).String()

	if err = p.database.UpdateUserPassword(context.Background(), username, passwordDigest); err != nil {
		return err
	}

	return nil
}

// loadUser load user info, returns error if user not exists or is disabled.
func (p *DBUserProvider) loadUser(username string) (*model.User, error) {
	var user model.User

	var err error

	// TODO: see howto inject a real context.
	var ctx = context.Background()

	if p.config.Search.Email {
		user, err = p.database.LoadUserByEmail(ctx, username)
		if err == nil && !user.Disabled {
			return &user, err
		}
	}

	if user, err = p.database.LoadUserByUsername(ctx, username); err != nil {
		return nil, err
	}

	if user.Disabled {
		return nil, ErrUserNotFound
	}

	return &user, nil
}
