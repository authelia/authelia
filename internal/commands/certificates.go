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
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
)

func newCertificatesCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "certificates",
		Short:             "Commands related to certificate generation",
		PersistentPreRunE: nil,
	}

	cmd.PersistentFlags().StringSlice("host", []string{}, "Comma-separated hostnames and IPs to generate a certificate for")

	err := cmd.MarkPersistentFlagRequired("host")
	if err != nil {
		log.Fatal(err)
	}

	cmd.AddCommand(newCertificatesGenerateCmd())

	return cmd
}

func newCertificatesGenerateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "generate",
		Short:             "Generate a self-signed certificate",
		RunE:              cmdCertificatesGenerateRunE,
		PersistentPreRunE: nil,
	}

	cmd.Flags().String("start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	cmd.Flags().Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
	cmd.Flags().Bool("ca", false, "Whether this cert should be its own Certificate Authority")
	cmd.Flags().Int("rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	cmd.Flags().String("ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521")
	cmd.Flags().Bool("ed25519", false, "Generate an Ed25519 key")
	cmd.Flags().String("dir", "", "Target directory where the certificate and keys will be stored")

	return cmd
}

func cmdCertificatesGenerateRunE(cmd *cobra.Command, _ []string) (err error) {
	// implementation retrieved from https://golang.org/src/crypto/tls/generate_cert.go
	var priv interface{}

	ecdsaCurve, err := cmd.Flags().GetString("ecdsa-curve")
	if err != nil {
		return err
	}

	ed25519Key, err := cmd.Flags().GetBool("ed25519")
	if err != nil {
		return err
	}

	rsaBits, err := cmd.Flags().GetInt("rsa-bits")
	if err != nil {
		return err
	}

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
		return fmt.Errorf("Unrecognized elliptic curve: %q", ecdsaCurve)
	}

	if err != nil {
		return fmt.Errorf("Failed to generate private key: %w", err)
	}

	validFrom, err := cmd.Flags().GetString("start-date")
	if err != nil {
		return err
	}

	var notBefore time.Time
	if len(validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", validFrom)
		if err != nil {
			return fmt.Errorf("Failed to parse creation date: %w", err)
		}
	}

	validFor, err := cmd.Flags().GetDuration("duration")
	if err != nil {
		return err
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

	hosts, err := cmd.PersistentFlags().GetStringSlice("host")
	if err != nil {
		return err
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	isCA, err := cmd.Flags().GetBool("ca")
	if err != nil {
		return err
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return fmt.Errorf("Failed to create certificate: %w", err)
	}

	certificateTargetDirectory, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	certPath := path.Join(certificateTargetDirectory, "cert.pem")
	certOut, err := os.Create(certPath)

	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %w", certPath, err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("Failed to write data to cert.pem: %w", err)
	}

	if err := certOut.Close(); err != nil {
		return fmt.Errorf("Error closing %s: %w", certPath, err)
	}

	log.Printf("wrote %s\n", certPath)

	keyPath := path.Join(certificateTargetDirectory, "key.pem")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %w", keyPath, err)
		return
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("Unable to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("Failed to write data to %s: %w", keyPath, err)
	}

	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("Error closing %s: %w", keyPath, err)
	}

	log.Printf("Certificate Key Written to %s\n", keyPath)

	return nil
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
