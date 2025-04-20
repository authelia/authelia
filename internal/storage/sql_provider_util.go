// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"database/sql"
	"fmt"
)

func checkSingleUpdateResult(result sql.Result) (err error) {
	var rowsAffected int64

	switch rowsAffected, err = result.RowsAffected(); {
	case err != nil:
		return fmt.Errorf("error occurred determining the number of affected rows: %w", err)
	case rowsAffected == 0:
		return ErrNoRowsAffected
	case rowsAffected > 1:
		return ErrMultipleRowsAffected
	default:
		return nil
	}
}

func (p *SQLProvider) rebind(transforms ...func(query string) string) {
	rebind := func(query string) string {
		query = p.db.Rebind(query)

		for _, transform := range transforms {
			query = transform(query)
		}

		return query
	}

	p.sqlFmtRenameTable = rebind(p.sqlFmtRenameTable)

	p.sqlSelectPreferred2FAMethod = rebind(p.sqlSelectPreferred2FAMethod)
	p.sqlSelectUserInfo = rebind(p.sqlSelectUserInfo)

	p.sqlInsertUserOpaqueIdentifier = rebind(p.sqlInsertUserOpaqueIdentifier)
	p.sqlSelectUserOpaqueIdentifier = rebind(p.sqlSelectUserOpaqueIdentifier)
	p.sqlSelectUserOpaqueIdentifierBySignature = rebind(p.sqlSelectUserOpaqueIdentifierBySignature)

	p.sqlInsertIdentityVerification = rebind(p.sqlInsertIdentityVerification)
	p.sqlConsumeIdentityVerification = rebind(p.sqlConsumeIdentityVerification)
	p.sqlRevokeIdentityVerification = rebind(p.sqlRevokeIdentityVerification)
	p.sqlSelectIdentityVerification = rebind(p.sqlSelectIdentityVerification)

	p.sqlInsertOneTimeCode = rebind(p.sqlInsertOneTimeCode)
	p.sqlConsumeOneTimeCode = rebind(p.sqlConsumeOneTimeCode)
	p.sqlRevokeOneTimeCode = rebind(p.sqlRevokeOneTimeCode)
	p.sqlSelectOneTimeCode = rebind(p.sqlSelectOneTimeCode)
	p.sqlSelectOneTimeCodeBySignature = rebind(p.sqlSelectOneTimeCodeBySignature)
	p.sqlSelectOneTimeCodeByID = rebind(p.sqlSelectOneTimeCodeByID)
	p.sqlSelectOneTimeCodeByPublicID = rebind(p.sqlSelectOneTimeCodeByPublicID)

	p.sqlSelectTOTPConfig = rebind(p.sqlSelectTOTPConfig)
	p.sqlUpdateTOTPConfigRecordSignIn = rebind(p.sqlUpdateTOTPConfigRecordSignIn)
	p.sqlUpdateTOTPConfigRecordSignInByUsername = rebind(p.sqlUpdateTOTPConfigRecordSignInByUsername)
	p.sqlDeleteTOTPConfig = rebind(p.sqlDeleteTOTPConfig)
	p.sqlSelectTOTPConfigs = rebind(p.sqlSelectTOTPConfigs)

	p.sqlInsertTOTPHistory = rebind(p.sqlInsertTOTPHistory)
	p.sqlSelectTOTPHistory = rebind(p.sqlSelectTOTPHistory)

	p.sqlInsertWebAuthnUser = rebind(p.sqlInsertWebAuthnUser)
	p.sqlSelectWebAuthnUser = rebind(p.sqlSelectWebAuthnUser)
	p.sqlSelectWebAuthnUserByUserID = rebind(p.sqlSelectWebAuthnUserByUserID)

	p.sqlInsertWebAuthnCredential = rebind(p.sqlInsertWebAuthnCredential)
	p.sqlSelectWebAuthnCredentials = rebind(p.sqlSelectWebAuthnCredentials)
	p.sqlSelectWebAuthnCredentialsByUsername = rebind(p.sqlSelectWebAuthnCredentialsByUsername)
	p.sqlSelectWebAuthnCredentialsByRPIDByUsername = rebind(p.sqlSelectWebAuthnCredentialsByRPIDByUsername)
	p.sqlSelectWebAuthnCredentialByID = rebind(p.sqlSelectWebAuthnCredentialByID)
	p.sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID = rebind(p.sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID)
	p.sqlUpdateWebAuthnCredentialRecordSignIn = rebind(p.sqlUpdateWebAuthnCredentialRecordSignIn)
	p.sqlDeleteWebAuthnCredential = rebind(p.sqlDeleteWebAuthnCredential)
	p.sqlDeleteWebAuthnCredentialByUsername = rebind(p.sqlDeleteWebAuthnCredentialByUsername)
	p.sqlDeleteWebAuthnCredentialByUsernameAndDisplayName = rebind(p.sqlDeleteWebAuthnCredentialByUsernameAndDisplayName)

	p.sqlSelectDuoDevice = rebind(p.sqlSelectDuoDevice)
	p.sqlDeleteDuoDevice = rebind(p.sqlDeleteDuoDevice)

	p.sqlInsertAuthenticationAttempt = rebind(p.sqlInsertAuthenticationAttempt)
	p.sqlSelectAuthenticationLogsRegulationRecordsByUsername = rebind(p.sqlSelectAuthenticationLogsRegulationRecordsByUsername)
	p.sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP = rebind(p.sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP)

	p.sqlInsertBannedUser = rebind(p.sqlInsertBannedUser)
	p.sqlSelectBannedUser = rebind(p.sqlSelectBannedUser)
	p.sqlSelectBannedUserByID = rebind(p.sqlSelectBannedUserByID)
	p.sqlSelectBannedUsers = rebind(p.sqlSelectBannedUsers)
	p.sqlSelectBannedUserLastTime = rebind(p.sqlSelectBannedUserLastTime)
	p.sqlRevokeBannedUser = rebind(p.sqlRevokeBannedUser)

	p.sqlInsertBannedIP = rebind(p.sqlInsertBannedIP)
	p.sqlSelectBannedIP = rebind(p.sqlSelectBannedIP)
	p.sqlSelectBannedIPByID = rebind(p.sqlSelectBannedIPByID)
	p.sqlSelectBannedIPs = rebind(p.sqlSelectBannedIPs)
	p.sqlSelectBannedIPLastTime = rebind(p.sqlSelectBannedIPLastTime)
	p.sqlRevokeBannedIP = rebind(p.sqlRevokeBannedIP)

	p.sqlSelectCachedData = rebind(p.sqlSelectCachedData)
	p.sqlDeleteCachedData = rebind(p.sqlDeleteCachedData)

	p.sqlSelectSession = rebind(p.sqlSelectSession)
	p.sqlSelectSessionByPublicID = rebind(p.sqlSelectSessionByPublicID)
	p.sqlSelectSessionSignatureByPublicID = rebind(p.sqlSelectSessionSignatureByPublicID)
	p.sqlSelectSessionSignaturesByUsername = rebind(p.sqlSelectSessionSignaturesByUsername)
	p.sqlUpdateSessionData = rebind(p.sqlUpdateSessionData)
	p.sqlUpdateSessionSignature = rebind(p.sqlUpdateSessionSignature)
	p.sqlUpdateSessionDestroyed = rebind(p.sqlUpdateSessionDestroyed)
	p.sqlDeleteSessionExpired = rebind(p.sqlDeleteSessionExpired)

	p.sqlInsertMigration = rebind(p.sqlInsertMigration)
	p.sqlSelectMigrations = rebind(p.sqlSelectMigrations)
	p.sqlSelectLatestMigration = rebind(p.sqlSelectLatestMigration)

	p.sqlInsertEncryptionValue = rebind(p.sqlInsertEncryptionValue)
	p.sqlSelectEncryptionValue = rebind(p.sqlSelectEncryptionValue)

	p.sqlSelectOAuth2ConsentPreConfigurations = rebind(p.sqlSelectOAuth2ConsentPreConfigurations)

	p.sqlInsertOAuth2ConsentSession = rebind(p.sqlInsertOAuth2ConsentSession)
	p.sqlUpdateOAuth2ConsentSessionResponseByID = rebind(p.sqlUpdateOAuth2ConsentSessionResponseByID)
	p.sqlUpdateOAuth2ConsentSessionResponseByChallengeID = rebind(p.sqlUpdateOAuth2ConsentSessionResponseByChallengeID)
	p.sqlUpdateOAuth2ConsentSessionGranted = rebind(p.sqlUpdateOAuth2ConsentSessionGranted)
	p.sqlSelectOAuth2ConsentSessionByChallengeID = rebind(p.sqlSelectOAuth2ConsentSessionByChallengeID)

	p.sqlInsertOAuth2AccessTokenSession = rebind(p.sqlInsertOAuth2AccessTokenSession)
	p.sqlRevokeOAuth2AccessTokenSession = rebind(p.sqlRevokeOAuth2AccessTokenSession)
	p.sqlRevokeOAuth2AccessTokenSessionByRequestID = rebind(p.sqlRevokeOAuth2AccessTokenSessionByRequestID)
	p.sqlDeactivateOAuth2AccessTokenSession = rebind(p.sqlDeactivateOAuth2AccessTokenSession)
	p.sqlDeactivateOAuth2AccessTokenSessionByRequestID = rebind(p.sqlDeactivateOAuth2AccessTokenSessionByRequestID)
	p.sqlSelectOAuth2AccessTokenSession = rebind(p.sqlSelectOAuth2AccessTokenSession)

	p.sqlInsertOAuth2AuthorizeCodeSession = rebind(p.sqlInsertOAuth2AuthorizeCodeSession)
	p.sqlRevokeOAuth2AuthorizeCodeSession = rebind(p.sqlRevokeOAuth2AuthorizeCodeSession)
	p.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID = rebind(p.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID)
	p.sqlDeactivateOAuth2AuthorizeCodeSession = rebind(p.sqlDeactivateOAuth2AuthorizeCodeSession)
	p.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID = rebind(p.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID)
	p.sqlSelectOAuth2AuthorizeCodeSession = rebind(p.sqlSelectOAuth2AuthorizeCodeSession)

	p.sqlInsertOAuth2DeviceCodeSession = rebind(p.sqlInsertOAuth2DeviceCodeSession)
	p.sqlSelectOAuth2DeviceCodeSession = rebind(p.sqlSelectOAuth2DeviceCodeSession)
	p.sqlUpdateOAuth2DeviceCodeSession = rebind(p.sqlUpdateOAuth2DeviceCodeSession)
	p.sqlUpdateOAuth2DeviceCodeSessionData = rebind(p.sqlUpdateOAuth2DeviceCodeSessionData)
	p.sqlDeactivateOAuth2DeviceCodeSession = rebind(p.sqlDeactivateOAuth2DeviceCodeSession)
	p.sqlSelectOAuth2DeviceCodeSessionByUserCode = rebind(p.sqlSelectOAuth2DeviceCodeSessionByUserCode)

	p.sqlInsertOAuth2OpenIDConnectSession = rebind(p.sqlInsertOAuth2OpenIDConnectSession)
	p.sqlRevokeOAuth2OpenIDConnectSession = rebind(p.sqlRevokeOAuth2OpenIDConnectSession)
	p.sqlRevokeOAuth2OpenIDConnectSessionByRequestID = rebind(p.sqlRevokeOAuth2OpenIDConnectSessionByRequestID)
	p.sqlDeactivateOAuth2OpenIDConnectSession = rebind(p.sqlDeactivateOAuth2OpenIDConnectSession)
	p.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID = rebind(p.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID)
	p.sqlSelectOAuth2OpenIDConnectSession = rebind(p.sqlSelectOAuth2OpenIDConnectSession)

	p.sqlInsertOAuth2PARContext = rebind(p.sqlInsertOAuth2PARContext)
	p.sqlUpdateOAuth2PARContext = rebind(p.sqlUpdateOAuth2PARContext)
	p.sqlRevokeOAuth2PARContext = rebind(p.sqlRevokeOAuth2PARContext)
	p.sqlSelectOAuth2PARContext = rebind(p.sqlSelectOAuth2PARContext)

	p.sqlInsertOAuth2PKCERequestSession = rebind(p.sqlInsertOAuth2PKCERequestSession)
	p.sqlRevokeOAuth2PKCERequestSession = rebind(p.sqlRevokeOAuth2PKCERequestSession)
	p.sqlRevokeOAuth2PKCERequestSessionByRequestID = rebind(p.sqlRevokeOAuth2PKCERequestSessionByRequestID)
	p.sqlDeactivateOAuth2PKCERequestSession = rebind(p.sqlDeactivateOAuth2PKCERequestSession)
	p.sqlDeactivateOAuth2PKCERequestSessionByRequestID = rebind(p.sqlDeactivateOAuth2PKCERequestSessionByRequestID)
	p.sqlSelectOAuth2PKCERequestSession = rebind(p.sqlSelectOAuth2PKCERequestSession)

	p.sqlInsertOAuth2RefreshTokenSession = rebind(p.sqlInsertOAuth2RefreshTokenSession)
	p.sqlRevokeOAuth2RefreshTokenSession = rebind(p.sqlRevokeOAuth2RefreshTokenSession)
	p.sqlRevokeOAuth2RefreshTokenSessionByRequestID = rebind(p.sqlRevokeOAuth2RefreshTokenSessionByRequestID)
	p.sqlDeactivateOAuth2RefreshTokenSession = rebind(p.sqlDeactivateOAuth2RefreshTokenSession)
	p.sqlDeactivateOAuth2RefreshTokenSessionByRequestID = rebind(p.sqlDeactivateOAuth2RefreshTokenSessionByRequestID)
	p.sqlSelectOAuth2RefreshTokenSession = rebind(p.sqlSelectOAuth2RefreshTokenSession)
	p.sqlSelectOAuth2RefreshTokenSessionAccessSignature = rebind(p.sqlSelectOAuth2RefreshTokenSessionAccessSignature)

	p.sqlSelectOAuth2BlacklistedJTI = rebind(p.sqlSelectOAuth2BlacklistedJTI)
}
