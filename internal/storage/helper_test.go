package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableShouldBeQuotedPerDBType(t *testing.T) {
	var tableName = "the_table"

	assert.Equal(t, fmt.Sprintf("`%s`", tableName), quoteTableName(tableName, "mysql"))
	assert.Equal(t, fmt.Sprintf(`"%s"`, tableName), quoteTableName(tableName, "postgres"))
	assert.Equal(t, fmt.Sprintf(`"%s"`, tableName), quoteTableName(tableName, "sqlite"))
	assert.Equal(t, tableName, quoteTableName(tableName, ""))
}
