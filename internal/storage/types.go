package storage

import (
	"github.com/jmoiron/sqlx"
)

// DBOrTx represents a database or transaction.
type DBOrTx interface {
	sqlx.Ext
	sqlx.ExtContext
	sqlx.Preparer
	sqlx.PreparerContext
}
