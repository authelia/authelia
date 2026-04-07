package authentication

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/argon2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// PGUserProvider is a user provider backed by PostgreSQL.
type PGUserProvider struct {
	db     *sqlx.DB
	config schema.AuthenticationBackendPostgreSQL
	hash   algorithm.Hash
}

type pgUser struct {
	Username    string    `db:"username"`
	DisplayName string    `db:"display_name"`
	Password    string    `db:"password_hash"`
	Email       string    `db:"email"`
	Groups      string    `db:"groups"`
	Disabled    bool      `db:"disabled"`
	Phone       string    `db:"phone"`
	CreatedAt   time.Time `db:"created_at"`
}

// NewPGUserProvider creates a new PostgreSQL-backed user provider.
func NewPGUserProvider(config schema.AuthenticationBackendPostgreSQL) (*PGUserProvider, error) {
	db, err := sqlx.Connect("postgres", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	hash, err := argon2.New(
		argon2.WithVariantName("argon2id"),
		argon2.WithT(3),
		argon2.WithM(65536),
		argon2.WithP(4),
		argon2.WithK(32),
		argon2.WithS(16),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create hash: %w", err)
	}

	return &PGUserProvider{db: db, config: config, hash: hash}, nil
}

// StartupCheck verifies the database connection and creates the users table if needed.
func (p *PGUserProvider) StartupCheck() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS authelia_users (
			username      VARCHAR(100) PRIMARY KEY,
			display_name  VARCHAR(255) NOT NULL DEFAULT '',
			password_hash TEXT NOT NULL,
			email         VARCHAR(255) NOT NULL DEFAULT '',
			groups        VARCHAR(500) NOT NULL DEFAULT 'users',
			disabled      BOOLEAN NOT NULL DEFAULT FALSE,
			phone         VARCHAR(20) NOT NULL DEFAULT '',
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure authelia_users table: %w", err)
	}

	_, _ = p.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_authelia_users_email ON authelia_users(email)`)

	return nil
}

func (p *PGUserProvider) getUser(username string) (*pgUser, error) {
	user := &pgUser{}

	err := p.db.Get(user, "SELECT username, display_name, password_hash, email, groups, disabled, phone FROM authelia_users WHERE username = $1", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("error retrieving user '%s': %w", username, err)
	}

	if user.Disabled {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// CheckUserPassword verifies a user's password.
func (p *PGUserProvider) CheckUserPassword(username string, password string) (valid bool, err error) {
	user, err := p.getUser(username)
	if err != nil {
		return false, err
	}

	digest, err := crypt.Decode(user.Password)
	if err != nil {
		return false, fmt.Errorf("error decoding password hash for user '%s': %w", username, err)
	}

	return digest.MatchAdvanced(password)
}

// GetDetails returns a user's basic details.
func (p *PGUserProvider) GetDetails(username string) (details *UserDetails, err error) {
	user, err := p.getUser(username)
	if err != nil {
		return nil, err
	}

	groups := parseGroups(user.Groups)
	emails := []string{}

	if user.Email != "" {
		emails = append(emails, user.Email)
	}

	return &UserDetails{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Emails:      emails,
		Groups:      groups,
	}, nil
}

// GetDetailsExtended returns a user's extended details.
func (p *PGUserProvider) GetDetailsExtended(username string) (details *UserDetailsExtended, err error) {
	user, err := p.getUser(username)
	if err != nil {
		return nil, err
	}

	groups := parseGroups(user.Groups)
	emails := []string{}

	if user.Email != "" {
		emails = append(emails, user.Email)
	}

	extended := &UserDetailsExtended{
		PhoneNumber: user.Phone,
	}
	extended.Username = user.Username
	extended.DisplayName = user.DisplayName
	extended.Emails = emails
	extended.Groups = groups

	return extended, nil
}

// UpdatePassword updates a user's password.
func (p *PGUserProvider) UpdatePassword(username string, newPassword string) (err error) {
	digest, err := p.hash.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	result, err := p.db.Exec("UPDATE authelia_users SET password_hash = $1, updated_at = NOW() WHERE username = $2", digest.Encode(), username)
	if err != nil {
		return fmt.Errorf("failed to update password for '%s': %w", username, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user '%s' not found", username)
	}

	return nil
}

// ChangePassword verifies old password then updates.
func (p *PGUserProvider) ChangePassword(username string, oldPassword string, newPassword string) (err error) {
	valid, err := p.CheckUserPassword(username, oldPassword)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("old password is incorrect")
	}

	return p.UpdatePassword(username, newPassword)
}

// Close closes the database connection.
func (p *PGUserProvider) Close() (err error) {
	if p.db != nil {
		return p.db.Close()
	}

	return nil
}

func parseGroups(groups string) []string {
	result := []string{}

	for _, g := range strings.Split(groups, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			result = append(result, g)
		}
	}

	return result
}
