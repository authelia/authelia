package storage

import (
	"database/sql"
	"fmt"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/clems4ever/authelia/logging"
	_ "github.com/go-sql-driver/mysql" // Load the MySQL Driver used in the connection string.
)

// MySQLProvider is a MySQL provider
type MySQLProvider struct {
	SQLProvider
}

// NewSQLProvider a SQL provider
func NewSQLProvider(configuration schema.SQLStorageConfiguration) *MySQLProvider {
	connectionString := configuration.Username

	if configuration.Password != "" {
		connectionString += fmt.Sprintf(":%s", configuration.Password)
	}

	if connectionString != "" {
		connectionString += "@"
	}

	address := configuration.Host
	if configuration.Port > 0 {
		address += fmt.Sprintf(":%d", configuration.Port)
	}
	connectionString += fmt.Sprintf("tcp(%s)", address)

	if configuration.Database != "" {
		connectionString += fmt.Sprintf("/%s", configuration.Database)
	}

	fmt.Println(connectionString)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		logging.Logger().Fatalf("Unable to connect to SQL database: %v", err)
	}

	provider := MySQLProvider{}
	if err := provider.initialize(db); err != nil {
		logging.Logger().Fatalf("Unable to initialize SQL database: %v", err)
	}
	return &provider
}
