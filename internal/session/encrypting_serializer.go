package session

import (
	"crypto/sha256"
	"fmt"

	"github.com/fasthttp/session/v2"

	"github.com/authelia/authelia/internal/utils"
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
		return nil, fmt.Errorf("Unable to marshal session: %v", err)
	}

	encryptedDst, err := utils.Encrypt(dst, &e.key)
	if err != nil {
		return nil, fmt.Errorf("Unable to encrypt session: %v", err)
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
		// If an error is thrown while decrypting, it's probably an old unencrypted session
		// so we just unmarshall it without decrypting. It's a way to avoid a breaking change
		// requiring to flush redis.
		// TODO(clems4ever): remove in few months
		_, uerr := dst.UnmarshalMsg(src)
		if uerr != nil {
			return fmt.Errorf("Unable to decrypt session: %s", err)
		}

		return nil
	}

	_, err = dst.UnmarshalMsg(decryptedSrc)

	return err
}
