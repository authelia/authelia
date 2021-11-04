package storage

// SchemaMigration represents an intended migration.
type SchemaMigration struct {
	Version  int
	Name     string
	Provider string
	Up       bool
	Query    string
}
