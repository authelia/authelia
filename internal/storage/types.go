package storage

// SchemaMigration represents an intended migration.
type SchemaMigration struct {
	Version  int
	Name     string
	Provider string
	Up       bool
	Query    string
}

func (m SchemaMigration) After() (after int) {
	if m.Up {
		return m.Version
	}

	return m.Version - 1
}
