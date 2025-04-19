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
		SQLProvider: NewSQLProvider(config, providerPostgres, "pgx", config.Storage.PostgreSQL.Schema, dsnPostgreSQL(config.Storage.PostgreSQL, caCertPool)),
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

	// PostgreSQL requires rebinding of any query that contains a '?' placeholder to use the '$#' notation placeholders.
	provider.rebind()

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
