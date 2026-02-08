package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/authelia/authelia/v4/internal/model"
)

// ConsentPreConfigRows holds and helps with retrieving multiple model.OAuth2ConsentSession rows.
type ConsentPreConfigRows struct {
	rows *sqlx.Rows
}

// Next is the row iterator.
func (r *ConsentPreConfigRows) Next() bool {
	if r.rows == nil {
		return false
	}

	return r.rows.Next()
}

// Close the rows.
func (r *ConsentPreConfigRows) Close() (err error) {
	if r.rows == nil {
		return nil
	}

	return r.rows.Close()
}

// Get returns the *model.OAuth2ConsentSession or scan error.
func (r *ConsentPreConfigRows) Get() (config *model.OAuth2ConsentPreConfig, err error) {
	if r.rows == nil {
		return nil, sql.ErrNoRows
	}

	config = &model.OAuth2ConsentPreConfig{}

	if err = r.rows.StructScan(config); err != nil {
		return nil, err
	}

	return config, nil
}
