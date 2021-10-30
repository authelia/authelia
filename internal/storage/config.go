package storage

// rebind the SQL statements for the specific provider.
func (p *SQLProvider) rebind() {
	p.sqlFmtRenameTable = p.db.Rebind(p.sqlFmtRenameTable)
	p.sqlSelectPreferred2FAMethod = p.db.Rebind(p.sqlSelectPreferred2FAMethod)
	p.sqlSelectExistsIdentityVerification = p.db.Rebind(p.sqlSelectExistsIdentityVerification)
	p.sqlInsertIdentityVerification = p.db.Rebind(p.sqlInsertIdentityVerification)
	p.sqlDeleteIdentityVerification = p.db.Rebind(p.sqlDeleteIdentityVerification)
	p.sqlSelectTOTPConfig = p.db.Rebind(p.sqlSelectTOTPConfig)
	p.sqlUpsertTOTPConfig = p.db.Rebind(p.sqlUpsertTOTPConfig)
	p.sqlDeleteTOTPConfig = p.db.Rebind(p.sqlDeleteTOTPConfig)
	p.sqlSelectU2FDevice = p.db.Rebind(p.sqlSelectU2FDevice)
	p.sqlInsertAuthenticationAttempt = p.db.Rebind(p.sqlInsertAuthenticationAttempt)
	p.sqlSelectAuthenticationAttemptsByUsername = p.db.Rebind(p.sqlSelectAuthenticationAttemptsByUsername)
	p.sqlInsertMigration = p.db.Rebind(p.sqlInsertMigration)
}
