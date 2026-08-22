package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/google/uuid"
	"golang.org/x/crypto/hkdf"

	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewCodec returns a new SecureCodec.
func NewCodec(rawKey string, hmacKey []byte, random random.Provider) (codec Codec, err error) {
	reader := hkdf.New(sha256.New, []byte(rawKey), nil, []byte(hkdfKeyInfoCodec))

	key := make([]byte, 32)

	if _, err = io.ReadFull(reader, key); err != nil {
		return nil, err
	}

	return &SecureCodec{
		encKey:  key,
		hmacKey: hmacKey,
		random:  random,

		charsetSessionID: randomSessionChars,
	}, nil
}

// SecureCodec is the default Codec which signs session identifiers with a HMAC key and seals session data with an
// authenticated encryption key derived from the session secret.
type SecureCodec struct {
	encKey  []byte
	hmacKey []byte

	random random.Provider

	charsetSessionID string
}

// GeneratePublicID returns a new random public identifier for a session. The public identifier is the value shared
// with consumers which need to reference a session without being able to derive its signature.
func (c *SecureCodec) GeneratePublicID() (id string, err error) {
	var pid uuid.UUID

	if pid, err = uuid.NewRandom(); err != nil {
		return "", err
	} else {
		return pid.String(), nil
	}
}

// GenerateSessionID returns a new random session identifier which is the value stored in the session cookie.
func (c *SecureCodec) GenerateSessionID() (id string, err error) {
	return c.random.StringCustomErr(32, c.charsetSessionID)
}

// Verify returns true if the given signature is the signature of the given data, comparing them in constant time.
func (c *SecureCodec) Verify(data []byte, signature string) bool {
	actual, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	expected := c.sign(data)

	return hmac.Equal(expected, actual)
}

// Sign returns the hex encoded HMAC signature of the given data.
func (c *SecureCodec) Sign(data []byte) string {
	return hex.EncodeToString(c.sign(data))
}

func (c *SecureCodec) sign(data []byte) []byte {
	mac := hmac.New(sha256.New, c.hmacKey)
	mac.Write(data)

	return mac.Sum(nil)
}

// Seal marshals and encrypts the given UserSession, binding the ciphertext to the domain and session identifier so it
// cannot be replayed against another domain or session.
func (c *SecureCodec) Seal(domain, id string, session UserSession) (data []byte, err error) {
	raw, err := session.MarshalMsg(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal session: %w", err)
	}

	if data, err = utils.Encrypt(raw, getAAD(domain, id), c.encKey); err != nil {
		return nil, fmt.Errorf("unable to encrypt session: %w", err)
	}

	return data, nil
}

// Open decrypts and unmarshals the data of the given Record into the given UserSession. A nil record, or one with no
// data, leaves the session untouched and returns no error.
func (c *SecureCodec) Open(domain string, record Record, session *UserSession) (err error) {
	if record == nil || len(record.GetSessionData()) == 0 {
		return nil
	}

	var data []byte

	if data, err = utils.Decrypt(record.GetSessionData(), getAAD(domain, record.GetSessionSignature()), c.encKey); err != nil {
		return fmt.Errorf("unable to decrypt session: %s", err)
	}

	_, err = session.UnmarshalMsg(data)

	return err
}

func getAAD(domain, id string) []byte {
	return []byte(fmt.Sprintf("authelia:session:v2:%s:%s", domain, id))
}
