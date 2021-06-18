package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

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
		RunE:  cmdRSAGenerateRunE,
	}

	cmd.Flags().StringP("dir", "d", "", "Target directory where the keypair will be stored")
	cmd.Flags().IntP("key-size", "b", 2048, "Sets the key size in bits")

	return cmd
}

func cmdRSAGenerateRunE(cmd *cobra.Command, _ []string) (err error) {
	bits, err := cmd.Flags().GetInt("key-size")
	if err != nil {
		return err
	}

	privateKey, publicKey := utils.GenerateRsaKeyPair(bits)

	rsaTargetDirectory, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	keyPath := path.Join(rsaTargetDirectory, "key.pem")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %w", keyPath, err)
	}

	_, err = keyOut.WriteString(utils.ExportRsaPrivateKeyAsPemStr(privateKey))
	if err != nil {
		return fmt.Errorf("Unable to write private key: %w", err)
	}

	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("Unable to close private key file: %w", err)
	}

	keyPath = path.Join(rsaTargetDirectory, "key.pub")
	keyOut, err = os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %w", keyPath, err)
	}

	publicPem, err := utils.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		return fmt.Errorf("Unable to marshal public key: %w", err)
	}

	_, err = keyOut.WriteString(publicPem)
	if err != nil {
		return fmt.Errorf("Unable to write private key: %v", err)
	}

	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("Unable to close public key file: %w", err)
	}

	return nil
}
