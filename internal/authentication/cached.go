package authentication

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"fmt"
	"hash"
	"sync"
	"time"
)

func NewCredentialCacheHMAC(h func() hash.Hash, lifespan time.Duration) *CredentialCacheHMAC {
	secret := make([]byte, h().BlockSize())

	_, _ = rand.Read(secret)

	return &CredentialCacheHMAC{
		mu:       sync.Mutex{},
		hash:     hmac.New(h, secret),
		lifespan: lifespan,

		values: map[string]CachedCredential{},
	}
}

type CredentialCacheHMAC struct {
	mu   sync.Mutex
	hash hash.Hash

	lifespan time.Duration

	values map[string]CachedCredential
}

func (c *CredentialCacheHMAC) Valid(username, password string) (valid, ok bool) {
	c.mu.Lock()

	defer c.mu.Unlock()

	var (
		entry CachedCredential
		err   error
	)

	if entry, ok = c.values[username]; ok {
		if entry.expires.Before(time.Now()) {
			delete(c.values, username)

			return false, false
		}
	}

	var value []byte

	if value, err = c.sum(username, password); err != nil {
		return false, false
	}

	valid = bytes.Equal(value, entry.value)

	c.hash.Reset()

	return valid, true
}

func (c *CredentialCacheHMAC) sum(username, password string) (sum []byte, err error) {
	defer c.hash.Reset()

	if _, err = c.hash.Write([]byte(password)); err != nil {
		return nil, fmt.Errorf("error occurred calculating cache hmac: %w", err)
	}

	if _, err = c.hash.Write([]byte(username)); err != nil {
		return nil, fmt.Errorf("error occurred calculating cache hmac: %w", err)
	}

	return c.hash.Sum(nil), nil
}

func (c *CredentialCacheHMAC) Put(username, password string) (err error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	var value []byte

	if value, err = c.sum(username, password); err != nil {
		return err
	}

	c.values[username] = CachedCredential{expires: time.Now().Add(c.lifespan), value: value}

	return nil
}

type CachedCredential struct {
	expires time.Time
	value   []byte
}
