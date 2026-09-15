// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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

// HMACKeys are the HMAC keys a SecureCodec signs values with. Each is used for a single purpose so that a signature
// made for one purpose is never valid for another, and each can be rotated independently of the others.
type HMACKeys struct {
	// Session is the key the session cookie value is signed with to derive the identifier the session is stored
	// against.
	Session []byte

	// SessionAuthz is the key the authz cookie value is derived and signed with.
	SessionAuthz []byte

	// SessionCSRF is the key the CSRF token is derived from the CSRF secret of a session with.
	SessionCSRF []byte
}

// NewCodec returns a new SecureCodec. Every HMAC key is required and each must be distinct.
func NewCodec(rawKey string, keys HMACKeys, random random.Provider) (codec Codec, err error) {
	switch {
	case len(keys.Session) == 0, len(keys.SessionAuthz) == 0, len(keys.SessionCSRF) == 0:
		return nil, fmt.Errorf("error occurred creating the session codec: every hmac key is required")
	case hmac.Equal(keys.Session, keys.SessionAuthz), hmac.Equal(keys.Session, keys.SessionCSRF), hmac.Equal(keys.SessionAuthz, keys.SessionCSRF):
		return nil, fmt.Errorf("error occurred creating the session codec: every hmac key must be distinct")
	}

	reader := hkdf.New(sha256.New, []byte(rawKey), nil, []byte(hkdfKeyInfoCodec))

	key := make([]byte, 32)

	if _, err = io.ReadFull(reader, key); err != nil {
		return nil, err
	}

	return &SecureCodec{
		encKey:       key,
		hmacKey:      keys.Session,
		hmacKeyAuthz: keys.SessionAuthz,
		hmacKeyCSRF:  keys.SessionCSRF,
		random:       random,
	}, nil
}

// SecureCodec is the default Codec which signs values with a HMAC key specific to their purpose and seals session data
// with an authenticated encryption key derived from the session secret.
type SecureCodec struct {
	encKey       []byte
	hmacKey      []byte
	hmacKeyAuthz []byte
	hmacKeyCSRF  []byte

	random random.Provider
}

// GeneratePublicID returns a new random public identifier for a session. The public identifier is the value shared
// with consumers which need to reference a session without being able to derive its signature.
func (c *SecureCodec) GeneratePublicID() (id string, err error) {
	var pid uuid.UUID

	if pid, err = uuid.NewV7(); err != nil {
		return "", err
	} else {
		return pid.String(), nil
	}
}

// GenerateSessionID returns a new random session identifier which is the raw value encoded into the session cookie.
func (c *SecureCodec) GenerateSessionID() (id []byte, err error) {
	id = make([]byte, sessionIDLength)

	if _, err = c.random.Read(id); err != nil {
		return nil, err
	}

	return id, nil
}

// GenerateCSRFSecret returns a new random CSRF secret which is stored in the session and signed to derive the CSRF token.
func (c *SecureCodec) GenerateCSRFSecret() (secret []byte, err error) {
	secret = make([]byte, 32)

	if _, err = io.ReadFull(c.random, secret); err != nil {
		return nil, err
	}

	return secret, nil
}

// Verify returns true if the given signature is the signature of the given data, comparing them in constant time.
func (c *SecureCodec) Verify(data []byte, signature string) bool {
	return verify(c.hmacKey, data, signature)
}

// Sign returns the hex encoded HMAC signature of the given data using the session key.
func (c *SecureCodec) Sign(data []byte) string {
	return hex.EncodeToString(sign(c.hmacKey, data))
}

// SignAuthz returns the hex encoded HMAC signature of the given data using the session authz key.
func (c *SecureCodec) SignAuthz(data []byte) string {
	return hex.EncodeToString(sign(c.hmacKeyAuthz, data))
}

// SignCSRF returns the hex encoded HMAC signature of the given data using the session CSRF key.
func (c *SecureCodec) SignCSRF(data []byte) string {
	return hex.EncodeToString(sign(c.hmacKeyCSRF, data))
}

// VerifyCSRF returns true if the given signature is the CSRF signature of the given data, comparing them in constant
// time.
func (c *SecureCodec) VerifyCSRF(data []byte, signature string) bool {
	return verify(c.hmacKeyCSRF, data, signature)
}

func verify(key, data []byte, signature string) bool {
	actual, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	return hmac.Equal(sign(key, data), actual)
}

func sign(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
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
