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

func cryptoCertificateGenFlags(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNamePathCA, "", "source directory of the certificate authority files, if not provided the certificate will be self-signed")
	cmd.Flags().String(cmdFlagNameFileCAPrivateKey, "ca.private.pem", "certificate authority private key to use to signing this certificate")
	cmd.Flags().String(cmdFlagNameFileCACertificate, "ca.public.crt", "certificate authority certificate to use when signing this certificate")
	cmd.Flags().String(cmdFlagNameFileCertificate, "public.crt", "name of the file to export the certificate data to")
	cmd.Flags().String(cmdFlagNameFileCSR, "request.csr", "name of the file to export the certificate request data to")
	cmd.Flags().StringSlice(cmdFlagNameExtendedUsage, nil, "specify the extended usage types of the certificate")

	cmd.Flags().String(cmdFlagNameSignature, "SHA256", "signature algorithm for the certificate")
	cmd.Flags().Bool(cmdFlagNameCA, false, "create the certificate as a certificate authority certificate")
	cmd.Flags().Bool(cmdFlagNameCSR, false, "create a certificate signing request instead of a certificate")

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

func cryptoPairGenFlags(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNameFilePublicKey, "public.pem", "name of the file to export the public key data to")
	cmd.Flags().Bool(cmdFlagNamePKCS8, false, "force PKCS #8 ASN.1 format")
}

func cryptoGenFlags(cmd *cobra.Command) {
	cmd.Flags().String(cmdFlagNameFilePrivateKey, "private.pem", "name of the file to export the private key data to")
	cmd.Flags().StringP(cmdFlagNameDirectory, "d", "", "directory where the generated keys, certificates, etc will be stored")
}

func cryptoRSAGenFlags(cmd *cobra.Command) {
	cmd.Flags().IntP(cmdFlagNameBits, "b", 2048, "number of RSA bits for the certificate")
}

func cryptoECDSAGenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(cmdFlagNameCurve, "b", "P256", "Sets the elliptic curve which can be P224, P256, P384, or P521")
}

func cryptoEd25519Flags(cmd *cobra.Command) {
}

func cryptoGetWritePathsFromCmd(cmd *cobra.Command) (privateKey, publicKey string, err error) {
	dir, err := cmd.Flags().GetString(cmdFlagNameDirectory)
	if err != nil {
		return "", "", err
	}

	ca, _ := cmd.Flags().GetBool(cmdFlagNameCA)

	var private, public string

	switch {
	case ca:
		private, err = cmd.Flags().GetString(cmdFlagNameFileCAPrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(cmdFlagNameFileCACertificate)
		if err != nil {
			return "", "", err
		}
	case cmd.Parent().Parent().Use == cmdUsePair:
		private, err = cmd.Flags().GetString(cmdFlagNameFilePrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(cmdFlagNameFilePublicKey)
		if err != nil {
			return "", "", err
		}
	default:
		private, err = cmd.Flags().GetString(cmdFlagNameFilePrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(cmdFlagNameFileCertificate)
		if err != nil {
			return "", "", err
		}
	}

	return filepath.Join(dir, private), filepath.Join(dir, public), nil
}

func cryptoGenPrivateKeyFromCmd(cmd *cobra.Command) (privateKey interface{}, err error) {
	var (
		bits     int
		curveStr string
		curve    elliptic.Curve
	)

	switch cmd.Parent().Use {
	case cmdUseRSA:
		bits, err = cmd.Flags().GetInt(cmdFlagNameBits)
		if err != nil {
			return nil, err
		}

		privateKey, err = rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, err
		}
	case cmdUseECDSA:
		curveStr, err = cmd.Flags().GetString(cmdFlagNameCurve)
		if err != nil {
			return nil, err
		}

		curve = utils.EllipticCurveFromString(curveStr)
		if curve == nil {
			return nil, fmt.Errorf("curve must be P224, P256, P384, or P521 but an invalid curve was specified: %s", curveStr)
		}

		privateKey, err = ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, err
		}
	case cmdUseEd25519:
		_, privateKey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
	}

	return privateKey, nil
}

func cryptoGetCAFromCmd(cmd *cobra.Command) (privateKey interface{}, certificate *x509.Certificate, err error) {
	if !cmd.Flags().Changed(cmdFlagNamePathCA) {
		return nil, nil, nil
	}

	pathCA, err := cmd.Flags().GetString(cmdFlagNamePathCA)
	if err != nil {
		return nil, nil, err
	}

	caPrivateKeyFileName, err := cmd.Flags().GetString(cmdFlagNameFileCAPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	caCertificateFileName, err := cmd.Flags().GetString(cmdFlagNameFileCACertificate)
	if err != nil {
		return nil, nil, err
	}

	bytesPrivateKey, err := os.ReadFile(filepath.Join(pathCA, caPrivateKeyFileName))
	if err != nil {
		return nil, nil, err
	}

	bytesCertificate, err := os.ReadFile(filepath.Join(pathCA, caCertificateFileName))
	if err != nil {
		return nil, nil, err
	}

	privateKey, err = utils.ParseX509FromPEM(bytesPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	certificate, err = utils.ParseX509CertificateFromPEM(bytesCertificate)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, certificate, nil
}

func cryptoGetCSRFromCmd(cmd *cobra.Command) (csr *x509.CertificateRequest, err error) {
	subject, err := cryptoGetSubjectFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	keyAlg, sigAlg := cryptoGetAlgFromCmd(cmd)

	dnsNames, ipAddresses, err := cryptoGetSANsFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	csr = &x509.CertificateRequest{
		Subject:            *subject,
		PublicKeyAlgorithm: keyAlg,
		SignatureAlgorithm: sigAlg,

		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
	}

	return csr, nil
}

func cryptoGetSANsFromCmd(cmd *cobra.Command) (dnsSANs []string, ipSANs []net.IP, err error) {
	sans, err := cmd.Flags().GetStringSlice(cmdFlagNameSANs)
	if err != nil {
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
	commonName, err := cmd.Flags().GetString(cmdFlagNameCommonName)
	if err != nil {
		return nil, err
	}

	organization, err := cmd.Flags().GetStringSlice(cmdFlagNameOrganization)
	if err != nil {
		return nil, err
	}

	organizationalUnit, err := cmd.Flags().GetStringSlice(cmdFlagNameOrganizationalUnit)
	if err != nil {
		return nil, err
	}

	country, err := cmd.Flags().GetStringSlice(cmdFlagNameCountry)
	if err != nil {
		return nil, err
	}

	locality, err := cmd.Flags().GetStringSlice(cmdFlagNameLocality)
	if err != nil {
		return nil, err
	}

	province, err := cmd.Flags().GetStringSlice(cmdFlagNameProvince)
	if err != nil {
		return nil, err
	}

	streetAddress, err := cmd.Flags().GetStringSlice(cmdFlagNameStreetAddress)
	if err != nil {
		return nil, err
	}

	postcode, err := cmd.Flags().GetStringSlice(cmdFlagNamePostcode)
	if err != nil {
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

func cryptoGetCertificateFromCmd(cmd *cobra.Command) (cert *x509.Certificate, err error) {
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
		notBefore    time.Time
		serialNumber *big.Int
		dnsSANs      []string
		ipSANs       []net.IP
	)

	switch len(notBeforeStr) {
	case 0:
		notBefore = time.Now()
	default:
		notBefore, err = time.Parse(timeLayoutCertificateNotBefore, notBeforeStr)
		if err != nil {
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

	extKeyUsages, _ := cmd.Flags().GetStringSlice(cmdFlagNameExtendedUsage)

	cert = &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      *subject,

		NotBefore: notBefore,
		NotAfter:  notBefore.Add(duration),

		IsCA: ca,

		KeyUsage:    cryptoGetKeyUsage(nil, ca),
		ExtKeyUsage: cryptoGetExtKeyUsage(extKeyUsages, ca),

		PublicKeyAlgorithm: keyAlg,
		SignatureAlgorithm: sigAlg,

		DNSNames:    dnsSANs,
		IPAddresses: ipSANs,

		BasicConstraintsValid: true,
	}

	return cert, nil
}

func cryptoGetExtKeyUsage(extKeyUsages []string, ca bool) (extKeyUsage []x509.ExtKeyUsage) {
	if len(extKeyUsages) == 0 {
		if ca {
			extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
		} else {
			extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		}

		return extKeyUsage
	}

loop:
	for _, extKeyUsageString := range extKeyUsages {
		switch strings.ToLower(extKeyUsageString) {
		case "any":
			extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
			break loop
		case "serverauth", "server_auth":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageServerAuth)
		case "clientauth", "client_auth":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageClientAuth)
		case "codesigning", "code_signing":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageCodeSigning)
		case "emailprotection", "email_protection":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageEmailProtection)
		case "ipsecendsystem", "ipsec_endsystem", "ipsec_end_system":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECEndSystem)
		case "ipsectunnel", "ipsec_tunnel":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECTunnel)
		case "ipsecuser", "ipsec_user":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECUser)
		case "ocspsigning", "ocsp_signing":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageOCSPSigning)
		}
	}

	return extKeyUsage
}

func cryptoGetKeyUsage(keyUsages []string, ca bool) (keyUsage x509.KeyUsage) {
	if len(keyUsages) == 0 {
		keyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		if ca {
			keyUsage |= x509.KeyUsageCertSign
		}

		return keyUsage
	}

	for _, keyUsageString := range keyUsages {
		switch strings.ToLower(keyUsageString) {
		case "digitalsignature", "digital_signature":
			keyUsage |= x509.KeyUsageDigitalSignature
		case "keyencipherment", "key_encipherment":
			keyUsage |= x509.KeyUsageKeyEncipherment
		case "dataencipherment", "data_encipherment":
			keyUsage |= x509.KeyUsageDataEncipherment
		case "keyagreement", "key_agreement":
			keyUsage |= x509.KeyUsageKeyAgreement
		case "certsign", "cert_sign", "certificatesign", "certificate_sign":
			keyUsage |= x509.KeyUsageCertSign
		case "crlsign", "crl_sign":
			keyUsage |= x509.KeyUsageCRLSign
		case "encipheronly", "encipher_only":
			keyUsage |= x509.KeyUsageEncipherOnly
		case "decipheronly", "decipher_only":
			keyUsage |= x509.KeyUsageDecipherOnly
		}
	}

	return keyUsage
}

/*
	Key Usage Values:
		digitalSignature
		keyEncipherment
		dataEncipherment
		keyAgreement
		certSign
		crlSign
		encipherOnly
		decipherOnly

	Extended Key Usage Values:
		serverAuth
		clientAuth
		codeSigning
		emailProtection
		ipsecEndSystem
		ipsecTunnel
		ipsecUser
		ocspSigning
*/
