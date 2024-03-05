package storage

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/authelia/authelia/v4/internal/model"
)

//go:embed migrations/*
var migrationsFS embed.FS

func latestMigrationVersion(providerName string) (version int, err error) {
	var (
		entries   []fs.DirEntry
		migration model.SchemaMigration
	)

	if entries, err = migrationsFS.ReadDir(path.Join(pathMigrations, providerName)); err != nil {
		return -1, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if migration, err = scanMigration(providerName, entry.Name()); err != nil {
			return -1, err
		}

		if !migration.Up {
			continue
		}

		if migration.Version > version {
			version = migration.Version
		}
	}

	return version, nil
}

// loadMigrations scans the migrations fs and loads the appropriate migrations for a given providerName, prior and
// target versions. If the target version is -1 this indicates the latest version. If the target version is 0
// this indicates the database zero state.
func loadMigrations(providerName string, prior, target int) (migrations []model.SchemaMigration, err error) {
	if prior == target {
		return nil, ErrMigrateCurrentVersionSameAsTarget
	}

	var (
		migration model.SchemaMigration
		entries   []fs.DirEntry
	)

	if entries, err = migrationsFS.ReadDir(path.Join(pathMigrations, providerName)); err != nil {
		return nil, err
	}

	up := prior < target

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if migration, err = scanMigration(providerName, entry.Name()); err != nil {
			return nil, err
		}

		if skipMigration(up, target, prior, &migration) {
			continue
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

func skipMigration(up bool, target, prior int, migration *model.SchemaMigration) (skip bool) {
	if up {
		if !migration.Up {
			// Skip if we wanted an Up migration but it isn't an Up migration.
			return true
		}

		if migration.Version > target || migration.Version <= prior {
			// Skip if the migration version is greater than the target or less than or equal to the previous version.
			return true
		}
	} else {
		if migration.Up {
			// Skip if we didn't want an Up migration but it is an Up migration.
			return true
		}

		if migration.Version <= target || migration.Version > prior {
			// Skip the migration if we want to go down and the migration version is less than or equal to the target
			// or greater than the previous version.
			return true
		}
	}

	return false
}

func scanMigration(providerName, m string) (migration model.SchemaMigration, err error) {
	if !reMigration.MatchString(m) {
		return model.SchemaMigration{}, errors.New("invalid migration: could not parse the format")
	}

	result := reMigration.FindStringSubmatch(m)

	migration = model.SchemaMigration{
		Name:     strings.ReplaceAll(result[reMigration.SubexpIndex("Name")], "_", " "),
		Provider: providerName,
	}

	var data []byte

	if data, err = migrationsFS.ReadFile(path.Join(pathMigrations, providerName, m)); err != nil {
		return model.SchemaMigration{}, err
	}

	migration.Query = string(data)

	switch direction := result[reMigration.SubexpIndex("Direction")]; direction {
	case "up":
		migration.Up = true
	case "down":
		migration.Up = false
	default:
		return model.SchemaMigration{}, fmt.Errorf("invalid migration: value in Direction group '%s' must be up or down", direction)
	}

	migration.Version, _ = strconv.Atoi(result[reMigration.SubexpIndex("Version")])

	return migration, nil
}
