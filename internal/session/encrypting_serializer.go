package session

import (
	"crypto/sha256"
	"fmt"

	"github.com/fasthttp/session/v2"

	"github.com/authelia/authelia/v4/internal/utils"
)

// EncryptingSerializer a serializer encrypting the data with AES-GCM with 256-bit keys.
type EncryptingSerializer struct {
	key [32]byte
}

// NewEncryptingSerializer return new encrypt instance.
func NewEncryptingSerializer(secret string) *EncryptingSerializer {
	key := sha256.Sum256([]byte(secret))
	return &EncryptingSerializer{key}
}

// Encode encode and encrypt session.
func (e *EncryptingSerializer) Encode(src session.Dict) ([]byte, error) {
	if len(src.D) == 0 {
		return nil, nil
	}

	dst, err := src.MarshalMsg(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal session: %v", err)
	}

	encryptedDst, err := utils.Encrypt(dst, &e.key)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt session: %v", err)
	}

	return encryptedDst, nil
}

// Decode decrypt and decode session.
func (e *EncryptingSerializer) Decode(dst *session.Dict, src []byte) error {
	if len(src) == 0 {
		return nil
	}

	dst.Reset()

	decryptedSrc, err := utils.Decrypt(src, &e.key)
	if err != nil {
		return fmt.Errorf("unable to decrypt session: %s", err)
	}

	_, err = dst.UnmarshalMsg(decryptedSrc)

	return err
}
