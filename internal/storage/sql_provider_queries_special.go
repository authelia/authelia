// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

const (
	queryFmtRenameTable = `
		ALTER TABLE %s
		RENAME TO %s;`

	queryFmtMySQLRenameTable = `
		ALTER TABLE %s
		RENAME %s;`

	queryFmtPostgreSQLLockTable = `LOCK TABLE %s IN %s MODE;`

	queryFmtSelectRowCount = `
		SELECT COUNT(id)
		FROM %s;`
)
