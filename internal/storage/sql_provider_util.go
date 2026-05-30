package storage

import (
	"database/sql"
	"fmt"
)

func checkSingleUpdateResult(result sql.Result) (err error) {
	var rowsAffected int64

	switch rowsAffected, err = result.RowsAffected(); {
	case err != nil:
		return fmt.Errorf("error occurred determining the number of affected rows: %w", err)
	case rowsAffected == 0:
		return ErrNoRowsAffected
	case rowsAffected > 1:
		return ErrMultipleRowsAffected
	default:
		return nil
	}
}
