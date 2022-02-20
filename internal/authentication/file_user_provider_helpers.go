package authentication

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/asaskevich/govalidator"
	"gopkg.in/yaml.v2"
)

//go:embed users_database.template.yml
var databaseTemplate []byte

func fileProviderEnsureDatabaseExists(path string) (err error) {
	var (
		fileInfo os.FileInfo
	)

	if fileInfo, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.WriteFile(path, databaseTemplate, 0600)
			if err != nil {
				return fmt.Errorf("unable to generate user database at path '%s': %w", path, err)
			}

			return fmt.Errorf("the user database was successfully generated at path '%s', please ensure you update it", path)
		}

		return fmt.Errorf("unknown error when trying to initiate the user database at path '%s': %w", path, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("the user database could not be loaded as the path '%s' is a directory", path)
	}

	return nil
}

func fileProviderReadPathToStruct(path string, database *DatabaseModel) (err error) {
	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return fmt.Errorf("unable to open user database file '%s': %w", path, err)
	}

	if err = yaml.Unmarshal(data, database); err != nil {
		return fmt.Errorf("unable to parse user database file '%s': %w", path, err)
	}

	return nil
}

func fileProviderValidateDatabaseSchema(path string, database *DatabaseModel) (err error) {
	var ok bool

	if ok, err = govalidator.ValidateStruct(database); err != nil && !ok {
		return fmt.Errorf("unable to parse user database file '%s': there is an issue with the schema: %w", path, err)
	} else if err != nil {
		return fmt.Errorf("unable to validate the user database file '%s': %w", path, err)
	}

	// Check user password hashes can be parsed.
	for username, user := range database.Users {
		_, err = ParseHash(user.HashedPassword)

		if err != nil {
			return fmt.Errorf("unable to parse user database file '%s': failed to parse password hash of user '%s': %w", path, username, err)
		}
	}

	return nil
}
