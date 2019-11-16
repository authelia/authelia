package storage

import (
	"database/sql"
	"fmt"
	"time"

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

	for i := 0; i < 3; i++ {
		if err = db.Ping(); err == nil {
			logging.Logger().Debug("Connection to the database is established")
			break
		}

		if i == 2 {
			logging.Logger().Fatal("Aborting because connection to database failed")
		}

		logging.Logger().Errorf("Unable to ping database retrying in 10 seconds. error: %v", err)
		time.Sleep(10 * time.Second)
	}

	provider := MySQLProvider{}
	if err := provider.initialize(db); err != nil {
		logging.Logger().Fatalf("Unable to initialize SQL database: %v", err)
	}
	return &provider
}
