package commands

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	host            string
	validFrom       string
	validFor        time.Duration
	isCA            bool
	rsaBits         int
	ecdsaCurve      string
	ed25519Key      bool
	targetDirectory string
)

func init() {
	CertificatesGenerateCmd.PersistentFlags().StringVar(&host, "host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	err := CertificatesGenerateCmd.MarkPersistentFlagRequired("host")

	if err != nil {
		log.Fatal(err)
	}

	CertificatesGenerateCmd.PersistentFlags().StringVar(&validFrom, "start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	CertificatesGenerateCmd.PersistentFlags().DurationVar(&validFor, "duration", 365*24*time.Hour, "Duration that certificate is valid for")
	CertificatesGenerateCmd.PersistentFlags().BoolVar(&isCA, "ca", false, "Whether this cert should be its own Certificate Authority")
	CertificatesGenerateCmd.PersistentFlags().IntVar(&rsaBits, "rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	CertificatesGenerateCmd.PersistentFlags().StringVar(&ecdsaCurve, "ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521")
	CertificatesGenerateCmd.PersistentFlags().BoolVar(&ed25519Key, "ed25519", false, "Generate an Ed25519 key")
	CertificatesGenerateCmd.PersistentFlags().StringVar(&targetDirectory, "dir", "", "Target directory where the certificate and keys will be stored")

	CertificatesCmd.AddCommand(CertificatesGenerateCmd)
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

//nolint:gocyclo // TODO: Consider refactoring/simplifying, time permitting.
func generateSelfSignedCertificate(cmd *cobra.Command, args []string) {
	// implementation retrieved from https://golang.org/src/crypto/tls/generate_cert.go
	var priv interface{}

	var err error

	switch ecdsaCurve {
	case "":
		if ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, rsaBits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		log.Fatalf("Unrecognized elliptic curve: %q", ecdsaCurve)
	}

	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	var notBefore time.Time
	if len(validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", validFrom)
		if err != nil {
			log.Fatalf("Failed to parse creation date: %v", err)
		}
	}

	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	certPath := path.Join(targetDirectory, "cert.pem")
	certOut, err := os.Create(certPath)

	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", certPath, err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to cert.pem: %v", err)
	}

	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", certPath, err)
	}

	log.Printf("wrote %s\n", certPath)

	keyPath := path.Join(targetDirectory, "key.pem")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", keyPath, err)
		return
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v", keyPath, err)
	}

	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", keyPath, err)
	}

	log.Printf("wrote %s\n", keyPath)
}

// CertificatesCmd certificate helper command.
var CertificatesCmd = &cobra.Command{
	Use:   "certificates",
	Short: "Commands related to certificate generation",
}

// CertificatesGenerateCmd certificate generation command.
var CertificatesGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a self-signed certificate",
	Run:   generateSelfSignedCertificate,
}
