package storage

import "fmt"

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
