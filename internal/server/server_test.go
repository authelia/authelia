package server

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// TemporaryCertificate contains the FD of 2 temporary files containing the PEM format of the certificate and private key.
type TemporaryCertificate struct {
	CertFile *os.File
	KeyFile  *os.File

	Certificate *x509.Certificate

	CertificatePEM []byte
	KeyPEM         []byte
}

func (tc TemporaryCertificate) TLSCertificate() (tls.Certificate, error) {
	return tls.LoadX509KeyPair(tc.CertFile.Name(), tc.KeyFile.Name())
}

func (tc *TemporaryCertificate) Close() {
	if tc.CertFile != nil {
		tc.CertFile.Close()
	}

	if tc.KeyFile != nil {
		tc.KeyFile.Close()
	}
}

type CertificateContext struct {
	Certificates      []TemporaryCertificate
	privateKeyBuilder utils.PrivateKeyBuilder
}

// NewCertificateContext instantiate a new certificate context used to easily generate certificates within tests.
func NewCertificateContext(privateKeyBuilder utils.PrivateKeyBuilder) (*CertificateContext, error) {
	certificateContext := new(CertificateContext)
	certificateContext.privateKeyBuilder = privateKeyBuilder

	cert, err := certificateContext.GenerateCertificate()
	if err != nil {
		return nil, err
	}

	certificateContext.Certificates = []TemporaryCertificate{*cert}

	return certificateContext, nil
}

// GenerateCertificate generate a new certificate in the context.
func (cc *CertificateContext) GenerateCertificate() (*TemporaryCertificate, error) {
	certBytes, keyBytes, err := utils.GenerateCertificate(cc.privateKeyBuilder,
		[]string{"authelia.com", "example.org", "local.example.com"},
		time.Now(), 3*time.Hour, false)
	if err != nil {
		return nil, fmt.Errorf("unable to generate certificate: %v", err)
	}

	tmpCertificate := new(TemporaryCertificate)

	certFile, err := os.CreateTemp("", "cert")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file for certificate: %v", err)
	}

	tmpCertificate.CertFile = certFile
	tmpCertificate.CertificatePEM = certBytes

	block, _ := pem.Decode(certBytes)

	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse certificate: %v", err)
	}

	tmpCertificate.Certificate = c

	err = os.WriteFile(tmpCertificate.CertFile.Name(), certBytes, 0600)
	if err != nil {
		tmpCertificate.Close()
		return nil, fmt.Errorf("unable to write certificates in file: %v", err)
	}

	keyFile, err := os.CreateTemp("", "key")
	if err != nil {
		tmpCertificate.Close()
		return nil, fmt.Errorf("unable to create temp file for private key: %v", err)
	}

	tmpCertificate.KeyFile = keyFile
	tmpCertificate.KeyPEM = keyBytes

	err = os.WriteFile(tmpCertificate.KeyFile.Name(), keyBytes, 0600)
	if err != nil {
		tmpCertificate.Close()
		return nil, fmt.Errorf("unable to write private key in file: %v", err)
	}

	cc.Certificates = append(cc.Certificates, *tmpCertificate)

	return tmpCertificate, nil
}

func (cc *CertificateContext) Close() {
	for _, tc := range cc.Certificates {
		tc.Close()
	}
}

type TLSServerContext struct {
	server *fasthttp.Server
	port   int
}

func NewTLSServerContext(configuration schema.Configuration) (serverContext *TLSServerContext, err error) {
	serverContext = new(TLSServerContext)

	providers := middlewares.NewProvidersBasic()

	providers.Random = random.NewMathematical()

	providers.Templates, err = templates.New(templates.Config{EmailTemplatesPath: configuration.Notifier.TemplatePath})
	if err != nil {
		return nil, err
	}

	s, listener, _, _, err := New(&configuration, providers)
	if err != nil {
		return nil, err
	}

	serverContext.server = s

	go func() {
		err := s.Serve(listener)
		if err != nil {
			logging.Logger().Fatal(err)
		}
	}()

	addrSplit := strings.Split(listener.Addr().String(), ":")
	if len(addrSplit) > 1 {
		port, err := strconv.ParseInt(addrSplit[len(addrSplit)-1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("unable to parse port from address: %v", err)
		}

		serverContext.port = int(port)
	}

	return serverContext, nil
}

func (sc *TLSServerContext) Port() int {
	return sc.port
}

func (sc *TLSServerContext) Close() error {
	return sc.server.Shutdown()
}

func TestShouldRaiseErrorWhenClientDoesNotSkipVerify(t *testing.T) {
	privateKeyBuilder := utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P256())
	certificateContext, err := NewCertificateContext(privateKeyBuilder)
	require.NoError(t, err)

	defer certificateContext.Close()

	tlsServerContext, err := NewTLSServerContext(schema.Configuration{
		Server: schema.Server{
			Address: &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues("tcp", "0.0.0.0", 9091)},
			TLS: schema.ServerTLS{
				Certificate: certificateContext.Certificates[0].CertFile.Name(),
				Key:         certificateContext.Certificates[0].KeyFile.Name(),
			},
		},
	})
	require.NoError(t, err)

	defer tlsServerContext.Close()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("https://local.example.com:%d", tlsServerContext.Port()), nil)
	require.NoError(t, err)

	_, err = http.DefaultClient.Do(req)
	require.Error(t, err)

	require.Contains(t, err.Error(), "x509: certificate signed by unknown authority")
}

func TestShouldServeOverTLSWhenClientDoesSkipVerify(t *testing.T) {
	privateKeyBuilder := utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P256())
	certificateContext, err := NewCertificateContext(privateKeyBuilder)
	require.NoError(t, err)

	defer certificateContext.Close()

	tlsServerContext, err := NewTLSServerContext(schema.Configuration{
		Server: schema.Server{
			Address: schema.DefaultServerConfiguration.Address,
			TLS: schema.ServerTLS{
				Certificate: certificateContext.Certificates[0].CertFile.Name(),
				Key:         certificateContext.Certificates[0].KeyFile.Name(),
			},
		},
	})
	require.NoError(t, err)

	defer tlsServerContext.Close()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("https://local.example.com:%d/api/notfound", tlsServerContext.Port()), nil)
	require.NoError(t, err)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Needs to be enabled in tests. Not used in production.
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "404 Not Found", res.Status)
}

func TestShouldServeOverTLSWhenClientHasProperRootCA(t *testing.T) {
	privateKeyBuilder := utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P256())
	certificateContext, err := NewCertificateContext(privateKeyBuilder)
	require.NoError(t, err)

	defer certificateContext.Close()

	tlsServerContext, err := NewTLSServerContext(schema.Configuration{
		Server: schema.Server{
			Address: schema.DefaultServerConfiguration.Address,
			TLS: schema.ServerTLS{
				Certificate: certificateContext.Certificates[0].CertFile.Name(),
				Key:         certificateContext.Certificates[0].KeyFile.Name(),
			},
		},
	})
	require.NoError(t, err)

	defer tlsServerContext.Close()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("https://local.example.com:%d/api/notfound", tlsServerContext.Port()), nil)
	require.NoError(t, err)

	block, _ := pem.Decode(certificateContext.Certificates[0].CertificatePEM)
	c, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	// Create a root CA for the client to properly validate server cert.
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(c)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    rootCAs,
			MinVersion: tls.VersionTLS13,
		},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "404 Not Found", res.Status)
}

func TestShouldRaiseWhenMutualTLSIsConfiguredAndClientIsNotAuthenticated(t *testing.T) {
	privateKeyBuilder := utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P256())
	certificateContext, err := NewCertificateContext(privateKeyBuilder)
	require.NoError(t, err)

	defer certificateContext.Close()

	clientCert, err := certificateContext.GenerateCertificate()
	require.NoError(t, err)

	tlsServerContext, err := NewTLSServerContext(schema.Configuration{
		Server: schema.Server{
			Address: schema.DefaultServerConfiguration.Address,
			TLS: schema.ServerTLS{
				Certificate:        certificateContext.Certificates[0].CertFile.Name(),
				Key:                certificateContext.Certificates[0].KeyFile.Name(),
				ClientCertificates: []string{clientCert.CertFile.Name()},
			},
		},
	})
	require.NoError(t, err)

	defer tlsServerContext.Close()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("https://local.example.com:%d/api/notfound", tlsServerContext.Port()), nil)
	require.NoError(t, err)

	// Create a root CA for the client to properly validate server cert.
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(certificateContext.Certificates[0].Certificate)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    rootCAs,
			MinVersion: tls.VersionTLS13,
		},
	}
	client := &http.Client{Transport: tr}

	_, err = client.Do(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "remote error: tls: certificate required")
}

func TestShouldServeProperlyWhenMutualTLSIsConfiguredAndClientIsAuthenticated(t *testing.T) {
	privateKeyBuilder := utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P256())
	certificateContext, err := NewCertificateContext(privateKeyBuilder)
	require.NoError(t, err)

	defer certificateContext.Close()

	clientCert, err := certificateContext.GenerateCertificate()
	require.NoError(t, err)

	tlsServerContext, err := NewTLSServerContext(schema.Configuration{
		Server: schema.Server{
			Address: schema.DefaultServerConfiguration.Address,
			TLS: schema.ServerTLS{
				Certificate:        certificateContext.Certificates[0].CertFile.Name(),
				Key:                certificateContext.Certificates[0].KeyFile.Name(),
				ClientCertificates: []string{clientCert.CertFile.Name()},
			},
		},
	})
	require.NoError(t, err)

	defer tlsServerContext.Close()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("https://local.example.com:%d/api/notfound", tlsServerContext.Port()), nil)
	require.NoError(t, err)

	// Create a root CA for the client to properly validate server cert.
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(certificateContext.Certificates[0].Certificate)

	cCert, err := certificateContext.Certificates[1].TLSCertificate()
	require.NoError(t, err)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      rootCAs,
			Certificates: []tls.Certificate{cCert},
			MinVersion:   tls.VersionTLS13,
		},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "404 Not Found", res.Status)
}
