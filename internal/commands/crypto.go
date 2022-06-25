package commands

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newCryptoCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "crypto",
		Short:   cmdAutheliaCryptoShort,
		Long:    cmdAutheliaCryptoLong,
		Example: cmdAutheliaCryptoExample,
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		newCryptoCertCmd(),
		newCryptoPairCmd(),
	)

	return cmd
}

func newCryptoGenerateCmd(category, algorithm string) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:  "generate",
		Args: cobra.NoArgs,
		RunE: cryptoGenerateRunE,
	}

	cryptoGenFlags(cmd)

	switch category {
	case cmdUseCertificate:
		cryptoCertificateGenFlags(cmd)

		switch algorithm {
		case cmdUseRSA:
			cmd.Short, cmd.Long = cmdAutheliaCryptoCertRSAGenerateShort, cmdAutheliaCryptoCertRSAGenerateLong
			cmd.Example = cmdAutheliaCryptoCertRSAGenerateExample

			cryptoRSAGenFlags(cmd)
		case cmdUseECDSA:
			cmd.Short, cmd.Long = cmdAutheliaCryptoCertECDSAGenerateShort, cmdAutheliaCryptoCertECDSAGenerateLong
			cmd.Example = cmdAutheliaCryptoCertECDSAGenerateExample

			cryptoECDSAGenFlags(cmd)
		case cmdUseEd25519:
			cmd.Short, cmd.Long = cmdAutheliaCryptoCertEd25519GenerateShort, cmdAutheliaCryptoCertEd25519GenerateLong
			cmd.Example = cmdAutheliaCryptoCertEd25519GenerateExample

			cryptoEd25519Flags(cmd)
		}
	case cmdUsePair:
		cryptoPairGenFlags(cmd)

		switch algorithm {
		case cmdUseRSA:
			cmd.Short, cmd.Long = cmdAutheliaCryptoPairRSAGenerateShort, cmdAutheliaCryptoPairRSAGenerateLong
			cmd.Example = cmdAutheliaCryptoPairRSAGenerateExample

			cryptoRSAGenFlags(cmd)
		case cmdUseECDSA:
			cmd.Short, cmd.Long = cmdAutheliaCryptoPairECDSAGenerateShort, cmdAutheliaCryptoPairECDSAGenerateLong
			cmd.Example = cmdAutheliaCryptoPairECDSAGenerateExample

			cryptoECDSAGenFlags(cmd)
		case cmdUseEd25519:
			cmd.Short, cmd.Long = cmdAutheliaCryptoPairEd25519GenerateShort, cmdAutheliaCryptoPairEd25519GenerateLong
			cmd.Example = cmdAutheliaCryptoPairEd25519GenerateExample

			cryptoEd25519Flags(cmd)
		}
	}

	return cmd
}

func newCryptoCertCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseCertificate,
		Short:   cmdAutheliaCryptoCertShort,
		Long:    cmdAutheliaCryptoCertLong,
		Example: cmdAutheliaCryptoCertExample,
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		newCryptoCertRSACmd(),
		newCryptoCertECDSACmd(),
		newCryptoCertEd25519Cmd(),
	)

	return cmd
}

func newCryptoCertRSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseRSA,
		Short:   cmdAutheliaCryptoCertRSAShort,
		Long:    cmdAutheliaCryptoCertRSALong,
		Example: cmdAutheliaCryptoCertRSAExample,
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(newCryptoGenerateCmd(cmdUseCertificate, cmdUseRSA))

	return cmd
}

func newCryptoCertECDSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseECDSA,
		Short:   cmdAutheliaCryptoCertECDSAShort,
		Long:    cmdAutheliaCryptoCertECDSALong,
		Example: cmdAutheliaCryptoCertECDSAExample,
		Args:    cobra.NoArgs,
		RunE:    cryptoGenerateRunE,
	}

	cmd.AddCommand(newCryptoGenerateCmd(cmdUseCertificate, cmdUseECDSA))

	return cmd
}

func newCryptoCertEd25519Cmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseEd25519,
		Short:   cmdAutheliaCryptoCertEd25519Short,
		Long:    cmdAutheliaCryptoCertEd25519Long,
		Example: cmdAutheliaCryptoCertEd25519Example,
		Args:    cobra.NoArgs,
		RunE:    cryptoGenerateRunE,
	}

	cmd.AddCommand(newCryptoGenerateCmd(cmdUseCertificate, cmdUseEd25519))

	return cmd
}

func newCryptoPairCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUsePair,
		Short:   cmdAutheliaCryptoPairShort,
		Long:    cmdAutheliaCryptoPairLong,
		Example: cmdAutheliaCryptoPairExample,
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		newCryptoPairRSACmd(),
		newCryptoPairECDSACmd(),
		newCryptoPairEd25519Cmd(),
	)

	return cmd
}

func newCryptoPairRSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseRSA,
		Short:   cmdAutheliaCryptoPairRSAShort,
		Long:    cmdAutheliaCryptoPairRSALong,
		Example: cmdAutheliaCryptoPairRSAExample,
		Args:    cobra.NoArgs,
		RunE:    cryptoGenerateRunE,
	}

	cmd.AddCommand(newCryptoGenerateCmd(cmdUsePair, cmdUseRSA))

	return cmd
}

func newCryptoPairECDSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseECDSA,
		Short:   cmdAutheliaCryptoPairECDSAShort,
		Long:    cmdAutheliaCryptoPairECDSALong,
		Example: cmdAutheliaCryptoPairECDSAExample,
		Args:    cobra.NoArgs,
		RunE:    cryptoGenerateRunE,
	}

	cmd.AddCommand(newCryptoGenerateCmd(cmdUsePair, cmdUseECDSA))

	return cmd
}

func newCryptoPairEd25519Cmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseEd25519,
		Short:   cmdAutheliaCryptoPairEd25519Short,
		Long:    cmdAutheliaCryptoPairEd25519Long,
		Example: cmdAutheliaCryptoPairEd25519Example,
		Args:    cobra.NoArgs,
		RunE:    cryptoGenerateRunE,
	}

	cmd.AddCommand(newCryptoGenerateCmd(cmdUsePair, cmdUseEd25519))

	return cmd
}

func cryptoGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		privateKey interface{}
	)

	if privateKey, err = cryptoGenPrivateKeyFromCmd(cmd); err != nil {
		return err
	}

	if cmd.Parent().Parent().Use == cmdUseCertificate {
		return cryptoCertificateGenerateRunE(cmd, args, privateKey)
	}

	return cryptoPairGenerateRunE(cmd, args, privateKey)
}

func cryptoPairGenerateRunE(cmd *cobra.Command, _ []string, privateKey interface{}) (err error) {
	var (
		privateKeyPath, publicKeyPath string
		pkcs8                         bool
	)

	if pkcs8, err = cmd.Flags().GetBool(cmdFlagNamePKCS8); err != nil {
		return err
	}

	if privateKeyPath, publicKeyPath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
		return err
	}

	b := strings.Builder{}

	b.WriteString("Generating key pair\n\n")

	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		b.WriteString(fmt.Sprintf("\tAlgorithm: RSA-%d %d bits\n\n", k.Size(), k.N.BitLen()))
	case *ecdsa.PrivateKey:
		b.WriteString(fmt.Sprintf("\tAlgorithm: ECDSA Curve %s\n\n", k.Curve.Params().Name))
	case ed25519.PrivateKey:
		b.WriteString("\tAlgorithm: Ed25519\n\n")
	}

	b.WriteString("Output Paths:\n")
	b.WriteString(fmt.Sprintf("\tPrivate Key: %s\n", privateKeyPath))
	b.WriteString(fmt.Sprintf("\tPublic Key: %s\n\n", publicKeyPath))

	fmt.Print(b.String())

	b.Reset()

	if err = utils.WriteKeyToPEM(privateKey, privateKeyPath, pkcs8); err != nil {
		return err
	}

	var publicKey interface{}

	if publicKey = utils.PublicKeyFromPrivateKey(privateKey); publicKey == nil {
		return fmt.Errorf("failed to obtain public key from private key")
	}

	if err = utils.WriteKeyToPEM(publicKey, publicKeyPath, pkcs8); err != nil {
		return err
	}

	return nil
}

func cryptoCertificateGenerateRunE(cmd *cobra.Command, args []string, privateKey interface{}) (err error) {
	var csr bool

	if csr, err = cmd.Flags().GetBool(cmdFlagNameCSR); err != nil {
		return err
	}

	if csr {
		return cryptoCertificateSigningRequestGenerateRunE(cmd, args, privateKey)
	}

	var (
		template, caCertificate, parent       *x509.Certificate
		publicKey, caPrivateKey, signatureKey interface{}
	)

	publicKey = utils.PublicKeyFromPrivateKey(privateKey)

	if caPrivateKey, caCertificate, err = cryptoGetCAFromCmd(cmd); err != nil {
		return err
	}

	signatureKey = privateKey

	if caPrivateKey != nil {
		signatureKey = caPrivateKey
	}

	if template, err = cryptoGetCertificateFromCmd(cmd); err != nil {
		return err
	}

	b := strings.Builder{}

	b.WriteString("Generating Certificate\n\n")

	b.WriteString(fmt.Sprintf("\tSerial: %x\n\n", template.SerialNumber))

	switch caCertificate {
	case nil:
		parent = template

		b.WriteString("Signed By:\n\tSelf-Signed\n")
	default:
		parent = caCertificate

		b.WriteString(fmt.Sprintf("Signed By:\n\t%s\n", caCertificate.Subject.CommonName))
		b.WriteString(fmt.Sprintf("\tSerial: %x, Expires: %s\n", caCertificate.SerialNumber, caCertificate.NotAfter.Format(time.RFC3339)))
	}

	b.WriteString("\nSubject:\n")
	b.WriteString(fmt.Sprintf("\tCommon Name: %s, Organization: %s, Organizational Unit: %s\n", template.Subject.CommonName, template.Subject.Organization, template.Subject.OrganizationalUnit))
	b.WriteString(fmt.Sprintf("\tCountry: %v, Province: %v, Street Address: %v, Postal Code: %v, Locality: %v\n\n", template.Subject.Country, template.Subject.Province, template.Subject.StreetAddress, template.Subject.PostalCode, template.Subject.Locality))

	b.WriteString("Properties:\n")
	b.WriteString(fmt.Sprintf("\tNot Before: %s, Not After: %s\n", template.NotBefore.Format(time.RFC3339), template.NotAfter.Format(time.RFC3339)))

	b.WriteString(fmt.Sprintf("\tCA: %v, CSR: %v, Signature Algorithm: %s, Public Key Algorithm: %s", template.IsCA, false, template.SignatureAlgorithm, template.PublicKeyAlgorithm))

	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		b.WriteString(fmt.Sprintf(", Bits: %d", k.N.BitLen()))
	case *ecdsa.PrivateKey:
		b.WriteString(fmt.Sprintf(", Elliptic Curve: %s", k.Curve.Params().Name))
	}

	b.WriteString(fmt.Sprintf("\n\tSubject Alternative Names: %s\n\n", strings.Join(cryptoSANsToString(template.DNSNames, template.IPAddresses), ", ")))

	var (
		privateKeyPath, certificatePath string
		certificate                     []byte
	)

	if privateKeyPath, certificatePath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
		return err
	}

	b.WriteString("Output Paths:\n")
	b.WriteString(fmt.Sprintf("\tPrivate Key: %s\n", privateKeyPath))
	b.WriteString(fmt.Sprintf("\tCertificate: %s\n\n", certificatePath))

	fmt.Print(b.String())

	b.Reset()

	if certificate, err = x509.CreateCertificate(rand.Reader, template, parent, publicKey, signatureKey); err != nil {
		return err
	}

	if err = utils.WriteKeyToPEM(privateKey, privateKeyPath, false); err != nil {
		return err
	}

	if err = utils.WriteCertificateBytesToPEM(certificate, certificatePath, false); err != nil {
		return err
	}

	return nil
}

func cryptoCertificateSigningRequestGenerateRunE(cmd *cobra.Command, _ []string, privateKey interface{}) (err error) {
	var (
		template                *x509.CertificateRequest
		csr                     []byte
		privateKeyPath, csrPath string
	)

	if template, err = cryptoGetCSRFromCmd(cmd); err != nil {
		return err
	}

	b := strings.Builder{}

	b.WriteString("Generating Certificate Signing Request\n\n")

	b.WriteString("Subject:\n")
	b.WriteString(fmt.Sprintf("\tCommon Name: %s, Organization: %s, Organizational Unit: %s\n", template.Subject.CommonName, template.Subject.Organization, template.Subject.OrganizationalUnit))
	b.WriteString(fmt.Sprintf("\tCountry: %v, Province: %v, Street Address: %v, Postal Code: %v, Locality: %v\n\n", template.Subject.Country, template.Subject.Province, template.Subject.StreetAddress, template.Subject.PostalCode, template.Subject.Locality))

	b.WriteString("Properties:\n")

	b.WriteString(fmt.Sprintf("\tSignature Algorithm: %s, Public Key Algorithm: %s", template.SignatureAlgorithm, template.PublicKeyAlgorithm))

	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		b.WriteString(fmt.Sprintf(", Bits: %d", k.N.BitLen()))
	case *ecdsa.PrivateKey:
		b.WriteString(fmt.Sprintf(", Elliptic Curve: %s", k.Curve.Params().Name))
	}

	b.WriteString(fmt.Sprintf("\n\tSubject Alternative Names: %s\n\n", strings.Join(cryptoSANsToString(template.DNSNames, template.IPAddresses), ", ")))

	if privateKeyPath, csrPath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
		return err
	}

	b.WriteString("Output Paths:\n")
	b.WriteString(fmt.Sprintf("\tPrivate Key: %s\n", privateKeyPath))
	b.WriteString(fmt.Sprintf("\tCertificate Signing Request: %s\n\n", csrPath))

	fmt.Print(b.String())

	b.Reset()

	if csr, err = x509.CreateCertificateRequest(rand.Reader, template, privateKey); err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	if privateKeyPath, csrPath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
		return err
	}

	if err = utils.WriteKeyToPEM(privateKey, privateKeyPath, false); err != nil {
		return err
	}

	if err = utils.WriteCertificateBytesToPEM(csr, csrPath, true); err != nil {
		return err
	}

	return nil
}
