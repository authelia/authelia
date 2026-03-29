package commands

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestCryptoSANsToString(t *testing.T) {
	testCases := []struct {
		name     string
		dnsSANs  []string
		ipSANs   []net.IP
		expected []string
	}{
		{
			"ShouldFormatDNSOnly",
			[]string{"example.com", "auth.example.com"},
			nil,
			[]string{"DNS.1:example.com", "DNS.2:auth.example.com"},
		},
		{
			"ShouldFormatIPOnly",
			nil,
			[]net.IP{net.ParseIP("192.168.1.1"), net.ParseIP("10.0.0.1")},
			[]string{"IP.1:192.168.1.1", "IP.2:10.0.0.1"},
		},
		{
			"ShouldFormatMixed",
			[]string{"example.com"},
			[]net.IP{net.ParseIP("192.168.1.1")},
			[]string{"DNS.1:example.com", "IP.1:192.168.1.1"},
		},
		{
			"ShouldReturnEmptyForNoSANs",
			nil,
			nil,
			[]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cryptoSANsToString(tc.dnsSANs, tc.ipSANs)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCryptoGetSANsFromCmd(t *testing.T) {
	testCases := []struct {
		name            string
		sans            []string
		expectedDNSSANs []string
		expectedIPSANs  []net.IP
	}{
		{
			"ShouldParseDNSAndIP",
			[]string{"example.com", "192.168.1.1", "auth.example.com", "10.0.0.1"},
			[]string{"example.com", "auth.example.com"},
			[]net.IP{net.ParseIP("192.168.1.1"), net.ParseIP("10.0.0.1")},
		},
		{
			"ShouldParseDNSOnly",
			[]string{"example.com"},
			[]string{"example.com"},
			nil,
		},
		{
			"ShouldParseIPOnly",
			[]string{"192.168.1.1"},
			nil,
			[]net.IP{net.ParseIP("192.168.1.1")},
		},
		{
			"ShouldReturnNilForEmpty",
			nil,
			nil,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().StringSlice(cmdFlagNameSANs, tc.sans, "")

			dnsSANs, ipSANs, err := cryptoGetSANsFromCmd(cmd)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedDNSSANs, dnsSANs)
			assert.Equal(t, tc.expectedIPSANs, ipSANs)
		})
	}

	t.Run("ShouldErrWhenSANsFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}

		_, _, err := cryptoGetSANsFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: sans")
	})
}

func TestCryptoGetSubjectFromCmd(t *testing.T) {
	testCases := []struct {
		name     string
		flags    map[string]string
		expected *pkix.Name
	}{
		{
			"ShouldReturnDefaults",
			nil,
			&pkix.Name{
				Organization:       []string{"Authelia"},
				OrganizationalUnit: []string{},
				Country:            []string{},
				Province:           []string{},
				Locality:           []string{},
				StreetAddress:      []string{},
				PostalCode:         []string{},
			},
		},
		{
			"ShouldReturnCustomValues",
			map[string]string{
				cmdFlagNameCommonName:         "Test CA",
				cmdFlagNameOrganization:       "TestOrg",
				cmdFlagNameOrganizationalUnit: "TestUnit",
				cmdFlagNameCountry:            "AU",
				cmdFlagNameProvince:           "NSW",
				cmdFlagNameLocality:           "Sydney",
				cmdFlagNameStreetAddress:      "123 Test St",
				cmdFlagNamePostcode:           "2000",
			},
			&pkix.Name{
				CommonName:         "Test CA",
				Organization:       []string{"TestOrg"},
				OrganizationalUnit: []string{"TestUnit"},
				Country:            []string{"AU"},
				Province:           []string{"NSW"},
				Locality:           []string{"Sydney"},
				StreetAddress:      []string{"123 Test St"},
				PostalCode:         []string{"2000"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmdFlagsCryptoCertificateCommon(cmd)

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			subject, err := cryptoGetSubjectFromCmd(cmd)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, subject)
		})
	}

	t.Run("ShouldErrWhenCommonNameFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: common-name")
	})

	t.Run("ShouldErrWhenOrganizationFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNameCommonName, "", "")

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: organization")
	})

	t.Run("ShouldErrWhenOrganizationalUnitFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNameCommonName, "", "")
		cmd.Flags().StringSlice(cmdFlagNameOrganization, nil, "")

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: organizational-unit")
	})

	t.Run("ShouldErrWhenCountryFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNameCommonName, "", "")
		cmd.Flags().StringSlice(cmdFlagNameOrganization, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameOrganizationalUnit, nil, "")

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: country")
	})

	t.Run("ShouldErrWhenLocalityFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNameCommonName, "", "")
		cmd.Flags().StringSlice(cmdFlagNameOrganization, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameOrganizationalUnit, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameCountry, nil, "")

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: locality")
	})

	t.Run("ShouldErrWhenProvinceFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNameCommonName, "", "")
		cmd.Flags().StringSlice(cmdFlagNameOrganization, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameOrganizationalUnit, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameCountry, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameLocality, nil, "")

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: province")
	})

	t.Run("ShouldErrWhenStreetAddressFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNameCommonName, "", "")
		cmd.Flags().StringSlice(cmdFlagNameOrganization, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameOrganizationalUnit, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameCountry, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameLocality, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameProvince, nil, "")

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: street-address")
	})

	t.Run("ShouldErrWhenPostcodeFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNameCommonName, "", "")
		cmd.Flags().StringSlice(cmdFlagNameOrganization, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameOrganizationalUnit, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameCountry, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameLocality, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameProvince, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameStreetAddress, nil, "")

		_, err := cryptoGetSubjectFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: postcode")
	})
}

func TestCryptoGetAlgFromCmd(t *testing.T) {
	testCases := []struct {
		name           string
		parentUse      string
		signature      string
		expectedKeyAlg x509.PublicKeyAlgorithm
		expectedSigAlg x509.SignatureAlgorithm
	}{
		{
			"ShouldReturnRSASHA256",
			cmdUseRSA,
			"SHA256",
			x509.RSA,
			x509.SHA256WithRSA,
		},
		{
			"ShouldReturnECDSASHA256",
			cmdUseECDSA,
			"SHA256",
			x509.ECDSA,
			x509.ECDSAWithSHA256,
		},
		{
			"ShouldReturnEd25519",
			cmdUseEd25519,
			"SHA256",
			x509.Ed25519,
			x509.PureEd25519,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parent := &cobra.Command{Use: tc.parentUse}
			cmd := &cobra.Command{Use: "generate"}

			parent.AddCommand(cmd)

			cmd.Flags().String(cmdFlagNameSignature, tc.signature, "")

			keyAlg, sigAlg := cryptoGetAlgFromCmd(cmd)

			assert.Equal(t, tc.expectedKeyAlg, keyAlg)
			assert.Equal(t, tc.expectedSigAlg, sigAlg)
		})
	}
}

// newTestCryptoGenerateCmd creates a cobra command hierarchy matching what cryptoGetWritePathsFromCmd expects.
func newTestCryptoGenerateCmd(parentUse, grandparentUse string, isCA, isCsr bool) *cobra.Command {
	grandparent := &cobra.Command{Use: grandparentUse}
	parent := &cobra.Command{Use: parentUse}

	var use string
	if isCsr {
		use = cmdUseRequest
	} else {
		use = "generate"
	}

	cmd := &cobra.Command{Use: use}

	grandparent.AddCommand(parent)
	parent.AddCommand(cmd)

	cmdFlagsCryptoPrivateKey(cmd)
	cmdFlagsCryptoPairGenerate(cmd)
	cmdFlagsCryptoCertificateGenerate(cmd)
	cmdFlagsCryptoCertificateRequest(cmd)

	if isCA {
		_ = cmd.Flags().Set(cmdFlagNameCA, "true")
	}

	return cmd
}

func TestCryptoGetWritePathsFromCmd(t *testing.T) {
	testCases := []struct {
		name                   string
		parentUse              string
		grandparentUse         string
		isCA                   bool
		isCsr                  bool
		dir                    string
		expectedPrivateKeyFile string
		expectedPublicFile     string
	}{
		{
			"ShouldReturnDefaultCertificatePaths",
			cmdUseRSA,
			"certificate",
			false,
			false,
			"",
			"private.pem",
			"public.crt",
		},
		{
			"ShouldReturnCAPaths",
			cmdUseRSA,
			"certificate",
			true,
			false,
			"",
			"ca.private.pem",
			"ca.public.crt",
		},
		{
			"ShouldReturnCSRPaths",
			cmdUseRSA,
			"certificate",
			false,
			true,
			"",
			"private.pem",
			"request.csr",
		},
		{
			"ShouldReturnCACSRPaths",
			cmdUseRSA,
			"certificate",
			true,
			true,
			"",
			"ca.private.pem",
			"request.csr",
		},
		{
			"ShouldReturnPairPaths",
			cmdUseRSA,
			cmdUsePair,
			false,
			false,
			"",
			"private.pem",
			"public.pem",
		},
		{
			"ShouldReturnPathsWithDirectory",
			cmdUseRSA,
			"certificate",
			false,
			false,
			"/tmp/certs",
			filepath.Join("/tmp/certs", "private.pem"),
			filepath.Join("/tmp/certs", "public.crt"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newTestCryptoGenerateCmd(tc.parentUse, tc.grandparentUse, tc.isCA, tc.isCsr)

			if tc.dir != "" {
				require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, tc.dir))
			}

			dir, privateKey, publicKey, err := cryptoGetWritePathsFromCmd(cmd)

			assert.NoError(t, err)

			if tc.dir != "" {
				assert.Equal(t, tc.dir, dir)
			}

			assert.Equal(t, tc.expectedPrivateKeyFile, privateKey)
			assert.Equal(t, tc.expectedPublicFile, publicKey)
		})
	}

	t.Run("ShouldErrWhenDirectoryFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{Use: "generate"}

		_, _, _, err := cryptoGetWritePathsFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: directory")
	})

	t.Run("ShouldErrWhenPrivateKeyFlagNotDefined", func(t *testing.T) {
		grandparent := &cobra.Command{Use: "certificate"}
		parent := &cobra.Command{Use: cmdUseRSA}
		cmd := &cobra.Command{Use: "generate"}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		cmd.Flags().String(cmdFlagNameDirectory, "", "")

		_, _, _, err := cryptoGetWritePathsFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: file.private-key")
	})

	t.Run("ShouldErrWhenPublicKeyFlagNotDefined", func(t *testing.T) {
		grandparent := &cobra.Command{Use: "certificate"}
		parent := &cobra.Command{Use: cmdUseRSA}
		cmd := &cobra.Command{Use: "generate"}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		cmd.Flags().String(cmdFlagNameDirectory, "", "")
		cmd.Flags().String(cmdFlagNameFilePrivateKey, "private.pem", "")

		_, _, _, err := cryptoGetWritePathsFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: file.certificate")
	})
}

func TestCryptoGenPrivateKeyFromCmd(t *testing.T) {
	testCases := []struct {
		name      string
		parentUse string
		flags     map[string]string
		err       string
		checkType func(t *testing.T, key any)
	}{
		{
			"ShouldGenerateRSAKey",
			cmdUseRSA,
			map[string]string{cmdFlagNameBits: "2048"},
			"",
			func(t *testing.T, key any) {
				_, ok := key.(*rsa.PrivateKey)
				assert.True(t, ok)
			},
		},
		{
			"ShouldGenerateECDSAP256Key",
			cmdUseECDSA,
			map[string]string{cmdFlagNameCurve: "P256"},
			"",
			func(t *testing.T, key any) {
				k, ok := key.(*ecdsa.PrivateKey)
				assert.True(t, ok)
				assert.Equal(t, elliptic.P256(), k.Curve)
			},
		},
		{
			"ShouldGenerateECDSAP384Key",
			cmdUseECDSA,
			map[string]string{cmdFlagNameCurve: "P384"},
			"",
			func(t *testing.T, key any) {
				k, ok := key.(*ecdsa.PrivateKey)
				assert.True(t, ok)
				assert.Equal(t, elliptic.P384(), k.Curve)
			},
		},
		{
			"ShouldGenerateEd25519Key",
			cmdUseEd25519,
			nil,
			"",
			func(t *testing.T, key any) {
				_, ok := key.(ed25519.PrivateKey)
				assert.True(t, ok)
			},
		},
		{
			"ShouldErrECDSAInvalidCurve",
			cmdUseECDSA,
			map[string]string{cmdFlagNameCurve: "P999"},
			"invalid curve 'P999' was specified",
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := NewCmdCtx()

			parent := &cobra.Command{Use: tc.parentUse}
			cmd := &cobra.Command{Use: "generate"}

			parent.AddCommand(cmd)

			switch tc.parentUse {
			case cmdUseRSA:
				cmdFlagsCryptoPrivateKeyRSA(cmd)
			case cmdUseECDSA:
				cmdFlagsCryptoPrivateKeyECDSA(cmd)
			case cmdUseEd25519:
				cmdFlagsCryptoPrivateKeyEd25519(cmd)
			}

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			key, err := cmdCtx.cryptoGenPrivateKeyFromCmd(cmd)

			if tc.err == "" {
				assert.NoError(t, err)
				require.NotNil(t, key)
				tc.checkType(t, key)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldErrWhenRSABitsFlagNotDefined", func(t *testing.T) {
		cmdCtx := NewCmdCtx()

		parent := &cobra.Command{Use: cmdUseRSA}
		cmd := &cobra.Command{Use: "generate"}

		parent.AddCommand(cmd)

		_, err := cmdCtx.cryptoGenPrivateKeyFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: bits")
	})

	t.Run("ShouldErrWhenECDSACurveFlagNotDefined", func(t *testing.T) {
		cmdCtx := NewCmdCtx()

		parent := &cobra.Command{Use: cmdUseECDSA}
		cmd := &cobra.Command{Use: "generate"}

		parent.AddCommand(cmd)

		_, err := cmdCtx.cryptoGenPrivateKeyFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: curve")
	})
}

func TestCryptoCertificateValidityFromCmd(t *testing.T) {
	testCases := []struct {
		name  string
		flags map[string]string
		err   string
	}{
		{
			"ShouldReturnDefaultDuration",
			nil,
			"",
		},
		{
			"ShouldReturnCustomDuration",
			map[string]string{cmdFlagNameDuration: "2y"},
			"",
		},
		{
			"ShouldReturnNotAfter",
			map[string]string{cmdFlagNameNotAfter: "Jan 2 15:04:05 2030"},
			"",
		},
		{
			"ShouldReturnNotBefore",
			map[string]string{cmdFlagNameNotBefore: "Jan 2 15:04:05 2020"},
			"",
		},
		{
			"ShouldErrBothNotAfterAndDuration",
			map[string]string{cmdFlagNameNotAfter: "Jan 2 15:04:05 2030", cmdFlagNameDuration: "1y"},
			"failed to determine not after",
		},
		{
			"ShouldErrInvalidDuration",
			map[string]string{cmdFlagNameDuration: "invalid"},
			"failed to parse duration string:",
		},
		{
			"ShouldErrInvalidNotBefore",
			map[string]string{cmdFlagNameNotBefore: "not-a-date"},
			"failed to parse not before:",
		},
		{
			"ShouldErrInvalidNotAfter",
			map[string]string{cmdFlagNameNotAfter: "not-a-date"},
			"failed to parse not after:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmdFlagsCryptoCertificateCommon(cmd)

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			notBefore, notAfter, err := cryptoCertificateValidityFromCmd(cmd)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.False(t, notBefore.IsZero())
				assert.True(t, notAfter.After(notBefore))
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldErrWhenDurationFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}

		_, _, err := cryptoCertificateValidityFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: duration")
	})
}

func TestCryptoGetCertificateFromCmd(t *testing.T) {
	testCases := []struct {
		name  string
		isCA  bool
		flags map[string]string
		err   string
	}{
		{
			"ShouldReturnCertificate",
			false,
			nil,
			"",
		},
		{
			"ShouldReturnCACertificate",
			true,
			nil,
			"",
		},
		{
			"ShouldReturnCertificateWithSANs",
			false,
			map[string]string{cmdFlagNameSANs: "example.com,192.168.1.1"},
			"",
		},
		{
			"ShouldReturnCertificateWithExtendedUsage",
			false,
			map[string]string{cmdFlagNameExtendedUsage: "server"},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := NewCmdCtx()

			parent := &cobra.Command{Use: cmdUseRSA}
			cmd := &cobra.Command{Use: "generate"}

			parent.AddCommand(cmd)

			cmdFlagsCryptoCertificateCommon(cmd)
			cmdFlagsCryptoCertificateGenerate(cmd)

			if tc.isCA {
				require.NoError(t, cmd.Flags().Set(cmdFlagNameCA, "true"))
			}

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			cert, err := cmdCtx.cryptoGetCertificateFromCmd(cmd)

			if tc.err == "" {
				assert.NoError(t, err)
				require.NotNil(t, cert)
				assert.Equal(t, tc.isCA, cert.IsCA)

				if tc.isCA {
					assert.NotZero(t, cert.KeyUsage&x509.KeyUsageCertSign)
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldErrWhenCAFlagNotDefined", func(t *testing.T) {
		cmdCtx := NewCmdCtx()

		parent := &cobra.Command{Use: cmdUseRSA}
		cmd := &cobra.Command{Use: "generate"}

		parent.AddCommand(cmd)

		_, err := cmdCtx.cryptoGetCertificateFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: ca")
	})

	t.Run("ShouldErrWhenExtendedUsageFlagNotDefined", func(t *testing.T) {
		cmdCtx := NewCmdCtx()

		parent := &cobra.Command{Use: cmdUseRSA}
		cmd := &cobra.Command{Use: "generate"}

		parent.AddCommand(cmd)

		cmdFlagsCryptoCertificateCommon(cmd)

		cmd.Flags().Bool(cmdFlagNameCA, false, "")

		_, err := cmdCtx.cryptoGetCertificateFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: extended-usage")
	})
}

func TestCryptoGetCSRFromCmd(t *testing.T) {
	testCases := []struct {
		name  string
		flags map[string]string
		err   string
	}{
		{
			"ShouldReturnCSR",
			nil,
			"",
		},
		{
			"ShouldReturnCSRWithSANs",
			map[string]string{cmdFlagNameSANs: "example.com,10.0.0.1", cmdFlagNameCommonName: "Test"},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parent := &cobra.Command{Use: cmdUseRSA}
			cmd := &cobra.Command{Use: cmdUseRequest}

			parent.AddCommand(cmd)

			cmdFlagsCryptoCertificateCommon(cmd)

			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			csr, err := cryptoGetCSRFromCmd(cmd)

			if tc.err == "" {
				assert.NoError(t, err)
				require.NotNil(t, csr)
				assert.Equal(t, x509.RSA, csr.PublicKeyAlgorithm)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldErrWhenSubjectFlagsNotDefined", func(t *testing.T) {
		parent := &cobra.Command{Use: cmdUseRSA}
		cmd := &cobra.Command{Use: cmdUseRequest}

		parent.AddCommand(cmd)

		cmd.Flags().String(cmdFlagNameSignature, "SHA256", "")

		_, err := cryptoGetCSRFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: common-name")
	})

	t.Run("ShouldErrWhenSANsFlagNotDefinedInCSR", func(t *testing.T) {
		parent := &cobra.Command{Use: cmdUseRSA}
		cmd := &cobra.Command{Use: cmdUseRequest}

		parent.AddCommand(cmd)

		cmd.Flags().String(cmdFlagNameSignature, "SHA256", "")
		cmd.Flags().String(cmdFlagNameCommonName, "", "")
		cmd.Flags().StringSlice(cmdFlagNameOrganization, []string{"Authelia"}, "")
		cmd.Flags().StringSlice(cmdFlagNameOrganizationalUnit, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameCountry, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameProvince, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameLocality, nil, "")
		cmd.Flags().StringSlice(cmdFlagNameStreetAddress, nil, "")
		cmd.Flags().StringSlice(cmdFlagNamePostcode, nil, "")

		_, err := cryptoGetCSRFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: sans")
	})
}

func TestCryptoGetCAFromCmd(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, cmd *cobra.Command)
		err   string
		isNil bool
	}{
		{
			"ShouldReturnNilWhenNotSet",
			nil,
			"",
			true,
		},
		{
			"ShouldSucceedLoadCA",
			func(t *testing.T, cmd *cobra.Command) {
				dir := t.TempDir()

				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				template := &x509.Certificate{
					SerialNumber:          big.NewInt(1),
					Subject:               pkix.Name{CommonName: "Test CA"},
					NotBefore:             time.Now().Add(-time.Hour),
					NotAfter:              time.Now().Add(time.Hour),
					IsCA:                  true,
					KeyUsage:              x509.KeyUsageCertSign,
					BasicConstraintsValid: true,
				}

				certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
				require.NoError(t, err)

				keyBytes, err := x509.MarshalECPrivateKey(key)
				require.NoError(t, err)

				require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.private.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}), 0600))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.public.crt"), pem.EncodeToMemory(&pem.Block{Type: utils.BlockTypeCertificate, Bytes: certBytes}), 0600))

				require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, dir))
			},
			"",
			false,
		},
		{
			"ShouldErrPrivateKeyNotFound",
			func(t *testing.T, cmd *cobra.Command) {
				dir := t.TempDir()

				require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, dir))
			},
			"could not read private key file",
			false,
		},
		{
			"ShouldErrInvalidPrivateKey",
			func(t *testing.T, cmd *cobra.Command) {
				dir := t.TempDir()

				require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.private.pem"), []byte("not a pem"), 0600))

				require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, dir))
			},
			"could not parse private key from file",
			false,
		},
		{
			"ShouldErrCertificateNotFound",
			func(t *testing.T, cmd *cobra.Command) {
				dir := t.TempDir()

				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				keyBytes, err := x509.MarshalECPrivateKey(key)
				require.NoError(t, err)

				require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.private.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}), 0600))

				require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, dir))
			},
			"could not read certificate file",
			false,
		},
		{
			"ShouldErrInvalidCertificate",
			func(t *testing.T, cmd *cobra.Command) {
				dir := t.TempDir()

				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				keyBytes, err := x509.MarshalECPrivateKey(key)
				require.NoError(t, err)

				require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.private.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}), 0600))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.public.crt"), []byte("not a pem"), 0600))

				require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, dir))
			},
			"could not parse certificate from file",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmdFlagsCryptoCertificateGenerate(cmd)

			if tc.setup != nil {
				tc.setup(t, cmd)
			}

			key, cert, err := cryptoGetCAFromCmd(cmd)

			if tc.err == "" {
				assert.NoError(t, err)

				if tc.isNil {
					assert.Nil(t, key)
					assert.Nil(t, cert)
				} else {
					assert.NotNil(t, key)
					assert.NotNil(t, cert)
				}
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}

	t.Run("ShouldErrWhenFileCAPrivateKeyFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNamePathCA, "", "")

		require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, t.TempDir()))

		_, _, err := cryptoGetCAFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: file.ca-private-key")
	})

	t.Run("ShouldErrWhenFileCACertificateFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String(cmdFlagNamePathCA, "", "")
		cmd.Flags().String(cmdFlagNameFileCAPrivateKey, "ca.private.pem", "")

		require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, t.TempDir()))

		_, _, err := cryptoGetCAFromCmd(cmd)

		assert.EqualError(t, err, "flag accessed but not defined: file.ca-certificate")
	})
}

func TestCryptoGenerateCertificateBundlesFromCmd(t *testing.T) {
	testCases := []struct {
		name     string
		bundles  string
		withCA   bool
		expected []string
		files    []string
	}{
		{
			"ShouldGenerateNoBundles",
			"",
			false,
			nil,
			nil,
		},
		{
			"ShouldGenerateChainBundle",
			"chain",
			false,
			[]string{"Certificate (chain):"},
			[]string{"public.chain.pem"},
		},
		{
			"ShouldGenerateChainBundleWithCA",
			"chain",
			true,
			[]string{"Certificate (chain):"},
			[]string{"public.chain.pem"},
		},
		{
			"ShouldGeneratePrivChainBundle",
			"priv-chain",
			false,
			[]string{"Certificate (priv-chain):"},
			[]string{"private.chain.pem"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			cmd := &cobra.Command{}
			cmdFlagsCryptoCertificateGenerate(cmd)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			if tc.bundles != "" {
				require.NoError(t, cmd.Flags().Set(cmdFlagNameBundles, tc.bundles))
			}

			key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			require.NoError(t, err)

			template := &x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject:      pkix.Name{CommonName: "Test"},
				NotBefore:    time.Now().Add(-time.Hour),
				NotAfter:     time.Now().Add(time.Hour),
			}

			certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
			require.NoError(t, err)

			var ca *x509.Certificate

			if tc.withCA {
				ca = template
			}

			err = cryptoGenerateCertificateBundlesFromCmd(cmd, dir, ca, certBytes, key)

			assert.NoError(t, err)

			for _, s := range tc.expected {
				assert.Contains(t, buf.String(), s)
			}

			for _, f := range tc.files {
				_, statErr := os.Stat(filepath.Join(dir, f))
				assert.NoError(t, statErr)
			}
		})
	}

	t.Run("ShouldErrWhenBundlesFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}

		err := cryptoGenerateCertificateBundlesFromCmd(cmd, "", nil, nil, nil)

		assert.EqualError(t, err, "flag accessed but not defined: bundles")
	})

	t.Run("ShouldErrWhenChainFileFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringSlice(cmdFlagNameBundles, nil, "")

		require.NoError(t, cmd.Flags().Set(cmdFlagNameBundles, "chain"))

		err := cryptoGenerateCertificateBundlesFromCmd(cmd, "", nil, nil, nil)

		assert.EqualError(t, err, "flag accessed but not defined: file.bundle.chain")
	})

	t.Run("ShouldErrWhenPrivChainFileFlagNotDefined", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringSlice(cmdFlagNameBundles, nil, "")

		require.NoError(t, cmd.Flags().Set(cmdFlagNameBundles, "priv-chain"))

		err := cryptoGenerateCertificateBundlesFromCmd(cmd, "", nil, nil, nil)

		assert.EqualError(t, err, "flag accessed but not defined: file.bundle.priv-chain")
	})
}

func TestFmtCryptoHashUse(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		expected string
	}{
		{
			"ShouldFormatArgon2",
			cmdUseHashArgon2,
			"Argon2",
		},
		{
			"ShouldFormatSHA2Crypt",
			cmdUseHashSHA2Crypt,
			"SHA2 Crypt",
		},
		{
			"ShouldFormatPBKDF2",
			cmdUseHashPBKDF2,
			"PBKDF2",
		},
		{
			"ShouldReturnDefaultForUnknown",
			"unknown",
			"unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, fmtCryptoHashUse(tc.use))
		})
	}
}

func TestFmtCryptoCertificateUse(t *testing.T) {
	testCases := []struct {
		name     string
		use      string
		expected string
	}{
		{
			"ShouldFormatEd25519",
			cmdUseEd25519,
			"Ed25519",
		},
		{
			"ShouldUppercaseRSA",
			cmdUseRSA,
			"RSA",
		},
		{
			"ShouldUppercaseECDSA",
			cmdUseECDSA,
			"ECDSA",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, fmtCryptoCertificateUse(tc.use))
		})
	}
}
