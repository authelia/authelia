package storage

const (
	queryFmtRenameTable = `
		ALTER TABLE %s
		RENAME TO %s;`

	queryFmtMySQLRenameTable = `
		ALTER TABLE %s
		RENAME %s;`
)
