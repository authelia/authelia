package commands

import (
	"log"
	"os"
	"path"

	"github.com/authelia/authelia/internal/utils"
	"github.com/spf13/cobra"
)

var rsaTargetDirectory string

func init() {
	RSAGenerateCmd.PersistentFlags().StringVar(&rsaTargetDirectory, "dir", "", "Target directory where the keypair will be stored")

	RSACmd.AddCommand(RSAGenerateCmd)
}

func generateRSAKeypair(cmd *cobra.Command, args []string) {
	privateKey, publicKey := utils.GenerateRsaKeyPair(2048)

	keyPath := path.Join(rsaTargetDirectory, "key.pem")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", keyPath, err)
		return
	}

	_, err = keyOut.WriteString(utils.ExportRsaPrivateKeyAsPemStr(privateKey))
	if err != nil {
		log.Fatalf("Unable to write private key: %v", err)
		return
	}

	if err := keyOut.Close(); err != nil {
		log.Fatalf("Unable to close private key file: %v", err)
		return
	}

	keyPath = path.Join(rsaTargetDirectory, "key.pub")
	keyOut, err = os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", keyPath, err)
		return
	}

	publicPem, err := utils.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		log.Fatalf("Unable to marshal public key: %v", err)
	}

	_, err = keyOut.WriteString(publicPem)
	if err != nil {
		log.Fatalf("Unable to write private key: %v", err)
		return
	}

	if err := keyOut.Close(); err != nil {
		log.Fatalf("Unable to close public key file: %v", err)
		return
	}
}

// RSACmd RSA helper command.
var RSACmd = &cobra.Command{
	Use:   "rsa",
	Short: "Commands related to rsa keypair generation",
}

// RSAGenerateCmd certificate generation command.
var RSAGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a RSA keypair",
	Run:   generateRSAKeypair,
}
