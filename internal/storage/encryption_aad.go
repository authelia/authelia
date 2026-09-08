package storage

import (
	"fmt"
)

// EncryptionAAD provides the Additional Authenticated Data for encrypted storage values.
type EncryptionAAD interface {
	// Get returns the Additional Authenticated Data for a value in the given table and column belonging to the
	// given row.
	Get(table, column, row string) []byte

	// GetIssuer returns the Additional Authenticated Data for a value in the given table and column belonging to
	// the given row and issued by the given issuer.
	GetIssuer(table, column, row, issuer string) []byte

	// IsEncryptionAAD just returns true and ensures this implements the interface.
	IsEncryptionAAD() bool
}

var (
	// aadNone is the scheme used by databases at schema 24 and below which do not use Additional Authenticated Data.
	aadNone EncryptionAAD = EncryptionAADNone{}

	// aadColumn is the scheme used by databases at schema 25 which bind values to their table and column only. Schema
	// 25 was never released so this exists solely to decrypt databases running an unreleased build.
	aadColumn EncryptionAAD = EncryptionAADColumn{}

	// aadRow is the scheme used by databases at schema 26 and above which bind values to their individual row.
	aadRow EncryptionAAD = EncryptionAADRow{}
)

// EncryptionAADNone is the EncryptionAAD implementation which produces no Additional Authenticated Data.
type EncryptionAADNone struct{}

// Get returns no Additional Authenticated Data.
func (EncryptionAADNone) Get(table, column, row string) []byte {
	return nil
}

// GetIssuer returns no Additional Authenticated Data.
func (EncryptionAADNone) GetIssuer(table, column, row, issuer string) []byte {
	return nil
}

// IsEncryptionAAD just returns true and ensures this implements the interface.
func (EncryptionAADNone) IsEncryptionAAD() bool {
	return true
}

// EncryptionAADColumn is the EncryptionAAD implementation which binds values to their table and column.
type EncryptionAADColumn struct{}

// Get returns the Additional Authenticated Data bound to the table and column.
func (EncryptionAADColumn) Get(table, column, row string) []byte {
	return []byte(fmt.Sprintf("authelia:storage:%s:%s", table, column))
}

// GetIssuer returns the Additional Authenticated Data bound to the table, issuer, and column. The issuer
// precedes the column, which is the ordering schema 25 wrote.
func (EncryptionAADColumn) GetIssuer(table, column, row, issuer string) []byte {
	return []byte(fmt.Sprintf("authelia:storage:%s:%s:%s", table, issuer, column))
}

// IsEncryptionAAD just returns true and ensures this implements the interface.
func (EncryptionAADColumn) IsEncryptionAAD() bool {
	return true
}

// EncryptionAADRow is the EncryptionAAD implementation which binds values to their individual row.
type EncryptionAADRow struct{}

// Get returns the Additional Authenticated Data bound to the table, column, and row.
func (EncryptionAADRow) Get(table, column, row string) []byte {
	return []byte(fmt.Sprintf("authelia:storage:%s:%s:%s", table, column, row))
}

// GetIssuer returns the Additional Authenticated Data bound to the table, column, row, and issuer.
func (EncryptionAADRow) GetIssuer(table, column, row, issuer string) []byte {
	return []byte(fmt.Sprintf("authelia:storage:%s:%s:%s:%s", table, column, row, issuer))
}

// IsEncryptionAAD just returns true and ensures this implements the interface.
func (EncryptionAADRow) IsEncryptionAAD() bool {
	return true
}

// aadForSchemaVersion returns the EncryptionAAD implementation used by a database at the given schema version.
func aadForSchemaVersion(version int) EncryptionAAD {
	switch {
	case version < schemaVersionEncryptionKeyDerivation:
		return aadNone
	case version < schemaVersionEncryptionAADRowScoped:
		return aadColumn
	default:
		return aadRow
	}
}
