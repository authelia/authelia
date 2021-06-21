package commands

import (
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// NewRSACmd returns a new RSA Cmd.
func NewRSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "rsa",
		Short: "Commands related to rsa keypair generation",
	}

	cmd.AddCommand(newRSAGenerateCmd())

	return cmd
}

func newRSAGenerateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a RSA keypair",
		Run:   cmdRSAGenerateRun,
	}

	cmd.Flags().StringP("dir", "d", "", "Target directory where the keypair will be stored")
	cmd.Flags().IntP("key-size", "b", 2048, "Sets the key size in bits")

	return cmd
}

func cmdRSAGenerateRun(cmd *cobra.Command, _ []string) {
	logger := logging.Logger()

	bits, err := cmd.Flags().GetInt("key-size")
	if err != nil {
		logger.Fatal(err)
	}

	privateKey, publicKey := utils.GenerateRsaKeyPair(bits)

	rsaTargetDirectory, err := cmd.Flags().GetString("dir")
	if err != nil {
		logger.Fatal(err)
	}

	keyPath := path.Join(rsaTargetDirectory, "key.pem")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		logger.Fatalf("Failed to open %s for writing: %w", keyPath, err)
	}

	_, err = keyOut.WriteString(utils.ExportRsaPrivateKeyAsPemStr(privateKey))
	if err != nil {
		logger.Fatalf("Unable to write private key: %w", err)
	}

	if err := keyOut.Close(); err != nil {
		logger.Fatalf("Unable to close private key file: %w", err)
	}

	keyPath = path.Join(rsaTargetDirectory, "key.pub")
	keyOut, err = os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		logger.Fatalf("Failed to open %s for writing: %w", keyPath, err)
	}

	publicPem, err := utils.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		logger.Fatalf("Unable to marshal public key: %w", err)
	}

	_, err = keyOut.WriteString(publicPem)
	if err != nil {
		logger.Fatalf("Unable to write private key: %v", err)
	}

	if err := keyOut.Close(); err != nil {
		logger.Fatalf("Unable to close public key file: %w", err)
	}
}
