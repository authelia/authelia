package storage

// rebind the SQL statements for the specific provider.
func (p *SQLProvider) rebind() {
	p.sqlRenameTable = p.db.Rebind(p.sqlRenameTable)
	p.sqlSelectPreferred2FAMethodByUsername = p.db.Rebind(p.sqlSelectPreferred2FAMethodByUsername)
	p.sqlSelectExistsIdentityVerificationToken = p.db.Rebind(p.sqlSelectExistsIdentityVerificationToken)
	p.sqlInsertIdentityVerificationToken = p.db.Rebind(p.sqlInsertIdentityVerificationToken)
	p.sqlDeleteIdentityVerificationToken = p.db.Rebind(p.sqlDeleteIdentityVerificationToken)
	p.sqlSelectTOTPSecretByUsername = p.db.Rebind(p.sqlSelectTOTPSecretByUsername)
	p.sqlUpsertTOTPSecret = p.db.Rebind(p.sqlUpsertTOTPSecret)
	p.sqlDeleteTOTPSecret = p.db.Rebind(p.sqlDeleteTOTPSecret)
	p.sqlSelectU2FDeviceByUsername = p.db.Rebind(p.sqlSelectU2FDeviceByUsername)
	p.sqlInsertAuthenticationAttempt = p.db.Rebind(p.sqlInsertAuthenticationAttempt)
	p.sqlSelectAuthenticationAttemptsByUsername = p.db.Rebind(p.sqlSelectAuthenticationAttemptsByUsername)
}
