package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/authelia/authelia/v4/internal/model"
)

// ConsentSessionRows holds and assists with retrieving multiple model.OAuth2ConsentSession rows.
type ConsentSessionRows struct {
	rows *sqlx.Rows
}

// Next is the row iterator.
func (r *ConsentSessionRows) Next() bool {
	if r.rows == nil {
		return false
	}

	return r.rows.Next()
}

// Close the rows.
func (r *ConsentSessionRows) Close() (err error) {
	if r.rows == nil {
		return nil
	}

	return r.rows.Close()
}

// Get returns the *model.OAuth2ConsentSession or scan error.
func (r *ConsentSessionRows) Get() (consent *model.OAuth2ConsentSession, err error) {
	if r.rows == nil {
		return nil, sql.ErrNoRows
	}

	consent = &model.OAuth2ConsentSession{}

	if err = r.rows.StructScan(consent); err != nil {
		return nil, err
	}

	return consent, nil
}
