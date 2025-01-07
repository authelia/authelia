package storage

import (
	"errors"
	"fmt"
	"strings"
)

func quoteTableName(tableName, dbType string) string {
	switch dbType {
	case providerMySQL:
		return fmt.Sprintf("`%s`", tableName)
	case providerPostgres:
		return fmt.Sprintf(`"%s"`, tableName)
	case providerSQLite:
		return fmt.Sprintf(`"%s"`, tableName)
	}

	return tableName
}

func isUserNotFoundError(err error) bool {
	return err != nil && err.Error() == errUserNotFound
}

func validateUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return errors.New("username can't be empty")
	}

	return nil
}

func validateGroupname(username string) error {
	if strings.TrimSpace(username) == "" {
		return errors.New("group name can't be empty")
	}

	return nil
}
