package storage

type schemaMigration struct {
	Version  int
	Name     string
	Provider string
	Up       bool
	Query    string
}
