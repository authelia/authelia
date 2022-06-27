package commands

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func cmdFlagsCryptoCertificateCommon(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNameSignature, "SHA256", "signature algorithm for the certificate")

	cmd.Flags().StringP(cmdFlagNameCommonName, "c", "", "certificate common name")
	cmd.Flags().StringSliceP(cmdFlagNameOrganization, "o", []string{"Authelia"}, "certificate organization")
	cmd.Flags().StringSlice(cmdFlagNameOrganizationalUnit, nil, "certificate organizational unit")
	cmd.Flags().StringSlice(cmdFlagNameCountry, nil, "certificate country")
	cmd.Flags().StringSlice(cmdFlagNameProvince, nil, "certificate province")
	cmd.Flags().StringSliceP(cmdFlagNameLocality, "l", nil, "certificate locality")
	cmd.Flags().StringSliceP(cmdFlagNameStreetAddress, "s", nil, "certificate street address")
	cmd.Flags().StringSliceP(cmdFlagNamePostcode, "p", nil, "certificate postcode")

	cmd.Flags().String(cmdFlagNameNotBefore, "", fmt.Sprintf("earliest date and time the certificate is considered valid formatted as %s (default is now)", timeLayoutCertificateNotBefore))
	cmd.Flags().Duration(cmdFlagNameDuration, 365*24*time.Hour, "duration of time the certificate is valid for")
	cmd.Flags().StringSlice(cmdFlagNameSANs, nil, "subject alternative names")
}

func cmdFlagsCryptoCertificateGenerate(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNamePathCA, "", "source directory of the certificate authority files, if not provided the certificate will be self-signed")
	cmd.Flags().String(cmdFlagNameFileCAPrivateKey, "ca.private.pem", "certificate authority private key to use to signing this certificate")
	cmd.Flags().String(cmdFlagNameFileCACertificate, "ca.public.crt", "certificate authority certificate to use when signing this certificate")
	cmd.Flags().String(cmdFlagNameFileCertificate, "public.crt", "name of the file to export the certificate data to")

	cmd.Flags().StringSlice(cmdFlagNameExtendedUsage, nil, "specify the extended usage types of the certificate")

	cmd.Flags().Bool(cmdFlagNameCA, false, "create the certificate as a certificate authority certificate")
}

func cmdFlagsCryptoCertificateRequest(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNameFileCSR, "request.csr", "name of the file to export the certificate request data to")
}

func cmdFlagsCryptoPairGenerate(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNameFilePublicKey, "public.pem", "name of the file to export the public key data to")
	cmd.Flags().Bool(cmdFlagNamePKCS8, false, "force PKCS #8 ASN.1 format")
}

func cmdFlagsCryptoPrivateKey(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNameFilePrivateKey, "private.pem", "name of the file to export the private key data to")
	cmd.Flags().StringP(cmdFlagNameDirectory, "d", "", "directory where the generated keys, certificates, etc will be stored")
}

func cmdFlagsCryptoPrivateKeyRSA(cmd *cobra.Command) {
	cmd.Flags().IntP(cmdFlagNameBits, "b", 2048, "number of RSA bits for the certificate")
}

func cmdFlagsCryptoPrivateKeyECDSA(cmd *cobra.Command) {
	cmd.Flags().StringP(cmdFlagNameCurve, "b", "P256", "Sets the elliptic curve which can be P224, P256, P384, or P521")
}

func cmdFlagsCryptoPrivateKeyEd25519(cmd *cobra.Command) {
}

func cryptoSANsToString(dnsSANs []string, ipSANs []net.IP) (sans []string) {
	sans = make([]string, len(dnsSANs)+len(ipSANs))

	j := 0

	for i, dnsSAN := range dnsSANs {
		sans[j] = fmt.Sprintf("DNS.%d:%s", i+1, dnsSAN)
		j++
	}

	for i, ipSAN := range ipSANs {
		sans[j] = fmt.Sprintf("IP.%d:%s", i+1, ipSAN)
		j++
	}

	return sans
}

func cryptoGetWritePathsFromCmd(cmd *cobra.Command) (privateKey, publicKey string, err error) {
	var dir string

	if dir, err = cmd.Flags().GetString(cmdFlagNameDirectory); err != nil {
		return "", "", err
	}

	ca, _ := cmd.Flags().GetBool(cmdFlagNameCA)
	csr := cmd.Use == cmdUseRequest

	var private, public string

	var flagPrivate, flagPublic string

	switch {
	case ca && csr:
		flagPrivate, flagPublic = cmdFlagNameFileCAPrivateKey, cmdFlagNameFileCSR
	case csr:
		flagPrivate, flagPublic = cmdFlagNameFilePrivateKey, cmdFlagNameFileCSR
	case ca:
		flagPrivate, flagPublic = cmdFlagNameFileCAPrivateKey, cmdFlagNameFileCACertificate
	case cmd.Parent().Parent().Use == cmdUsePair:
		flagPrivate, flagPublic = cmdFlagNameFilePrivateKey, cmdFlagNameFilePublicKey
	default:
		flagPrivate, flagPublic = cmdFlagNameFilePrivateKey, cmdFlagNameFileCertificate
	}

	if private, err = cmd.Flags().GetString(flagPrivate); err != nil {
		return "", "", err
	}

	if public, err = cmd.Flags().GetString(flagPublic); err != nil {
		return "", "", err
	}

	return filepath.Join(dir, private), filepath.Join(dir, public), nil
}

func cryptoGenPrivateKeyFromCmd(cmd *cobra.Command) (privateKey interface{}, err error) {
	switch cmd.Parent().Use {
	case cmdUseRSA:
		var (
			bits int
		)

		if bits, err = cmd.Flags().GetInt(cmdFlagNameBits); err != nil {
			return nil, err
		}

		if privateKey, err = rsa.GenerateKey(rand.Reader, bits); err != nil {
			return nil, fmt.Errorf("generating RSA private key resulted in an error: %w", err)
		}
	case cmdUseECDSA:
		var (
			curveStr string
			curve    elliptic.Curve
		)

		if curveStr, err = cmd.Flags().GetString(cmdFlagNameCurve); err != nil {
			return nil, err
		}

		if curve = utils.EllipticCurveFromString(curveStr); curve == nil {
			return nil, fmt.Errorf("invalid curve '%s' was specified: curve must be P224, P256, P384, or P521", curveStr)
		}

		if privateKey, err = ecdsa.GenerateKey(curve, rand.Reader); err != nil {
			return nil, fmt.Errorf("generating ECDSA private key resulted in an error: %w", err)
		}
	case cmdUseEd25519:
		if _, privateKey, err = ed25519.GenerateKey(rand.Reader); err != nil {
			return nil, fmt.Errorf("generating Ed25519 private key resulted in an error: %w", err)
		}
	}

	return privateKey, nil
}

func cryptoGetCAFromCmd(cmd *cobra.Command) (privateKey interface{}, cert *x509.Certificate, err error) {
	if !cmd.Flags().Changed(cmdFlagNamePathCA) {
		return nil, nil, nil
	}

	var (
		dir, filePrivateKey, fileCertificate string

		ok bool

		certificate interface{}
	)

	if dir, err = cmd.Flags().GetString(cmdFlagNamePathCA); err != nil {
		return nil, nil, err
	}

	if filePrivateKey, err = cmd.Flags().GetString(cmdFlagNameFileCAPrivateKey); err != nil {
		return nil, nil, err
	}

	if fileCertificate, err = cmd.Flags().GetString(cmdFlagNameFileCACertificate); err != nil {
		return nil, nil, err
	}

	var (
		bytesPrivateKey, bytesCertificate []byte
	)

	pathPrivateKey := filepath.Join(dir, filePrivateKey)
	if bytesPrivateKey, err = os.ReadFile(pathPrivateKey); err != nil {
		return nil, nil, fmt.Errorf("could not read private key file '%s': %w", pathPrivateKey, err)
	}

	if privateKey, err = utils.ParseX509FromPEM(bytesPrivateKey); err != nil {
		return nil, nil, fmt.Errorf("could not parse private key from file '%s': %w", pathPrivateKey, err)
	}

	if privateKey == nil || !utils.IsX509PrivateKey(privateKey) {
		return nil, nil, fmt.Errorf("could not parse private key from file '%s': does not appear to be a private key", pathPrivateKey)
	}

	pathCertificate := filepath.Join(dir, fileCertificate)
	if bytesCertificate, err = os.ReadFile(pathCertificate); err != nil {
		return nil, nil, fmt.Errorf("could not read certificate file '%s': %w", pathCertificate, err)
	}

	if certificate, err = utils.ParseX509FromPEM(bytesCertificate); err != nil {
		return nil, nil, fmt.Errorf("could not parse certificate from file '%s': %w", pathCertificate, err)
	}

	if cert, ok = utils.CastX509AsCertificate(certificate); !ok {
		return nil, nil, fmt.Errorf("could not parse certificate from file '%s': does not appear to be a certificate", pathCertificate)
	}

	return privateKey, cert, nil
}

func cryptoGetCSRFromCmd(cmd *cobra.Command) (csr *x509.CertificateRequest, err error) {
	var (
		subject *pkix.Name
		dnsSANs []string
		ipSANs  []net.IP
	)

	if subject, err = cryptoGetSubjectFromCmd(cmd); err != nil {
		return nil, err
	}

	keyAlg, sigAlg := cryptoGetAlgFromCmd(cmd)

	if dnsSANs, ipSANs, err = cryptoGetSANsFromCmd(cmd); err != nil {
		return nil, err
	}

	csr = &x509.CertificateRequest{
		Subject:            *subject,
		PublicKeyAlgorithm: keyAlg,
		SignatureAlgorithm: sigAlg,

		DNSNames:    dnsSANs,
		IPAddresses: ipSANs,
	}

	return csr, nil
}

func cryptoGetSANsFromCmd(cmd *cobra.Command) (dnsSANs []string, ipSANs []net.IP, err error) {
	var (
		sans []string
	)

	if sans, err = cmd.Flags().GetStringSlice(cmdFlagNameSANs); err != nil {
		return nil, nil, err
	}

	for _, san := range sans {
		if ipSAN := net.ParseIP(san); ipSAN != nil {
			ipSANs = append(ipSANs, ipSAN)

			continue
		}

		dnsSANs = append(dnsSANs, san)
	}

	return dnsSANs, ipSANs, nil
}

func cryptoGetAlgFromCmd(cmd *cobra.Command) (keyAlg x509.PublicKeyAlgorithm, sigAlg x509.SignatureAlgorithm) {
	sigAlgStr, _ := cmd.Flags().GetString(cmdFlagNameSignature)
	keyAlgStr := cmd.Parent().Use

	return utils.KeySigAlgorithmFromString(keyAlgStr, sigAlgStr)
}

func cryptoGetSubjectFromCmd(cmd *cobra.Command) (subject *pkix.Name, err error) {
	var (
		commonName                                                                             string
		organization, organizationalUnit, country, locality, province, streetAddress, postcode []string
	)

	if commonName, err = cmd.Flags().GetString(cmdFlagNameCommonName); err != nil {
		return nil, err
	}

	if organization, err = cmd.Flags().GetStringSlice(cmdFlagNameOrganization); err != nil {
		return nil, err
	}

	if organizationalUnit, err = cmd.Flags().GetStringSlice(cmdFlagNameOrganizationalUnit); err != nil {
		return nil, err
	}

	if country, err = cmd.Flags().GetStringSlice(cmdFlagNameCountry); err != nil {
		return nil, err
	}

	if locality, err = cmd.Flags().GetStringSlice(cmdFlagNameLocality); err != nil {
		return nil, err
	}

	if province, err = cmd.Flags().GetStringSlice(cmdFlagNameProvince); err != nil {
		return nil, err
	}

	if streetAddress, err = cmd.Flags().GetStringSlice(cmdFlagNameStreetAddress); err != nil {
		return nil, err
	}

	if postcode, err = cmd.Flags().GetStringSlice(cmdFlagNamePostcode); err != nil {
		return nil, err
	}

	return &pkix.Name{
		CommonName:         commonName,
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Country:            country,
		Locality:           locality,
		Province:           province,
		StreetAddress:      streetAddress,
		PostalCode:         postcode,
	}, nil
}

func cryptoGetCertificateFromCmd(cmd *cobra.Command) (certificate *x509.Certificate, err error) {
	var (
		ca           bool
		subject      *pkix.Name
		notBeforeStr string
		duration     time.Duration
	)

	if ca, err = cmd.Flags().GetBool(cmdFlagNameCA); err != nil {
		return nil, err
	}

	if subject, err = cryptoGetSubjectFromCmd(cmd); err != nil {
		return nil, err
	}

	if notBeforeStr, err = cmd.Flags().GetString(cmdFlagNameNotBefore); err != nil {
		return nil, err
	}

	if duration, err = cmd.Flags().GetDuration(cmdFlagNameDuration); err != nil {
		return nil, err
	}

	var (
		notBefore             time.Time
		serialNumber          *big.Int
		dnsSANs, extKeyUsages []string
		ipSANs                []net.IP
	)

	switch len(notBeforeStr) {
	case 0:
		notBefore = time.Now()
	default:
		if notBefore, err = time.Parse(timeLayoutCertificateNotBefore, notBeforeStr); err != nil {
			return nil, fmt.Errorf("failed to parse not before: %w", err)
		}
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	if serialNumber, err = rand.Int(rand.Reader, serialNumberLimit); err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	if dnsSANs, ipSANs, err = cryptoGetSANsFromCmd(cmd); err != nil {
		return nil, err
	}

	keyAlg, sigAlg := cryptoGetAlgFromCmd(cmd)

	if extKeyUsages, err = cmd.Flags().GetStringSlice(cmdFlagNameExtendedUsage); err != nil {
		return nil, err
	}

	certificate = &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      *subject,

		NotBefore: notBefore,
		NotAfter:  notBefore.Add(duration),

		IsCA: ca,

		KeyUsage:    utils.X509ParseKeyUsage(nil, ca),
		ExtKeyUsage: utils.X509ParseExtendedKeyUsage(extKeyUsages, ca),

		PublicKeyAlgorithm: keyAlg,
		SignatureAlgorithm: sigAlg,

		DNSNames:    dnsSANs,
		IPAddresses: ipSANs,

		BasicConstraintsValid: true,
	}

	return certificate, nil
}

func fmtCryptoUse(use string) string {
	switch use {
	case cmdUseEd25519:
		return "Ed25519"
	default:
		return strings.ToUpper(use)
	}
}
