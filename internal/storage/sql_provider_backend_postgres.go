package storage

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// PostgreSQLProvider is a PostgreSQL provider.
type PostgreSQLProvider struct {
	SQLProvider
}

// NewPostgreSQLProvider a PostgreSQL provider.
func NewPostgreSQLProvider(config *schema.Configuration, caCertPool *x509.CertPool) (provider *PostgreSQLProvider) {
	provider = &PostgreSQLProvider{
		SQLProvider: NewSQLProvider(config, providerPostgres, "pgx", dsnPostgreSQL(config.Storage.PostgreSQL, caCertPool)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryPostgreSelectExistingTables

	// Specific alterations to this provider.
	// PostgreSQL doesn't have a UPSERT statement but has an ON CONFLICT operation instead.
	provider.sqlUpsertDuoDevice = fmt.Sprintf(queryFmtUpsertDuoDevicePostgreSQL, tableDuoDevices)
	provider.sqlUpsertTOTPConfig = fmt.Sprintf(queryFmtUpsertTOTPConfigurationPostgreSQL, tableTOTPConfigurations)
	provider.sqlUpsertPreferred2FAMethod = fmt.Sprintf(queryFmtUpsertPreferred2FAMethodPostgreSQL, tableUserPreferences)
	provider.sqlUpsertEncryptionValue = fmt.Sprintf(queryFmtUpsertEncryptionValuePostgreSQL, tableEncryption)
	provider.sqlUpsertOAuth2BlacklistedJTI = fmt.Sprintf(queryFmtUpsertOAuth2BlacklistedJTIPostgreSQL, tableOAuth2BlacklistedJTI)
	provider.sqlInsertOAuth2ConsentPreConfiguration = fmt.Sprintf(queryFmtInsertOAuth2ConsentPreConfigurationPostgreSQL, tableOAuth2ConsentPreConfiguration)
	provider.sqlUpsertCachedData = fmt.Sprintf(queryFmtUpsertCachedDataPostgreSQL, tableCachedData)
	provider.sqlUpsertSession = fmt.Sprintf(queryFmtUpsertSessionPostgreSQL, tableSessions)

	// PostgreSQL requires rebinding of any query that contains a '?' placeholder to use the '$#' notation placeholders.
	provider.sqlFmtRenameTable = provider.db.Rebind(provider.sqlFmtRenameTable)

	provider.sqlSelectPreferred2FAMethod = provider.db.Rebind(provider.sqlSelectPreferred2FAMethod)
	provider.sqlSelectUserInfo = provider.db.Rebind(provider.sqlSelectUserInfo)

	provider.sqlInsertUserOpaqueIdentifier = provider.db.Rebind(provider.sqlInsertUserOpaqueIdentifier)
	provider.sqlSelectUserOpaqueIdentifier = provider.db.Rebind(provider.sqlSelectUserOpaqueIdentifier)
	provider.sqlSelectUserOpaqueIdentifierBySignature = provider.db.Rebind(provider.sqlSelectUserOpaqueIdentifierBySignature)

	provider.sqlInsertIdentityVerification = provider.db.Rebind(provider.sqlInsertIdentityVerification)
	provider.sqlConsumeIdentityVerification = provider.db.Rebind(provider.sqlConsumeIdentityVerification)
	provider.sqlRevokeIdentityVerification = provider.db.Rebind(provider.sqlRevokeIdentityVerification)
	provider.sqlSelectIdentityVerification = provider.db.Rebind(provider.sqlSelectIdentityVerification)

	provider.sqlInsertOneTimeCode = provider.db.Rebind(provider.sqlInsertOneTimeCode)
	provider.sqlConsumeOneTimeCode = provider.db.Rebind(provider.sqlConsumeOneTimeCode)
	provider.sqlRevokeOneTimeCode = provider.db.Rebind(provider.sqlRevokeOneTimeCode)
	provider.sqlSelectOneTimeCode = provider.db.Rebind(provider.sqlSelectOneTimeCode)
	provider.sqlSelectOneTimeCodeBySignature = provider.db.Rebind(provider.sqlSelectOneTimeCodeBySignature)
	provider.sqlSelectOneTimeCodeByID = provider.db.Rebind(provider.sqlSelectOneTimeCodeByID)
	provider.sqlSelectOneTimeCodeByPublicID = provider.db.Rebind(provider.sqlSelectOneTimeCodeByPublicID)

	provider.sqlSelectTOTPConfig = provider.db.Rebind(provider.sqlSelectTOTPConfig)
	provider.sqlUpdateTOTPConfigRecordSignIn = provider.db.Rebind(provider.sqlUpdateTOTPConfigRecordSignIn)
	provider.sqlUpdateTOTPConfigRecordSignInByUsername = provider.db.Rebind(provider.sqlUpdateTOTPConfigRecordSignInByUsername)
	provider.sqlDeleteTOTPConfig = provider.db.Rebind(provider.sqlDeleteTOTPConfig)
	provider.sqlSelectTOTPConfigs = provider.db.Rebind(provider.sqlSelectTOTPConfigs)

	provider.sqlInsertTOTPHistory = provider.db.Rebind(provider.sqlInsertTOTPHistory)
	provider.sqlSelectTOTPHistory = provider.db.Rebind(provider.sqlSelectTOTPHistory)

	provider.sqlInsertWebAuthnUser = provider.db.Rebind(provider.sqlInsertWebAuthnUser)
	provider.sqlSelectWebAuthnUser = provider.db.Rebind(provider.sqlSelectWebAuthnUser)
	provider.sqlSelectWebAuthnUserByUserID = provider.db.Rebind(provider.sqlSelectWebAuthnUserByUserID)

	provider.sqlInsertWebAuthnCredential = provider.db.Rebind(provider.sqlInsertWebAuthnCredential)
	provider.sqlSelectWebAuthnCredentials = provider.db.Rebind(provider.sqlSelectWebAuthnCredentials)
	provider.sqlSelectWebAuthnCredentialsByUsername = provider.db.Rebind(provider.sqlSelectWebAuthnCredentialsByUsername)
	provider.sqlSelectWebAuthnCredentialsByRPIDByUsername = provider.db.Rebind(provider.sqlSelectWebAuthnCredentialsByRPIDByUsername)
	provider.sqlSelectWebAuthnCredentialByID = provider.db.Rebind(provider.sqlSelectWebAuthnCredentialByID)
	provider.sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID = provider.db.Rebind(provider.sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID)
	provider.sqlUpdateWebAuthnCredentialRecordSignIn = provider.db.Rebind(provider.sqlUpdateWebAuthnCredentialRecordSignIn)
	provider.sqlDeleteWebAuthnCredential = provider.db.Rebind(provider.sqlDeleteWebAuthnCredential)
	provider.sqlDeleteWebAuthnCredentialByUsername = provider.db.Rebind(provider.sqlDeleteWebAuthnCredentialByUsername)
	provider.sqlDeleteWebAuthnCredentialByUsernameAndDisplayName = provider.db.Rebind(provider.sqlDeleteWebAuthnCredentialByUsernameAndDisplayName)

	provider.sqlSelectDuoDevice = provider.db.Rebind(provider.sqlSelectDuoDevice)
	provider.sqlDeleteDuoDevice = provider.db.Rebind(provider.sqlDeleteDuoDevice)

	provider.sqlInsertAuthenticationAttempt = provider.db.Rebind(provider.sqlInsertAuthenticationAttempt)
	provider.sqlSelectAuthenticationLogsRegulationRecordsByUsername = provider.db.Rebind(provider.sqlSelectAuthenticationLogsRegulationRecordsByUsername)
	provider.sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP = provider.db.Rebind(provider.sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP)

	provider.sqlInsertBannedUser = provider.db.Rebind(provider.sqlInsertBannedUser)
	provider.sqlSelectBannedUser = provider.db.Rebind(provider.sqlSelectBannedUser)
	provider.sqlSelectBannedUserByID = provider.db.Rebind(provider.sqlSelectBannedUserByID)
	provider.sqlSelectBannedUsers = provider.db.Rebind(provider.sqlSelectBannedUsers)
	provider.sqlSelectBannedUserLastTime = provider.db.Rebind(provider.sqlSelectBannedUserLastTime)
	provider.sqlRevokeBannedUser = provider.db.Rebind(provider.sqlRevokeBannedUser)

	provider.sqlInsertBannedIP = provider.db.Rebind(provider.sqlInsertBannedIP)
	provider.sqlSelectBannedIP = provider.db.Rebind(provider.sqlSelectBannedIP)
	provider.sqlSelectBannedIPByID = provider.db.Rebind(provider.sqlSelectBannedIPByID)
	provider.sqlSelectBannedIPs = provider.db.Rebind(provider.sqlSelectBannedIPs)
	provider.sqlSelectBannedIPLastTime = provider.db.Rebind(provider.sqlSelectBannedIPLastTime)
	provider.sqlRevokeBannedIP = provider.db.Rebind(provider.sqlRevokeBannedIP)

	provider.sqlSelectCachedData = provider.db.Rebind(provider.sqlSelectCachedData)
	provider.sqlDeleteCachedData = provider.db.Rebind(provider.sqlDeleteCachedData)

	provider.sqlSelectSession = provider.db.Rebind(provider.sqlSelectSession)
	provider.sqlDeleteSession = provider.db.Rebind(provider.sqlDeleteSession)
	provider.sqlDeleteExpiredSessions = provider.db.Rebind(provider.sqlDeleteExpiredSessions)
	provider.sqlCountSessions = provider.db.Rebind(provider.sqlCountSessions)

	provider.sqlInsertMigration = provider.db.Rebind(provider.sqlInsertMigration)
	provider.sqlSelectMigrations = provider.db.Rebind(provider.sqlSelectMigrations)
	provider.sqlSelectLatestMigration = provider.db.Rebind(provider.sqlSelectLatestMigration)

	provider.sqlSelectEncryptionValue = provider.db.Rebind(provider.sqlSelectEncryptionValue)

	provider.sqlSelectOAuth2ConsentPreConfigurations = provider.db.Rebind(provider.sqlSelectOAuth2ConsentPreConfigurations)

	provider.sqlInsertOAuth2ConsentSession = provider.db.Rebind(provider.sqlInsertOAuth2ConsentSession)
	provider.sqlUpdateOAuth2ConsentSessionResponse = provider.db.Rebind(provider.sqlUpdateOAuth2ConsentSessionResponse)
	provider.sqlUpdateOAuth2ConsentSessionGranted = provider.db.Rebind(provider.sqlUpdateOAuth2ConsentSessionGranted)
	provider.sqlSelectOAuth2ConsentSessionByChallengeID = provider.db.Rebind(provider.sqlSelectOAuth2ConsentSessionByChallengeID)

	provider.sqlInsertOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlInsertOAuth2AccessTokenSession)
	provider.sqlRevokeOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlRevokeOAuth2AccessTokenSession)
	provider.sqlRevokeOAuth2AccessTokenSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2AccessTokenSessionByRequestID)
	provider.sqlDeactivateOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlDeactivateOAuth2AccessTokenSession)
	provider.sqlDeactivateOAuth2AccessTokenSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2AccessTokenSessionByRequestID)
	provider.sqlSelectOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlSelectOAuth2AccessTokenSession)

	provider.sqlInsertOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlInsertOAuth2AuthorizeCodeSession)
	provider.sqlRevokeOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlRevokeOAuth2AuthorizeCodeSession)
	provider.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID)
	provider.sqlDeactivateOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlDeactivateOAuth2AuthorizeCodeSession)
	provider.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID)
	provider.sqlSelectOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlSelectOAuth2AuthorizeCodeSession)

	provider.sqlInsertOAuth2DeviceCodeSession = provider.db.Rebind(provider.sqlInsertOAuth2DeviceCodeSession)
	provider.sqlSelectOAuth2DeviceCodeSession = provider.db.Rebind(provider.sqlSelectOAuth2DeviceCodeSession)
	provider.sqlUpdateOAuth2DeviceCodeSession = provider.db.Rebind(provider.sqlUpdateOAuth2DeviceCodeSession)
	provider.sqlUpdateOAuth2DeviceCodeSessionData = provider.db.Rebind(provider.sqlUpdateOAuth2DeviceCodeSessionData)
	provider.sqlDeactivateOAuth2DeviceCodeSession = provider.db.Rebind(provider.sqlDeactivateOAuth2DeviceCodeSession)
	provider.sqlSelectOAuth2DeviceCodeSessionByUserCode = provider.db.Rebind(provider.sqlSelectOAuth2DeviceCodeSessionByUserCode)

	provider.sqlInsertOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlInsertOAuth2OpenIDConnectSession)
	provider.sqlRevokeOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlRevokeOAuth2OpenIDConnectSession)
	provider.sqlRevokeOAuth2OpenIDConnectSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2OpenIDConnectSessionByRequestID)
	provider.sqlDeactivateOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlDeactivateOAuth2OpenIDConnectSession)
	provider.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID)
	provider.sqlSelectOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlSelectOAuth2OpenIDConnectSession)

	provider.sqlInsertOAuth2PARContext = provider.db.Rebind(provider.sqlInsertOAuth2PARContext)
	provider.sqlUpdateOAuth2PARContext = provider.db.Rebind(provider.sqlUpdateOAuth2PARContext)
	provider.sqlRevokeOAuth2PARContext = provider.db.Rebind(provider.sqlRevokeOAuth2PARContext)
	provider.sqlSelectOAuth2PARContext = provider.db.Rebind(provider.sqlSelectOAuth2PARContext)

	provider.sqlInsertOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlInsertOAuth2PKCERequestSession)
	provider.sqlRevokeOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlRevokeOAuth2PKCERequestSession)
	provider.sqlRevokeOAuth2PKCERequestSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2PKCERequestSessionByRequestID)
	provider.sqlDeactivateOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlDeactivateOAuth2PKCERequestSession)
	provider.sqlDeactivateOAuth2PKCERequestSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2PKCERequestSessionByRequestID)
	provider.sqlSelectOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlSelectOAuth2PKCERequestSession)

	provider.sqlInsertOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlInsertOAuth2RefreshTokenSession)
	provider.sqlRevokeOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlRevokeOAuth2RefreshTokenSession)
	provider.sqlRevokeOAuth2RefreshTokenSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2RefreshTokenSessionByRequestID)
	provider.sqlDeactivateOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlDeactivateOAuth2RefreshTokenSession)
	provider.sqlDeactivateOAuth2RefreshTokenSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2RefreshTokenSessionByRequestID)
	provider.sqlSelectOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlSelectOAuth2RefreshTokenSession)

	provider.sqlSelectOAuth2BlacklistedJTI = provider.db.Rebind(provider.sqlSelectOAuth2BlacklistedJTI)

	provider.schema = config.Storage.PostgreSQL.Schema

	return provider
}

func dsnPostgreSQL(config *schema.StoragePostgreSQL, globalCACertPool *x509.CertPool) (dsn string) {
	dsnConfig, _ := pgx.ParseConfig("")

	dsnConfig.Host, dsnConfig.Port = dsnPostgreSQLHostPort(config.Address)
	dsnConfig.Database = config.Database
	dsnConfig.User = config.Username
	dsnConfig.Password = config.Password
	dsnConfig.TLSConfig = loadPostgreSQLTLSConfig(config, globalCACertPool)
	dsnConfig.ConnectTimeout = config.Timeout
	dsnConfig.RuntimeParams = map[string]string{
		"application_name": fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		"search_path":      config.Schema,
	}

	if len(config.Servers) != 0 {
		dsnPostgreSQLFallbacks(config, globalCACertPool, dsnConfig)
	}

	return stdlib.RegisterConnConfig(dsnConfig)
}

func dsnPostgreSQLHostPort(address *schema.AddressTCP) (host string, port uint16) {
	if !address.IsUnixDomainSocket() {
		return address.SocketHostname(), address.Port()
	}

	host, port = address.SocketHostname(), address.Port()

	if port == 0 {
		port = 5432
	}

	dir, base := filepath.Dir(host), filepath.Base(host)

	matches := rePostgreSQLUnixDomainSocket.FindStringSubmatch(base)

	if len(matches) != 2 {
		return host, port
	}

	if raw, err := strconv.ParseUint(matches[1], 10, 16); err == nil {
		host = dir
		port = uint16(raw)
	}

	return host, port
}

func dsnPostgreSQLFallbacks(config *schema.StoragePostgreSQL, globalCACertPool *x509.CertPool, dsnConfig *pgx.ConnConfig) {
	dsnConfig.Fallbacks = make([]*pgconn.FallbackConfig, len(config.Servers))

	for i, server := range config.Servers {
		fallback := &pgconn.FallbackConfig{
			TLSConfig: loadPostgreSQLModernTLSConfig(server.TLS, globalCACertPool),
		}

		fallback.Host, fallback.Port = dsnPostgreSQLHostPort(server.Address)

		if fallback.Port == 0 && !server.Address.IsUnixDomainSocket() {
			fallback.Port = 5432
		}

		dsnConfig.Fallbacks[i] = fallback
	}
}

func loadPostgreSQLTLSConfig(config *schema.StoragePostgreSQL, globalCACertPool *x509.CertPool) (tlsConfig *tls.Config) {
	if config.TLS != nil {
		return loadPostgreSQLModernTLSConfig(config.TLS, globalCACertPool)
	} else if config.SSL != nil { //nolint:staticcheck
		return loadPostgreSQLLegacyTLSConfig(config, globalCACertPool)
	}

	return nil
}

func loadPostgreSQLModernTLSConfig(config *schema.TLS, globalCACertPool *x509.CertPool) (tlsConfig *tls.Config) {
	return utils.NewTLSConfig(config, globalCACertPool)
}

//nolint:staticcheck // Used for legacy purposes.
func loadPostgreSQLLegacyTLSConfig(config *schema.StoragePostgreSQL, globalCACertPool *x509.CertPool) (tlsConfig *tls.Config) {
	var (
		ca    *x509.Certificate
		certs []tls.Certificate
	)

	ca, certs = loadPostgreSQLLegacyTLSConfigFiles(config)

	switch config.SSL.Mode {
	case "disable":
		return nil
	default:
		var caCertPool *x509.CertPool

		switch ca {
		case nil:
			caCertPool = globalCACertPool
		default:
			caCertPool = globalCACertPool
			caCertPool.AddCert(ca)
		}

		tlsConfig = &tls.Config{
			Certificates:       certs,
			RootCAs:            caCertPool,
			InsecureSkipVerify: true, //nolint:gosec
		}

		switch {
		case config.SSL.Mode == "require" && config.SSL.RootCertificate != "" || config.SSL.Mode == "verify-ca":
			tlsConfig.VerifyPeerCertificate = newPostgreSQLVerifyCAFunc(tlsConfig)
		case config.SSL.Mode == "verify-full":
			tlsConfig.InsecureSkipVerify = false
			tlsConfig.ServerName = config.Address.Hostname()
		}
	}

	return tlsConfig
}

//nolint:staticcheck // Used for legacy purposes.
func loadPostgreSQLLegacyTLSConfigFiles(config *schema.StoragePostgreSQL) (ca *x509.Certificate, certs []tls.Certificate) {
	var (
		err error
	)

	if config.SSL.RootCertificate != "" {
		var (
			data  []byte
			block *pem.Block
		)

		if data, err = os.ReadFile(config.SSL.RootCertificate); err != nil {
			return nil, nil
		}

		block, _ = pem.Decode(data)

		if ca, err = x509.ParseCertificate(block.Bytes); err != nil {
			return nil, nil
		}
	}

	if config.SSL.Certificate != "" && config.SSL.Key != "" {
		var (
			dataKey, dataCert []byte
		)

		if dataKey, err = os.ReadFile(config.SSL.Key); err != nil {
			return nil, nil
		}

		if dataCert, err = os.ReadFile(config.SSL.Certificate); err != nil {
			return nil, nil
		}

		var cert tls.Certificate

		if cert, err = tls.X509KeyPair(dataCert, dataKey); err != nil {
			return nil, nil
		}

		certs = []tls.Certificate{cert}
	}

	return ca, certs
}

func newPostgreSQLVerifyCAFunc(config *tls.Config) func(certificates [][]byte, _ [][]*x509.Certificate) (err error) {
	return func(certificates [][]byte, _ [][]*x509.Certificate) (err error) {
		certs := make([]*x509.Certificate, len(certificates))

		var cert *x509.Certificate

		for i, asn1Data := range certificates {
			if cert, err = x509.ParseCertificate(asn1Data); err != nil {
				return errors.New("failed to parse certificate from server: " + err.Error())
			}

			certs[i] = cert
		}

		// Leave DNSName empty to skip hostname verification.
		opts := x509.VerifyOptions{
			Roots:         config.RootCAs,
			Intermediates: x509.NewCertPool(),
		}

		// Skip the first cert because it's the leaf. All others
		// are intermediates.
		for _, cert = range certs[1:] {
			opts.Intermediates.AddCert(cert)
		}

		_, err = certs[0].Verify(opts)

		return err
	}
}
