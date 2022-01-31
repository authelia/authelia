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
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func cryptoCertificateGenFlags(cmd *cobra.Command) {
	cmd.Flags().String("ca-path", "", "source directory of the CA files, if not provided the certificate will be self-signed")
	cmd.Flags().String(flagNameCAPrivateKey, "ca.private.pem", "CA key to use to signing this certificate (you must also set --ca-path)")
	cmd.Flags().String(flagNameCACertificate, "ca.public.crt", "CA certificate to use when signing this certificate (you must also set --ca-path)")
	cmd.Flags().String(flagNameCertificate, "public.crt", "name of the file to export the certificate data to")
	cmd.Flags().String("csr", "csr", "name of the file to export the CSR data to")

	cmd.Flags().String("signature", "SHA1", "signature algorithm for the certificate")
	cmd.Flags().Bool("ca", false, "create the certificate as a CA certificate")
	cmd.Flags().Bool("create-csr", false, "create a CSR instead of a certificate")

	cmd.Flags().StringP("common-name", "c", "", "certificate common name")
	cmd.Flags().StringSliceP("organization", "o", []string{"Authelia"}, "certificate organization")
	cmd.Flags().StringSlice("organizational-unit", nil, "certificate organizational unit")
	cmd.Flags().StringSlice("country", nil, "certificate country")
	cmd.Flags().StringSlice("province", nil, "certificate province")
	cmd.Flags().StringSliceP("locality", "l", nil, "certificate locality")
	cmd.Flags().StringSliceP("street-address", "s", nil, "certificate street address")
	cmd.Flags().StringSliceP("postcode", "p", nil, "certificate postcode")
	cmd.Flags().String("not-before", "", fmt.Sprintf("earliest date and time the certificate is considered valid formatted as %s (default is now)", timeLayoutCertificateNotBefore))
	cmd.Flags().Duration("duration", 365*24*time.Hour, "duration of time the certificate is valid for")
	cmd.Flags().StringSlice("sans", nil, "subject alternative names")
}

func cryptoPairGenFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagNamePublicKey, "public.pem", "name of the file to export the public key data to")
	cmd.Flags().Bool("pkcs8", false, "force PKCS #8 ASN.1 format")
}

func cryptoGenFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagNamePrivateKey, "private.pem", "name of the file to export the private key data to")
	cmd.Flags().String("directory", "", "directory where the generated keys, certificates, etc will be stored")
}

func cryptoRSAGenFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("bits", "b", 2048, "number of RSA bits for the certificate")
}

func cryptoECDSAGenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("curve", "b", "P256", "Sets the elliptic curve which can be P224, P256, P384, or P521")
}

func cryptoGetWritePathsFromCmd(cmd *cobra.Command) (privateKey, publicKey string, err error) {
	dir, err := cmd.Flags().GetString("directory")
	if err != nil {
		return "", "", err
	}

	ca, _ := cmd.Flags().GetBool("ca")

	var private, public string

	switch {
	case ca:
		private, err = cmd.Flags().GetString(flagNameCAPrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(flagNameCACertificate)
		if err != nil {
			return "", "", err
		}
	case cmd.Parent().Use == "pair":
		private, err = cmd.Flags().GetString(flagNamePrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(flagNamePublicKey)
		if err != nil {
			return "", "", err
		}
	default:
		private, err = cmd.Flags().GetString(flagNamePrivateKey)
		if err != nil {
			return "", "", err
		}

		public, err = cmd.Flags().GetString(flagNameCertificate)
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
		bits, err = cmd.Flags().GetInt("bits")
		if err != nil {
			return nil, err
		}

		privateKey, err = rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, err
		}
	case "ecdsa":
		curveStr, err = cmd.Flags().GetString("curve")
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
			fmt.Println("gen ed25519")

			return nil, err
		}
	}

	return privateKey, nil
}

func cryptoGetCAFromCmd(cmd *cobra.Command) (privateKey interface{}, cert *x509.Certificate, err error) {
	if !cmd.Flags().Changed("ca-path") {
		return nil, nil, nil
	}

	caPath, err := cmd.Flags().GetString("ca-path")
	if err != nil {
		return nil, nil, err
	}

	keyName, err := cmd.Flags().GetString(flagNameCAPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	certName, err := cmd.Flags().GetString(flagNameCACertificate)
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
	sans, err := cmd.Flags().GetStringSlice("sans")
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
	sigAlgStr, _ := cmd.Flags().GetString("signature")
	keyAlgStr := cmd.Use

	return utils.KeySigAlgorithmFromString(keyAlgStr, sigAlgStr)
}

func cryptoGetSubjectFromCmd(cmd *cobra.Command) (subject *pkix.Name, err error) {
	commonName, err := cmd.Flags().GetString("common-name")
	if err != nil {
		return nil, err
	}

	organization, err := cmd.Flags().GetStringSlice("organization")
	if err != nil {
		return nil, err
	}

	organizationalUnit, err := cmd.Flags().GetStringSlice("organizational-unit")
	if err != nil {
		return nil, err
	}

	country, err := cmd.Flags().GetStringSlice("country")
	if err != nil {
		return nil, err
	}

	locality, err := cmd.Flags().GetStringSlice("locality")
	if err != nil {
		return nil, err
	}

	province, err := cmd.Flags().GetStringSlice("province")
	if err != nil {
		return nil, err
	}

	streetAddress, err := cmd.Flags().GetStringSlice("street-address")
	if err != nil {
		return nil, err
	}

	postcode, err := cmd.Flags().GetStringSlice("postcode")
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
	ca, err := cmd.Flags().GetBool("ca")
	if err != nil {
		return nil, err
	}

	subject, err := cryptoGetSubjectFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	notBeforeStr, err := cmd.Flags().GetString("not-before")
	if err != nil {
		return nil, err
	}

	duration, err := cmd.Flags().GetDuration("duration")
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

	cert = &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      *subject,

		NotBefore: notBefore,
		NotAfter:  notBefore.Add(duration),

		IsCA: ca,

		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		PublicKeyAlgorithm: keyAlg,
		SignatureAlgorithm: sigAlg,

		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,

		BasicConstraintsValid: true,
	}

	if cert.IsCA {
		cert.KeyUsage |= x509.KeyUsageCertSign
	}

	return cert, nil
}
