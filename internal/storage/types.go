package storage

import (
	"github.com/jmoiron/sqlx"
)

// SQLXClient is a *sqlx.DB or *sqlx.Tx.
type SQLXClient interface {
	sqlx.Execer
	sqlx.ExecerContext

	sqlx.Preparer
	sqlx.PreparerContext

	sqlx.Queryer
	sqlx.QueryerContext

	sqlx.Ext
	sqlx.ExtContext
}
