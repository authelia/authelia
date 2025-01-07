package authentication

import (
	"context"
	"errors"
	"time"

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
	var user *model.User

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	if user, err = p.loadUser(ctx, username); err != nil {
		return false, err
	}

	if user.Disabled {
		return false, ErrUserNotFound
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

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	if user, err = p.loadUser(ctx, username); err != nil {
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	if user, err = p.loadUser(ctx, username); err != nil {
		return err
	}

	if user.Disabled {
		return ErrUserNotFound
	}

	var digest algorithm.Digest

	if digest, err = p.hash.Hash(newPassword); err != nil {
		return err
	}

	// the user real username.
	username = user.Username

	var passwordDigest = schema.NewPasswordDigest(digest).String()

	if err = p.database.UpdateUserPassword(ctx, username, passwordDigest); err != nil {
		return err
	}

	return nil
}

// loadUser load user info, returns error if user not exists or is disabled.
func (p *DBUserProvider) loadUser(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	var err error

	if p.config.Search.Email {
		user, err = p.database.LoadUserByEmail(ctx, username)
		if err == nil && !user.Disabled {
			return &user, err
		}
	}

	if user, err = p.database.LoadUserByUsername(ctx, username); err != nil {
		return nil, err
	}

	return &user, nil
}
