package storage

import (
	"embed"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*
var migrationsFS embed.FS

func latestMigrationVersion(providerName string) (version int, err error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return -1, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		m, err := scanMigration(entry.Name())
		if err != nil {
			return -1, err
		}

		if m.Provider != providerName {
			continue
		}

		if !m.Up {
			continue
		}

		if m.Version > version {
			version = m.Version
		}
	}

	return version, nil
}

func loadMigration(providerName string, version int, up bool) (migration *SchemaMigration, err error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		m, err := scanMigration(entry.Name())
		if err != nil {
			return nil, err
		}

		migration = &m

		if up != migration.Up {
			continue
		}

		if migration.Provider != providerAll && migration.Provider != providerName {
			continue
		}

		if version != migration.Version {
			continue
		}

		return migration, nil
	}

	return nil, errors.New("migration not found")
}

// loadMigrations scans the migrations fs and loads the appropriate migrations for a given providerName, prior and
// target versions. If the target version is -1 this indicates the latest version. If the target version is 0
// this indicates the database zero state.
func loadMigrations(providerName string, prior, target int) (migrations []SchemaMigration, err error) {
	if prior == target && (prior != -1 || target != -1) {
		return nil, errors.New("cannot migrate to the same version as prior")
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	up := prior < target

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		migration, err := scanMigration(entry.Name())
		if err != nil {
			return nil, err
		}

		if migration.Provider != providerAll && migration.Provider != providerName {
			continue
		}

		if up {
			if !migration.Up {
				continue
			}

			if target != -1 && (migration.Version > target || migration.Version <= prior) {
				continue
			}
		} else {
			if migration.Up {
				continue
			}

			// If we're targeting pre1 we want to skip the down migration 1.
			if migration.Version == 1 && target == -1 {
				continue
			}

			if migration.Version <= target || migration.Version > prior {
				continue
			}
		}

		migrations = append(migrations, migration)
	}

	if up {
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Version < migrations[j].Version
		})
	} else {
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Version > migrations[j].Version
		})
	}

	return migrations, nil
}

func scanMigration(m string) (migration SchemaMigration, err error) {
	result := reMigration.FindStringSubmatch(m)

	if result == nil || len(result) != 5 {
		return SchemaMigration{}, errors.New("invalid migration: could not parse the format")
	}

	migration = SchemaMigration{
		Name:     strings.ReplaceAll(result[2], "_", " "),
		Provider: result[3],
	}

	data, err := migrationsFS.ReadFile(fmt.Sprintf("migrations/%s", m))
	if err != nil {
		return SchemaMigration{}, err
	}

	migration.Query = string(data)

	switch result[4] {
	case "up":
		migration.Up = true
	case "down":
		migration.Up = false
	default:
		return SchemaMigration{}, fmt.Errorf("invalid migration: value in position 4 '%s' must be up or down", result[4])
	}

	migration.Version, _ = strconv.Atoi(result[1])

	switch migration.Provider {
	case providerAll, provideerSQLite, providerMySQL, providerPostgres:
		break
	default:
		return SchemaMigration{}, fmt.Errorf("invalid migration: value in position 3 '%s' must be all, sqlite, postgres, or mysql", result[3])
	}

	return migration, nil
}
