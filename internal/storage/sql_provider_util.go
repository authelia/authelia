package storage

import "fmt"

func fmtQuery(query, table, provider, schema string) string {
	switch provider {
	case providerMSSQL:
		return fmt.Sprintf(query, fmt.Sprintf("%s.%s", schema, table))
	default:
		return fmt.Sprintf(query, table)
	}
}

func (p *SQLProvider) rebind() {
	p.sqlFmtRenameTable = p.db.Rebind(p.sqlFmtRenameTable)

	p.sqlSelectPreferred2FAMethod = p.db.Rebind(p.sqlSelectPreferred2FAMethod)
	p.sqlSelectUserInfo = p.db.Rebind(p.sqlSelectUserInfo)

	p.sqlInsertUserOpaqueIdentifier = p.db.Rebind(p.sqlInsertUserOpaqueIdentifier)
	p.sqlSelectUserOpaqueIdentifier = p.db.Rebind(p.sqlSelectUserOpaqueIdentifier)
	p.sqlSelectUserOpaqueIdentifierBySignature = p.db.Rebind(p.sqlSelectUserOpaqueIdentifierBySignature)

	p.sqlInsertIdentityVerification = p.db.Rebind(p.sqlInsertIdentityVerification)
	p.sqlConsumeIdentityVerification = p.db.Rebind(p.sqlConsumeIdentityVerification)
	p.sqlRevokeIdentityVerification = p.db.Rebind(p.sqlRevokeIdentityVerification)
	p.sqlSelectIdentityVerification = p.db.Rebind(p.sqlSelectIdentityVerification)

	p.sqlInsertOneTimeCode = p.db.Rebind(p.sqlInsertOneTimeCode)
	p.sqlConsumeOneTimeCode = p.db.Rebind(p.sqlConsumeOneTimeCode)
	p.sqlRevokeOneTimeCode = p.db.Rebind(p.sqlRevokeOneTimeCode)
	p.sqlSelectOneTimeCode = p.db.Rebind(p.sqlSelectOneTimeCode)
	p.sqlSelectOneTimeCodeBySignature = p.db.Rebind(p.sqlSelectOneTimeCodeBySignature)
	p.sqlSelectOneTimeCodeByID = p.db.Rebind(p.sqlSelectOneTimeCodeByID)
	p.sqlSelectOneTimeCodeByPublicID = p.db.Rebind(p.sqlSelectOneTimeCodeByPublicID)

	p.sqlSelectTOTPConfig = p.db.Rebind(p.sqlSelectTOTPConfig)
	p.sqlUpdateTOTPConfigRecordSignIn = p.db.Rebind(p.sqlUpdateTOTPConfigRecordSignIn)
	p.sqlUpdateTOTPConfigRecordSignInByUsername = p.db.Rebind(p.sqlUpdateTOTPConfigRecordSignInByUsername)
	p.sqlDeleteTOTPConfig = p.db.Rebind(p.sqlDeleteTOTPConfig)
	p.sqlSelectTOTPConfigs = p.db.Rebind(p.sqlSelectTOTPConfigs)

	p.sqlInsertTOTPHistory = p.db.Rebind(p.sqlInsertTOTPHistory)
	p.sqlSelectTOTPHistory = p.db.Rebind(p.sqlSelectTOTPHistory)

	p.sqlInsertWebAuthnUser = p.db.Rebind(p.sqlInsertWebAuthnUser)
	p.sqlSelectWebAuthnUser = p.db.Rebind(p.sqlSelectWebAuthnUser)
	p.sqlSelectWebAuthnUserByUserID = p.db.Rebind(p.sqlSelectWebAuthnUserByUserID)

	p.sqlInsertWebAuthnCredential = p.db.Rebind(p.sqlInsertWebAuthnCredential)
	p.sqlSelectWebAuthnCredentials = p.db.Rebind(p.sqlSelectWebAuthnCredentials)
	p.sqlSelectWebAuthnCredentialsByUsername = p.db.Rebind(p.sqlSelectWebAuthnCredentialsByUsername)
	p.sqlSelectWebAuthnCredentialsByRPIDByUsername = p.db.Rebind(p.sqlSelectWebAuthnCredentialsByRPIDByUsername)
	p.sqlSelectWebAuthnCredentialByID = p.db.Rebind(p.sqlSelectWebAuthnCredentialByID)
	p.sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID = p.db.Rebind(p.sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID)
	p.sqlUpdateWebAuthnCredentialRecordSignIn = p.db.Rebind(p.sqlUpdateWebAuthnCredentialRecordSignIn)
	p.sqlDeleteWebAuthnCredential = p.db.Rebind(p.sqlDeleteWebAuthnCredential)
	p.sqlDeleteWebAuthnCredentialByUsername = p.db.Rebind(p.sqlDeleteWebAuthnCredentialByUsername)
	p.sqlDeleteWebAuthnCredentialByUsernameAndDisplayName = p.db.Rebind(p.sqlDeleteWebAuthnCredentialByUsernameAndDisplayName)

	p.sqlSelectDuoDevice = p.db.Rebind(p.sqlSelectDuoDevice)
	p.sqlDeleteDuoDevice = p.db.Rebind(p.sqlDeleteDuoDevice)

	p.sqlInsertAuthenticationAttempt = p.db.Rebind(p.sqlInsertAuthenticationAttempt)
	p.sqlSelectAuthenticationLogsRegulationRecordsByUsername = p.db.Rebind(p.sqlSelectAuthenticationLogsRegulationRecordsByUsername)
	p.sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP = p.db.Rebind(p.sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP)

	p.sqlInsertBannedUser = p.db.Rebind(p.sqlInsertBannedUser)
	p.sqlSelectBannedUser = p.db.Rebind(p.sqlSelectBannedUser)
	p.sqlSelectBannedUserByID = p.db.Rebind(p.sqlSelectBannedUserByID)
	p.sqlSelectBannedUsers = p.db.Rebind(p.sqlSelectBannedUsers)
	p.sqlSelectBannedUserLastTime = p.db.Rebind(p.sqlSelectBannedUserLastTime)
	p.sqlRevokeBannedUser = p.db.Rebind(p.sqlRevokeBannedUser)

	p.sqlInsertBannedIP = p.db.Rebind(p.sqlInsertBannedIP)
	p.sqlSelectBannedIP = p.db.Rebind(p.sqlSelectBannedIP)
	p.sqlSelectBannedIPByID = p.db.Rebind(p.sqlSelectBannedIPByID)
	p.sqlSelectBannedIPs = p.db.Rebind(p.sqlSelectBannedIPs)
	p.sqlSelectBannedIPLastTime = p.db.Rebind(p.sqlSelectBannedIPLastTime)
	p.sqlRevokeBannedIP = p.db.Rebind(p.sqlRevokeBannedIP)

	p.sqlSelectCachedData = p.db.Rebind(p.sqlSelectCachedData)
	p.sqlDeleteCachedData = p.db.Rebind(p.sqlDeleteCachedData)

	p.sqlInsertMigration = p.db.Rebind(p.sqlInsertMigration)
	p.sqlSelectMigrations = p.db.Rebind(p.sqlSelectMigrations)
	p.sqlSelectLatestMigration = p.db.Rebind(p.sqlSelectLatestMigration)

	p.sqlSelectEncryptionValue = p.db.Rebind(p.sqlSelectEncryptionValue)

	p.sqlSelectOAuth2ConsentPreConfigurations = p.db.Rebind(p.sqlSelectOAuth2ConsentPreConfigurations)

	p.sqlInsertOAuth2ConsentSession = p.db.Rebind(p.sqlInsertOAuth2ConsentSession)
	p.sqlUpdateOAuth2ConsentSessionSubject = p.db.Rebind(p.sqlUpdateOAuth2ConsentSessionSubject)
	p.sqlUpdateOAuth2ConsentSessionResponse = p.db.Rebind(p.sqlUpdateOAuth2ConsentSessionResponse)
	p.sqlUpdateOAuth2ConsentSessionGranted = p.db.Rebind(p.sqlUpdateOAuth2ConsentSessionGranted)
	p.sqlSelectOAuth2ConsentSessionByChallengeID = p.db.Rebind(p.sqlSelectOAuth2ConsentSessionByChallengeID)

	p.sqlInsertOAuth2AccessTokenSession = p.db.Rebind(p.sqlInsertOAuth2AccessTokenSession)
	p.sqlRevokeOAuth2AccessTokenSession = p.db.Rebind(p.sqlRevokeOAuth2AccessTokenSession)
	p.sqlRevokeOAuth2AccessTokenSessionByRequestID = p.db.Rebind(p.sqlRevokeOAuth2AccessTokenSessionByRequestID)
	p.sqlDeactivateOAuth2AccessTokenSession = p.db.Rebind(p.sqlDeactivateOAuth2AccessTokenSession)
	p.sqlDeactivateOAuth2AccessTokenSessionByRequestID = p.db.Rebind(p.sqlDeactivateOAuth2AccessTokenSessionByRequestID)
	p.sqlSelectOAuth2AccessTokenSession = p.db.Rebind(p.sqlSelectOAuth2AccessTokenSession)

	p.sqlInsertOAuth2AuthorizeCodeSession = p.db.Rebind(p.sqlInsertOAuth2AuthorizeCodeSession)
	p.sqlRevokeOAuth2AuthorizeCodeSession = p.db.Rebind(p.sqlRevokeOAuth2AuthorizeCodeSession)
	p.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID = p.db.Rebind(p.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID)
	p.sqlDeactivateOAuth2AuthorizeCodeSession = p.db.Rebind(p.sqlDeactivateOAuth2AuthorizeCodeSession)
	p.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID = p.db.Rebind(p.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID)
	p.sqlSelectOAuth2AuthorizeCodeSession = p.db.Rebind(p.sqlSelectOAuth2AuthorizeCodeSession)

	p.sqlInsertOAuth2DeviceCodeSession = p.db.Rebind(p.sqlInsertOAuth2DeviceCodeSession)
	p.sqlSelectOAuth2DeviceCodeSession = p.db.Rebind(p.sqlSelectOAuth2DeviceCodeSession)
	p.sqlUpdateOAuth2DeviceCodeSession = p.db.Rebind(p.sqlUpdateOAuth2DeviceCodeSession)
	p.sqlDeactivateOAuth2DeviceCodeSession = p.db.Rebind(p.sqlDeactivateOAuth2DeviceCodeSession)
	p.sqlSelectOAuth2DeviceCodeSessionByUserCode = p.db.Rebind(p.sqlSelectOAuth2DeviceCodeSessionByUserCode)

	p.sqlInsertOAuth2OpenIDConnectSession = p.db.Rebind(p.sqlInsertOAuth2OpenIDConnectSession)
	p.sqlRevokeOAuth2OpenIDConnectSession = p.db.Rebind(p.sqlRevokeOAuth2OpenIDConnectSession)
	p.sqlRevokeOAuth2OpenIDConnectSessionByRequestID = p.db.Rebind(p.sqlRevokeOAuth2OpenIDConnectSessionByRequestID)
	p.sqlDeactivateOAuth2OpenIDConnectSession = p.db.Rebind(p.sqlDeactivateOAuth2OpenIDConnectSession)
	p.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID = p.db.Rebind(p.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID)
	p.sqlSelectOAuth2OpenIDConnectSession = p.db.Rebind(p.sqlSelectOAuth2OpenIDConnectSession)

	p.sqlInsertOAuth2PARContext = p.db.Rebind(p.sqlInsertOAuth2PARContext)
	p.sqlUpdateOAuth2PARContext = p.db.Rebind(p.sqlUpdateOAuth2PARContext)
	p.sqlRevokeOAuth2PARContext = p.db.Rebind(p.sqlRevokeOAuth2PARContext)
	p.sqlSelectOAuth2PARContext = p.db.Rebind(p.sqlSelectOAuth2PARContext)

	p.sqlInsertOAuth2PKCERequestSession = p.db.Rebind(p.sqlInsertOAuth2PKCERequestSession)
	p.sqlRevokeOAuth2PKCERequestSession = p.db.Rebind(p.sqlRevokeOAuth2PKCERequestSession)
	p.sqlRevokeOAuth2PKCERequestSessionByRequestID = p.db.Rebind(p.sqlRevokeOAuth2PKCERequestSessionByRequestID)
	p.sqlDeactivateOAuth2PKCERequestSession = p.db.Rebind(p.sqlDeactivateOAuth2PKCERequestSession)
	p.sqlDeactivateOAuth2PKCERequestSessionByRequestID = p.db.Rebind(p.sqlDeactivateOAuth2PKCERequestSessionByRequestID)
	p.sqlSelectOAuth2PKCERequestSession = p.db.Rebind(p.sqlSelectOAuth2PKCERequestSession)

	p.sqlInsertOAuth2RefreshTokenSession = p.db.Rebind(p.sqlInsertOAuth2RefreshTokenSession)
	p.sqlRevokeOAuth2RefreshTokenSession = p.db.Rebind(p.sqlRevokeOAuth2RefreshTokenSession)
	p.sqlRevokeOAuth2RefreshTokenSessionByRequestID = p.db.Rebind(p.sqlRevokeOAuth2RefreshTokenSessionByRequestID)
	p.sqlDeactivateOAuth2RefreshTokenSession = p.db.Rebind(p.sqlDeactivateOAuth2RefreshTokenSession)
	p.sqlDeactivateOAuth2RefreshTokenSessionByRequestID = p.db.Rebind(p.sqlDeactivateOAuth2RefreshTokenSessionByRequestID)
	p.sqlSelectOAuth2RefreshTokenSession = p.db.Rebind(p.sqlSelectOAuth2RefreshTokenSession)

	p.sqlSelectOAuth2BlacklistedJTI = p.db.Rebind(p.sqlSelectOAuth2BlacklistedJTI)
}
