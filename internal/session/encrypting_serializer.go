package session

import (
	"crypto/sha256"
	"fmt"

	"github.com/fasthttp/session/v2"

	"github.com/authelia/authelia/v4/internal/utils"
)

// Serializer is a function that can serialize session information.
type Serializer interface {
	Encode(src session.Dict) (data []byte, err error)
	Decode(dst *session.Dict, src []byte) (err error)
}

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
func (e *EncryptingSerializer) Encode(src session.Dict) (data []byte, err error) {
	if len(src.KV) == 0 {
		return nil, nil
	}

	dst, err := src.MarshalMsg(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal session: %v", err)
	}

	if data, err = utils.Encrypt(dst, &e.key); err != nil {
		return nil, fmt.Errorf("unable to encrypt session: %v", err)
	}

	return data, nil
}

// Decode decrypt and decode session.
func (e *EncryptingSerializer) Decode(dst *session.Dict, src []byte) (err error) {
	if len(src) == 0 {
		return nil
	}

	for k := range dst.KV {
		delete(dst.KV, k)
	}

	var data []byte

	if data, err = utils.Decrypt(src, &e.key); err != nil {
		return fmt.Errorf("unable to decrypt session: %s", err)
	}

	_, err = dst.UnmarshalMsg(data)

	return err
}
