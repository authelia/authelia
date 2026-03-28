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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestNewCrypto(t *testing.T) {
	var cmd *cobra.Command

	cmd = newCryptoCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newCryptoCertificateCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newCryptoPairCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newCryptoPairSubCmd(&CmdCtx{}, "generate")
	assert.NotNil(t, cmd)

	cmd = newCryptoPairSubCmd(&CmdCtx{}, "verify")
	assert.NotNil(t, cmd)
}

func TestCryptoRandRunE(t *testing.T) {
	t.Run("ShouldSucceedPrint", func(t *testing.T) {
		cmdCtx := NewCmdCtx()

		cmd := newCryptoRandCmd(cmdCtx)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := cmdCtx.CryptoRandRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Random Value:")
	})

	t.Run("ShouldSucceedPrintWithArgs", func(t *testing.T) {
		cmdCtx := NewCmdCtx()

		cmd := newCryptoRandCmd(cmdCtx)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameLength, "16"))

		err := cmdCtx.CryptoRandRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Random Value:")
	})

	t.Run("ShouldSucceedWriteFiles", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoRandCmd(cmdCtx)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		file1 := filepath.Join(dir, "secret1.txt")
		file2 := filepath.Join(dir, "secret2.txt")

		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, file1+","+file2))

		err := cmdCtx.CryptoRandRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Created 2 files")

		_, err = os.Stat(file1)
		assert.NoError(t, err)

		_, err = os.Stat(file2)
		assert.NoError(t, err)
	})

	t.Run("ShouldSucceedWriteFilesInSubdir", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoRandCmd(cmdCtx)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		file := filepath.Join(dir, "subdir", "secret.txt")

		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, file))

		err := cmdCtx.CryptoRandRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Created 1 files")

		_, err = os.Stat(file)
		assert.NoError(t, err)
	})

	t.Run("ShouldSucceedWriteFilesFromArgs", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoRandCmd(cmdCtx)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		file := filepath.Join(dir, "secret.txt")

		err := cmdCtx.CryptoRandRunE(cmd, []string{file})

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Created 1 files")
	})

	t.Run("ShouldErrArgsWithFileFlag", func(t *testing.T) {
		cmdCtx := NewCmdCtx()

		cmd := newCryptoRandCmd(cmdCtx)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameFile, "/tmp/test.txt"))

		err := cmdCtx.CryptoRandRunE(cmd, []string{"extra"})

		assert.ErrorContains(t, err, "arguments may not be specified at the same time as the files flag")
	})
}

func TestRunCryptoRandPrint(t *testing.T) {
	t.Run("ShouldSucceedAlphanumeric", func(t *testing.T) {
		cmd := newCryptoRandCmd(NewCmdCtx())

		buf := new(bytes.Buffer)

		err := runCryptoRandPrint(buf, cmd.Flags())

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Random Value:")
	})

	t.Run("ShouldSucceedURLEncoded", func(t *testing.T) {
		cmd := newCryptoRandCmd(NewCmdCtx())

		require.NoError(t, cmd.Flags().Set(cmdFlagNameCharSet, "rfc3986"))

		buf := new(bytes.Buffer)

		err := runCryptoRandPrint(buf, cmd.Flags())

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Random Value:")
	})
}

func TestRunCryptoRandFiles(t *testing.T) {
	t.Run("ShouldErrInvalidFileMode", func(t *testing.T) {
		cmd := newCryptoRandCmd(NewCmdCtx())

		require.NoError(t, cmd.Flags().Set(cmdFlagNameModeFiles, "xyz"))

		buf := new(bytes.Buffer)

		err := runCryptoRandFiles(buf, cmd.Flags(), []string{filepath.Join(t.TempDir(), "test.txt")})

		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("ShouldErrInvalidDirMode", func(t *testing.T) {
		cmd := newCryptoRandCmd(NewCmdCtx())

		require.NoError(t, cmd.Flags().Set(cmdFlagNameModeDirectories, "xyz"))

		buf := new(bytes.Buffer)

		err := runCryptoRandFiles(buf, cmd.Flags(), []string{filepath.Join(t.TempDir(), "test.txt")})

		assert.ErrorContains(t, err, "invalid syntax")
	})
}

func TestCryptoGenerateRunE(t *testing.T) {
	t.Run("ShouldSucceedCertificateRSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUseCertificate, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}
		grandparent := &cobra.Command{Use: cmdUseCertificate}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoGenerateRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating Certificate")

		_, statErr := os.Stat(filepath.Join(dir, "private.pem"))
		assert.NoError(t, statErr)
	})

	t.Run("ShouldSucceedCertificateECDSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUseCertificate, cmdUseECDSA)
		parent := &cobra.Command{Use: cmdUseECDSA}
		grandparent := &cobra.Command{Use: cmdUseCertificate}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoGenerateRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Elliptic Curve")
	})

	t.Run("ShouldSucceedCertificateEd25519", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUseCertificate, cmdUseEd25519)
		parent := &cobra.Command{Use: cmdUseEd25519}
		grandparent := &cobra.Command{Use: cmdUseCertificate}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoGenerateRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating Certificate")
	})

	t.Run("ShouldSucceedPairRSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUsePair, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}
		grandparent := &cobra.Command{Use: cmdUsePair}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoGenerateRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating key pair")
		assert.Contains(t, buf.String(), "RSA")

		_, statErr := os.Stat(filepath.Join(dir, "private.pem"))
		assert.NoError(t, statErr)

		_, statErr = os.Stat(filepath.Join(dir, "public.pem"))
		assert.NoError(t, statErr)
	})

	t.Run("ShouldSucceedPairECDSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUsePair, cmdUseECDSA)
		parent := &cobra.Command{Use: cmdUseECDSA}
		grandparent := &cobra.Command{Use: cmdUsePair}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoGenerateRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "ECDSA")
	})

	t.Run("ShouldSucceedPairEd25519", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUsePair, cmdUseEd25519)
		parent := &cobra.Command{Use: cmdUseEd25519}
		grandparent := &cobra.Command{Use: cmdUsePair}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoGenerateRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Ed25519")
	})
}

func TestCryptoCertificateRequestRunE(t *testing.T) {
	t.Run("ShouldSucceedRSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoCertificateRequestCmd(cmdCtx, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}

		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))
		require.NoError(t, cmd.Flags().Set(cmdFlagNameCommonName, "Test"))
		require.NoError(t, cmd.Flags().Set(cmdFlagNameSANs, "example.com,10.0.0.1"))

		err := cmdCtx.CryptoCertificateRequestRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating Certificate Request")
		assert.Contains(t, buf.String(), "Test")
		assert.Contains(t, buf.String(), "Bits:")

		_, statErr := os.Stat(filepath.Join(dir, "private.pem"))
		assert.NoError(t, statErr)

		_, statErr = os.Stat(filepath.Join(dir, "request.csr"))
		assert.NoError(t, statErr)
	})

	t.Run("ShouldSucceedECDSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoCertificateRequestCmd(cmdCtx, cmdUseECDSA)
		parent := &cobra.Command{Use: cmdUseECDSA}

		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoCertificateRequestRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Elliptic Curve")
	})

	t.Run("ShouldSucceedEd25519", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoCertificateRequestCmd(cmdCtx, cmdUseEd25519)
		parent := &cobra.Command{Use: cmdUseEd25519}

		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		err := cmdCtx.CryptoCertificateRequestRunE(cmd, nil)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating Certificate Request")
	})

	t.Run("ShouldSucceedRSAWithLegacy", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoCertificateRequestCmd(cmdCtx, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}

		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))
		require.NoError(t, cmd.Flags().Set(cmdFlagNameLegacy, "true"))

		err := cmdCtx.CryptoCertificateRequestRunE(cmd, nil)

		assert.NoError(t, err)

		_, statErr := os.Stat(filepath.Join(dir, "private.legacy.pem"))
		assert.NoError(t, statErr)
	})
}

func TestCryptoCertificateGenerateRunE(t *testing.T) {
	t.Run("ShouldSucceedSelfSignedRSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUseCertificate, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}
		grandparent := &cobra.Command{Use: cmdUseCertificate}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))
		require.NoError(t, cmd.Flags().Set(cmdFlagNameSANs, "example.com"))

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		err = cmdCtx.CryptoCertificateGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating Certificate")
		assert.Contains(t, buf.String(), "Self-Signed")
		assert.Contains(t, buf.String(), "Bits: 2048")

		_, statErr := os.Stat(filepath.Join(dir, "public.crt"))
		assert.NoError(t, statErr)
	})

	t.Run("ShouldSucceedWithCA", func(t *testing.T) {
		dir := t.TempDir()
		caDir := t.TempDir()

		caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		caTemplate := &x509.Certificate{
			SerialNumber:          big.NewInt(1),
			Subject:               pkix.Name{CommonName: "Test CA"},
			NotBefore:             time.Now().Add(-time.Hour),
			NotAfter:              time.Now().Add(time.Hour),
			IsCA:                  true,
			KeyUsage:              x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		}

		caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
		require.NoError(t, err)

		caKeyBytes, err := x509.MarshalECPrivateKey(caKey)
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(filepath.Join(caDir, "ca.private.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyBytes}), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(caDir, "ca.public.crt"), pem.EncodeToMemory(&pem.Block{Type: utils.BlockTypeCertificate, Bytes: caCertBytes}), 0600))

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUseCertificate, cmdUseECDSA)
		parent := &cobra.Command{Use: cmdUseECDSA}
		grandparent := &cobra.Command{Use: cmdUseCertificate}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))
		require.NoError(t, cmd.Flags().Set(cmdFlagNamePathCA, caDir))

		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		err = cmdCtx.CryptoCertificateGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Test CA")
		assert.NotContains(t, buf.String(), "Self-Signed")
	})

	t.Run("ShouldSucceedEd25519", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUseCertificate, cmdUseEd25519)
		parent := &cobra.Command{Use: cmdUseEd25519}
		grandparent := &cobra.Command{Use: cmdUseCertificate}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		_, key, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		err = cmdCtx.CryptoCertificateGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating Certificate")
	})

	t.Run("ShouldSucceedWithLegacy", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUseCertificate, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}
		grandparent := &cobra.Command{Use: cmdUseCertificate}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))
		require.NoError(t, cmd.Flags().Set(cmdFlagNameLegacy, "true"))

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		err = cmdCtx.CryptoCertificateGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)

		_, statErr := os.Stat(filepath.Join(dir, "private.legacy.pem"))
		assert.NoError(t, statErr)
	})
}

func TestCryptoPairGenerateRunE(t *testing.T) {
	t.Run("ShouldSucceedRSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUsePair, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}
		grandparent := &cobra.Command{Use: cmdUsePair}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		err = cmdCtx.CryptoPairGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Generating key pair")
		assert.Contains(t, buf.String(), "RSA")
	})

	t.Run("ShouldSucceedECDSA", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUsePair, cmdUseECDSA)
		parent := &cobra.Command{Use: cmdUseECDSA}
		grandparent := &cobra.Command{Use: cmdUsePair}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		err = cmdCtx.CryptoPairGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "ECDSA")
	})

	t.Run("ShouldSucceedEd25519", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUsePair, cmdUseEd25519)
		parent := &cobra.Command{Use: cmdUseEd25519}
		grandparent := &cobra.Command{Use: cmdUsePair}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))

		_, key, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		err = cmdCtx.CryptoPairGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Ed25519")
	})

	t.Run("ShouldSucceedWithLegacy", func(t *testing.T) {
		dir := t.TempDir()

		cmdCtx := NewCmdCtx()

		cmd := newCryptoGenerateCmd(cmdCtx, cmdUsePair, cmdUseRSA)
		parent := &cobra.Command{Use: cmdUseRSA}
		grandparent := &cobra.Command{Use: cmdUsePair}

		grandparent.AddCommand(parent)
		parent.AddCommand(cmd)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		require.NoError(t, cmd.Flags().Set(cmdFlagNameDirectory, dir))
		require.NoError(t, cmd.Flags().Set(cmdFlagNameLegacy, "true"))

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		err = cmdCtx.CryptoPairGenerateRunE(cmd, nil, key)

		assert.NoError(t, err)

		_, statErr := os.Stat(filepath.Join(dir, "private.legacy.pem"))
		assert.NoError(t, statErr)

		_, statErr = os.Stat(filepath.Join(dir, "public.legacy.pem"))
		assert.NoError(t, statErr)
	})
}

func TestRunCryptoPairGenerate(t *testing.T) {
	t.Run("ShouldSucceedRSA", func(t *testing.T) {
		dir := t.TempDir()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		buf := new(bytes.Buffer)

		err = runCryptoPairGenerate(buf, false, key, dir, filepath.Join(dir, "private.pem"), "", filepath.Join(dir, "public.pem"), "", "legacy")

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "RSA")
	})

	t.Run("ShouldSucceedECDSA", func(t *testing.T) {
		dir := t.TempDir()

		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		buf := new(bytes.Buffer)

		err = runCryptoPairGenerate(buf, false, key, dir, filepath.Join(dir, "private.pem"), "", filepath.Join(dir, "public.pem"), "", "legacy")

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "ECDSA")
	})

	t.Run("ShouldSucceedEd25519", func(t *testing.T) {
		dir := t.TempDir()

		_, key, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		buf := new(bytes.Buffer)

		err = runCryptoPairGenerate(buf, false, key, dir, filepath.Join(dir, "private.pem"), "", filepath.Join(dir, "public.pem"), "", "legacy")

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Ed25519")
	})

	t.Run("ShouldShowDirectory", func(t *testing.T) {
		dir := t.TempDir()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		buf := new(bytes.Buffer)

		err = runCryptoPairGenerate(buf, false, key, dir, filepath.Join(dir, "private.pem"), "", filepath.Join(dir, "public.pem"), "", "legacy")

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Directory:")
	})
}

func TestNewCryptoCertificateSubCmd(t *testing.T) {
	testCases := []struct {
		name string
		use  string
	}{
		{"ShouldCreateRSA", cmdUseRSA},
		{"ShouldCreateECDSA", cmdUseECDSA},
		{"ShouldCreateEd25519", cmdUseEd25519},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newCryptoCertificateSubCmd(&CmdCtx{}, tc.use)

			assert.NotNil(t, cmd)
			assert.Equal(t, tc.use, cmd.Use)
		})
	}
}

func TestNewCryptoCertificateRequestCmd(t *testing.T) {
	testCases := []struct {
		name string
		alg  string
	}{
		{"ShouldCreateRSA", cmdUseRSA},
		{"ShouldCreateECDSA", cmdUseECDSA},
		{"ShouldCreateEd25519", cmdUseEd25519},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newCryptoCertificateRequestCmd(&CmdCtx{}, tc.alg)

			assert.NotNil(t, cmd)
			assert.Equal(t, cmdUseRequest, cmd.Use)
		})
	}
}

func TestNewCryptoGenerateCmd(t *testing.T) {
	testCases := []struct {
		name     string
		category string
		alg      string
	}{
		{"ShouldCreateCertificateRSA", cmdUseCertificate, cmdUseRSA},
		{"ShouldCreateCertificateECDSA", cmdUseCertificate, cmdUseECDSA},
		{"ShouldCreateCertificateEd25519", cmdUseCertificate, cmdUseEd25519},
		{"ShouldCreatePairRSA", cmdUsePair, cmdUseRSA},
		{"ShouldCreatePairECDSA", cmdUsePair, cmdUseECDSA},
		{"ShouldCreatePairEd25519", cmdUsePair, cmdUseEd25519},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newCryptoGenerateCmd(&CmdCtx{}, tc.category, tc.alg)

			assert.NotNil(t, cmd)
			assert.Equal(t, cmdUseGenerate, cmd.Use)
		})
	}
}
