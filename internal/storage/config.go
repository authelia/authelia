package storage

// rebind the SQL statements for the specific provider.
func (p *SQLProvider) rebind() {
	p.sqlFmtRenameTable = p.db.Rebind(p.sqlFmtRenameTable)
	p.sqlSelectPreferred2FAMethodByUsername = p.db.Rebind(p.sqlSelectPreferred2FAMethodByUsername)
	p.sqlSelectExistsIdentityVerification = p.db.Rebind(p.sqlSelectExistsIdentityVerification)
	p.sqlInsertIdentityVerification = p.db.Rebind(p.sqlInsertIdentityVerification)
	p.sqlDeleteIdentityVerification = p.db.Rebind(p.sqlDeleteIdentityVerification)
	p.sqlSelectTOTPConfigByUsername = p.db.Rebind(p.sqlSelectTOTPConfigByUsername)
	p.sqlUpsertTOTPConfig = p.db.Rebind(p.sqlUpsertTOTPConfig)
	p.sqlDeleteTOTPConfig = p.db.Rebind(p.sqlDeleteTOTPConfig)
	p.sqlSelectU2FDeviceByUsername = p.db.Rebind(p.sqlSelectU2FDeviceByUsername)
	p.sqlInsertAuthenticationAttempt = p.db.Rebind(p.sqlInsertAuthenticationAttempt)
	p.sqlSelectAuthenticationAttemptsByUsername = p.db.Rebind(p.sqlSelectAuthenticationAttemptsByUsername)
	p.sqlInsertMigration = p.db.Rebind(p.sqlInsertMigration)
}
