package commands

import (
	"crypto/elliptic"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"

	"github.com/spf13/cobra"
)

// NewCertificatesCmd returns a new Certificates Cmd.
func NewCertificatesCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "certificates",
		Short: "Commands related to certificate generation",
		Args:  cobra.NoArgs,
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
		Use:   "generate",
		Short: "Generate a self-signed certificate",
		Args:  cobra.NoArgs,
		Run:   cmdCertificatesGenerateRun,
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

func cmdCertificatesGenerateRun(cmd *cobra.Command, _ []string) {
	// implementation retrieved from https://golang.org/src/crypto/tls/generate_cert.go
	ecdsaCurve, err := cmd.Flags().GetString("ecdsa-curve")
	if err != nil {
		fmt.Printf("Failed to parse ecdsa-curve flag: %v\n", err)
		os.Exit(1)
	}

	ed25519Key, err := cmd.Flags().GetBool("ed25519")
	if err != nil {
		fmt.Printf("Failed to parse ed25519 flag: %v\n", err)
		os.Exit(1)
	}

	rsaBits, err := cmd.Flags().GetInt("rsa-bits")
	if err != nil {
		fmt.Printf("Failed to parse rsa-bits flag: %v\n", err)
		os.Exit(1)
	}

	hosts, err := cmd.Flags().GetStringSlice("host")
	if err != nil {
		fmt.Printf("Failed to parse host flag: %v\n", err)
		os.Exit(1)
	}

	validFrom, err := cmd.Flags().GetString("start-date")
	if err != nil {
		fmt.Printf("Failed to parse start-date flag: %v\n", err)
		os.Exit(1)
	}

	validFor, err := cmd.Flags().GetDuration("duration")
	if err != nil {
		fmt.Printf("Failed to parse duration flag: %v\n", err)
		os.Exit(1)
	}

	isCA, err := cmd.Flags().GetBool("ca")
	if err != nil {
		fmt.Printf("Failed to parse ca flag: %v\n", err)
		os.Exit(1)
	}

	certificateTargetDirectory, err := cmd.Flags().GetString("dir")
	if err != nil {
		fmt.Printf("Failed to parse dir flag: %v\n", err)
		os.Exit(1)
	}

	cmdCertificatesGenerateRunExtended(hosts, ecdsaCurve, validFrom, certificateTargetDirectory, ed25519Key, isCA, rsaBits, validFor)
}

func cmdCertificatesGenerateRunExtended(hosts []string, ecdsaCurve, validFrom, certificateTargetDirectory string, ed25519Key, isCA bool, rsaBits int, validFor time.Duration) {
	certPath := filepath.Join(certificateTargetDirectory, "cert.pem")
	keyPath := filepath.Join(certificateTargetDirectory, "key.pem")

	var (
		notBefore time.Time
		err       error
	)

	switch len(validFrom) {
	case 0:
		notBefore = time.Now()
	default:
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", validFrom)
		if err != nil {
			log.Fatalf("failed to parse start date: %v", err)
		}
	}

	var privateKeyBuilder utils.PrivateKeyBuilder

	switch ecdsaCurve {
	case "":
		if ed25519Key {
			privateKeyBuilder = utils.Ed25519KeyBuilder{}
		} else {
			privateKeyBuilder = utils.RSAKeyBuilder{}.WithKeySize(rsaBits)
		}
	case "P224":
		privateKeyBuilder = utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P224())
	case "P256":
		privateKeyBuilder = utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P256())
	case "384":
		privateKeyBuilder = utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P384())
	case "521":
		privateKeyBuilder = utils.ECDSAKeyBuilder{}.WithCurve(elliptic.P521())
	}

	certBytes, keyBytes, err := utils.GenerateCertificate(privateKeyBuilder, hosts, notBefore, validFor, isCA)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(certPath, certBytes, 0600)
	if err != nil {
		log.Fatalf("failed to write %s for writing: %v", certPath, err)
	}

	fmt.Printf("Certificate written to %s\n", certPath)

	err = ioutil.WriteFile(keyPath, keyBytes, 0600)
	if err != nil {
		log.Fatalf("failed to write %s for writing: %v", certPath, err)
	}

	fmt.Printf("Private Key written to %s\n", keyPath)
}
