package storage

import (
	"bytes"
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

func latestMigrationVersion(provider string) (version int, err error) {
	var (
		entries   []fs.DirEntry
		migration model.SchemaMigration
	)

	if entries, err = migrationsFS.ReadDir(path.Join(pathMigrations, provider)); err != nil {
		return -1, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if migration, err = scanMigration(provider, entry.Name()); err != nil {
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
func loadMigrations(provider, schema string, prior, target int) (migrations []model.SchemaMigration, err error) {
	if prior == target {
		return nil, ErrMigrateCurrentVersionSameAsTarget
	}

	start := providerMigrationStart[provider]

	if start != 1 {
		switch {
		case prior > target && target != 0 && target < start:
			return nil, fmt.Errorf("migrations between %d (current) and %d (target) are invalid as the '%s' provider only has migrations starting at %d meaning the minimum target version when migrating down is %d with the exception of 0", prior, target, provider, start, start)
		case prior < target && target < start:
			return nil, fmt.Errorf("migrations between %d (current) and %d (target) are invalid as the '%s' provider only has migrations starting at %d meaning the minimum target version when migrating up is %d", prior, target, provider, start, start)
		}
	}

	var (
		migration model.SchemaMigration
		entries   []fs.DirEntry
	)

	if entries, err = migrationsFS.ReadDir(path.Join(pathMigrations, provider)); err != nil {
		return nil, err
	}

	up := prior < target

	var filters []migrationFilter

	switch provider {
	case providerMSSQL:
		switch schema {
		case "", "dbo":
			break
		default:
			filters = append(filters, migrationFilterReplace([]byte("[dbo]"), []byte(fmt.Sprintf("[%s]", schema))))
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if migration, err = scanMigration(provider, entry.Name(), filters...); err != nil {
			return nil, err
		}

		if skipMigration(up, start, target, prior, &migration) {
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

func skipMigration(up bool, start, target, prior int, migration *model.SchemaMigration) (skip bool) {
	if up {
		if !migration.Up {
			// Skip if we wanted an Up migration but it isn't an Up migration.
			return true
		}

		if migration.Version > target || migration.Version <= prior {

			// Skip the migration if:
			//  - the version is greater than the target.
			//  - the version less than or equal to the previous version.
			return true
		}
	} else {
		if migration.Up {
			// Skip if we didn't want an Up migration but it is an Up migration.
			return true
		}

		if migration.Version <= target || migration.Version > prior {
			// Skip the migration if:
			//  - the version is less than or equal to the target.
			//  - the version greater than the prior version.
			return true
		}
	}

	return false
}

func scanMigration(providerName, m string, filters ...migrationFilter) (migration model.SchemaMigration, err error) {
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

	for _, filter := range filters {
		data = filter(data)
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

type migrationFilter func(s []byte) []byte

func migrationFilterReplace(old, new []byte) migrationFilter {
	return func(s []byte) []byte {
		return bytes.ReplaceAll(s, old, new)
	}
}
