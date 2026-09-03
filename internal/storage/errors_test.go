package storage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestIsSerializationFailure(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{name: "ShouldNotMatchNil", err: nil, expected: false},
		{name: "ShouldNotMatchGeneric", err: errors.New("some error"), expected: false},
		{name: "ShouldMatchSQLiteBusy", err: sqlite3.Error{Code: sqlite3.ErrBusy}, expected: true},
		{name: "ShouldMatchSQLiteLocked", err: sqlite3.Error{Code: sqlite3.ErrLocked}, expected: true},
		{name: "ShouldMatchSQLiteBusyWrapped", err: fmt.Errorf("error revoking oauth2 refresh token session: %w", sqlite3.Error{Code: sqlite3.ErrBusy}), expected: true},
		{name: "ShouldNotMatchSQLiteConstraint", err: sqlite3.Error{Code: sqlite3.ErrConstraint}, expected: false},
		{name: "ShouldMatchMySQLDeadlock", err: &mysql.MySQLError{Number: codeMySQLLockDeadlock}, expected: true},
		{name: "ShouldMatchMySQLLockWaitTimeout", err: &mysql.MySQLError{Number: codeMySQLLockWaitTimeout}, expected: true},
		{name: "ShouldMatchMySQLDeadlockWrapped", err: fmt.Errorf("error inserting oauth2 access token session: %w", &mysql.MySQLError{Number: codeMySQLLockDeadlock}), expected: true},
		{name: "ShouldNotMatchMySQLDuplicateEntry", err: &mysql.MySQLError{Number: 1062}, expected: false},
		{name: "ShouldMatchPostgresSerializationFailure", err: &pgconn.PgError{Code: codePostgresSerializationFailure}, expected: true},
		{name: "ShouldMatchPostgresDeadlockDetected", err: &pgconn.PgError{Code: codePostgresDeadlockDetected}, expected: true},
		{name: "ShouldMatchPostgresSerializationFailureWrapped", err: fmt.Errorf("error inserting oauth2 access token session: %w", &pgconn.PgError{Code: codePostgresSerializationFailure}), expected: true},
		{name: "ShouldNotMatchPostgresUniqueViolation", err: &pgconn.PgError{Code: "23505"}, expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsSerializationFailure(tc.err))
		})
	}
}
