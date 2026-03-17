package session2

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash"
	"sync"

	"github.com/hashicorp/go-msgpack/v2/codec"

	"github.com/authelia/authelia/v4/internal/random"
)

func NewEncoder(rand random.Provider, key, secret []byte) (encoder *Encoder, err error) {
	k := sha256.Sum256(key)

	block, err := aes.NewCipher(k[:])
	if err != nil {
		return nil, fmt.Errorf("error occurred creating aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating gcm aead cipher: %w", err)
	}

	return &Encoder{
		rand:    rand,
		encrypt: gcm,
		lock:    &sync.Mutex{},
		hmac:    sync.Pool{New: func() interface{} { return hmac.New(sha256.New, secret) }},
	}, nil
}

type Encoder struct {
	rand    random.Provider
	encrypt cipher.AEAD
	lock    sync.Locker
	hmac    sync.Pool
}

func (e *Encoder) DecodeSessionData(ciphertext []byte, v any) (err error) {
	if len(ciphertext) < e.encrypt.NonceSize() {
		return fmt.Errorf("error decrypting session data: malformed ciphertext")
	}

	cleartext, err := e.encrypt.Open(nil, ciphertext[:e.encrypt.NonceSize()], ciphertext[e.encrypt.NonceSize():], nil)
	if err != nil {
		return fmt.Errorf("error decrypting session data: %w", err)
	}

	decoder := codec.NewDecoderBytes(cleartext, &codec.MsgpackHandle{})

	if err = decoder.Decode(v); err != nil {
		return fmt.Errorf("error decoding session data: %w", err)
	}

	return nil
}

func (e *Encoder) Encode(id []byte, v any) (sum, ciphertext []byte, err error) {
	e.lock.Lock()

	defer e.lock.Unlock()

	buf := bytes.NewBuffer(nil)

	encoder := codec.NewEncoder(buf, &codec.MsgpackHandle{})

	if err = encoder.Encode(v); err != nil {
		return nil, nil, fmt.Errorf("error occurred encoding session data: %w", err)
	}

	nonce := make([]byte, e.encrypt.NonceSize())
	if _, err = e.rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("error occurred encrypting session data: error generating nonce: %w", err)
	}

	return e.sum(id), e.encrypt.Seal(nil, nonce, buf.Bytes(), nil), nil
}

func (e *Encoder) EncodeSessionID(id []byte) (ciphertext []byte) {
	e.lock.Lock()

	defer e.lock.Unlock()

	return e.sum(id)
}

func (e *Encoder) sum(id []byte) (sum []byte) {
	h := e.hmac.Get().(hash.Hash)

	defer func() {
		h.Reset()
		e.hmac.Put(h)
	}()

	h.Write(id)

	return h.Sum(nil)
}
