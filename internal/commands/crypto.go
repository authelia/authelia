package commands

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewCryptoCmd creates a new crypto command.
func NewCryptoCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "crypto",
		Short: "Commands related to cryptography",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		newCryptoCertificateCmd(),
		newCryptoPairCmd(),
	)

	return cmd
}

func newCryptoCertificateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "cert",
		Short: "Commands related to certificates",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		newCryptoRSACertCmd(),
		newCryptoECDSACertCmd(),
		newCryptoEd25519CertCmd(),
	)

	return cmd
}

func newCryptoPairCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "pair",
		Short: "Commands related to key pairs",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		newCryptoRSACmd(),
		newCryptoECDSACmd(),
		newCryptoEd25519Cmd(),
	)

	return cmd
}

func newCryptoRSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "rsa",
		Short: "Generate RSA key pairs",
		Args:  cobra.NoArgs,
		RunE:  cryptoGenRunE,
	}

	cryptoGenFlags(cmd)
	cryptoPairGenFlags(cmd)
	cryptoRSAGenFlags(cmd)

	return cmd
}

func newCryptoECDSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "ecdsa",
		Short: "Generate ECDSA key pairs",
		Args:  cobra.NoArgs,
		RunE:  cryptoGenRunE,
	}

	cryptoGenFlags(cmd)
	cryptoPairGenFlags(cmd)
	cryptoECDSAGenFlags(cmd)

	return cmd
}

func newCryptoEd25519Cmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "ed25519",
		Short: "Generate Ed25519 key pairs",
		Args:  cobra.NoArgs,
		RunE:  cryptoGenRunE,
	}

	cryptoGenFlags(cmd)
	cryptoPairGenFlags(cmd)

	return cmd
}

func newCryptoRSACertCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "rsa",
		Short: "Generate RSA certificates",
		Args:  cobra.NoArgs,
		RunE:  cryptoGenRunE,
	}

	cryptoGenFlags(cmd)
	cryptoCertificateGenFlags(cmd)
	cryptoRSAGenFlags(cmd)

	return cmd
}

func newCryptoECDSACertCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "ecdsa",
		Short: "Generate ECDSA certificates",
		Args:  cobra.NoArgs,
		RunE:  cryptoGenRunE,
	}

	cryptoGenFlags(cmd)
	cryptoCertificateGenFlags(cmd)
	cryptoECDSAGenFlags(cmd)

	return cmd
}

func newCryptoEd25519CertCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "ed25519",
		Short: "Generate Ed25519 certificates",
		Args:  cobra.NoArgs,
		RunE:  cryptoGenRunE,
	}

	cryptoGenFlags(cmd)
	cryptoCertificateGenFlags(cmd)

	return cmd
}

func cryptoGenRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		privateKey interface{}
	)

	privateKey, err = cryptoGenPrivateKeyFromCmd(cmd)
	if err != nil {
		return err
	}

	if cmd.Parent().Use == "cert" {
		return cryptoCertificateGenRunE(cmd, args, privateKey)
	}

	return cryptoPairGenRunE(cmd, args, privateKey)
}

func cryptoPairGenRunE(cmd *cobra.Command, _ []string, newPrivateKey interface{}) (err error) {
	fmt.Printf("Generating key pair\n\n")

	switch privateKey := newPrivateKey.(type) {
	case *rsa.PrivateKey:
		fmt.Printf("\tAlgorithm: RSA-%d %d bits\n\n", privateKey.Size(), privateKey.N.BitLen())
	case *ecdsa.PrivateKey:
		fmt.Printf("\tAlgorithm: ECDSA Curve %s\n\n", privateKey.Curve.Params().Name)
	case ed25519.PrivateKey:
		fmt.Printf("\tAlgorithm: Ed25519\n\n")
	}

	privateKeyPath, publicKeyPath, err := cryptoGetWritePathsFromCmd(cmd)
	if err != nil {
		return err
	}

	pkcs8, _ := cmd.Flags().GetBool("pkcs8")

	fmt.Printf("Writing private key to %s\n", privateKeyPath)

	err = utils.WriteKeyToPEM(newPrivateKey, privateKeyPath, pkcs8)
	if err != nil {
		return err
	}

	newPublicKey := utils.PublicKeyFromPrivateKey(newPrivateKey)
	if newPublicKey == nil {
		return fmt.Errorf("failed to obtain public key from private key")
	}

	fmt.Printf("Writing public key to %s\n", publicKeyPath)

	err = utils.WriteKeyToPEM(newPublicKey, publicKeyPath, pkcs8)
	if err != nil {
		return err
	}

	return nil
}

func cryptoCertificateGenRunE(cmd *cobra.Command, args []string, newPrivateKey interface{}) (err error) {
	isCSR, err := cmd.Flags().GetBool("create-csr")
	if err != nil {
		return err
	}

	if isCSR {
		return cryptoCertificateGenCSRRunE(cmd, args, newPrivateKey)
	}

	var (
		template, caCert, parent *x509.Certificate
		priv, pub, caPrivateKey  interface{}
		data                     []byte
	)

	priv, pub = newPrivateKey, utils.PublicKeyFromPrivateKey(newPrivateKey)

	caPrivateKey, caCert, err = cryptoGetCAFromCmd(cmd)
	if err != nil {
		return err
	}

	if caPrivateKey != nil {
		priv = caPrivateKey
	}

	template, err = cryptoGetCertificateFromCmd(cmd)
	if err != nil {
		return err
	}

	sans := make([]string, len(template.DNSNames)+len(template.IPAddresses))

	j := 0

	for i, dns := range template.DNSNames {
		sans[i] = fmt.Sprintf("DNS.%d:%s", i+1, dns)
		j = i
	}

	for i, ip := range template.IPAddresses {
		sans[i+j+1] = fmt.Sprintf("IP.%d:%s", i+1, ip)
	}

	fmt.Printf("Generating Certificate with serial %x\n\n", template.SerialNumber)

	switch caCert {
	case nil:
		parent = template

		fmt.Println("\tSigned By: Self-Signed")
		fmt.Println("")
	default:
		parent = caCert

		fmt.Printf("Signed By: %s\n", caCert.Subject.CommonName)
		fmt.Printf("\tSerial: %x, Expires: %v\n\n", caCert.SerialNumber, caCert.NotAfter)
	}

	fmt.Println("Subject:")
	fmt.Printf("\tCommon Name: %s, Organization: %s, Organizational Unit: %s\n", template.Subject.CommonName, template.Subject.Organization, template.Subject.OrganizationalUnit)
	fmt.Printf("\tCountry: %v, Province: %v, Street Address: %v, Postal Code: %v, Locality: %v\n\n", template.Subject.Country, template.Subject.Province, template.Subject.StreetAddress, template.Subject.PostalCode, template.Subject.Locality)

	fmt.Println("Properties:")
	fmt.Printf("\tNot Before: %v, Not After: %v\n", template.NotBefore, template.NotAfter)

	var extra string

	switch privateKey := newPrivateKey.(type) {
	case *rsa.PrivateKey:
		extra = fmt.Sprintf(", Bits: %d", privateKey.N.BitLen())
	case *ecdsa.PrivateKey:
		extra = fmt.Sprintf(", Elliptic Curve: %s", privateKey.Curve.Params().Name)
	}

	fmt.Printf("\tCA: %v, CSR: %v, Signature Algorithm: %s, Public Key Algorithm: %s%s\n", template.IsCA, isCSR, template.SignatureAlgorithm, template.PublicKeyAlgorithm, extra)
	fmt.Printf("\tSubject Alternative Names: %s\n\n", strings.Join(sans, ", "))

	data, err = x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		return err
	}

	privateKeyPath, certificatePath, err := cryptoGetWritePathsFromCmd(cmd)
	if err != nil {
		return err
	}

	fmt.Printf("Writing private key to %s\n", privateKeyPath)

	err = utils.WriteKeyToPEM(newPrivateKey, privateKeyPath, false)
	if err != nil {
		return err
	}

	fmt.Printf("Writing certificate to %s\n", certificatePath)

	err = utils.WriteCertificateBytesToPEM(data, certificatePath, false)
	if err != nil {
		return err
	}

	return nil
}

func cryptoCertificateGenCSRRunE(cmd *cobra.Command, _ []string, newPrivateKey interface{}) (err error) {
	var (
		template *x509.CertificateRequest
		data     []byte
	)

	template, err = cryptoGetCSRFromCmd(cmd)
	if err != nil {
		return err
	}

	data, err = x509.CreateCertificateRequest(rand.Reader, template, newPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	key, err := cmd.Flags().GetString("key")
	if err != nil {
		return err
	}

	csr, err := cmd.Flags().GetString("csr")
	if err != nil {
		return err
	}

	err = utils.WriteKeyToPEM(newPrivateKey, filepath.Join(dir, key), false)
	if err != nil {
		return err
	}

	err = utils.WriteCertificateBytesToPEM(data, filepath.Join(dir, csr), false)
	if err != nil {
		return err
	}

	return nil
}
