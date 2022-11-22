package storage

import (
	"github.com/jmoiron/sqlx"
)

// SQLXConnection is a *sqlx.DB or *sqlx.Tx.
type SQLXConnection interface {
	sqlx.Execer
	sqlx.ExecerContext

	sqlx.Preparer
	sqlx.PreparerContext

	sqlx.Queryer
	sqlx.QueryerContext

	sqlx.Ext
	sqlx.ExtContext
}
