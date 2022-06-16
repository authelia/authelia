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
	cmd.Flags().String(flagNamePathCA, "", "source directory of the certificate authority files, if not provided the certificate will be self-signed")
	cmd.Flags().String(flagNameFileCAPrivateKey, "ca.private.pem", "certificate authority private key to use to signing this certificate")
	cmd.Flags().String(flagNameFileCACertificate, "ca.public.crt", "certificate authority certificate to use when signing this certificate")
	cmd.Flags().String(flagNameFileCertificate, "public.crt", "name of the file to export the certificate data to")
	cmd.Flags().String(flagNameFileCSR, "request.csr", "name of the file to export the certificate request data to")
	cmd.Flags().StringSlice(flagNameExtendedUsage, nil, "specify the extended usage types of the certificate")

	cmd.Flags().String(flagNameSignature, "SHA256", "signature algorithm for the certificate")
	cmd.Flags().Bool(flagNameCA, false, "create the certificate as a certificate authority certificate")
	cmd.Flags().Bool(flagNameCSR, false, "create a certificate signing request instead of a certificate")

	cmd.Flags().StringP(flagNameCommonName, "c", "", "certificate common name")
	cmd.Flags().StringSliceP(flagNameOrganization, "o", []string{"Authelia"}, "certificate organization")
	cmd.Flags().StringSlice(flagNameOrganizationalUnit, nil, "certificate organizational unit")
	cmd.Flags().StringSlice(flagNameCountry, nil, "certificate country")
	cmd.Flags().StringSlice(flagNameProvince, nil, "certificate province")
	cmd.Flags().StringSliceP(flagNameLocality, "l", nil, "certificate locality")
	cmd.Flags().StringSliceP(flagNameStreetAddress, "s", nil, "certificate street address")
	cmd.Flags().StringSliceP(flagNamePostcode, "p", nil, "certificate postcode")

	cmd.Flags().String(flagNameNotBefore, "", fmt.Sprintf("earliest date and time the certificate is considered valid formatted as %s (default is now)", timeLayoutCertificateNotBefore))
	cmd.Flags().Duration(flagNameDuration, 365*24*time.Hour, "duration of time the certificate is valid for")
	cmd.Flags().StringSlice(flagNameSANs, nil, "subject alternative names")
}

func cryptoPairGenFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagNameFilePublicKey, "public.pem", "name of the file to export the public key data to")
	cmd.Flags().Bool(flagNamePKCS8, false, "force PKCS #8 ASN.1 format")
}

func cryptoGenFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagNameFilePrivateKey, "private.pem", "name of the file to export the private key data to")
	cmd.Flags().StringP(flagNameDirectory, "d", "", "directory where the generated keys, certificates, etc will be stored")
}

func cryptoRSAGenFlags(cmd *cobra.Command) {
	cmd.Flags().IntP(flagNameBits, "b", 2048, "number of RSA bits for the certificate")
}

func cryptoECDSAGenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(flagNameCurve, "b", "P256", "Sets the elliptic curve which can be P224, P256, P384, or P521")
}

func cryptoGetWritePathsFromCmd(cmd *cobra.Command) (privateKey, publicKey string, err error) {
	dir, err := cmd.Flags().GetString(flagNameDirectory)
	if err != nil {
		return "", "", err
	}

	ca, _ := cmd.Flags().GetBool(flagNameCA)

	var private, public string

	switch {
	case ca:
		private, err = cmd.Flags().GetString(flagNameFileCAPrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(flagNameFileCACertificate)
		if err != nil {
			return "", "", err
		}
	case cmd.Parent().Use == "pair":
		private, err = cmd.Flags().GetString(flagNameFilePrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(flagNameFilePublicKey)
		if err != nil {
			return "", "", err
		}
	default:
		private, err = cmd.Flags().GetString(flagNameFilePrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(flagNameFileCertificate)
		if err != nil {
			return "", "", err
		}
	}

	/*
		csr, _ := cmd.Flags().GetBool("create-csr")

		if csr {
			publicExt = ".csr"
		}
	*/

	return filepath.Join(dir, private), filepath.Join(dir, public), nil
}

func cryptoGenPrivateKeyFromCmd(cmd *cobra.Command) (privateKey interface{}, err error) {
	var (
		bits     int
		curveStr string
		curve    elliptic.Curve
	)

	switch cmd.Use {
	case "rsa":
		bits, err = cmd.Flags().GetInt(flagNameBits)
		if err != nil {
			return nil, err
		}

		privateKey, err = rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, err
		}
	case "ecdsa":
		curveStr, err = cmd.Flags().GetString(flagNameCurve)
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
	case "ed25519":
		_, privateKey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
	}

	return privateKey, nil
}

func cryptoGetCAFromCmd(cmd *cobra.Command) (privateKey interface{}, cert *x509.Certificate, err error) {
	if !cmd.Flags().Changed(flagNamePathCA) {
		return nil, nil, nil
	}

	caPath, err := cmd.Flags().GetString(flagNamePathCA)
	if err != nil {
		return nil, nil, err
	}

	keyName, err := cmd.Flags().GetString(flagNameFileCAPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	certName, err := cmd.Flags().GetString(flagNameFileCACertificate)
	if err != nil {
		return nil, nil, err
	}

	keyBytes, err := os.ReadFile(filepath.Join(caPath, keyName))
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := os.ReadFile(filepath.Join(caPath, certName))
	if err != nil {
		return nil, nil, err
	}

	privateKey, err = utils.ParseX509FromPEM(keyBytes)
	if err != nil {
		return nil, nil, err
	}

	cert, err = utils.ParseX509CertificateFromPEM(certBytes)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, cert, nil
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
	sans, err := cmd.Flags().GetStringSlice(flagNameSANs)
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
	sigAlgStr, _ := cmd.Flags().GetString(flagNameSignature)
	keyAlgStr := cmd.Use

	return utils.KeySigAlgorithmFromString(keyAlgStr, sigAlgStr)
}

func cryptoGetSubjectFromCmd(cmd *cobra.Command) (subject *pkix.Name, err error) {
	commonName, err := cmd.Flags().GetString(flagNameCommonName)
	if err != nil {
		return nil, err
	}

	organization, err := cmd.Flags().GetStringSlice(flagNameOrganization)
	if err != nil {
		return nil, err
	}

	organizationalUnit, err := cmd.Flags().GetStringSlice(flagNameOrganizationalUnit)
	if err != nil {
		return nil, err
	}

	country, err := cmd.Flags().GetStringSlice(flagNameCountry)
	if err != nil {
		return nil, err
	}

	locality, err := cmd.Flags().GetStringSlice(flagNameLocality)
	if err != nil {
		return nil, err
	}

	province, err := cmd.Flags().GetStringSlice(flagNameProvince)
	if err != nil {
		return nil, err
	}

	streetAddress, err := cmd.Flags().GetStringSlice(flagNameStreetAddress)
	if err != nil {
		return nil, err
	}

	postcode, err := cmd.Flags().GetStringSlice(flagNamePostcode)
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
	ca, err := cmd.Flags().GetBool(flagNameCA)
	if err != nil {
		return nil, err
	}

	subject, err := cryptoGetSubjectFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	notBeforeStr, err := cmd.Flags().GetString(flagNameNotBefore)
	if err != nil {
		return nil, err
	}

	duration, err := cmd.Flags().GetDuration(flagNameDuration)
	if err != nil {
		return nil, err
	}

	var notBefore time.Time

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

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	dnsNames, ipAddresses, err := cryptoGetSANsFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	keyAlg, sigAlg := cryptoGetAlgFromCmd(cmd)

	extKeyUsages, _ := cmd.Flags().GetStringSlice(flagNameExtendedUsage)

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

		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,

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
