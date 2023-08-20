package model

import (
	"strings"
)

// SchemaMigration represents an intended migration.
type SchemaMigration struct {
	Version  int
	Name     string
	Provider string
	Up       bool
	Query    string
}

// NotEmpty returns true if the SchemaMigration is not an empty string.
func (m SchemaMigration) NotEmpty() bool {
	return len(strings.TrimSpace(m.Query)) != 0
}

// Before returns the version the schema should be at Before the migration is applied.
func (m SchemaMigration) Before() (before int) {
	if m.Up {
		return m.Version - 1
	}

	return m.Version
}

// After returns the version the schema will be at After the migration is applied.
func (m SchemaMigration) After() (after int) {
	if m.Up {
		return m.Version
	}

	return m.Version - 1
}
