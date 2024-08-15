package commands

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newCryptoCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseCrypto,
		Short:   cmdAutheliaCryptoShort,
		Long:    cmdAutheliaCryptoLong,
		Example: cmdAutheliaCryptoExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newCryptoRandCmd(ctx),
		newCryptoCertificateCmd(ctx),
		newCryptoHashCmd(ctx),
		newCryptoPairCmd(ctx),
	)

	return cmd
}

func newCryptoRandCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseRand,
		Short:   cmdAutheliaCryptoRandShort,
		Long:    cmdAutheliaCryptoRandLong,
		Example: cmdAutheliaCryptoRandExample,
		Args:    cobra.NoArgs,
		RunE:    ctx.CryptoRandRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(cmdFlagNameCharSet, "x", cmdFlagValueCharSet, cmdFlagUsageCharset)
	cmd.Flags().String(cmdFlagNameCharacters, "", cmdFlagUsageCharacters)
	cmd.Flags().IntP(cmdFlagNameLength, "n", 72, cmdFlagUsageLength)
	cmd.Flags().BoolP(cmdFlagNameQuiet, "q", false, cmdFlagUsageQuiet)
	cmd.Flags().StringP(cmdFlagNameOutputFile, "o", "", cmdFlagUsageOutputFile)
	cmd.Flags().StringP(cmdFlagNameDirectory, "d", "", cmdFlagUsageDirectory)

	return cmd
}

func newCryptoCertificateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUseCertificate,
		Short:   cmdAutheliaCryptoCertificateShort,
		Long:    cmdAutheliaCryptoCertificateLong,
		Example: cmdAutheliaCryptoCertificateExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newCryptoCertificateSubCmd(ctx, cmdUseRSA),
		newCryptoCertificateSubCmd(ctx, cmdUseECDSA),
		newCryptoCertificateSubCmd(ctx, cmdUseEd25519),
	)

	return cmd
}

func newCryptoCertificateSubCmd(ctx *CmdCtx, use string) (cmd *cobra.Command) {
	useFmt := fmtCryptoCertificateUse(use)

	cmd = &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf(fmtCmdAutheliaCryptoCertificateSubShort, useFmt),
		Long:    fmt.Sprintf(fmtCmdAutheliaCryptoCertificateSubLong, useFmt, useFmt),
		Example: fmt.Sprintf(fmtCmdAutheliaCryptoCertificateSubExample, use),
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newCryptoGenerateCmd(ctx, cmdUseCertificate, use), newCryptoCertificateRequestCmd(ctx, use))

	return cmd
}

func newCryptoCertificateRequestCmd(ctx *CmdCtx, algorithm string) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:  cmdUseRequest,
		Args: cobra.NoArgs,
		RunE: ctx.CryptoCertificateRequestRunE,

		DisableAutoGenTag: true,
	}

	cmdFlagsCryptoPrivateKey(cmd)
	cmdFlagsCryptoCertificateCommon(cmd)
	cmdFlagsCryptoCertificateRequest(cmd)

	algorithmFmt := fmtCryptoCertificateUse(algorithm)

	cmd.Short = fmt.Sprintf(fmtCmdAutheliaCryptoCertificateGenerateRequestShort, algorithmFmt, cryptoCertCSROut)
	cmd.Long = fmt.Sprintf(fmtCmdAutheliaCryptoCertificateGenerateRequestLong, algorithmFmt, cryptoCertCSROut, algorithmFmt, cryptoCertCSROut)

	switch algorithm {
	case cmdUseRSA:
		cmd.Example = cmdAutheliaCryptoCertificateRSARequestExample

		cmdFlagsCryptoPrivateKeyRSA(cmd)
	case cmdUseECDSA:
		cmd.Example = cmdAutheliaCryptoCertificateECDSARequestExample

		cmdFlagsCryptoPrivateKeyECDSA(cmd)
	case cmdUseEd25519:
		cmd.Example = cmdAutheliaCryptoCertificateEd25519RequestExample

		cmdFlagsCryptoPrivateKeyEd25519(cmd)
	}

	return cmd
}

func newCryptoPairCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     cmdUsePair,
		Short:   cmdAutheliaCryptoPairShort,
		Long:    cmdAutheliaCryptoPairLong,
		Example: cmdAutheliaCryptoPairExample,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newCryptoPairSubCmd(ctx, cmdUseRSA),
		newCryptoPairSubCmd(ctx, cmdUseECDSA),
		newCryptoPairSubCmd(ctx, cmdUseEd25519),
	)

	return cmd
}

func newCryptoPairSubCmd(ctx *CmdCtx, use string) (cmd *cobra.Command) {
	var (
		example, useFmt string
	)

	useFmt = fmtCryptoCertificateUse(use)

	switch use {
	case cmdUseRSA:
		example = cmdAutheliaCryptoPairRSAExample
	case cmdUseECDSA:
		example = cmdAutheliaCryptoPairECDSAExample
	case cmdUseEd25519:
		example = cmdAutheliaCryptoPairEd25519Example
	}

	cmd = &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf(cmdAutheliaCryptoPairSubShort, useFmt),
		Long:    fmt.Sprintf(cmdAutheliaCryptoPairSubLong, useFmt, useFmt),
		Example: example,
		Args:    cobra.NoArgs,
		RunE:    ctx.CryptoGenerateRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newCryptoGenerateCmd(ctx, cmdUsePair, use))

	return cmd
}

func newCryptoGenerateCmd(ctx *CmdCtx, category, algorithm string) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:  cmdUseGenerate,
		Args: cobra.NoArgs,
		RunE: ctx.CryptoGenerateRunE,

		DisableAutoGenTag: true,
	}

	cmdFlagsCryptoPrivateKey(cmd)

	algorithmFmt := fmtCryptoCertificateUse(algorithm)

	switch category {
	case cmdUseCertificate:
		cmdFlagsCryptoCertificateCommon(cmd)
		cmdFlagsCryptoCertificateGenerate(cmd)

		cmd.Short = fmt.Sprintf(fmtCmdAutheliaCryptoCertificateGenerateRequestShort, algorithmFmt, cryptoCertPubCertOut)
		cmd.Long = fmt.Sprintf(fmtCmdAutheliaCryptoCertificateGenerateRequestLong, algorithmFmt, cryptoCertPubCertOut, algorithmFmt, cryptoCertPubCertOut)

		switch algorithm {
		case cmdUseRSA:
			cmd.Example = cmdAutheliaCryptoCertificateRSAGenerateExample

			cmdFlagsCryptoPrivateKeyRSA(cmd)
		case cmdUseECDSA:
			cmd.Example = cmdAutheliaCryptoCertificateECDSAGenerateExample

			cmdFlagsCryptoPrivateKeyECDSA(cmd)
		case cmdUseEd25519:
			cmd.Example = cmdAutheliaCryptoCertificateEd25519GenerateExample

			cmdFlagsCryptoPrivateKeyEd25519(cmd)
		}
	case cmdUsePair:
		cmdFlagsCryptoPairGenerate(cmd)

		cmd.Short = fmt.Sprintf(fmtCmdAutheliaCryptoPairGenerateShort, algorithmFmt)
		cmd.Long = fmt.Sprintf(fmtCmdAutheliaCryptoPairGenerateLong, algorithmFmt, algorithmFmt)

		switch algorithm {
		case cmdUseRSA:
			cmd.Example = cmdAutheliaCryptoPairRSAGenerateExample

			cmdFlagsCryptoPrivateKeyRSA(cmd)
		case cmdUseECDSA:
			cmd.Example = cmdAutheliaCryptoPairECDSAGenerateExample

			cmdFlagsCryptoPrivateKeyECDSA(cmd)
		case cmdUseEd25519:
			cmd.Example = cmdAutheliaCryptoPairEd25519GenerateExample

			cmdFlagsCryptoPrivateKeyEd25519(cmd)
		}
	}

	return cmd
}

// CryptoRandRunE is the RunE for the authelia crypto rand command.
func (ctx *CmdCtx) CryptoRandRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		random     string
		quiet      bool
		outputPath string
	)

	if random, err = flagsGetRandomCharacters(cmd.Flags(), cmdFlagNameLength, cmdFlagNameCharSet, cmdFlagNameCharacters); err != nil {
		return err
	}

	quiet, _ = cmd.Flags().GetBool(cmdFlagNameQuiet)

	if outputPath, err = flagsGetFilePath(cmd.Flags(), cmdFlagNameOutputFile, cmdFlagNameDirectory); err != nil {
		return fmt.Errorf("error getting output file path: %v", err)
	}

	if outputPath != "" {
		dir := filepath.Dir(outputPath)

		if err = os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		if err = os.WriteFile(outputPath, []byte(random), 0600); err != nil {
			return fmt.Errorf("failed to write random value to file: %v", err)
		}

		if !quiet {
			fmt.Printf("Random value written to file: %s\n", outputPath)
		}
	} else {
		if !quiet {
			fmt.Printf("Random Value: %s\n", random)

			if value := url.QueryEscape(random); random != value {
				fmt.Printf("Random Value (URL Encoded): %s\n", value)
			}
		} else {
			fmt.Println(random)
		}
	}

	return nil
}

// CryptoGenerateRunE is the RunE for the authelia crypto [pair|certificate] [rsa|ecdsa|ed25519] commands.
func (ctx *CmdCtx) CryptoGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		privateKey any
	)

	if privateKey, err = ctx.cryptoGenPrivateKeyFromCmd(cmd); err != nil {
		return err
	}

	if cmd.Parent().Parent().Use == cmdUseCertificate {
		return ctx.CryptoCertificateGenerateRunE(cmd, args, privateKey)
	}

	return ctx.CryptoPairGenerateRunE(cmd, args, privateKey)
}

// CryptoCertificateRequestRunE is the RunE for the authelia crypto certificate request command.
//
//nolint:gocyclo
func (ctx *CmdCtx) CryptoCertificateRequestRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		template                                      *x509.CertificateRequest
		privateKey                                    any
		csr                                           []byte
		privateKeyPath, privateKeyLegacyPath, csrPath string
		legacy                                        bool
		dir, extLegacy                                string
	)

	if legacy, err = cmd.Flags().GetBool(cmdFlagNameLegacy); err != nil {
		return err
	}

	if extLegacy, err = cmd.Flags().GetString(cmdFlagNameFileExtensionLegacy); err != nil {
		return err
	}

	if privateKey, err = ctx.cryptoGenPrivateKeyFromCmd(cmd); err != nil {
		return err
	}

	if template, err = cryptoGetCSRFromCmd(cmd); err != nil {
		return err
	}

	b := strings.Builder{}

	b.WriteString("Generating Certificate Request\n\n")

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
	case ed25519.PrivateKey:
		// Legacy format is not available for Ed25519.
		legacy = false
	}

	b.WriteString(fmt.Sprintf("\n\tSubject Alternative Names: %s\n\n", strings.Join(cryptoSANsToString(template.DNSNames, template.IPAddresses), ", ")))

	if dir, privateKeyPath, csrPath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
		return err
	}

	privateKeyPaths := []string{filepath.Base(privateKeyPath)}

	if legacy {
		if ext := path.Ext(privateKeyPath); len(ext) == 0 {
			privateKeyLegacyPath = fmt.Sprintf("%s.%s", privateKeyPath, extLegacy)
		} else {
			privateKeyLegacyPath = fmt.Sprintf("%s.%s%s", strings.TrimSuffix(privateKeyPath, ext), extLegacy, ext)
		}

		privateKeyPaths = append(privateKeyPaths, filepath.Base(privateKeyLegacyPath))
	}

	b.WriteString("Output Paths:\n")

	if cdir := filepath.Clean(dir); len(cdir) != 0 {
		b.WriteString(fmt.Sprintf("\tDirectory: %s\n", cdir))
	}

	b.WriteString(fmt.Sprintf("\tPrivate Key: %s\n", strings.Join(privateKeyPaths, ", ")))
	b.WriteString(fmt.Sprintf("\tCertificate Request: %s\n\n", filepath.Base(csrPath)))

	if csr, err = x509.CreateCertificateRequest(ctx.providers.Random, template, privateKey); err != nil {
		return fmt.Errorf("failed to create certificate request: %w", err)
	}

	if err = utils.WriteKeyToPEM(privateKey, privateKeyPath, false); err != nil {
		return err
	}

	if legacy {
		if err = utils.WriteKeyToPEM(privateKey, privateKeyLegacyPath, true); err != nil {
			return err
		}
	}

	if err = utils.WriteCertificateBytesAsPEMToPath(csrPath, true, csr); err != nil {
		return err
	}

	b.WriteString("\n")

	fmt.Print(b.String())

	b.Reset()

	return nil
}

// CryptoCertificateGenerateRunE is the RunE for the authelia crypto certificate [rsa|ecdsa|ed25519] commands.
//
//nolint:gocyclo
func (ctx *CmdCtx) CryptoCertificateGenerateRunE(cmd *cobra.Command, _ []string, privateKey any) (err error) {
	var (
		template, caCertificate, parent       *x509.Certificate
		publicKey, caPrivateKey, signatureKey any
		legacy                                bool
		extLegacy                             string
	)

	if legacy, err = cmd.Flags().GetBool(cmdFlagNameLegacy); err != nil {
		return err
	}

	if extLegacy, err = cmd.Flags().GetString(cmdFlagNameFileExtensionLegacy); err != nil {
		return err
	}

	if publicKey = utils.PublicKeyFromPrivateKey(privateKey); publicKey == nil {
		return fmt.Errorf("failed to obtain public key from private key")
	}

	if caPrivateKey, caCertificate, err = cryptoGetCAFromCmd(cmd); err != nil {
		return err
	}

	signatureKey = privateKey

	if caPrivateKey != nil {
		signatureKey = caPrivateKey
	}

	if template, err = ctx.cryptoGetCertificateFromCmd(cmd); err != nil {
		return err
	}

	b := &strings.Builder{}

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
	case ed25519.PrivateKey:
		// Legacy format is not available for Ed25519.
		legacy = false
	}

	b.WriteString(fmt.Sprintf("\n\tSubject Alternative Names: %s\n\n", strings.Join(cryptoSANsToString(template.DNSNames, template.IPAddresses), ", ")))

	var (
		dir, privateKeyPath, privateKeyLegacyPath, certificatePath string

		certificate []byte
	)

	if dir, privateKeyPath, certificatePath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
		return err
	}

	privateKeyPaths := []string{filepath.Base(privateKeyPath)}

	if legacy {
		if ext := path.Ext(privateKeyPath); len(ext) == 0 {
			privateKeyLegacyPath = fmt.Sprintf("%s.%s", privateKeyPath, extLegacy)
		} else {
			privateKeyLegacyPath = fmt.Sprintf("%s.%s%s", strings.TrimSuffix(privateKeyPath, ext), extLegacy, ext)
		}

		privateKeyPaths = append(privateKeyPaths, filepath.Base(privateKeyLegacyPath))
	}

	b.WriteString("Output Paths:\n")

	if cdir := filepath.Clean(dir); len(cdir) != 0 {
		b.WriteString(fmt.Sprintf("\tDirectory: %s\n", cdir))
	}

	b.WriteString(fmt.Sprintf("\tPrivate Key: %s\n", strings.Join(privateKeyPaths, ", ")))
	b.WriteString(fmt.Sprintf("\tCertificate: %s\n", filepath.Base(certificatePath)))

	if certificate, err = x509.CreateCertificate(ctx.providers.Random, template, parent, publicKey, signatureKey); err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	if err = utils.WriteKeyToPEM(privateKey, privateKeyPath, false); err != nil {
		return err
	}

	if legacy {
		if err = utils.WriteKeyToPEM(privateKey, privateKeyLegacyPath, true); err != nil {
			return err
		}
	}

	if err = utils.WriteCertificateBytesAsPEMToPath(certificatePath, false, certificate); err != nil {
		return err
	}

	if cmd.Flags().Changed(cmdFlagNameBundles) {
		if err = cryptoGenerateCertificateBundlesFromCmd(cmd, b, dir, caCertificate, certificate, privateKey); err != nil {
			return err
		}
	}

	b.WriteString("\n")

	fmt.Print(b.String())

	b.Reset()

	return nil
}

// CryptoPairGenerateRunE is the RunE for the authelia crypto pair [rsa|ecdsa|ed25519] commands.
//
//nolint:gocyclo
func (ctx *CmdCtx) CryptoPairGenerateRunE(cmd *cobra.Command, _ []string, privateKey any) (err error) {
	var (
		privateKeyPath, publicKeyPath             string
		privateKeyLegacyPath, publicKeyLegacyPath string
		dir, extLegacy                            string

		legacy bool
	)

	if legacy, err = cmd.Flags().GetBool(cmdFlagNameLegacy); err != nil {
		return err
	}

	if extLegacy, err = cmd.Flags().GetString(cmdFlagNameFileExtensionLegacy); err != nil {
		return err
	}

	if dir, privateKeyPath, publicKeyPath, err = cryptoGetWritePathsFromCmd(cmd); err != nil {
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

		// Legacy format is not available for Ed25519.
		legacy = false
	}

	privateKeyPaths := []string{filepath.Base(privateKeyPath)}
	publicKeyPaths := []string{filepath.Base(publicKeyPath)}

	if legacy {
		if ext := path.Ext(privateKeyPath); len(ext) == 0 {
			privateKeyLegacyPath = fmt.Sprintf("%s.%s", privateKeyPath, extLegacy)
		} else {
			privateKeyLegacyPath = fmt.Sprintf("%s.%s%s", strings.TrimSuffix(privateKeyPath, ext), extLegacy, ext)
		}

		if ext := path.Ext(publicKeyPath); len(ext) == 0 {
			publicKeyLegacyPath = fmt.Sprintf("%s.%s", publicKeyPath, extLegacy)
		} else {
			publicKeyLegacyPath = fmt.Sprintf("%s.%s%s", strings.TrimSuffix(publicKeyPath, ext), extLegacy, ext)
		}

		privateKeyPaths = append(privateKeyPaths, filepath.Base(privateKeyLegacyPath))
		publicKeyPaths = append(publicKeyPaths, filepath.Base(publicKeyLegacyPath))
	}

	b.WriteString("Output Paths:\n")

	if cdir := filepath.Clean(dir); len(cdir) != 0 {
		b.WriteString(fmt.Sprintf("\tDirectory: %s\n", cdir))
	}

	b.WriteString(fmt.Sprintf("\tPrivate Key: %s\n", strings.Join(privateKeyPaths, ", ")))
	b.WriteString(fmt.Sprintf("\tPublic Key: %s\n\n", strings.Join(publicKeyPaths, ", ")))

	if err = utils.WriteKeyToPEM(privateKey, privateKeyPath, false); err != nil {
		return err
	}

	if legacy {
		if err = utils.WriteKeyToPEM(privateKey, privateKeyLegacyPath, true); err != nil {
			return err
		}
	}

	var publicKey any

	if publicKey = utils.PublicKeyFromPrivateKey(privateKey); publicKey == nil {
		return fmt.Errorf("failed to obtain public key from private key")
	}

	if err = utils.WriteKeyToPEM(publicKey, publicKeyPath, false); err != nil {
		return err
	}

	if legacy {
		if err = utils.WriteKeyToPEM(publicKey, publicKeyLegacyPath, true); err != nil {
			return err
		}
	}

	b.WriteString("\n")

	fmt.Print(b.String())

	b.Reset()

	return nil
}
