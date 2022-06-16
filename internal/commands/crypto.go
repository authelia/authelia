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

	//cryptoGenFlags(cmd)
	//cryptoPairGenFlags(cmd)
	//cryptoRSAGenFlags(cmd)

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

	privateKey, err = cryptoGenPrivateKeyFromCmd(cmd)
	if err != nil {
		return err
	}

	if cmd.Parent().Parent().Use == cmdUseCertificate {
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

	pkcs8, _ := cmd.Flags().GetBool(cmdFlagNamePKCS8)

	fmt.Printf("Writing private key to %s\n", privateKeyPath)

	if err = utils.WriteKeyToPEM(newPrivateKey, privateKeyPath, pkcs8); err != nil {
		return err
	}

	newPublicKey := utils.PublicKeyFromPrivateKey(newPrivateKey)
	if newPublicKey == nil {
		return fmt.Errorf("failed to obtain public key from private key")
	}

	fmt.Printf("Writing public key to %s\n", publicKeyPath)

	if err = utils.WriteKeyToPEM(newPublicKey, publicKeyPath, pkcs8); err != nil {
		return err
	}

	return nil
}

func cryptoCertificateGenRunE(cmd *cobra.Command, args []string, newPrivateKey interface{}) (err error) {
	isCSR, err := cmd.Flags().GetBool(cmdFlagNameCSR)
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

	if caPrivateKey, caCert, err = cryptoGetCAFromCmd(cmd); err != nil {
		return err
	}

	if caPrivateKey != nil {
		priv = caPrivateKey
	}

	if template, err = cryptoGetCertificateFromCmd(cmd); err != nil {
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

	if data, err = x509.CreateCertificate(rand.Reader, template, parent, pub, priv); err != nil {
		return err
	}

	var (
		privateKeyPath, certificatePath string
	)

	if privateKeyPath, certificatePath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
		return err
	}

	fmt.Printf("Writing private key to %s\n", privateKeyPath)

	if err = utils.WriteKeyToPEM(newPrivateKey, privateKeyPath, false); err != nil {
		return err
	}

	fmt.Printf("Writing certificate to %s\n", certificatePath)

	if err = utils.WriteCertificateBytesToPEM(data, certificatePath, false); err != nil {
		return err
	}

	return nil
}

func cryptoCertificateGenCSRRunE(cmd *cobra.Command, _ []string, newPrivateKey interface{}) (err error) {
	var (
		template                           *x509.CertificateRequest
		data                               []byte
		directory, privateKeyFile, csrFile string
	)

	if template, err = cryptoGetCSRFromCmd(cmd); err != nil {
		return err
	}

	if data, err = x509.CreateCertificateRequest(rand.Reader, template, newPrivateKey); err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	if directory, err = cmd.Flags().GetString(cmdFlagNameDirectory); err != nil {
		return err
	}

	if privateKeyFile, err = cmd.Flags().GetString(cmdFlagNameFilePrivateKey); err != nil {
		return err
	}

	if csrFile, err = cmd.Flags().GetString(cmdFlagNameFileCSR); err != nil {
		return err
	}

	if err = utils.WriteKeyToPEM(newPrivateKey, filepath.Join(directory, privateKeyFile), false); err != nil {
		return err
	}

	if err = utils.WriteCertificateBytesToPEM(data, filepath.Join(directory, csrFile), false); err != nil {
		return err
	}

	return nil
}
